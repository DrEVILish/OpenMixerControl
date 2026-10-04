// Package bridge connects OMC's X32 surface protocol to a PRO surface
// driver. It keeps the state OMC last set for every X32 element, so the PRO
// can show any page of it (EXTEND) and redraw after a reconnect.
package bridge

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/driver"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/mapping"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/pro"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/x32full"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/x32proto"
)

// Bridge translates in both directions. Its methods are safe for concurrent use.
type Bridge struct {
	mu      sync.Mutex
	log     *slog.Logger
	table   *x32full.Table
	surface *pro.Surface
	m       mapping.Map
	drv     driver.Driver
	send    func([]byte)

	page    int
	touched map[string]bool

	// state as OMC set it, by X32 element name
	led   map[string]bool
	fader map[string]uint16
	lcd   map[string]driver.LCD
	meter map[string]uint8
	ring  map[string]uint16

	byControl map[string][]mapping.Rule
	byElement map[string][]mapping.Rule
	byLCD     map[string][]mapping.Rule
}

// New validates the mapping and builds a bridge. Call SetSender before
// traffic flows to OMC and Run to consume driver events.
func New(log *slog.Logger, t *x32full.Table, s *pro.Surface, m mapping.Map, drv driver.Driver) (*Bridge, error) {
	if err := m.Validate(s, t); err != nil {
		return nil, err
	}
	b := &Bridge{
		log: log, table: t, surface: s, m: m, drv: drv,
		page:      1,
		touched:   map[string]bool{},
		led:       map[string]bool{},
		fader:     map[string]uint16{},
		lcd:       map[string]driver.LCD{},
		meter:     map[string]uint8{},
		ring:      map[string]uint16{},
		byControl: map[string][]mapping.Rule{},
		byElement: map[string][]mapping.Rule{},
		byLCD:     map[string][]mapping.Rule{},
	}
	for _, r := range m.Rules {
		b.byControl[r.Control] = append(b.byControl[r.Control], r)
		b.byElement[r.Element] = append(b.byElement[r.Element], r)
		if r.LCD != "" {
			b.byLCD[r.LCD] = append(b.byLCD[r.LCD], r)
		}
	}
	b.redrawLocked()
	return b, nil
}

// SetSender sets where encoded events for OMC go; nil drops them (no OMC connected).
func (b *Bridge) SetSender(send func([]byte)) {
	b.mu.Lock()
	b.send = send
	b.mu.Unlock()
}

// Page returns the current input page (1-based).
func (b *Bridge) Page() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.page
}

// Run feeds driver events into the bridge until ctx ends or the driver stops.
func (b *Bridge) Run(ctx context.Context) {
	ev := b.drv.Events()
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-ev:
			if !ok {
				return
			}
			b.HandleEvent(e)
		}
	}
}

func (b *Bridge) activeRule(control string) (mapping.Rule, bool) {
	for _, r := range b.byControl[control] {
		if r.Active(b.page) {
			return r, true
		}
	}
	return mapping.Rule{}, false
}

func (b *Bridge) kind(control string) pro.Kind {
	c, _ := b.surface.Get(control)
	return c.Kind
}

// HandleCommand applies one frame from OMC.
func (b *Bridge) HandleCommand(c x32proto.Command) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch c.Class {
	case x32proto.ClassLED:
		l, ok := x32proto.ParseLED(c)
		if !ok {
			break
		}
		for _, e := range b.table.LEDs(l.Board, l.Index) {
			b.led[e.Name] = l.On
			for _, r := range b.byElement[e.Name] {
				if k := b.kind(r.Control); r.Active(b.page) && (k == pro.Button || k == pro.LCDButton) {
					b.drv.SetLED(r.Control, l.On)
				}
			}
		}
	case x32proto.ClassFader:
		f, ok := x32proto.ParseFader(c)
		if !ok {
			break
		}
		if e := b.table.Fader(f.Board, f.Index); e != nil {
			b.fader[e.Name] = f.Position
			for _, r := range b.byElement[e.Name] {
				if r.Active(b.page) && b.kind(r.Control) == pro.Fader && !b.touched[r.Control] {
					b.drv.SetFader(r.Control, f.Position)
				}
			}
		}
	case x32proto.ClassDisplay:
		l, ok := x32proto.ParseLCD(c)
		if !ok {
			break
		}
		if e := b.table.LCD(l.Board, l.Index); e != nil {
			d := driver.LCD{Color: l.Color}
			for _, t := range l.Texts {
				d.Lines = append(d.Lines, t.Text)
			}
			b.lcd[e.Name] = d
			for _, r := range b.byLCD[e.Name] {
				if r.Active(b.page) {
					b.drv.SetLCD(r.Control, d)
				}
			}
		}
	case x32proto.ClassMeter:
		m, ok := x32proto.ParseMeter(c)
		if !ok {
			break
		}
		if e := b.table.Fader(m.Board, m.Index); e != nil {
			b.meter[e.Name] = m.LEDs
			for _, r := range b.byElement[e.Name] {
				if r.Active(b.page) && b.kind(r.Control) == pro.Meter {
					b.drv.SetMeter(r.Control, m.LEDs)
				}
			}
		}
	case x32proto.ClassRing:
		g, ok := x32proto.ParseRing(c)
		if !ok {
			break
		}
		if e := b.table.Encoder(g.Board, g.Index); e != nil {
			b.ring[e.Name] = g.LEDs
			for _, r := range b.byElement[e.Name] {
				if r.Active(b.page) && b.kind(r.Control) == pro.Rotary {
					b.drv.SetRing(r.Control, g.LEDs)
				}
			}
		}
	case x32proto.ClassControl:
		// brightness and contrast; the PRO has its own dimmer
	default:
		b.log.Debug("unhandled OMC command", "cmd", c.String())
	}
}

