package bridge

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/driver"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/mapping"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/pro"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/x32full"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/x32proto"
)

type fakeDriver struct {
	mu    sync.Mutex
	led   map[string]bool
	fader map[string]uint16
	lcd   map[string]driver.LCD
	meter map[string]uint8
	ring  map[string]uint16
	calls []string
}

func newFake() *fakeDriver {
	return &fakeDriver{led: map[string]bool{}, fader: map[string]uint16{}, lcd: map[string]driver.LCD{}, meter: map[string]uint8{}, ring: map[string]uint16{}}
}

func (f *fakeDriver) Events() <-chan driver.Event { return nil }
func (f *fakeDriver) SetLED(id string, on bool)   { f.led[id] = on }
func (f *fakeDriver) SetFader(id string, p uint16) {
	f.fader[id] = p
	f.calls = append(f.calls, fmt.Sprintf("fader %s %d", id, p))
}
func (f *fakeDriver) SetLCD(id string, l driver.LCD) { f.lcd[id] = l }
func (f *fakeDriver) SetMeter(id string, m uint8)    { f.meter[id] = m }
func (f *fakeDriver) SetRing(id string, r uint16)    { f.ring[id] = r }
func (f *fakeDriver) Close() error                   { return nil }

func setup(t *testing.T, m pro.Model) (*Bridge, *fakeDriver, *bytes.Buffer) {
	t.Helper()
	drv := newFake()
	b, err := New(slog.New(slog.NewTextHandler(io.Discard, nil)), x32full.NewTable(), pro.New(m), mapping.Default(m), drv)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	b.SetSender(func(p []byte) { out.Write(p) })
	return b, drv, &out
}

func cmd(board x32proto.Board, class, index byte, data ...byte) x32proto.Command {
	return x32proto.Command{Board: board, Class: class, Index: index, Data: data}
}

func TestOMCToSurface(t *testing.T) {
	b, drv, _ := setup(t, pro.PRO2)

	b.HandleCommand(cmd(x32proto.BoardM, 'L', 0x80, 0x80|0x42))  // BOARD_M_MUTE_3 on
	b.HandleCommand(cmd(x32proto.BoardR, 'F', 0x08, 0x00, 0x0C)) // main fader
	b.HandleCommand(cmd(x32proto.BoardL, 'D', 0x01, 0x05, 0, 0, 0, 0x03, 0, 0, 'K', 'i', 'k'))
	b.HandleCommand(cmd(x32proto.BoardL, 'M', 0x00, 0x1F))
	b.HandleCommand(cmd(x32proto.BoardMain, 'R', 0x0D, 0x7F, 0x80)) // display encoder 1

	if !drv.led["IN11.MUTE"] {
		t.Error("board M mute 3 should light IN11.MUTE on a PRO2")
	}
	if drv.fader["MASTER_L.FADER"] != 0x0C00 || drv.fader["MASTER_R.FADER"] != 0x0C00 {
		t.Errorf("masters %d %d", drv.fader["MASTER_L.FADER"], drv.fader["MASTER_R.FADER"])
	}
	if l := drv.lcd["IN2.SEL"]; l.Color != 5 || len(l.Lines) != 1 || l.Lines[0] != "Kik" {
		t.Errorf("lcd %+v", l)
	}
	if drv.meter["IN1.METER"] != 0x1F {
		t.Errorf("meter %x", drv.meter["IN1.METER"])
	}
	if drv.ring["ROT1"] != 0x7F {
		t.Errorf("ring %x", drv.ring["ROT1"])
	}
}

func TestSurfaceToOMC(t *testing.T) {
	b, _, out := setup(t, pro.PRO2)

	b.HandleEvent(driver.Event{Control: "VCA2.SEL", Kind: driver.Press})
	if !bytes.Equal(out.Bytes(), x32proto.EncodeButton(x32proto.BoardR, 0x21, true)) {
		t.Errorf("VCA2.SEL press sent % X", out.Bytes())
	}
	out.Reset()
	b.HandleEvent(driver.Event{Control: "IN9.FADER", Kind: driver.Move, Value: 1234})
	if !bytes.Equal(out.Bytes(), x32proto.EncodeFader(x32proto.BoardM, 0, 1234)) {
		t.Errorf("IN9 fader sent % X", out.Bytes())
	}
	out.Reset()
	b.HandleEvent(driver.Event{Control: "ROT7", Kind: driver.Turn, Value: -1})
	if !bytes.Equal(out.Bytes(), x32proto.EncodeEncoder(x32proto.BoardMain, 0x00, -1)) {
		t.Errorf("ROT7 sent % X", out.Bytes())
	}
	out.Reset()
	b.HandleEvent(driver.Event{Control: "MASTER_M.FADER", Kind: driver.Move, Value: 100})
	if out.Len() != 0 {
		t.Errorf("unmapped mono master sent % X", out.Bytes())
	}
}

