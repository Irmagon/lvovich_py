package lvovich

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type fixtureT struct {
	Fio []struct {
		Input struct {
			SurName    string `json:"SurName"`
			FirstName  string `json:"FirstName"`
			SecondName string `json:"SecondName"`
		} `json:"input"`
		Gender *string `json:"gender"`
		Cases  map[string]struct {
			Gender     *string `json:"gender"`
			FirstName  *string `json:"FirstName"`
			SurName    *string `json:"SurName"`
			SecondName *string `json:"SecondName"`
			Initials   *string `json:"initials"`
		} `json:"cases"`
	} `json:"fio"`
	Cities []struct {
		Input string `json:"input"`
		In    string `json:"in"`
		InF   string `json:"inF"`
		From  string `json:"from"`
		FromF string `json:"fromF"`
		To    string `json:"to"`
	} `json:"cities"`
}

func loadFixture(t *testing.T) *fixtureT {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "fixture.json"))
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	var fx fixtureT
	if err := json.Unmarshal(data, &fx); err != nil {
		t.Fatalf("fixture unmarshal: %v", err)
	}
	return &fx
}

func cmpStr(t *testing.T, ctx, field, got string, exp *string) {
	t.Helper()
	if exp == nil {
		if got != "" {
			t.Errorf("%s: %s: ожидалось отсутствие, получено %q", ctx, field, got)
		}
		return
	}
	if got != *exp {
		t.Errorf("%s: %s: получено %q, ожидалось %q", ctx, field, got, *exp)
	}
}

func TestFixtureFio(t *testing.T) {
	fx := loadFixture(t)
	for i, f := range fx.Fio {
		p := Person{SurName: f.Input.SurName, FirstName: f.Input.FirstName, SecondName: f.Input.SecondName}
		ctx := inputDesc(t, i, f.Input.SurName, f.Input.FirstName, f.Input.SecondName)

		gender := GetGender(p)
		t.Logf("%s gender: (%q,%q,%q) -> %q", ctx, f.Input.SurName, f.Input.FirstName, f.Input.SecondName, gender)
		cmpStr(t, ctx, "gender", gender, f.Gender)

		for decl, exp := range f.Cases {
			if decl == "initials" {
				got := Incline(p, "genitive", "initials")
				t.Logf("%s/initials: (%q,%q,%q) -> %q %q", ctx, f.Input.SurName, f.Input.FirstName, f.Input.SecondName, got.SurName, got.Initials)
				cmpStr(t, ctx+"/initials", "SurName", got.SurName, exp.SurName)
				cmpStr(t, ctx+"/initials", "initials", got.Initials, exp.Initials)
				continue
			}
			got := Incline(p, decl, "")
			t.Logf("%s/%s: (%q,%q,%q) -> %q %q %q gender=%q", ctx, decl, f.Input.SurName, f.Input.FirstName, f.Input.SecondName, got.FirstName, got.SurName, got.SecondName, got.Gender)
			cmpStr(t, ctx+"/"+decl, "FirstName", got.FirstName, exp.FirstName)
			cmpStr(t, ctx+"/"+decl, "SurName", got.SurName, exp.SurName)
			cmpStr(t, ctx+"/"+decl, "SecondName", got.SecondName, exp.SecondName)
			cmpStr(t, ctx+"/"+decl, "gender", got.Gender, exp.Gender)
		}
	}
}

func TestFixtureCities(t *testing.T) {
	fx := loadFixture(t)
	for i, c := range fx.Cities {
		ctx := "city[" + c.Input + "]"
		_ = i
		t.Logf("%s in: %q -> %q", ctx, c.Input, CityIn(c.Input, ""))
		cmpStr(t, ctx, "in", CityIn(c.Input, ""), strPtr(c.In))
		t.Logf("%s inF: %q -> %q", ctx, c.Input, CityIn(c.Input, "female"))
		cmpStr(t, ctx, "inF", CityIn(c.Input, "female"), strPtr(c.InF))
		t.Logf("%s from: %q -> %q", ctx, c.Input, CityFrom(c.Input, ""))
		cmpStr(t, ctx, "from", CityFrom(c.Input, ""), strPtr(c.From))
		t.Logf("%s fromF: %q -> %q", ctx, c.Input, CityFrom(c.Input, "female"))
		cmpStr(t, ctx, "fromF", CityFrom(c.Input, "female"), strPtr(c.FromF))
		t.Logf("%s to: %q -> %q", ctx, c.Input, CityTo(c.Input))
		cmpStr(t, ctx, "to", CityTo(c.Input), strPtr(c.To))
	}
}

func strPtr(s string) *string { return &s }

func inputDesc(t *testing.T, i int, sur, first, second string) string {
	t.Helper()
	return "fio[" + string(rune('0'+i%10)) + "] " + sur + " " + first + " " + second
}
