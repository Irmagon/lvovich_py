package server

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// soapNode — элемент XML с сохранением порядка атрибутов/детей.
type soapNode struct {
	local    string
	attrKeys []string
	attrs    map[string]string
	children []*soapNode
	text     string
}

// logSoap пишет строку лога SOAP-операции. При отключённом логировании
// (Enabled()==false) args не сериализуется и модуль логирования не вызывается.
func (s *Server) logSoap(op string, args *soapNode) {
	if !s.log.Enabled() {
		return
	}
	s.log.Log("", op+" - "+soapArgsJSON(args))
}

// parseSoapDoc разбирает SOAP-конверт: операция = первый элемент внутри soap:Body.
func parseSoapDoc(body []byte) (op string, opEl *soapNode, err error) {
	if len(strings.TrimSpace(string(body))) == 0 {
		return "", nil, fmt.Errorf("empty body")
	}
	dec := xml.NewDecoder(strings.NewReader(string(body)))
	var root *soapNode
	var stack []*soapNode
	for {
		tok, terr := dec.Token()
		if terr == io.EOF {
			break
		}
		if terr != nil {
			return "", nil, terr
		}
		switch t := tok.(type) {
		case xml.StartElement:
			n := &soapNode{local: localName(t.Name), attrs: map[string]string{}}
			for _, a := range t.Attr {
				if a.Name.Space == "xmlns" || a.Name.Local == "xmlns" {
					continue
				}
				n.attrKeys = append(n.attrKeys, a.Name.Local)
				n.attrs[a.Name.Local] = a.Value
			}
			if len(stack) > 0 {
				stack[len(stack)-1].children = append(stack[len(stack)-1].children, n)
			} else if root == nil {
				root = n
			} else {
				root.children = append(root.children, n)
			}
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].text += string(t)
			}
		}
	}
	if root == nil {
		return "", nil, fmt.Errorf("no root element")
	}
	bodyEl := findBody(root)
	if bodyEl == nil || len(bodyEl.children) == 0 {
		return "", nil, fmt.Errorf("no operation element")
	}
	opEl = bodyEl.children[0]
	return opEl.local, opEl, nil
}

func localName(name xml.Name) string {
	if name.Space != "" {
		return name.Local
	}
	return name.Local
}

func findBody(n *soapNode) *soapNode {
	if n.local == "Body" {
		return n
	}
	for _, c := range n.children {
		if b := findBody(c); b != nil {
			return b
		}
	}
	return nil
}

// soapOps — известные операции.
var soapOps = map[string]bool{
	"Incline": true, "GetGender": true,
	"CityIn": true, "CityFrom": true, "CityTo": true,
	"OrgIn": true, "OrgFrom": true, "OrgTo": true,
}

// soapActions — SOAPAction -> операция (из WSDL-binding).
var soapActions = map[string]string{
	"urn:Incline": "Incline", "urn:GetGender": "GetGender",
	"urn:CityIn": "CityIn", "urn:CityFrom": "CityFrom", "urn:CityTo": "CityTo",
	"urn:OrgIn": "OrgIn", "urn:OrgFrom": "OrgFrom", "urn:OrgTo": "OrgTo",
}

// handleSoap — POST /soap.
func (s *Server) handleSoap(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(io.LimitReader(r.Body, maxBody))
	if err != nil {
		s.soapStack500(w, "Error reading request body")
		return
	}
	op, opEl, err := parseSoapDoc(data)
	if err != nil {
		s.soapStack500(w, err.Error())
		return
	}

	action := strings.Trim(r.Header.Get("SOAPAction"), `" `)
	errKind := ""
	switch {
	case action != "":
		exp, known := soapActions[action]
		switch {
		case !known:
			errKind = "style"
		case exp != op:
			errKind = "description"
		}
	case !soapOps[op]:
		errKind = "description"
	}

	switch errKind {
	case "style":
		s.soapFault(w, "TypeError: Cannot read properties of undefined (reading 'style')")
		return
	case "description":
		s.soapStack500(w, "TypeError: Cannot read properties of undefined (reading 'description')")
		return
	}

	s.soapCall(w, op, opEl)
}

// soapCall выполняет операцию и пишет ответ.
func (s *Server) soapCall(w http.ResponseWriter, op string, opEl *soapNode) {
	args := opEl
	if len(opEl.children) == 0 {
		args = nil // как в node-soap: пустой <tns:Op/> -> args null
	}
	switch op {
	case "Incline":
		s.soapIncline(w, opEl, args)
	case "GetGender":
		s.soapGetGender(w, opEl, args)
	case "CityIn", "CityFrom", "CityTo":
		s.soapCity(w, opEl, args, op)
	default: // OrgIn, OrgFrom, OrgTo
		s.soapOrg(w, opEl, args, op)
	}
}