func TestMasterFadersFollowEachOther(t *testing.T) {
	b, drv, _ := setup(t, pro.PRO2)
	b.HandleEvent(driver.Event{Control: "MASTER_L.FADER", Kind: driver.Touch})
	b.HandleEvent(driver.Event{Control: "MASTER_L.FADER", Kind: driver.Move, Value: 3000})
	if drv.fader["MASTER_R.FADER"] != 3000 {
		t.Errorf("right master at %d, want 3000", drv.fader["MASTER_R.FADER"])
	}
}

func TestTouchedFaderIsNotDriven(t *testing.T) {
	b, drv, _ := setup(t, pro.PRO2)
	b.HandleEvent(driver.Event{Control: "IN1.FADER", Kind: driver.Touch})
	b.HandleEvent(driver.Event{Control: "IN1.FADER", Kind: driver.Move, Value: 2000})
	drv.calls = nil
	b.HandleCommand(cmd(x32proto.BoardL, 'F', 0x00, 0x00, 0x01)) // OMC wants 256
	if len(drv.calls) != 0 {
		t.Errorf("motor driven while touched: %v", drv.calls)
	}
	b.HandleEvent(driver.Event{Control: "IN1.FADER", Kind: driver.Untouch})
	if drv.fader["IN1.FADER"] != 256 {
		t.Errorf("after release fader at %d, want 256", drv.fader["IN1.FADER"])
	}
}

func TestExtendPagesOnEightStripDesks(t *testing.T) {
	b, drv, out := setup(t, pro.PRO2C)

	b.HandleCommand(cmd(x32proto.BoardL, 'F', 0x00, 0x10, 0x00)) // L1 = 16
	b.HandleCommand(cmd(x32proto.BoardM, 'F', 0x00, 0x20, 0x00)) // M1 = 32
	b.HandleCommand(cmd(x32proto.BoardM, 'L', 0x80, 0x80|0x40))  // M mute 1 on
	if drv.fader["IN1.FADER"] != 16 || drv.led["IN1.MUTE"] {
		t.Fatalf("page 1: fader %d mute %v", drv.fader["IN1.FADER"], drv.led["IN1.MUTE"])
	}

	b.HandleEvent(driver.Event{Control: "EXTEND", Kind: driver.Press})
	if b.Page() != 2 || !drv.led["EXTEND"] {
		t.Fatalf("page %d, extend LED %v", b.Page(), drv.led["EXTEND"])
	}
	if drv.fader["IN1.FADER"] != 32 || !drv.led["IN1.MUTE"] {
		t.Errorf("page 2: fader %d mute %v", drv.fader["IN1.FADER"], drv.led["IN1.MUTE"])
	}
	if out.Len() != 0 {
		t.Errorf("EXTEND must stay local, sent % X", out.Bytes())
	}
	b.HandleEvent(driver.Event{Control: "IN1.SEL", Kind: driver.Press})
	if !bytes.Equal(out.Bytes(), x32proto.EncodeButton(x32proto.BoardM, 0x20, true)) {
		t.Errorf("page 2 select sent % X", out.Bytes())
	}

	b.HandleEvent(driver.Event{Control: "EXTEND", Kind: driver.Press})
	if b.Page() != 1 || drv.led["EXTEND"] || drv.fader["IN1.FADER"] != 16 {
		t.Errorf("back on page %d, fader %d", b.Page(), drv.fader["IN1.FADER"])
	}
}

func TestSharedLEDNumberLightsTheLED(t *testing.T) {
	// OMC numbers EQ_PEQ_LED and DISPLAY_ENCODER_BUTTON_4 both 0x1A on the
	// main board; the frame must not light the PRO's rotary button 4.
	b, drv, _ := setup(t, pro.PRO2)
	b.HandleCommand(cmd(x32proto.BoardMain, 'L', 0x80, 0x80|0x1A))
	if drv.led["ROT4.BTN"] {
		t.Error("EQ LED frame lit ROT4.BTN")
	}
}
