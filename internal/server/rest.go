package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

const maxBody = 100 * 1024 // как express.json() limit=100kb

type bodyKey struct{}

// bodyInfo — результат обработки тела запроса как express.json().
type bodyInfo struct {
	value *JValue // nil, если тело не JSON (req.body === undefined)
}

func withBodyInfo(r *http.Request, bi *bodyInfo) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), bodyKey{}, bi))
}

func getBodyInfo(r *http.Request) *bodyInfo {
	bi, _ := r.Context().Value(bodyKey{}).(*bodyInfo)
	return bi
}

// parseBody читает и разбирает тело запроса, как express.json():
//   - без Content-Type application/json -> req.body === undefined (value=nil)
//   - пустое тело -> {}
//   - битый JSON -> 400 со страницей express "SyntaxError".
//
// Возвращает (bi, handled): handled=true означает, что ответ уже записан.
func (s *Server) parseBody(w http.ResponseWriter, r *http.Request) (*bodyInfo, bool) {
	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if err != nil {
		return nil, false
	}
	if len(data) > maxBody {
		s.errPayloadTooLarge(w)
		return nil, true
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return &bodyInfo{value: Obj()}, false
	}
	v, perr := ParseJSON(string(data))
	if perr != nil {
		errPage(w, http.StatusBadRequest, "Error", syntaxErrorHTML(string(data)))
		return nil, true
	}
	return &bodyInfo{value: v}, false
}

// reqBody возвращает разобранное тело или nil (req.body === undefined).
func reqBody(r *http.Request) *JValue {
	bi := getBodyInfo(r)
	if bi == nil {
		return nil
	}
	return bi.value
}

// bodyLogText — то, что пишется в лог для запроса (JSON.stringify(body)).
func bodyLogText(r *http.Request) string {
	bi := getBodyInfo(r)
	if bi == nil || bi.value == nil {
		return "undefined"
	}
	return JStringify(reorderFIO(bi.value))
}

// reorderFIO — переставляет ключи ФИО (SurName, FirstName, SecondName) в начало
// объекта, сохраняя порядок остальных ключей.
func reorderFIO(v *JValue) *JValue {
	if v == nil || v.Type != JObj {
		return v
	}
	out := Obj()
	for _, k := range []string{"SurName", "FirstName", "SecondName"} {
		if val := v.Get(k); val != nil {
			out.Set(k, val)
		}
	}
	for _, k := range v.Keys() {
		if k == "SurName" || k == "FirstName" || k == "SecondName" {
			continue
		}
		if val := v.Get(k); val != nil {
			out.Set(k, val)
		}
	}
	return out
}

// syntaxErrorHTML — сообщение body-parser в express-форме.
// Позиция ошибки считается в UTF-16 юнитах (как в V8), 0-based.
func syntaxErrorHTML(data string) string {
	pos := jsonErrorPosition(data)
	col := pos + 1
	return fmt.Sprintf("SyntaxError: Expected double-quoted property name in JSON at position %d (line 1 column %d)", pos, col)
}

// jsonErrorPosition — позиция (в UTF-16 юнитах), где парсер ожидал
// имя свойства (после '{' или ','), но встретил другой токен.
// Если ошибок такого рода нет, возвращается длина строки в юнитах.
func jsonErrorPosition(data string) int {
	sc := &jsonScanner{s: data}
	sc.skipWS()
	if !sc.has() {
		return sc.units
	}
	if sc.peek() != '{' {
		return sc.units
	}
	sc.take()
	return sc.parseObject()
}

type jsonScanner struct {
	s     string
	i     int
	units int
}

func (sc *jsonScanner) has() bool { return sc.i < len(sc.s) }

func (sc *jsonScanner) peek() byte { return sc.s[sc.i] }

// take считывает один руна и учитывает UTF-16 длину.
func (sc *jsonScanner) take() byte {
	if sc.s[sc.i] < 0x80 {
		b := sc.s[sc.i]
		sc.i++
		sc.units++
		return b
	}
	r, sz := utf8.DecodeRuneInString(sc.s[sc.i:])
	sc.i += sz
	sc.units += utf16.RuneLen(r)
	return 0
}