func (s *Server) soapIncline(w http.ResponseWriter, opEl *soapNode, args *soapNode) {
	s.logSoap("SOAP Incline", args)

	var lastName, firstName, middleName, decl, format string
	if args != nil {
		if err := soapReadString(args, "LastName", &lastName); err != nil {
			s.soapFieldFault(w, "InclineResponse", err)
			return
		}
		firstName = soapGetString(args, "FirstName")
		middleName = soapGetString(args, "MiddleName")
		decl = soapGetString(args, "Declension")
		format = soapGetString(args, "Format")
	} else {
		s.soapFieldFault(w, "InclineResponse", fmt.Errorf("Cannot read properties of null (reading 'LastName')"))
		return
	}
	if format == "" {
		format = "full"
	}

	res := s.core.Incline(s.person(lastName, firstName, middleName), decl, format)

	var fields []soapField
	if format == "initials" {
		fields = appendIf(fields, "LastName", res.SurName)
		fields = appendIf(fields, "Initials", res.Initials)
	} else {
		fields = appendIf(fields, "LastName", res.SurName)
		fields = appendIf(fields, "FirstName", res.FirstName)
		fields = appendIf(fields, "MiddleName", res.SecondName)
		fields = appendIf(fields, "Gender", res.Gender)
	}
	s.soapResponse(w, "InclineResponse", fields)
}

func (s *Server) soapGetGender(w http.ResponseWriter, opEl *soapNode, args *soapNode) {
	s.logSoap("SOAP GetGender", args)

	var lastName, firstName, middleName string
	if args == nil {
		s.soapFieldFault(w, "GetGenderResponse", fmt.Errorf("Cannot read properties of null (reading 'LastName')"))
		return
	}
	lastName = soapGetString(args, "LastName")
	firstName = soapGetString(args, "FirstName")
	middleName = soapGetString(args, "MiddleName")

	g := s.core.GetGender(s.person(lastName, firstName, middleName))
	s.soapResponse(w, "GetGenderResponse", appendIf(nil, "Gender", g))
}

func (s *Server) soapCity(w http.ResponseWriter, opEl *soapNode, args *soapNode, op string) {
	s.logSoap("SOAP "+op, args)

	respEl := op + "Response"
	name, present := "", false
	if args != nil {
		name, present = soapGetStringChecked(args, "Name")
	} else {
		s.soapFieldFault(w, respEl, fmt.Errorf("Cannot read properties of null (reading 'Name')"))
		return
	}
	if !present {
		s.soapFieldFault(w, respEl, fmt.Errorf("Cannot read properties of undefined (reading 'toLowerCase')"))
		return
	}
	gender := soapGetString(args, "Gender")
	var out string
	switch op {
	case "CityIn":
		out = s.core.CityIn(name, gender)
	case "CityFrom":
		out = s.core.CityFrom(name, gender)
	default:
		out = s.core.CityTo(name)
	}
	s.soapResponse(w, respEl, appendIf(nil, "Name", out))
}

func (s *Server) soapOrg(w http.ResponseWriter, opEl *soapNode, args *soapNode, op string) {
	s.logSoap("SOAP "+op, args)

	respEl := op + "Response"
	name, present := "", false
	if args != nil {
		name, present = soapGetStringChecked(args, "Name")
	} else {
		s.soapFieldFault(w, respEl, fmt.Errorf("Cannot read properties of null (reading 'Name')"))
		return
	}
	if !present {
		s.soapFieldFault(w, respEl, fmt.Errorf("Cannot read properties of undefined (reading 'toLowerCase')"))
		return
	}
	var out string
	switch op {
	case "OrgIn":
		out = s.core.OrgIn(name)
	case "OrgFrom":
		out = s.core.OrgFrom(name)
	default:
		out = s.core.OrgTo(name)
	}
	s.soapResponse(w, respEl, appendIf(nil, "Name", out))
}

// soapReadString — как args.LastName в JS: null-args даёт TypeError.
func soapReadString(n *soapNode, name string, dst *string) error {
	if n == nil {
		return fmt.Errorf("Cannot read properties of null (reading '%s')", name)
	}
	*dst = soapGetString(n, name)
	return nil
}

// soapGetString — args.Name (строка или "" если нет/null/не текст).
func soapGetString(n *soapNode, name string) string {
	if n == nil {
		return ""
	}
	for _, c := range n.children {
		if c.local == name {
			return strings.TrimSpace(c.text)
		}
	}
	return ""
}

func soapGetStringChecked(n *soapNode, name string) (string, bool) {
	if n == nil {
		return "", false
	}
	for _, c := range n.children {
		if c.local == name {
			return strings.TrimSpace(c.text), true
		}
	}
	return "", false
}

// soapFieldFault — ошибка обработчика: <Fault>внутри</Fault> (HTTP 200).
func (s *Server) soapFieldFault(w http.ResponseWriter, respEl string, err error) {
	s.soapResponse(w, respEl, []soapField{{name: "Fault", value: err.Error()}})
}

