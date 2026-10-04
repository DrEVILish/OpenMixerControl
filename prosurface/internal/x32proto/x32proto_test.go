package x32proto

import (
	"bytes"
	"os"
	"testing"
)

// testdata/omc-startup.bin is the first 4 KiB OMC (PC build, --model X32)
// wrote to its surface port at startup, captured with socat -r.
func TestDecodeOMCCapture(t *testing.T) {
	raw, err := os.ReadFile("testdata/omc-startup.bin")
	if err != nil {
		t.Fatal(err)
	}
	var d Decoder
	cmds := d.Feed(raw)
	if len(cmds) < 100 {
		t.Fatalf("decoded %d commands, want at least 100", len(cmds))
	}
	if d.Errors() != 0 {
		t.Errorf("%d malformed frames", d.Errors())
	}
	classes := map[byte]int{}
	for _, c := range cmds {
		if !c.ChecksumOK {
			t.Errorf("bad checksum: %s", c)
		}
		classes[c.Class]++
	}
	for _, want := range []byte{ClassLED, ClassFader, ClassDisplay, ClassMeter, ClassRing} {
		if classes[want] == 0 {
			t.Errorf("no %c frames in capture", want)
		}
	}

	// the same bytes fed one at a time decode identically
	var d2 Decoder
	var n int
	for _, b := range raw {
		n += len(d2.Feed([]byte{b}))
	}
	if n != len(cmds) {
		t.Errorf("byte-wise feed decoded %d commands, want %d", n, len(cmds))
	}
}

func TestEncodeDecodeStuffing(t *testing.T) {
	in := Command{Board: BoardM, Class: ClassFader, Index: 3, Data: []byte{0xFE, 0x0F}}
	raw := EncodeCommand(in)
	want := []byte{0xFE, 0x85, 'F', 3, 0xFE, 0xFF, 0x0F, 0xFE}
	if !bytes.Equal(raw[:len(want)], want) {
		t.Fatalf("encoded % X, want prefix % X", raw, want)
	}
	var d Decoder
	out := d.Feed(raw)
	if len(out) != 1 || !out[0].ChecksumOK || !bytes.Equal(out[0].Data, in.Data) || out[0].Board != BoardM {
		t.Fatalf("decoded %+v", out)
	}
	f, ok := ParseFader(out[0])
	if !ok || f.Position != 0x0FFE {
		t.Errorf("fader %+v", f)
	}
}

func TestParseLCD(t *testing.T) {
	// a scribble strip frame as OMC sent it in the capture
	c := Command{Board: BoardL, Class: ClassDisplay, Index: 0, Data: []byte{
		0x03, 0x00, 0x00, 0x00,
		0x07, 0x00, 0x00, '-', '-', '-', '|', '-', '-', '-',
		0x26, 0x00, 0x14, 'K', 'a', 'n', 'a', 'l', '1',
		0x03, 0x23, 0x33, 'C', 'H', '1',
	}}
	l, ok := ParseLCD(c)
	if !ok || l.Color != 3 || len(l.Texts) != 3 {
		t.Fatalf("lcd %+v", l)
	}
	if l.Texts[1].Text != "Kanal1" || !l.Texts[1].Large || l.Texts[2].Text != "CH1" {
		t.Errorf("texts %+v", l.Texts)
	}
}

func TestParseLEDAndRing(t *testing.T) {
	l, ok := ParseLED(Command{Board: BoardR, Class: ClassLED, Index: 0x80, Data: []byte{0x80 | 0x28}})
	if !ok || !l.On || l.Index != 0x28 {
		t.Errorf("led %+v", l)
	}
	r, ok := ParseRing(Command{Board: BoardMain, Class: ClassRing, Index: 0x0C, Data: []byte{0x40, 0x80}})
	if !ok || r.LEDs != 0x40 || !r.Backlight {
		t.Errorf("ring %+v", r)
	}
}

// The events below were captured from prosurfaced and accepted by OMC
// (button press and release, encoder +5, fader 2000).
func TestEncodeEvents(t *testing.T) {
	cases := []struct {
		got, want []byte
	}{
		{EncodeButton(BoardL, 0x22, true), []byte{0xFE, 0x84, 0x62, 0x22, 0xA2, 0xFE, 0x53}},
		{EncodeButton(BoardL, 0x22, false), []byte{0xFE, 0x84, 0x62, 0x22, 0x22, 0xFE, 0x53}},
		{EncodeEncoder(BoardMain, 0x0E, 5), []byte{0xFE, 0x81, 0x65, 0x0E, 0x05, 0xFE, 0x04}},
		{EncodeFader(BoardL, 3, 2000), []byte{0xFE, 0x84, 0x66, 0x03, 0xD0, 0x07, 0xFE, 0x38}},
	}
	for i, c := range cases {
		if !bytes.Equal(c.got, c.want) {
			t.Errorf("case %d: % X, want % X", i, c.got, c.want)
		}
	}
}

func TestEventsNeverEmbedFrameMarker(t *testing.T) {
	// OMC splits events on 0xFE, so only the two marker positions may hold it.
	check := func(name string, p []byte, frameLen int) {
		for i := 0; i < len(p); i += frameLen {
			f := p[i : i+frameLen]
			for j, b := range f {
				if b == 0xFE && j != 0 && j != frameLen-2 {
					t.Errorf("%s: 0xFE at %d in % X", name, j, f)
				}
			}
		}
	}
	check("fader", EncodeFader(BoardR, 0, 0x01FE), 8)
	check("encoder", EncodeEncoder(BoardMain, 1, -2), 7)
	if n := len(EncodeEncoder(BoardMain, 1, -2)) / 7; n != 2 {
		t.Errorf("-2 split into %d events, want 2", n)
	}
	check("encoder", EncodeEncoder(BoardMain, 1, 300), 7)
}