func (sc *jsonScanner) skipWS() {
	for sc.has() {
		switch sc.s[sc.i] {
		case ' ', '\t', '\n', '\r':
			sc.take()
		default:
			return
		}
	}
}

func (sc *jsonScanner) parseString() bool {
	for sc.has() {
		c := sc.take()
		if c == '\\' {
			if !sc.has() {
				return false
			}
			sc.take()
			continue
		}
		if c == '"' {
			return true
		}
	}
	return false
}

func (sc *jsonScanner) parseIdentifier() {
	for sc.has() {
		c := sc.s[sc.i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.' {
			sc.take()
			continue
		}
		return
	}
}

func (sc *jsonScanner) skipValue() bool {
	sc.skipWS()
	if !sc.has() {
		return false
	}
	switch sc.peek() {
	case '"':
		sc.take()
		return sc.parseString()
	case '{':
		sc.take()
		err := sc.parseObject()
		return err < 0
	case '[':
		sc.take()
		for {
			sc.skipWS()
			if !sc.has() {
				return false
			}
			if sc.peek() == ']' {
				sc.take()
				return true
			}
			if !sc.skipValue() {
				return false
			}
			sc.skipWS()
			if sc.peek() == ',' {
				sc.take()
				continue
			}
			if sc.peek() == ']' {
				sc.take()
				return true
			}
			return false
		}
	case 't':
		sc.parseIdentifier()
		return true
	case 'f':
		sc.parseIdentifier()
		return true
	case 'n':
		sc.parseIdentifier()
		return true
	default:
		sc.parseIdentifier()
		return true
	}
}

// parseObject разбирает тело объекта (после '{'); возвращает
// позицию ошибки или -1 при успехе.
func (sc *jsonScanner) parseObject() int {
	sc.skipWS()
	if !sc.has() {
		return sc.units
	}
	if sc.peek() == '}' {
		sc.take()
		return -1
	}
	for {
		sc.skipWS()
		if !sc.has() {
			return sc.units
		}
		if sc.peek() != '"' {
			return sc.units
		}
		sc.take()
		if !sc.parseString() {
			return sc.units
		}
		sc.skipWS()
		if !sc.has() || sc.peek() != ':' {
			return sc.units
		}
		sc.take()
		if !sc.skipValue() {
			return sc.units
		}
		sc.skipWS()
		if !sc.has() {
			return sc.units
		}
		switch sc.peek() {
		case ',':
			sc.take()
			continue
		case '}':
			sc.take()
			return -1
		default:
			return sc.units
		}
	}
}

// exprStrField — достаёт строковое поле (как JS: field || ”).
// Если поле есть, но не строка -> (false, ok=false) — JS TypeError "X.trim is not a function".
func exprStrField(body *JValue, name string) (string, bool) {
	if body == nil {
		return "", true
	}
	v := body.Get(name)
	if v == nil {
		return "", true
	}
	if !v.IsString() {
		return "", false
	}
	s, _ := v.StrVal()
	return s, true
}

// person — собрать ФИО.
func (s *Server) person(sur, first, second string) *Person {
	return &Person{SurName: sur, FirstName: first, SecondName: second}
}

func (s *Server) handleRestIncline(w http.ResponseWriter, r *http.Request) {
	body := reqBody(r)
	if body == nil || body.Type != JObj {
		s.err500(w, r, "Cannot read properties of undefined (reading 'declension')")
		return
	}
	declension, _ := exprStrField(body, "declension")
	format, _ := exprStrField(body, "format")

	sur, ok := exprStrField(body, "SurName")
	if !ok {
		s.err500(w, r, "SurName.trim is not a function")
		return
	}
	first, ok := exprStrField(body, "FirstName")
	if !ok {
		s.err500(w, r, "FirstName.trim is not a function")
		return
	}
	second, ok := exprStrField(body, "SecondName")
	if !ok {
		s.err500(w, r, "SecondName.trim is not a function")
		return
	}

	res := s.core.Incline(s.person(sur, first, second), declension, format)
	obj := Obj()
	if format == "initials" {
		obj.Set("SurName", Str(res.SurName))
		obj.Set("initials", Str(res.Initials))
	} else {
		if sur != "" {
			obj.Set("SurName", Str(res.SurName))
		}
		if first != "" {
			obj.Set("FirstName", Str(res.FirstName))
		}
		if second != "" {
			obj.Set("SecondName", Str(res.SecondName))
		}
		if res.Gender == "" {
			obj.Set("gender", Null())
		} else {
			obj.Set("gender", Str(res.Gender))
		}
	}
	sendJSON(w, r, http.StatusOK, JStringify(obj))
}

