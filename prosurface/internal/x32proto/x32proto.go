// Package x32proto implements the serial protocol between OMC and the
// Behringer X32 / Midas M32 surface boards, as used by
// src/surface-controller-xm32.cpp and src/surface.cpp in OMC.
//
// OMC -> surface (commands):
//
//	FE  80+board  class  index  data...  FE  checksum
//
// A data byte of 0xFE is sent as FE FF (byte stuffing). Text in LCD frames
// is ASCII and is not stuffed. The checksum is
// (0xFE - sum(all bytes before the end marker) - (len-3)) & 0x7F, where len
// counts every byte up to and including the end marker.
//
// Surface -> OMC (events):
//
//	FE  80+board                       board header, remembered by OMC
//	class  index  value  FE  checksum  short event (buttons, encoders)
//	class  index  lsb  msb  FE  chk    long event (faders)
//
// OMC does not verify event checksums, but we send the documented formula
// anyway so a real X32 would accept the frames too.
package x32proto

import "fmt"

// Board addresses a surface board on the X32 serial bus.
type Board uint8

const (
	BoardExtra Board = 0x00 // assign section, scenes, mute groups
	BoardMain  Board = 0x01 // channel strip, display encoders, view buttons
	BoardL     Board = 0x04 // left fader bank
	BoardM     Board = 0x05 // middle fader bank (X32 Full and M32 only)
	BoardR     Board = 0x08 // right fader bank and main fader
)

func (b Board) String() string {
	switch b {
	case BoardExtra:
		return "EXTRA"
	case BoardMain:
		return "MAIN"
	case BoardL:
		return "L"
	case BoardM:
		return "M"
	case BoardR:
		return "R"
	}
	return fmt.Sprintf("0x%02X", uint8(b))
}

// Command classes sent by OMC.
const (
	ClassLED     byte = 'L'
	ClassFader   byte = 'F'
	ClassDisplay byte = 'D'
	ClassMeter   byte = 'M'
	ClassRing    byte = 'R'
	ClassControl byte = 'C' // index 'B' brightness, 'C' contrast
)

// Event classes sent by the surface.
const (
	EventFader   byte = 'f'
	EventButton  byte = 'b'
	EventEncoder byte = 'e'
)

// FaderMax is the full-scale fader value on the X32 bus (12 bit).
const FaderMax = 0x0FFF

// Command is one decoded frame from OMC, with byte stuffing removed.
type Command struct {
	Board Board
	Class byte
	Index byte
	Data  []byte
	// ChecksumOK reports whether the received checksum matched.
	ChecksumOK bool
}

func (c Command) String() string {
	return fmt.Sprintf("%s %c idx=0x%02X data=% X", c.Board, c.Class, c.Index, c.Data)
}

// Checksum computes the OMC frame checksum over raw (stuffed) bytes that
// start with the 0xFE start marker and end with the 0xFE end marker.
func Checksum(raw []byte) byte {
	sum := 0xFE
	for _, b := range raw[:len(raw)-1] {
		sum -= int(b)
	}
	sum -= len(raw) - 3
	return byte(sum & 0x7F)
}

// EncodeCommand builds a frame the way OMC's SurfaceMessage does. It is used
// by tests and by tools that pretend to be OMC.
func EncodeCommand(c Command) []byte {
	raw := []byte{0xFE, 0x80 | byte(c.Board), c.Class, c.Index}
	for _, b := range c.Data {
		raw = append(raw, b)
		if b == 0xFE {
			raw = append(raw, 0xFF)
		}
	}
	raw = append(raw, 0xFE)
	return append(raw, Checksum(raw))
}

// Decoder turns a byte stream from OMC into commands. It keeps partial
// frames between Feed calls, so reads may split frames anywhere.
type Decoder struct {
	state  int
	raw    []byte
	board  Board
	body   []byte
	errors int
}

const (
	stIdle = iota
	stHeader
	stBody
	stMaybeEnd
)

// Errors returns the number of malformed frames skipped so far.
func (d *Decoder) Errors() int { return d.errors }

// Feed consumes bytes and returns every command completed by them.
func (d *Decoder) Feed(p []byte) []Command {
	var out []Command
	for _, b := range p {
		switch d.state {
		case stIdle:
			if b == 0xFE {
				d.raw = append(d.raw[:0], b)
				d.state = stHeader
			}
		case stHeader:
			if b&0x80 == 0 || b == 0xFE {
				d.errors++
				d.state = stIdle
				if b == 0xFE {
					d.raw = append(d.raw[:0], b)
					d.state = stHeader
				}
				continue
			}
			d.raw = append(d.raw, b)
			d.board = Board(b & 0x7F)
			d.body = d.body[:0]
			d.state = stBody
		case stBody:
			d.raw = append(d.raw, b)
			if b == 0xFE {
				d.state = stMaybeEnd
			} else {
				d.body = append(d.body, b)
			}
		case stMaybeEnd:
			if b == 0xFF {
				// stuffed 0xFE data byte
				d.raw = append(d.raw, b)
				d.body = append(d.body, 0xFE)
				d.state = stBody
				continue
			}
			// b is the checksum
			d.state = stIdle
			if len(d.body) < 2 {
				d.errors++
				continue
			}
			c := Command{
				Board:      d.board,
				Class:      d.body[0],
				Index:      d.body[1],
				Data:       append([]byte(nil), d.body[2:]...),
				ChecksumOK: Checksum(d.raw) == b,
			}
			out = append(out, c)
		}
	}
	return out
}

