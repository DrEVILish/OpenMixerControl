// Package pro describes the control surfaces of the Midas PRO1, PRO2 and
// PRO2C as far as they are publicly documented (Midas brochures and owner's
// manuals). Control IDs are our own; the hardware driver translates them to
// whatever the PRO's scan processor uses once that protocol is known.
//
// Sections marked provisional are educated guesses about which buttons exist
// and must be checked against a real desk.
package pro

import "fmt"

// Model is a PRO console model.
type Model string

const (
	PRO1  Model = "pro1"
	PRO2  Model = "pro2"
	PRO2C Model = "pro2c"
)

// ParseModel accepts "pro1", "pro2" or "pro2c".
func ParseModel(s string) (Model, error) {
	switch m := Model(s); m {
	case PRO1, PRO2, PRO2C:
		return m, nil
	}
	return "", fmt.Errorf("unknown model %q (want pro1, pro2 or pro2c)", s)
}

// Kind is the type of a physical control.
type Kind int

const (
	// Button is a push button, usually with an LED.
	Button Kind = iota
	// LCDButton is a channel select button with a colour LCD in its cap.
	LCDButton
	// Fader is a 100 mm motorised, touch-sensitive fader.
	Fader
	// Rotary is an endless encoder.
	Rotary
	// Meter is a strip LED meter. Not confirmed on PRO hardware.
	Meter
)

func (k Kind) String() string {
	return [...]string{"button", "lcd-button", "fader", "rotary", "meter"}[k]
}

// Control is one physical control on the surface.
type Control struct {
	ID      string
	Kind    Kind
	Label   string
	Section string
	// Strip groups the controls of one channel strip (for the simulator layout).
	Strip string
	// Provisional marks controls whose presence or position is unconfirmed.
	Provisional bool
}

// Surface is a model's full list of controls.
type Surface struct {
	Model       Model
	InputStrips int
	Controls    []Control
	byID        map[string]int
}

// Get returns a control by ID.
func (s *Surface) Get(id string) (Control, bool) {
	i, ok := s.byID[id]
	if !ok {
		return Control{}, false
	}
	return s.Controls[i], true
}

// Section returns the controls of one section in layout order.
func (s *Surface) Section(name string) []Control {
	var out []Control
	for _, c := range s.Controls {
		if c.Section == name {
			out = append(out, c)
		}
	}
	return out
}

// Strips returns the strip names of a section in layout order.
func (s *Surface) Strips(section string) []string {
	var out []string
	seen := map[string]bool{}
	for _, c := range s.Controls {
		if c.Section == section && c.Strip != "" && !seen[c.Strip] {
			seen[c.Strip] = true
			out = append(out, c.Strip)
		}
	}
	return out
}

// Section names.
const (
	SecInputs     = "inputs"
	SecVCA        = "vca"
	SecMasters    = "masters"
	SecRotaries   = "rotaries"
	SecNavigation = "navigation"
	SecScenes     = "scenes"
	SecMuteGroups = "mute-groups"
)

// New returns the surface of a model.
func New(m Model) *Surface {
	s := &Surface{Model: m, byID: map[string]int{}}
	add := func(c Control) {
		s.byID[c.ID] = len(s.Controls)
		s.Controls = append(s.Controls, c)
	}
	strip := func(sec, prefix, label string, n int, meter bool) {
		for i := 1; i <= n; i++ {
			st := fmt.Sprintf("%s%d", prefix, i)
			l := fmt.Sprintf("%s %d", label, i)
			add(Control{ID: st + ".SEL", Kind: LCDButton, Label: l, Section: sec, Strip: st})
			add(Control{ID: st + ".SOLO", Kind: Button, Label: "Solo", Section: sec, Strip: st})
			add(Control{ID: st + ".MUTE", Kind: Button, Label: "Mute", Section: sec, Strip: st})
			if meter {
				add(Control{ID: st + ".METER", Kind: Meter, Label: "Meter", Section: sec, Strip: st, Provisional: true})
			}
			add(Control{ID: st + ".FADER", Kind: Fader, Label: l, Section: sec, Strip: st})
		}
	}

	switch m {
	case PRO2:
		s.InputStrips = 16
	default:
		s.InputStrips = 8
	}
	strip(SecInputs, "IN", "Input", s.InputStrips, true)
	strip(SecVCA, "VCA", "VCA", 8, false)

	for _, ms := range []struct{ id, label string }{{"MASTER_L", "Left"}, {"MASTER_R", "Right"}, {"MASTER_M", "Mono"}} {
		add(Control{ID: ms.id + ".SEL", Kind: LCDButton, Label: ms.label, Section: SecMasters, Strip: ms.id})
		add(Control{ID: ms.id + ".SOLO", Kind: Button, Label: "Solo", Section: SecMasters, Strip: ms.id})
		add(Control{ID: ms.id + ".MUTE", Kind: Button, Label: "Mute", Section: SecMasters, Strip: ms.id})
		add(Control{ID: ms.id + ".FADER", Kind: Fader, Label: ms.label, Section: SecMasters, Strip: ms.id})
	}

	for i := 1; i <= 8; i++ {
		st := fmt.Sprintf("ROT%d", i)
		add(Control{ID: st, Kind: Rotary, Label: fmt.Sprintf("Rotary %d", i), Section: SecRotaries, Strip: st})
		add(Control{ID: st + ".BTN", Kind: Button, Label: fmt.Sprintf("Button %d", i), Section: SecRotaries, Strip: st})
	}

	for _, n := range []struct{ id, label string }{
		{"HOME", "Home"}, {"METERS", "Meters"}, {"ROUTING", "Routing"}, {"SETUP", "Setup"},
		{"LIBRARY", "Library"}, {"EFFECTS", "Effects"}, {"UTILITY", "Utility"},
		{"UP", "Up"}, {"DOWN", "Down"}, {"LEFT", "Left"}, {"RIGHT", "Right"},
		{"EXTEND", "Extend"}, {"FLIP", "Flip"},
	} {
		add(Control{ID: n.id, Kind: Button, Label: n.label, Section: SecNavigation, Provisional: n.id != "EXTEND" && n.id != "FLIP"})
	}
	for _, n := range []struct{ id, label string }{{"SCENE_PREV", "Prev"}, {"SCENE_NEXT", "Next"}, {"SCENE_GO", "Go"}} {
		add(Control{ID: n.id, Kind: Button, Label: n.label, Section: SecScenes, Provisional: true})
	}
	for i := 1; i <= 6; i++ {
		add(Control{ID: fmt.Sprintf("MUTEGRP%d", i), Kind: Button, Label: fmt.Sprintf("Mute %d", i), Section: SecMuteGroups, Provisional: true})
	}
	return s
}
