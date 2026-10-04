// Package driver defines the boundary between the bridge and the physical
// (or simulated) PRO surface. Controls are addressed by the IDs from package
// pro; fader positions use the X32 range 0..4095 so drivers convert once.
package driver

// EventKind is what happened to a control.
type EventKind int

const (
	Press EventKind = iota
	Release
	// Move carries a fader position in Value (0..4095).
	Move
	// Touch and Untouch report the fader's touch sensor.
	Touch
	Untouch
	// Turn carries a signed rotary step count in Value.
	Turn
)

func (k EventKind) String() string {
	return [...]string{"press", "release", "move", "touch", "untouch", "turn"}[k]
}

// Event is input from the surface.
type Event struct {
	Control string
	Kind    EventKind
	Value   int
}

// LCD is what an LCD select button shows.
type LCD struct {
	// Color: bit 0 red, bit 1 green, bit 2 blue, bit 3 inverted (X32 encoding).
	Color uint8
	Lines []string
}

// Driver is implemented by the real scan-processor driver and the simulator.
// Output methods must not block for long; the bridge calls them while
// handling OMC traffic.
type Driver interface {
	// Events delivers surface input. It is closed when the driver stops.
	Events() <-chan Event
	SetLED(control string, on bool)
	SetFader(control string, pos uint16)
	SetLCD(control string, lcd LCD)
	SetMeter(control string, leds uint8)
	// SetRing shows an encoder ring: 13 LEDs, bit 0 leftmost.
	SetRing(control string, leds uint16)
	Close() error
}
