// Package simui serves a browser simulator of the PRO surface, built with
// Go templates and htmx 4 (vendored under static/, BSD-0-Clause). Input goes
// to the sim driver; every state change the bridge makes is pushed back over
// Server-Sent Events as <hx-partial> fragments.
package simui

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/driver"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/driver/sim"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/pro"
)

//go:embed templates/*.html static/*
var files embed.FS

// Status describes the daemon for the page header.
type Status struct {
	OMCConnected bool
	OMCAddr      string
	Page, Pages  int
}

// Server is the simulator's HTTP handler.
type Server struct {
	log     *slog.Logger
	surface *pro.Surface
	drv     *sim.Driver
	status  func() Status
	tmpl    *template.Template
	mux     *http.ServeMux
}

// New builds the simulator for a surface.
func New(log *slog.Logger, s *pro.Surface, drv *sim.Driver, status func() Status) (*Server, error) {
	srv := &Server{log: log, surface: s, drv: drv, status: status, mux: http.NewServeMux()}
	t, err := template.New("").Funcs(template.FuncMap{
		"domid": domID,
		"state": drv.Snapshot,
		"color": lcdColor,
		"ink":   lcdInk,
		"pct":   func(v uint16) int { return int(v) * 100 / 4095 },
		"bit":   func(v uint16, i int) bool { return v&(1<<i) != 0 },
		"bit8":  func(v uint8, i int) bool { return v&(1<<i) != 0 },
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
		"rseq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = n - 1 - i
			}
			return s
		},
		"strips":  s.Strips,
		"section": s.Section,
		"control": func(id string) pro.Control { c, _ := s.Get(id); return c },
	}).ParseFS(files, "templates/*.html")
	if err != nil {
		return nil, err
	}
	srv.tmpl = t

	srv.mux.Handle("GET /static/", http.FileServerFS(files))
	srv.mux.HandleFunc("GET /{$}", srv.index)
	srv.mux.HandleFunc("GET /status", srv.statusFragment)
	srv.mux.HandleFunc("GET /events", srv.events)
	srv.mux.HandleFunc("POST /ev/{id}/{kind}", srv.input)
	return srv, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

// domID turns a control ID like "IN1.SEL" into a CSS-safe element id.
func domID(id string) string { return "c-" + strings.ReplaceAll(id, ".", "-") }

// lcdColor maps the X32 scribble strip colour bits (R=1, G=2, B=4,
// inverted=8) to a CSS background.
func lcdColor(c uint8) template.CSS {
	rgb := c & 7
	if rgb == 0 || c&8 != 0 {
		return "#111"
	}
	return template.CSS(fmt.Sprintf("rgb(%d,%d,%d)", 255*int(rgb&1), 255*int(rgb>>1&1), 255*int(rgb>>2&1)))
}

// lcdInk picks the text colour; inverted strips show coloured text on black.
func lcdInk(c uint8) template.CSS {
	rgb := c & 7
	if c&8 != 0 && rgb != 0 {
		return template.CSS(fmt.Sprintf("rgb(%d,%d,%d)", 255*int(rgb&1), 255*int(rgb>>1&1), 255*int(rgb>>2&1)))
	}
	if rgb == 0 {
		return "#777"
	}
	return "#000"
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.render(w, "page", map[string]any{"Surface": s.surface, "Status": s.status()})
}

func (s *Server) statusFragment(w http.ResponseWriter, r *http.Request) {
	s.render(w, "status", s.status())
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, name, data); err != nil {
		s.log.Error("render", "template", name, "err", err)
	}
}

// input handles POST /ev/{id}/{kind}: press, release, touch, untouch,
// move (form value v = 0..4095) and turn (form value d = steps).
func (s *Server) input(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, ok := s.surface.Get(id)
	if !ok {
		http.Error(w, "unknown control", http.StatusNotFound)
		return
	}
	e := driver.Event{Control: c.ID}
	switch r.PathValue("kind") {
	case "press":
		e.Kind = driver.Press
	case "release":
		e.Kind = driver.Release
	case "touch":
		e.Kind = driver.Touch
	case "untouch":
		e.Kind = driver.Untouch
	case "move":
		v, err := strconv.Atoi(r.FormValue("v"))
		if err != nil {
			http.Error(w, "bad v", http.StatusBadRequest)
			return
		}
		e.Kind, e.Value = driver.Move, v
	case "turn":
		d, err := strconv.Atoi(r.FormValue("d"))
		if err != nil || d == 0 {
			http.Error(w, "bad d", http.StatusBadRequest)
			return
		}
		e.Kind, e.Value = driver.Turn, d
	default:
		http.Error(w, "unknown event", http.StatusNotFound)
		return
	}
	s.drv.Inject(e)
	w.WriteHeader(http.StatusNoContent)
}

// events streams control updates. Each SSE message carries one or more
// <hx-partial> elements that replace the changed controls by id.
func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	ch, cancel := s.drv.Subscribe()
	defer cancel()
	// send headers now so the browser's fetch resolves before the first update
	w.WriteHeader(http.StatusOK)
	fl.Flush()

	// Coalesce bursts (a scene recall touches every control) into one
	// message per 30 ms.
	tick := time.NewTicker(30 * time.Millisecond)
	defer tick.Stop()
	dirty := map[string]bool{}
	var buf bytes.Buffer
	for {
		select {
		case <-r.Context().Done():
			return
		case id := <-ch:
			dirty[id] = true
		case <-tick.C:
			if len(dirty) == 0 {
				continue
			}
			buf.Reset()
			for id := range dirty {
				c, ok := s.surface.Get(id)
				if !ok {
					continue
				}
				fmt.Fprintf(&buf, `<hx-partial hx-target="#%s" hx-swap="outerHTML">`, domID(id))
				if err := s.tmpl.ExecuteTemplate(&buf, "ctl", c); err != nil {
					s.log.Error("render", "control", id, "err", err)
				}
				buf.WriteString("</hx-partial>")
			}
			clear(dirty)
			// SSE data lines cannot contain newlines.
			msg := strings.ReplaceAll(buf.String(), "\n", "")
			if _, err := fmt.Fprintf(w, "data: %s\n\n", msg); err != nil {
				return
			}
			fl.Flush()
		}
	}
}