func (s *Server) handleRestGender(w http.ResponseWriter, r *http.Request) {
	body := reqBody(r)
	if body != nil && body.Type != JObj {
		s.err500(w, r, "Cannot read properties of undefined (reading 'SurName')")
		return
	}
	var sur, first, second string
	if body != nil {
		var ok bool
		if sur, ok = exprStrField(body, "SurName"); !ok {
			s.err500(w, r, "SurName.trim is not a function")
			return
		}
		if first, ok = exprStrField(body, "FirstName"); !ok {
			s.err500(w, r, "FirstName.trim is not a function")
			return
		}
		if second, ok = exprStrField(body, "SecondName"); !ok {
			s.err500(w, r, "SecondName.trim is not a function")
			return
		}
	}
	g := s.core.GetGender(s.person(sur, first, second))
	obj := Obj()
	if g == "" {
		obj.Set("gender", Null())
	} else {
		obj.Set("gender", Str(g))
	}
	sendJSON(w, r, http.StatusOK, JStringify(obj))
}

func (s *Server) handleRestCity(w http.ResponseWriter, r *http.Request, kind string) {
	body := reqBody(r)
	var nameV *JValue
	if body != nil {
		nameV = body.Get("name")
	}
	if nameV == nil || !nameV.IsString() {
		if kind == "to" {
			// JS: cityTo(undefined) -> undefined -> {name: undefined} -> {}
			sendJSON(w, r, http.StatusOK, "{}")
			return
		}
		s.err500(w, r, "Cannot read properties of undefined (reading 'toLowerCase')")
		return
	}
	name, _ := nameV.StrVal()
	gender, _ := exprStrField(body, "gender")
	var out string
	switch kind {
	case "in":
		out = s.core.CityIn(name, gender)
	case "from":
		out = s.core.CityFrom(name, gender)
	default:
		out = s.core.CityTo(name)
	}
	obj := Obj()
	obj.Set("name", Str(out))
	sendJSON(w, r, http.StatusOK, JStringify(obj))
}

func (s *Server) handleRestOrg(w http.ResponseWriter, r *http.Request, kind string) {
	body := reqBody(r)
	var nameV *JValue
	if body != nil {
		nameV = body.Get("name")
	}
	if nameV == nil || !nameV.IsString() {
		sendJSON(w, r, http.StatusOK, "{}")
		return
	}
	name, _ := nameV.StrVal()
	var out string
	switch kind {
	case "in":
		out = s.core.OrgIn(name)
	case "from":
		out = s.core.OrgFrom(name)
	default:
		out = s.core.OrgTo(name)
	}
	obj := Obj()
	obj.Set("name", Str(out))
	sendJSON(w, r, http.StatusOK, JStringify(obj))
}

func (s *Server) err500(w http.ResponseWriter, r *http.Request, msg string) {
	sendJSON(w, r, http.StatusInternalServerError, `{"error":`+quoteJSON(msg)+`}`)
}

func quoteJSON(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// utf16Count — длина строки в UTF-16 юнитах (позиции в сообщениях JS).
func utf16Count(s string) int {
	n := 0
	for _, r := range s {
		if r < 0x10000 {
			n++
		} else {
			n += 2
		}
	}
	return n
}

// errPayloadTooLarge — аналог ошибки 413 в express/body-parser.
func (s *Server) errPayloadTooLarge(w http.ResponseWriter) {
	errPage(w, http.StatusRequestEntityTooLarge, "Error", "PayloadTooLargeError: request entity too large")
}
