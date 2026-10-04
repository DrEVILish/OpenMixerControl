// Command prosurfaced runs on a Midas PRO1/PRO2 (or any Linux box, with the
// simulator) and presents the PRO surface to OMC as an X32 Full surface.
//
// OMC connects over TCP. On the OMC host, bridge the TCP stream to a pty and
// start OMC on it:
//
//	socat pty,raw,echo=0,link=/tmp/ttyPRO tcp:<pro-ip>:10000
//	omc -b --model X32 --surface-tty /tmp/ttyPRO
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/bridge"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/driver"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/driver/scanproc"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/driver/sim"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/mapping"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/pro"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/simui"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/x32full"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/x32proto"
)

func main() {
	var (
		modelFlag   = flag.String("model", "pro2", "PRO model: pro1, pro2 or pro2c")
		listen      = flag.String("listen", ":10000", "TCP address OMC connects to")
		drvFlag     = flag.String("driver", "sim", "surface driver: sim or scanproc")
		dev         = flag.String("dev", "", "scan processor device (scanproc driver)")
		httpAddr    = flag.String("http", ":8080", "simulator web UI address (sim driver); empty disables it")
		mapPath     = flag.String("mapping", "", "JSON mapping file; default is the built-in mapping")
		dumpMapping = flag.Bool("dump-mapping", false, "print the built-in mapping as JSON and exit")
		verbose     = flag.Bool("v", false, "log every frame")
	)
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	model, err := pro.ParseModel(*modelFlag)
	if err != nil {
		fatal(log, err)
	}
	m := mapping.Default(model)
	if *dumpMapping {
		b, _ := m.JSON()
		fmt.Println(string(b))
		return
	}
	if *mapPath != "" {
		if m, err = mapping.Load(*mapPath); err != nil {
			fatal(log, err)
		}
		if m.Model != model {
			fatal(log, fmt.Errorf("mapping is for %s, not %s", m.Model, model))
		}
	}

	surface := pro.New(model)
	var drv driver.Driver
	var simDrv *sim.Driver
	switch *drvFlag {
	case "sim":
		simDrv = sim.New()
		drv = simDrv
	case "scanproc":
		if drv, err = scanproc.Open(*dev); err != nil {
			fatal(log, err)
		}
	default:
		fatal(log, fmt.Errorf("unknown driver %q", *drvFlag))
	}
	defer drv.Close()

	br, err := bridge.New(log, x32full.NewTable(), surface, m, drv)
	if err != nil {
		fatal(log, err)
	}
	log.Info("prosurfaced", "surface", br.Describe(), "driver", *drvFlag)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go br.Run(ctx)

	omc := &omcLink{log: log, br: br, addr: *listen}
	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		fatal(log, err)
	}
	go omc.serve(ctx, ln)
	log.Info("waiting for OMC", "listen", ln.Addr().String())

	if simDrv != nil && *httpAddr != "" {
		ui, err := simui.New(log, surface, simDrv, func() simui.Status {
			connected, addr := omc.status()
			return simui.Status{OMCConnected: connected, OMCAddr: addr, Page: br.Page(), Pages: m.Pages}
		})
		if err != nil {
			fatal(log, err)
		}
		hs := &http.Server{Addr: *httpAddr, Handler: ui}
		go func() {
			<-ctx.Done()
			hs.Close()
		}()
		log.Info("simulator", "url", "http://localhost"+*httpAddr)
		if err := hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fatal(log, err)
		}
		return
	}
	<-ctx.Done()
}

func fatal(log *slog.Logger, err error) {
	log.Error(err.Error())
	os.Exit(1)
}

// omcLink accepts one OMC connection at a time; a new connection replaces
// the old one, so restarting OMC or socat just works.
type omcLink struct {
	log  *slog.Logger
	br   *bridge.Bridge
	addr string

	mu   sync.Mutex
	conn net.Conn
}

func (o *omcLink) status() (bool, string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.conn != nil {
		return true, o.conn.RemoteAddr().String()
	}
	return false, o.addr
}

func (o *omcLink) serve(ctx context.Context, ln net.Listener) {
	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		o.mu.Lock()
		if o.conn != nil {
			o.conn.Close()
		}
		o.conn = c
		o.mu.Unlock()
		o.log.Info("OMC connected", "from", c.RemoteAddr().String())
		o.br.SetSender(func(p []byte) {
			if _, err := c.Write(p); err != nil {
				o.log.Debug("write to OMC", "err", err)
			}
		})
		go o.read(c)
	}
}

func (o *omcLink) read(c net.Conn) {
	defer func() {
		o.mu.Lock()
		if o.conn == c {
			o.conn = nil
			o.br.SetSender(nil)
		}
		o.mu.Unlock()
		c.Close()
		o.log.Info("OMC disconnected", "from", c.RemoteAddr().String())
	}()
	var dec x32proto.Decoder
	buf := make([]byte, 4096)
	for {
		n, err := c.Read(buf)
		if err != nil {
			return
		}
		for _, cmd := range dec.Feed(buf[:n]) {
			o.log.Debug("from OMC", "cmd", cmd.String(), "checksum", cmd.ChecksumOK)
			o.br.HandleCommand(cmd)
		}
	}
}
