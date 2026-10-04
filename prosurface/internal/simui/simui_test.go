package simui

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/driver"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/driver/sim"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/pro"
)

func newServer(t *testing.T) (*httptest.Server, *sim.Driver) {
	t.Helper()
	drv := sim.New()
	ui, err := New(slog.New(slog.NewTextHandler(io.Discard, nil)), pro.New(pro.PRO2), drv, func() Status { return Status{OMCAddr: ":10000", Page: 1, Pages: 1} })
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(ui)
	t.Cleanup(ts.Close)
	return ts, drv
}

func TestPageRendersEveryControl(t *testing.T) {
	ts, drv := newServer(t)
	drv.SetLCD("IN1.SEL", driver.LCD{Color: 2, Lines: []string{"Vox"}})
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	page := string(body)
	for _, c := range pro.New(pro.PRO2).Controls {
		if !strings.Contains(page, `id="`+domID(c.ID)+`"`) {
			t.Errorf("control %s missing from page", c.ID)
		}
	}
	if !strings.Contains(page, "background:rgb(0,255,0)") || !strings.Contains(page, "Vox") {
		t.Error("LCD colour or text not rendered")
	}
	if !strings.Contains(page, `hx-sse:connect="/events"`) {
		t.Error("page does not open the event stream")
	}
}

func TestInputReachesDriver(t *testing.T) {
	ts, drv := newServer(t)
	resp, err := http.Post(ts.URL+"/ev/IN4.FADER/move", "application/x-www-form-urlencoded", strings.NewReader("v=1000"))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d", resp.StatusCode)
	}
	select {
	case e := <-drv.Events():
		if e.Control != "IN4.FADER" || e.Kind != driver.Move || e.Value != 1000 {
			t.Errorf("event %+v", e)
		}
	case <-time.After(time.Second):
		t.Fatal("no event")
	}
	if resp, _ := http.Post(ts.URL+"/ev/NOPE/press", "", nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown control status %d", resp.StatusCode)
	}
}

func TestEventStreamSendsPartials(t *testing.T) {
	ts, drv := newServer(t)
	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	time.Sleep(50 * time.Millisecond) // let the handler subscribe
	drv.SetLED("VCA3.MUTE", true)
	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	msg := string(buf[:n])
	if !strings.HasPrefix(msg, "data: ") || !strings.Contains(msg, `<hx-partial hx-target="#c-VCA3-MUTE" hx-swap="outerHTML">`) || !strings.Contains(msg, "btn on") {
		t.Errorf("unexpected message %q", msg)
	}
}