// soapResponse — стандартный конверт-ответ (как node-soap).
func (s *Server) soapResponse(w http.ResponseWriter, respEl string, fields []soapField) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="utf-8"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"  xmlns:tns="urn:LvovichService"><soap:Body><tns:`)
	b.WriteString(respEl)
	b.WriteString(`><`)
	b.WriteString(respEl)
	b.WriteString(`>`)
	for _, f := range fields {
		b.WriteString("<")
		b.WriteString(f.name)
		b.WriteString(">")
		b.WriteString(xmlEscape(f.value))
		b.WriteString("</")
		b.WriteString(f.name)
		b.WriteString(">")
	}
	b.WriteString(`</`)
	b.WriteString(respEl)
	b.WriteString(`></tns:`)
	b.WriteString(respEl)
	b.WriteString(`></soap:Body></soap:Envelope>`)
	serveSoapText(w, http.StatusOK, b.String())
}

// soapFault — SOAP-Fault (HTTP 500), как выбрасывает node-soap.
func (s *Server) soapFault(w http.ResponseWriter, reason string) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="utf-8"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"  xmlns:tns="urn:LvovichService"><soap:Body><soap:Fault><soap:Code><soap:Value>SOAP-ENV:Server</soap:Value><soap:Subcode><soap:Value>InternalServerError</soap:Value></soap:Subcode></soap:Code><soap:Reason><soap:Text>`)
	b.WriteString(xmlEscape(reason))
	b.WriteString(`</soap:Text></soap:Reason></soap:Fault></soap:Body></soap:Envelope>`)
	serveSoapText(w, http.StatusInternalServerError, b.String())
}

// soapStack500 — HTTP 500 c текстом (как стек node).
func (s *Server) soapStack500(w http.ResponseWriter, errText string) {
	body := errText + "\n    at Server._process (soap:server)\n    at Server._processRequestXml (soap:server)"
	serveSoapText(w, http.StatusInternalServerError, body)
}

func serveSoapText(w http.ResponseWriter, status int, body string) {
	h := w.Header()
	h.Set("Content-Type", "text/xml")
	h.Set("X-Powered-By", "Express")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// soapField — поле ответа.
type soapField struct {
	name  string
	value string
}

func appendIf(fields []soapField, name, value string) []soapField {
	if value == "" {
		return fields
	}
	return append(fields, soapField{name: name, value: value})
}

func xmlEscape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&apos;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// soapArgsJSON — JSON.stringify(args) в представлении node-soap:
// ключи ФИО (LastName, FirstName, MiddleName) идут первыми, затем остальные
// в порядке документа.
func soapArgsJSON(args *soapNode) string {
	if args == nil {
		return "null"
	}
	var rest []*soapNode
	var b strings.Builder
	b.WriteByte('{')
	first := true
	write := func(n *soapNode) {
		if !first {
			b.WriteByte(',')
		}
		first = false
		b.WriteByte('"')
		jsEscapeStr(n.local, &b)
		b.WriteString(`":`)
		soapWriteElem(&b, n)
	}
	for _, k := range []string{"LastName", "FirstName", "MiddleName"} {
		for _, c := range args.children {
			if c.local == k {
				write(c)
				break
			}
		}
	}
	for _, c := range args.children {
		if c.local == "LastName" || c.local == "FirstName" || c.local == "MiddleName" {
			continue
		}
		rest = append(rest, c)
	}
	for _, c := range rest {
		write(c)
	}
	b.WriteByte('}')
	return b.String()
}

func soapWriteElem(b *strings.Builder, n *soapNode) {
	txt := strings.TrimSpace(n.text)
	if len(n.children) == 0 && len(n.attrKeys) == 0 {
		if txt == "" {
			b.WriteString("null")
		} else {
			soapWriteJSONStr(b, txt)
		}
		return
	}
	b.WriteByte('{')
	first := true
	if len(n.attrKeys) > 0 {
		first = false
		b.WriteString(`"attributes":{`)
		for i, k := range n.attrKeys {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteByte('"')
			jsEscapeStr(k, b)
			b.WriteString(`":`)
			soapWriteJSONStr(b, n.attrs[k])
		}
		b.WriteByte('}')
	}
	if txt != "" || len(n.children) == 0 {
		if !first {
			b.WriteByte(',')
		}
		first = false
		b.WriteString(`"$value":`)
		soapWriteJSONStr(b, txt)
	}
	for _, c := range n.children {
		if !first {
			b.WriteByte(',')
		}
		first = false
		b.WriteByte('"')
		jsEscapeStr(c.local, b)
		b.WriteString(`":`)
		soapWriteElem(b, c)
	}
	b.WriteByte('}')
}

func soapWriteJSONStr(b *strings.Builder, s string) {
	b.WriteByte('"')
	jsEscapeStr(s, b)
	b.WriteByte('"')
}