// HandleEvent applies one input event from the surface.
func (b *Bridge) HandleEvent(e driver.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if e.Control == b.m.PageButton && b.m.PageButton != "" {
		if e.Kind == driver.Press && b.m.Pages > 1 {
			b.page = b.page%b.m.Pages + 1
			b.log.Info("input page", "page", b.page)
			b.redrawLocked()
		}
		return
	}

	r, ok := b.activeRule(e.Control)
	if !ok {
		b.log.Debug("unmapped control", "control", e.Control, "event", e.Kind)
		return
	}
	el, err := b.table.Get(r.Element)
	if err != nil {
		return
	}

	switch e.Kind {
	case driver.Press, driver.Release:
		b.out(x32proto.EncodeButton(el.Board, el.Index, e.Kind == driver.Press))
	case driver.Move:
		pos := clampFader(e.Value)
		b.fader[el.Name] = pos
		b.out(x32proto.EncodeFader(el.Board, el.Index, pos))
		// other faders on the same element follow (Left and Right masters)
		for _, o := range b.byElement[el.Name] {
			if o.Control != e.Control && o.Active(b.page) && b.kind(o.Control) == pro.Fader && !b.touched[o.Control] {
				b.drv.SetFader(o.Control, pos)
			}
		}
	case driver.Touch:
		b.touched[e.Control] = true
	case driver.Untouch:
		delete(b.touched, e.Control)
		// snap to where OMC wants the fader now
		b.drv.SetFader(e.Control, b.fader[el.Name])
	case driver.Turn:
		b.out(x32proto.EncodeEncoder(el.Board, el.Index, e.Value))
	}
}

func clampFader(v int) uint16 {
	if v < 0 {
		return 0
	}
	if v > x32proto.FaderMax {
		return x32proto.FaderMax
	}
	return uint16(v)
}

func (b *Bridge) out(p []byte) {
	if b.send != nil {
		b.send(p)
	}
}

// Redraw pushes the whole state to the driver, for example after it reconnects.
func (b *Bridge) Redraw() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.redrawLocked()
}

func (b *Bridge) redrawLocked() {
	for _, c := range b.surface.Controls {
		if c.ID == b.m.PageButton {
			b.drv.SetLED(c.ID, b.page > 1)
			continue
		}
		r, ok := b.activeRule(c.ID)
		switch c.Kind {
		case pro.Button:
			b.drv.SetLED(c.ID, ok && b.led[r.Element])
		case pro.LCDButton:
			b.drv.SetLED(c.ID, ok && b.led[r.Element])
			var l driver.LCD
			if ok {
				l = b.lcd[r.LCD]
			}
			b.drv.SetLCD(c.ID, l)
		case pro.Fader:
			if !b.touched[c.ID] {
				var p uint16
				if ok {
					p = b.fader[r.Element]
				}
				b.drv.SetFader(c.ID, p)
			}
		case pro.Meter:
			var m uint8
			if ok {
				m = b.meter[r.Element]
			}
			b.drv.SetMeter(c.ID, m)
		case pro.Rotary:
			var g uint16
			if ok {
				g = b.ring[r.Element]
			}
			b.drv.SetRing(c.ID, g)
		}
	}
}

// Describe returns a one-line summary for logs.
func (b *Bridge) Describe() string {
	return fmt.Sprintf("%s, %d rules, %d page(s)", b.surface.Model, len(b.m.Rules), b.m.Pages)
}
