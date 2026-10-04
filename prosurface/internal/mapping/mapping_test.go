package mapping

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/pro"
	"github.com/DrEVILish/OpenMixerControl/prosurface/internal/x32full"
)

func TestDefaultsValidate(t *testing.T) {
	tbl := x32full.NewTable()
	for _, m := range []pro.Model{pro.PRO1, pro.PRO2, pro.PRO2C} {
		if err := Default(m).Validate(pro.New(m), tbl); err != nil {
			t.Errorf("%s: %v", m, err)
		}
	}
}

func TestEveryStripControlIsMapped(t *testing.T) {
	for _, m := range []pro.Model{pro.PRO1, pro.PRO2} {
		s := pro.New(m)
		mapped := map[string]bool{}
		for _, r := range Default(m).Rules {
			mapped[r.Control] = true
		}
		for _, sec := range []string{pro.SecInputs, pro.SecVCA} {
			for _, c := range s.Section(sec) {
				if !mapped[c.ID] {
					t.Errorf("%s: %s is not mapped", m, c.ID)
				}
			}
		}
	}
}

func TestJSONRoundTrip(t *testing.T) {
	mp := Default(pro.PRO1)
	b, err := mp.JSON()
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "pro1.json")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.Pages != 2 || len(got.Rules) != len(mp.Rules) || got.PageButton != "EXTEND" {
		t.Errorf("loaded %d rules, %d pages", len(got.Rules), got.Pages)
	}
}

func TestValidateRejectsKindMismatch(t *testing.T) {
	mp := Default(pro.PRO2)
	mp.Rules = append(mp.Rules, Rule{Control: "IN1.FADER", Element: "HOME"})
	if err := mp.Validate(pro.New(pro.PRO2), x32full.NewTable()); err == nil {
		t.Error("fader bound to a button was accepted")
	}
}
