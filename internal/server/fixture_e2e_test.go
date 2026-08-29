package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Сквозной прогон реальных данных (fixture lvovich) через живой HTTP-сервер:
// REST /api/incline, /api/gender, /api/city/* и SOAP /soap.

type e2eFixture struct {
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

func loadE2EFixture(t *testing.T) *e2eFixture {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "lvovich", "testdata", "fixture.json"))
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	var fx e2eFixture
	if err := json.Unmarshal(data, &fx); err != nil {
		t.Fatalf("fixture unmarshal: %v", err)
	}
	return &fx
}

// ptrOr — значение указателя (или fallback, или "").
func ptrOr(primary, fallback *string) string {
	if primary != nil {
		return *primary
	}
	if fallback != nil {
		return *fallback
	}
	return ""
}

func TestE2EFixtureRestFio(t *testing.T) {
	base := startServer(t, testCfg())
	fx := loadE2EFixture(t)
	for i, f := range fx.Fio {
		ctx := "fio[" + string(rune('0'+i%10)) + "]"
		input := `{"SurName":` + quoteGo(f.Input.SurName) +
			`,"FirstName":` + quoteGo(f.Input.FirstName) +
			`,"SecondName":` + quoteGo(f.Input.SecondName) + `}`
		first, sur, second := f.Input.FirstName, f.Input.SurName, f.Input.SecondName

		for decl, c := range f.Cases {
			if decl == "initials" {
				continue
			}
			payload := `{"SurName":` + quoteGo(sur) +
				`,"FirstName":` + quoteGo(first) +
				`,"SecondName":` + quoteGo(second) +
				`,"declension":` + quoteGo(decl) +
				`,"format":"full"}`
			r := doReq(t, "POST", base+"/api/incline", map[string]string{
				"Authorization": authHdr, "Content-Type": "application/json",
			}, payload)
			want := Obj()
			if sur != "" {
				want.Set("SurName", Str(ptrOr(c.SurName, nil)))
			}
			if first != "" {
				want.Set("FirstName", Str(ptrOr(c.FirstName, nil)))
			}
			if second != "" {
				want.Set("SecondName", Str(ptrOr(c.SecondName, nil)))
			}
			if c.Gender == nil || *c.Gender == "" {
				want.Set("gender", Null())
			} else {
				want.Set("gender", Str(*c.Gender))
			}
			w := JStringify(want)
			t.Logf("%s/%s: %s -> %s", ctx, decl, payload, r.body)
			if r.status != 200 || r.body != w {
				t.Errorf("%s/%s: status=%d\nbody=%s\nwant=%s", ctx, decl, r.status, r.body, w)
			}
		}

		// format=initials
		ci, ok := f.Cases["initials"]
		if !ok {
			t.Errorf("%s: нет кейса initials", ctx)
			continue
		}
		payload := `{"SurName":` + quoteGo(sur) +
			`,"FirstName":` + quoteGo(first) +
			`,"SecondName":` + quoteGo(second) +
			`,"declension":"genitive","format":"initials"}`
		r := doReq(t, "POST", base+"/api/incline", map[string]string{
			"Authorization": authHdr, "Content-Type": "application/json",
		}, payload)
		want := Obj()
		want.Set("SurName", Str(ptrOr(ci.SurName, nil)))
		initialsVal := ""
		if ci.Initials != nil {
			initialsVal = *ci.Initials
		}
		want.Set("initials", Str(initialsVal))
		w := JStringify(want)
		t.Logf("%s/initials: %s -> %s", ctx, payload, r.body)
		if r.status != 200 || r.body != w {
			t.Errorf("%s/initials: status=%d\nbody=%s\nwant=%s", ctx, r.status, r.body, w)
		}

		// /api/gender
		gr := doReq(t, "POST", base+"/api/gender", map[string]string{
			"Authorization": authHdr, "Content-Type": "application/json",
		}, input)
		gw := Obj()
		if f.Gender == nil || *f.Gender == "" {
			gw.Set("gender", Null())
		} else {
			gw.Set("gender", Str(*f.Gender))
		}
		g := JStringify(gw)
		t.Logf("%s/gender: %s -> %s", ctx, input, gr.body)
		if gr.status != 200 || gr.body != g {
			t.Errorf("%s/gender: status=%d body=%s want=%s", ctx, gr.status, gr.body, g)
		}
	}
}

func TestE2EFixtureRestCities(t *testing.T) {
	base := startServer(t, testCfg())
	fx := loadE2EFixture(t)
	cases := []struct {
		url, field, gender string
	}{
		{"/api/city/in", "In", ""},
		{"/api/city/in", "InF", "female"},
		{"/api/city/from", "From", ""},
		{"/api/city/from", "FromF", "female"},
		{"/api/city/to", "To", ""},
	}
	for _, c := range fx.Cities {
		ctx := "city[" + c.Input + "]"
		for _, tc := range cases {
			var payload string
			if tc.field == "To" {
				payload = `{"name":` + quoteGo(c.Input) + `}`
			} else {
				payload = `{"name":` + quoteGo(c.Input) + `,"gender":` + quoteGo(tc.gender) + `}`
			}
			r := doReq(t, "POST", base+tc.url, map[string]string{
				"Authorization": authHdr, "Content-Type": "application/json",
			}, payload)
			var want string
			switch tc.field {
			case "In":
				want = c.In
			case "InF":
				want = c.InF
			case "From":
				want = c.From
			case "FromF":
				want = c.FromF
			default:
				want = c.To
			}
			body := `{"name":` + quoteJSON(want) + `}`
			t.Logf("%s/%s: %s -> %s", ctx, tc.field, payload, r.body)
			if r.status != 200 || r.body != body {
				t.Errorf("%s/%s: status=%d body=%s want=%s", ctx, tc.field, r.status, r.body, body)
			}
		}
	}
}

func TestE2EFixtureSoap(t *testing.T) {
	base := startServer(t, testCfg())
	fx := loadE2EFixture(t)
	for i, f := range fx.Fio {
		if i > 5 {
			break
		}
		ctx := "soap[" + string(rune('0'+i%10)) + "]"
		decl := "genitive"
		c, ok := f.Cases[decl]
		if !ok {
			t.Errorf("%s: нет генитива", ctx)
			continue
		}
		inner := `<LastName>` + f.Input.SurName + `</LastName>` +
			`<FirstName>` + f.Input.FirstName + `</FirstName>` +
			`<MiddleName>` + f.Input.SecondName + `</MiddleName>` +
			`<Declension>` + decl + `</Declension><Format>full</Format>`
		req := `<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="urn:LvovichService"><soap:Body><tns:Incline>` + inner + `</tns:Incline></soap:Body></soap:Envelope>`
		r := doReq(t, "POST", base+"/soap", map[string]string{
			"Authorization": authHdr, "Content-Type": "text/xml", "SOAPAction": "urn:Incline",
		}, req)
		t.Logf("%s/%s: Incline(%q,%q,%q) -> %s", ctx, decl, f.Input.SurName, f.Input.FirstName, f.Input.SecondName, r.body)
		if r.status != 200 || !strings.Contains(r.body, "FirstName>"+ptrOr(c.FirstName, nil)+"<") || !strings.Contains(r.body, "LastName>"+ptrOr(c.SurName, nil)+"<") {
			t.Errorf("%s: status=%d body=%s", ctx, r.status, r.body)
		}
	}
}

// quoteGo — JSON-строка с корректным экранированием.
func quoteGo(s string) string {
	return quoteJSON(s)
}