// eventChecksum follows the formula published by the OpenX32 project:
// (0xFE - board - class - index - sum(data) - len(data)) & 0x7F.
func eventChecksum(board Board, class, index byte, data []byte) byte {
	sum := 0xFE - int(board) - int(class) - int(index)
	for _, b := range data {
		sum -= int(b)
	}
	sum -= len(data)
	return byte(sum & 0x7F)
}

func event(board Board, class, index byte, data ...byte) []byte {
	out := []byte{0xFE, 0x80 | byte(board), class, index}
	out = append(out, data...)
	out = append(out, 0xFE, eventChecksum(board, class, index, data))
	return out
}

// EncodeFader reports a fader position (0..FaderMax) to OMC.
//
// OMC's event parser splits on 0xFE and has no unstuffing, so a low byte of
// 0xFE would corrupt the stream; it is sent as 0xFD instead (one step of
// 4096, inaudible).
func EncodeFader(board Board, index uint8, pos uint16) []byte {
	if pos > FaderMax {
		pos = FaderMax
	}
	lsb := byte(pos & 0xFF)
	if lsb == 0xFE {
		lsb = 0xFD
	}
	return event(board, EventFader, index, lsb, byte(pos>>8))
}

// EncodeButton reports a button press or release. OMC looks the button up
// by the low 7 bits of the value and reads the press state from bit 7.
func EncodeButton(board Board, code uint8, pressed bool) []byte {
	v := code & 0x7F
	if pressed {
		v |= 0x80
	}
	return event(board, EventButton, code&0x7F, v)
}

// EncodeEncoder reports a relative encoder turn. Large deltas are split into
// several events; a step of -2 (0xFE) is sent as two steps of -1 so the
// byte never collides with the frame marker.
func EncodeEncoder(board Board, index uint8, delta int) []byte {
	var out []byte
	for delta != 0 {
		step := delta
		if step > 127 {
			step = 127
		}
		if step < -127 {
			step = -127
		}
		if step == -2 {
			step = -1
		}
		out = append(out, event(board, EventEncoder, index, byte(int8(step)))...)
		delta -= step
	}
	return out
}

// LED is a decoded 'L' command.
type LED struct {
	Board Board
	Index uint8
	On    bool
}

// ParseLED decodes an 'L' command: index is fixed at 0x80 and the data byte
// carries the LED number in bits 0-6 and the state in bit 7.
func ParseLED(c Command) (LED, bool) {
	if c.Class != ClassLED || len(c.Data) < 1 {
		return LED{}, false
	}
	return LED{Board: c.Board, Index: c.Data[0] & 0x7F, On: c.Data[0]&0x80 != 0}, true
}

// Fader is a decoded 'F' command (motor fader target position).
type Fader struct {
	Board    Board
	Index    uint8
	Position uint16
}

// ParseFader decodes an 'F' command.
func ParseFader(c Command) (Fader, bool) {
	if c.Class != ClassFader || len(c.Data) < 2 {
		return Fader{}, false
	}
	return Fader{Board: c.Board, Index: c.Index, Position: uint16(c.Data[0]) | uint16(c.Data[1]&0x0F)<<8}, true
}

// LCDText is one text item of a scribble strip.
type LCDText struct {
	Large bool
	X, Y  uint8
	Text  string
}

// LCD is a decoded 'D' command for a scribble strip.
type LCD struct {
	Board Board
	Index uint8
	// Color: bit 0 red, bit 1 green, bit 2 blue, bit 3 inverted.
	Color uint8
	Icon  uint8
	Texts []LCDText
}

// ParseLCD decodes a 'D' command. The X32 Rack 7-segment frame (index 0x80)
// is not a scribble strip and is rejected.
func ParseLCD(c Command) (LCD, bool) {
	if c.Class != ClassDisplay || c.Index == 0x80 || len(c.Data) < 4 {
		return LCD{}, false
	}
	l := LCD{Board: c.Board, Index: c.Index, Color: c.Data[0] & 0x0F, Icon: c.Data[1]}
	d := c.Data[4:]
	for len(d) >= 3 {
		sizeLen := d[0]
		large := sizeLen&0x20 != 0
		n := int(sizeLen &^ 0x20)
		if len(d) < 3+n {
			n = len(d) - 3
		}
		l.Texts = append(l.Texts, LCDText{Large: large, X: d[1], Y: d[2], Text: string(d[3 : 3+n])})
		d = d[3+n:]
	}
	return l, true
}

// Meter is a decoded 'M' command for a channel strip meter.
type Meter struct {
	Board Board
	Index uint8
	// LEDs: bit 0 -60 dB ... bit 4 -6 dB, bit 5 clip, bit 6 gate, bit 7 comp.
	LEDs uint8
}

// ParseMeter decodes a strip 'M' command. The main meter frames on the main
// board carry a different layout and are rejected.
func ParseMeter(c Command) (Meter, bool) {
	if c.Class != ClassMeter || len(c.Data) != 1 {
		return Meter{}, false
	}
	return Meter{Board: c.Board, Index: c.Index, LEDs: c.Data[0]}, true
}

// Ring is a decoded 'R' command for an encoder LED ring.
type Ring struct {
	Board     Board
	Index     uint8
	LEDs      uint16 // 13 ring LEDs, bit 0 = leftmost
	Backlight bool
}

// ParseRing decodes an 'R' command.
func ParseRing(c Command) (Ring, bool) {
	if c.Class != ClassRing || len(c.Data) < 2 {
		return Ring{}, false
	}
	return Ring{
		Board:     c.Board,
		Index:     c.Index,
		LEDs:      uint16(c.Data[0]) | uint16(c.Data[1]&0x7F)<<8,
		Backlight: c.Data[1]&0x80 != 0,
	}, true
}
