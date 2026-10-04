// Package sim is an in-memory PRO surface. It records what the bridge shows
// on each control and accepts input from the browser simulator (package
// simui), so the whole path to OMC can be exercised without a desk.
package sim

import (
	"slices"
	"sync"

	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/driver"
)

// State is what one control currently shows.
type State struct {
	LED     bool
	Fader   uint16
	LCD     driver.LCD
	Meter   uint8
	Ring    uint16
	Touched bool
}

// Driver implements driver.Driver.
type Driver struct {
	mu     sync.Mutex
	state  map[string]*State
	events chan driver.Event
	subs   map[chan string]struct{}

	// closeMu guards closed and sends on events; it is separate from mu so
	// a full events channel never blocks the bridge's output calls.
	closeMu sync.RWMutex
	closed  bool
}

// New returns an empty simulated surface.
func New() *Driver {
	return &Driver{
		state:  map[string]*State{},
		events: make(chan driver.Event, 256),
		subs:   map[chan string]struct{}{},
	}
}

func (d *Driver) get(id string) *State {
	s, ok := d.state[id]
	if !ok {
		s = &State{}
		d.state[id] = s
	}
	return s
}

// Snapshot returns a copy of a control's state.
func (d *Driver) Snapshot(id string) State {
	d.mu.Lock()
	defer d.mu.Unlock()
	s := *d.get(id)
	s.LCD.Lines = slices.Clone(s.LCD.Lines)
	return s
}

// Subscribe returns a channel of control IDs whose state changed. A slow
// subscriber misses updates rather than blocking the bridge.
func (d *Driver) Subscribe() (<-chan string, func()) {
	ch := make(chan string, 1024)
	d.mu.Lock()
	d.subs[ch] = struct{}{}
	d.mu.Unlock()
	return ch, func() {
		d.mu.Lock()
		delete(d.subs, ch)
		d.mu.Unlock()
	}
}

// update applies f to a control's state and notifies subscribers if it changed.
func (d *Driver) update(id string, f func(*State)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	s := d.get(id)
	before := *s
	before.LCD.Lines = slices.Clone(s.LCD.Lines)
	f(s)
	if before.LED == s.LED && before.Fader == s.Fader && before.Meter == s.Meter &&
		before.Ring == s.Ring && before.Touched == s.Touched &&
		before.LCD.Color == s.LCD.Color && slices.Equal(before.LCD.Lines, s.LCD.Lines) {
		return
	}
	for ch := range d.subs {
		select {
		case ch <- id:
		default:
		}
	}
}

// Inject delivers input from the simulator UI to the bridge.
func (d *Driver) Inject(e driver.Event) {
	switch e.Kind {
	case driver.Move:
		d.update(e.Control, func(s *State) { s.Fader = uint16(max(0, min(e.Value, 4095))) })
	case driver.Touch:
		d.update(e.Control, func(s *State) { s.Touched = true })
	case driver.Untouch:
		d.update(e.Control, func(s *State) { s.Touched = false })
	}
	d.closeMu.RLock()
	defer d.closeMu.RUnlock()
	if !d.closed {
		d.events <- e
	}
}

func (d *Driver) Events() <-chan driver.Event { return d.events }

func (d *Driver) SetLED(id string, on bool) { d.update(id, func(s *State) { s.LED = on }) }

func (d *Driver) SetFader(id string, pos uint16) { d.update(id, func(s *State) { s.Fader = pos }) }

func (d *Driver) SetLCD(id string, l driver.LCD) {
	d.update(id, func(s *State) { s.LCD = driver.LCD{Color: l.Color, Lines: slices.Clone(l.Lines)} })
}

func (d *Driver) SetMeter(id string, leds uint8) { d.update(id, func(s *State) { s.Meter = leds }) }

func (d *Driver) SetRing(id string, leds uint16) { d.update(id, func(s *State) { s.Ring = leds }) }

func (d *Driver) Close() error {
	d.closeMu.Lock()
	defer d.closeMu.Unlock()
	if !d.closed {
		d.closed = true
		close(d.events)
	}
	return nil
}
