// Package mapping ties PRO surface controls to X32 Full surface elements.
//
// A mapping is a list of rules. Each rule binds one PRO control to one X32
// element, optionally on one page only. Several controls may share an
// element (the PRO's Left and Right masters both drive OMC's main fader).
// On the 8-strip desks (PRO1, PRO2C) the EXTEND button flips the input
// strips between page 1 (X32 board L) and page 2 (X32 board M).
package mapping

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/pro"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/x32full"
)

// Rule binds a PRO control to an X32 element.
type Rule struct {
	Control string `json:"control"`
	Element string `json:"element"`
	// LCD names the X32 scribble strip shown on an LCD select button.
	LCD string `json:"lcd,omitempty"`
	// Page restricts the rule to one page (1-based); 0 means every page.
	Page int `json:"page,omitempty"`
}

// Map is a complete mapping for one model.
type Map struct {
	Model pro.Model `json:"model"`
	// Pages is the number of input pages; 1 disables paging.
	Pages int `json:"pages"`
	// PageButton is the PRO control that cycles pages. It is handled locally
	// and never sent to OMC.
	PageButton string `json:"pageButton,omitempty"`
	Rules      []Rule `json:"rules"`
}

// Default returns the built-in mapping for a model.
func Default(m pro.Model) Map {
	mp := Map{Model: m, Pages: 1, PageButton: "EXTEND"}
	add := func(r Rule) { mp.Rules = append(mp.Rules, r) }
	stripRules := func(ctl string, board string, n int, page int, meter bool) {
		x := fmt.Sprintf("BOARD_%s_", board)
		add(Rule{Control: ctl + ".SEL", Element: fmt.Sprintf("%sSELECT_%d", x, n), LCD: fmt.Sprintf("%sLCD_%d", x, n), Page: page})
		add(Rule{Control: ctl + ".SOLO", Element: fmt.Sprintf("%sSOLO_%d", x, n), Page: page})
		add(Rule{Control: ctl + ".MUTE", Element: fmt.Sprintf("%sMUTE_%d", x, n), Page: page})
		add(Rule{Control: ctl + ".FADER", Element: fmt.Sprintf("%sFADER_%d", x, n), Page: page})
		if meter {
			add(Rule{Control: ctl + ".METER", Element: fmt.Sprintf("%sFADER_%d", x, n), Page: page})
		}
	}

	surface := pro.New(m)
	if surface.InputStrips == 16 {
		for i := 1; i <= 8; i++ {
			stripRules(fmt.Sprintf("IN%d", i), "L", i, 0, true)
			stripRules(fmt.Sprintf("IN%d", i+8), "M", i, 0, true)
		}
	} else {
		mp.Pages = 2
		for i := 1; i <= 8; i++ {
			stripRules(fmt.Sprintf("IN%d", i), "L", i, 1, true)
			stripRules(fmt.Sprintf("IN%d", i), "M", i, 2, true)
		}
	}
	for i := 1; i <= 8; i++ {
		stripRules(fmt.Sprintf("VCA%d", i), "R", i, 0, false)
	}
	for _, ms := range []string{"MASTER_L", "MASTER_R"} {
		add(Rule{Control: ms + ".SEL", Element: "BOARD_R_SELECT_MAIN", LCD: "BOARD_R_LCD_MAIN"})
		add(Rule{Control: ms + ".SOLO", Element: "BOARD_R_SOLO_MAIN"})
		add(Rule{Control: ms + ".MUTE", Element: "BOARD_R_MUTE_MAIN"})
		add(Rule{Control: ms + ".FADER", Element: "BOARD_R_FADER_MAIN"})
	}
	for i := 1; i <= 6; i++ {
		add(Rule{Control: fmt.Sprintf("ROT%d", i), Element: fmt.Sprintf("DISPLAY_ENCODER_%d", i)})
		add(Rule{Control: fmt.Sprintf("ROT%d.BTN", i), Element: fmt.Sprintf("DISPLAY_ENCODER_BUTTON_%d", i)})
	}
	add(Rule{Control: "ROT7", Element: "GAIN_ENCODER"})
	add(Rule{Control: "ROT8", Element: "PAN_BAL_ENCODER"})
	for _, n := range []string{"HOME", "METERS", "ROUTING", "SETUP", "LIBRARY", "EFFECTS", "UTILITY", "UP", "DOWN", "LEFT", "RIGHT"} {
		add(Rule{Control: n, Element: n})
	}
	add(Rule{Control: "FLIP", Element: "SEND_ON_FADER"})
	add(Rule{Control: "SCENE_PREV", Element: "SCENES_PREV"})
	add(Rule{Control: "SCENE_NEXT", Element: "SCENES_NEXT"})
	add(Rule{Control: "SCENE_GO", Element: "SCENES_GO"})
	for i := 1; i <= 6; i++ {
		add(Rule{Control: fmt.Sprintf("MUTEGRP%d", i), Element: fmt.Sprintf("MUTE_GROUP_%d", i)})
	}
	return mp
}

// Load reads a mapping from a JSON file.
func Load(path string) (Map, error) {
	var mp Map
	b, err := os.ReadFile(path)
	if err != nil {
		return mp, err
	}
	if err := json.Unmarshal(b, &mp); err != nil {
		return mp, fmt.Errorf("%s: %w", path, err)
	}
	if mp.Pages < 1 {
		mp.Pages = 1
	}
	return mp, nil
}

// JSON returns the mapping as indented JSON, for editing.
func (mp Map) JSON() ([]byte, error) { return json.MarshalIndent(mp, "", "  ") }

func compatible(c pro.Kind, e x32full.Kind) bool {
	switch c {
	case pro.Button, pro.LCDButton:
		return e == x32full.Button
	case pro.Fader, pro.Meter:
		return e == x32full.Fader
	case pro.Rotary:
		return e == x32full.Encoder
	}
	return false
}

// Validate checks every rule against the surface and the X32 table.
func (mp Map) Validate(s *pro.Surface, t *x32full.Table) error {
	if mp.PageButton != "" {
		if _, ok := s.Get(mp.PageButton); !ok {
			return fmt.Errorf("page button %q is not a %s control", mp.PageButton, s.Model)
		}
	}
	for _, r := range mp.Rules {
		c, ok := s.Get(r.Control)
		if !ok {
			return fmt.Errorf("rule %s -> %s: no such %s control", r.Control, r.Element, s.Model)
		}
		if r.Control == mp.PageButton {
			return fmt.Errorf("rule %s -> %s: the page button cannot be mapped", r.Control, r.Element)
		}
		e, err := t.Get(r.Element)
		if err != nil {
			return fmt.Errorf("rule %s: %w", r.Control, err)
		}
		if !compatible(c.Kind, e.Kind) {
			return fmt.Errorf("rule %s -> %s: a %s cannot drive a %s", r.Control, r.Element, c.Kind, e.Kind)
		}
		if r.LCD != "" {
			l, err := t.Get(r.LCD)
			if err != nil || l.Kind != x32full.Lcd {
				return fmt.Errorf("rule %s: %q is not an X32 scribble strip", r.Control, r.LCD)
			}
		}
		if r.Page < 0 || r.Page > mp.Pages {
			return fmt.Errorf("rule %s: page %d outside 1..%d", r.Control, r.Page, mp.Pages)
		}
	}
	return nil
}

// Active reports whether a rule applies on a page (1-based).
func (r Rule) Active(page int) bool { return r.Page == 0 || r.Page == page }
