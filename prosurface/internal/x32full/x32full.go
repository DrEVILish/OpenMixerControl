// Package x32full holds the X32 Full surface element table that OMC uses
// (src/x32config.cpp), and lookups from wire addresses to element names.
package x32full

//go:generate go run ./gen -src ../../../src/x32config.cpp -out elements_gen.go

import (
	"fmt"

	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/x32proto"
)

// Kind is the type of a surface element.
type Kind int

const (
	Button Kind = iota
	Led
	Encoder
	Fader
	Lcd
)

func (k Kind) String() string {
	return [...]string{"button", "led", "encoder", "fader", "lcd"}[k]
}

// Element is one X32 surface element, named after OMC's SurfaceElementId.
type Element struct {
	Name  string
	Label string
	Kind  Kind
	Board x32proto.Board
	// Index is the button/LED number, encoder number, fader number or LCD number.
	Index uint8
	// Backlight is the LED number of an encoder's backlight, 0 if none.
	Backlight uint8
	// NoLED marks buttons without an LED.
	NoLED bool
}

type addr struct {
	board x32proto.Board
	index uint8
}

// Table indexes the elements by name and by wire address.
type Table struct {
	byName   map[string]*Element
	leds     map[addr][]*Element // buttons with LEDs, plain LEDs; OMC reuses some numbers
	faders   map[addr]*Element
	encoders map[addr]*Element
	lcds     map[addr]*Element
}

// Elements returns all X32 Full elements in OMC's definition order.
func Elements() []Element { return elements }

// NewTable builds the lookup table.
func NewTable() *Table {
	t := &Table{
		byName:   map[string]*Element{},
		leds:     map[addr][]*Element{},
		faders:   map[addr]*Element{},
		encoders: map[addr]*Element{},
		lcds:     map[addr]*Element{},
	}
	for i := range elements {
		e := &elements[i]
		t.byName[e.Name] = e
		a := addr{e.Board, e.Index}
		switch e.Kind {
		case Button:
			if !e.NoLED {
				t.leds[a] = append(t.leds[a], e)
			}
		case Led:
			t.leds[a] = append(t.leds[a], e)
		case Fader:
			t.faders[a] = e
		case Encoder:
			t.encoders[a] = e
		case Lcd:
			t.lcds[a] = e
		}
	}
	// Where a plain LED and a button share a number (the EQ type LEDs and the
	// display encoder buttons on the main board), the frame is for the LED.
	for a, els := range t.leds {
		var onlyLEDs []*Element
		for _, e := range els {
			if e.Kind == Led {
				onlyLEDs = append(onlyLEDs, e)
			}
		}
		if len(onlyLEDs) > 0 {
			t.leds[a] = onlyLEDs
		}
	}
	return t
}

// Get returns the element with the given name.
func (t *Table) Get(name string) (*Element, error) {
	e, ok := t.byName[name]
	if !ok {
		return nil, fmt.Errorf("unknown X32 element %q", name)
	}
	return e, nil
}

// LEDs returns the buttons and LEDs lit by an 'L' command. Several buttons
// can share one LED number (OMC reuses a few), so there can be more than one.
func (t *Table) LEDs(b x32proto.Board, index uint8) []*Element { return t.leds[addr{b, index}] }

// Fader returns the fader moved by an 'F' command, or the strip addressed by an 'M' command.
func (t *Table) Fader(b x32proto.Board, index uint8) *Element { return t.faders[addr{b, index}] }

// Encoder returns the encoder addressed by an 'R' command.
func (t *Table) Encoder(b x32proto.Board, index uint8) *Element { return t.encoders[addr{b, index}] }

// LCD returns the scribble strip addressed by a 'D' command.
func (t *Table) LCD(b x32proto.Board, index uint8) *Element { return t.lcds[addr{b, index}] }
