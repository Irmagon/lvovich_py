package server

import (
	"fmt"
	"strings"
	"unicode/utf16"
)

// JType — тип JSON-значения.
type JType byte

const (
	JObj   JType = 'o'
	JArr   JType = 'a'
	JStr   JType = 's'
	JNum   JType = 'n'
	JBool  JType = 'b'
	JNull  JType = 'z'
	JUndef JType = 'u'
)

// JValue — JSON-значение, сохраняющее порядок ключей (как JS-объект).
type JValue struct {
	Type JType
	// объект
	keys []string
	vals map[string]*JValue
	// массив
	items []*JValue
	// строка / число / bool (сырое слово)
	str string
	raw string
}

// Obj создаёт объект.
func Obj() *JValue { return &JValue{Type: JObj, vals: map[string]*JValue{}} }

// Arr создаёт массив.
func Arr() *JValue { return &JValue{Type: JArr} }

// Str создаёт строковое значение.
func Str(s string) *JValue { return &JValue{Type: JStr, str: s} }

// Num создаёт числовое значение (raw — исходное число).
func Num(raw string) *JValue { return &JValue{Type: JNum, raw: raw} }

// BoolVal создаёт булево значение.
func BoolVal(raw string) *JValue { return &JValue{Type: JBool, raw: raw} }

// Null создаёт null.
func Null() *JValue { return &JValue{Type: JNull} }

// Undef создаёт undefined (при сериализации в объекте пропускается).
func Undef() *JValue { return &JValue{Type: JUndef} }

// Set помещает ключ в объект.
func (v *JValue) Set(k string, val *JValue) {
	if v.keys == nil {
		v.keys = nil
	}
	v.keys = append(v.keys, k)
	v.vals[k] = val
}

// Get возвращает значение ключа объекта или nil.
func (v *JValue) Get(k string) *JValue {
	if v.Type != JObj {
		return nil
	}
	return v.vals[k]
}

// Keys возвращает ключи объекта в порядке вставки.
func (v *JValue) Keys() []string {
	if v.Type != JObj {
		return nil
	}
	return v.keys
}

// StrVal возвращает строковое значение (если это строка).
func (v *JValue) StrVal() (string, bool) {
	if v == nil || v.Type != JStr {
		return "", false
	}
	return v.str, true
}

// AsString возвращает строку или "" для не-строк.
func (v *JValue) AsString() string {
	if v == nil {
		return ""
	}
	if v.Type == JStr {
		return v.str
	}
	if v.Type == JNum || v.Type == JBool {
		return v.raw
	}
	return ""
}

// IsString сообщает, что значение — строка.
func (v *JValue) IsString() bool { return v != nil && v.Type == JStr }

// jParser — парсер JSON с сохранением порядка ключей.
type jParser struct {
	s   string
	pos int
}

