// Package scanproc will drive the real PRO1/PRO2 surface through its scan
// processor board. The protocol is not known yet: it needs captures from a
// desk running the stock software (see prosurface/README.md, "Desk
// checklist"). Until then Open always fails, so the daemon refuses to start
// with -driver=scanproc instead of silently doing nothing.
//
// Expected shape once the protocol is decoded:
//
//   - Open the transport the scan board uses (USB bulk/HID, a serial port,
//     or a UDP socket) and run its init/firmware sequence if the stock
//     software sends one.
//   - Translate scan reports into driver.Event with pro control IDs:
//     buttons to Press/Release, faders to Move (scaled to 0..4095) plus
//     Touch/Untouch from the touch sensors, rotaries to Turn.
//   - Translate SetLED/SetFader/SetLCD/SetMeter/SetRing into scan board
//     commands, batching LED updates per refresh if the board wants that.
package scanproc

import (
	"errors"

	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/driver"
)

// ErrUnknownProtocol is returned until the scan processor protocol is decoded.
var ErrUnknownProtocol = errors.New("scanproc: the PRO scan processor protocol is not decoded yet; use -driver=sim, and see the desk checklist in prosurface/README.md")

// Open connects to the scan processor at dev (device path or address).
func Open(dev string) (driver.Driver, error) {
	return nil, ErrUnknownProtocol
}