func (p *jParser) skipWS() {
	for p.pos < len(p.s) {
		switch p.s[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *jParser) parseValue() (*JValue, error) {
	p.skipWS()
	if p.pos >= len(p.s) {
		return nil, fmt.Errorf("unexpected end")
	}
	switch p.s[p.pos] {
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case '"':
		sv, err := p.parseString()
		if err != nil {
			return nil, err
		}
		return Str(sv), nil
	case 't':
		if strings.HasPrefix(p.s[p.pos:], "true") {
			p.pos += 4
			return BoolVal("true"), nil
		}
	case 'f':
		if strings.HasPrefix(p.s[p.pos:], "false") {
			p.pos += 5
			return BoolVal("false"), nil
		}
	case 'n':
		if strings.HasPrefix(p.s[p.pos:], "null") {
			p.pos += 4
			return Null(), nil
		}
	default:
		// число
		start := p.pos
		for p.pos < len(p.s) {
			c := p.s[p.pos]
			if (c >= '0' && c <= '9') || c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E' {
				p.pos++
			} else {
				break
			}
		}
		if p.pos > start {
			return Num(p.s[start:p.pos]), nil
		}
	}
	return nil, fmt.Errorf("unexpected char at %d", p.pos)
}

func (p *jParser) parseObject() (*JValue, error) {
	obj := Obj()
	p.pos++ // '{'
	for {
		p.skipWS()
		if p.pos >= len(p.s) {
			return nil, fmt.Errorf("unexpected end of object")
		}
		if p.s[p.pos] == '}' {
			p.pos++
			return obj, nil
		}
		k, err := p.parseString()
		if err != nil {
			return nil, err
		}
		p.skipWS()
		if p.pos >= len(p.s) || p.s[p.pos] != ':' {
			return nil, fmt.Errorf("expected ':'")
		}
		p.pos++
		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		obj.Set(k, v)
		p.skipWS()
		if p.pos >= len(p.s) {
			return nil, fmt.Errorf("unexpected end of object")
		}
		if p.s[p.pos] == ',' {
			p.pos++
			continue
		}
		if p.s[p.pos] == '}' {
			p.pos++
			return obj, nil
		}
		return nil, fmt.Errorf("expected ',' or '}'")
	}
}

func (p *jParser) parseArray() (*JValue, error) {
	arr := Arr()
	p.pos++ // '['
	for {
		p.skipWS()
		if p.pos >= len(p.s) {
			return nil, fmt.Errorf("unexpected end of array")
		}
		if p.s[p.pos] == ']' {
			p.pos++
			return arr, nil
		}
		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		arr.items = append(arr.items, v)
		p.skipWS()
		if p.pos >= len(p.s) {
			return nil, fmt.Errorf("unexpected end of array")
		}
		if p.s[p.pos] == ',' {
			p.pos++
			continue
		}
		if p.s[p.pos] == ']' {
			p.pos++
			return arr, nil
		}
		return nil, fmt.Errorf("expected ',' or ']'")
	}
}

func (p *jParser) parseString() (string, error) {
	if p.pos >= len(p.s) || p.s[p.pos] != '"' {
		return "", fmt.Errorf("expected string")
	}
	p.pos++
	var b strings.Builder
	for p.pos < len(p.s) {
		c := p.s[p.pos]
		if c == '"' {
			p.pos++
			return b.String(), nil
		}
		if c == '\\' {
			p.pos++
			if p.pos >= len(p.s) {
				return "", fmt.Errorf("bad escape")
			}
			e := p.s[p.pos]
			p.pos++
			switch e {
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			case '/':
				b.WriteByte('/')
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case 'u':
				if p.pos+4 > len(p.s) {
					return "", fmt.Errorf("bad unicode escape")
				}
				var r rune
				for _, hc := range p.s[p.pos : p.pos+4] {
					r *= 16
					switch {
					case hc >= '0' && hc <= '9':
						r += rune(hc - '0')
					case hc >= 'a' && hc <= 'f':
						r += rune(hc-'a') + 10
					case hc >= 'A' && hc <= 'F':
						r += rune(hc-'A') + 10
					default:
						return "", fmt.Errorf("bad unicode escape")
					}
				}
				p.pos += 4
				if utf16.IsSurrogate(r) && p.pos+6 <= len(p.s) && p.s[p.pos] == '\\' && p.s[p.pos+1] == 'u' {
					// surrogate pair
					var r2 rune
					for _, hc := range p.s[p.pos+2 : p.pos+6] {
						r2 *= 16
						switch {
						case hc >= '0' && hc <= '9':
							r2 += rune(hc - '0')
						case hc >= 'a' && hc <= 'f':
							r2 += rune(hc-'a') + 10
						case hc >= 'A' && hc <= 'F':
							r2 += rune(hc-'A') + 10
						}
					}
					if utf16.IsSurrogate(r2) {
						dec := utf16.DecodeRune(r, r2)
						b.WriteRune(dec)
						p.pos += 6
						continue
					}
				}
				b.WriteRune(r)
			default:
				return "", fmt.Errorf("bad escape %q", e)
			}
			continue
		}
		b.WriteByte(c)
		p.pos++
	}
	return "", fmt.Errorf("unterminated string")
}

// ParseJSON разбирает JSON, сохраняя порядок ключей.
func ParseJSON(s string) (*JValue, error) {
	p := &jParser{s: s}
	v, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	return v, nil
}

// jsEscapeStr экранирует строку как JSON.stringify.
func jsEscapeStr(s string, b *strings.Builder) {
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				b.WriteString(fmt.Sprintf(`\u%04x`, r))
			} else {
				b.WriteRune(r)
			}
		}
	}
}

// JStringify сериализует значение как JSON.stringify (порядок ключей сохраняется).
func JStringify(v *JValue) string {
	var b strings.Builder
	jWrite(&b, v)
	return b.String()
}

func jWrite(b *strings.Builder, v *JValue) {
	if v == nil || v.Type == JNull {
		b.WriteString("null")
		return
	}
	if v.Type == JUndef {
		b.WriteString("null")
		return
	}
	switch v.Type {
	case JStr:
		b.WriteByte('"')
		jsEscapeStr(v.str, b)
		b.WriteByte('"')
	case JNum, JBool:
		b.WriteString(v.raw)
	case JArr:
		b.WriteByte('[')
		for i, it := range v.items {
			if i > 0 {
				b.WriteByte(',')
			}
			jWrite(b, it)
		}
		b.WriteByte(']')
	case JObj:
		b.WriteByte('{')
		first := true
		for _, k := range v.keys {
			val := v.vals[k]
			if val != nil && val.Type == JUndef {
				continue
			}
			if !first {
				b.WriteByte(',')
			}
			first = false
			b.WriteByte('"')
			jsEscapeStr(k, b)
			b.WriteString(`":`)
			jWrite(b, val)
		}
		b.WriteByte('}')
	}
}
