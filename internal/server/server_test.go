package server

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// startServer запускает сервер на случайном порту.
func startServer(t *testing.T, cfg Config) string {
	t.Helper()
	h := NewServer(cfg, filepath.Join(t.TempDir(), "server.log"))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: h}
	go srv.Serve(ln)
	t.Cleanup(func() {
		_ = srv.Close()
		h.Close()
	})
	return "http://" + ln.Addr().String()
}

type resp struct {
	status  int
	ct      string
	etag    string
	body    string
	headers http.Header
}

func doReq(t *testing.T, method, url string, headers map[string]string, body string) resp {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rr, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Body.Close()
	b, _ := io.ReadAll(rr.Body)
	return resp{status: rr.StatusCode, ct: rr.Header.Get("Content-Type"), etag: rr.Header.Get("ETag"), body: string(b), headers: rr.Header}
}

const authHdr = "Bearer testtoken999"

func testCfg() Config {
	return Config{Address: "127.0.0.1", Port: 0, Token: "testtoken999", AllowedIPs: []string{"127.0.0.1", "::1"}, Swagger: true, Logging: true, LogMode: "async", FlushMs: 50, BufferKB: 64}
}

func TestRestInclineFull(t *testing.T) {
	base := startServer(t, testCfg())
	r := doReq(t, "POST", base+"/api/incline", map[string]string{
		"Authorization":  authHdr,
		"Content-Type":   "application/json",
	}, `{"SurName":"Иванов","FirstName":"Иван","SecondName":"Иванович","declension":"dative"}`)
	if r.status != 200 {
		t.Fatalf("status=%d", r.status)
	}
	want := `{"SurName":"Иванову","FirstName":"Ивану","SecondName":"Ивановичу","gender":"male"}`
	if r.body != want {
		t.Errorf("body=%q want=%q", r.body, want)
	}
	if r.ct != "application/json; charset=utf-8" {
		t.Errorf("ct=%q", r.ct)
	}
	if r.etag != `W/"67-cRkPwL0pcV2myUWCffwApguWrOw"` {
		t.Errorf("etag=%q", r.etag)
	}
}

func TestRestInclineInitials(t *testing.T) {
	base := startServer(t, testCfg())
	r := doReq(t, "POST", base+"/api/incline", map[string]string{
		"Authorization": authHdr, "Content-Type": "application/json",
	}, `{"SurName":"Иванов","FirstName":"Иван","SecondName":"Иванович","declension":"dative","format":"initials"}`)
	want := `{"SurName":"Иванову","initials":"И.И."}`
	if r.body != want {
		t.Errorf("body=%q want=%q", r.body, want)
	}
	if r.etag != `W/"30-ucq2GNzQnxi1NYREvpubS6ntemM"` {
		t.Errorf("etag=%q", r.etag)
	}
}

func TestRestGender(t *testing.T) {
	base := startServer(t, testCfg())
	r := doReq(t, "POST", base+"/api/gender", map[string]string{
		"Authorization": authHdr, "Content-Type": "application/json",
	}, `{"SurName":"Смирнова","FirstName":"Анна","SecondName":""}`)
	if r.body != `{"gender":"female"}` {
		t.Errorf("body=%q", r.body)
	}
	if r.etag != `W/"13-mQQQl849EeBKBSmIUdhIwMgE31o"` {
		t.Errorf("etag=%q", r.etag)
	}
}

func TestRestCities(t *testing.T) {
	base := startServer(t, testCfg())
	cases := []struct {
		url, body, want string
	}{
		{"/api/city/in", `{"name":"Москва","gender":"female"}`, `{"name":"Москве"}`},
		{"/api/city/from", `{"name":"Орел","gender":"female"}`, `{"name":"Орла"}`},
		{"/api/city/to", `{"name":"Москва"}`, `{"name":"Москву"}`},
	}
	for _, c := range cases {
		r := doReq(t, "POST", base+c.url, map[string]string{
			"Authorization": authHdr, "Content-Type": "application/json",
		}, c.body)
		if r.body != c.want {
			t.Errorf("%s: body=%q want=%q", c.url, r.body, c.want)
		}
	}
}

func TestRestOrgs(t *testing.T) {
	base := startServer(t, testCfg())
	cases := []struct {
		url, body, want string
	}{
		{"/api/org/in", `{"name":"ООО «Ромашка»"}`, `{"name":"ООО «Ромашке»"}`},
		{"/api/org/from", `{"name":"ООО «Ромашка»"}`, `{"name":"ООО «Ромашки»"}`},
		{"/api/org/to", `{"name":"ООО «Ромашка»"}`, `{"name":"ООО «Ромашку»"}`},
	}
	for _, c := range cases {
		r := doReq(t, "POST", base+c.url, map[string]string{
			"Authorization": authHdr, "Content-Type": "application/json",
		}, c.body)
		if r.body != c.want {
			t.Errorf("%s: body=%q want=%q", c.url, r.body, c.want)
		}
	}
}

func TestRestAuth(t *testing.T) {
	base := startServer(t, testCfg())
	r := doReq(t, "POST", base+"/api/incline", map[string]string{
		"Content-Type": "application/json",
	}, `{"SurName":"Иванов","FirstName":"Иван","SecondName":"Иванович"}`)
	if r.status != 401 || r.body != `{"error":"Unauthorized"}` {
		t.Errorf("noauth: status=%d body=%q", r.status, r.body)
	}
	r = doReq(t, "POST", base+"/api/incline", map[string]string{
		"Authorization": "Bearer wrong", "Content-Type": "application/json",
	}, `{"SurName":"Иванов"}`)
	if r.status != 401 || r.body != `{"error":"Unauthorized"}` {
		t.Errorf("badtoken: status=%d body=%q", r.status, r.body)
	}
}

func TestRest404AndBadJSON(t *testing.T) {
	base := startServer(t, testCfg())
	notFoundBody := "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<title>Error</title>\n</head>\n<body>\n<pre>Cannot GET /api/incline</pre>\n</body>\n</html>\n"
	r := doReq(t, "GET", base+"/api/incline", map[string]string{"Authorization": authHdr}, "")
	if r.status != 404 || r.body != notFoundBody {
		t.Errorf("GET /api/incline: status=%d body=%q", r.status, r.body)
	}
	non := doReq(t, "GET", base+"/nope", map[string]string{"Authorization": authHdr}, "")
	if non.body != strings.ReplaceAll(notFoundBody, "/api/incline", "/nope") {
		t.Errorf("GET /nope: body=%q", non.body)
	}

	r = doReq(t, "POST", base+"/api/incline", map[string]string{
		"Authorization": authHdr, "Content-Type": "application/json",
	}, `{"SurName": "Иванов", undeclared`)
	if r.status != 400 {
		t.Fatalf("badjson status=%d", r.status)
	}
	if !strings.Contains(r.body, "SyntaxError: Expected double-quoted property name in JSON at position 22 (line 1 column 23)") {
		t.Errorf("badjson msg missing, got %q", r.body)
	}
	if r.headers.Get("Content-Security-Policy") != "default-src 'none'" {
		t.Errorf("csp missing")
	}
}

func TestWsdl(t *testing.T) {
	base := startServer(t, testCfg())
	r := doReq(t, "GET", base+"/wsdl", map[string]string{"Authorization": authHdr}, "")
	if r.status != 200 || r.ct != "text/xml; charset=utf-8" {
		t.Fatalf("wsdl: status=%d ct=%q", r.status, r.ct)
	}
	if r.body != string(wsdlData) {
		t.Error("wsdl body mismatch")
	}
	r = doReq(t, "GET", base+"/soap?wsdl", map[string]string{"Authorization": authHdr}, "")
	if r.status != 200 || r.ct != "application/xml" || r.body != string(wsdlData) {
		t.Fatalf("soap?wsdl: status=%d ct=%q bodylen=%d", r.status, r.ct, len(r.body))
	}
	r = doReq(t, "GET", base+"/soap", map[string]string{"Authorization": authHdr}, "")
	if r.status != 200 || r.body != "" || r.ct != "" {
		t.Fatalf("GET /soap: status=%d ct=%q body=%q", r.status, r.ct, r.body)
	}
}

const soapEnvelope = `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="urn:LvovichService">
  <soap:Body>
    <tns:%s>%s</tns:%s>
  </soap:Body>
</soap:Envelope>`

func TestSoapInclineFull(t *testing.T) {
	base := startServer(t, testCfg())
	xmlReq := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="urn:LvovichService">
  <soap:Body>
    <tns:Incline>
      <LastName>Иванов</LastName>
      <FirstName>Иван</FirstName>
      <MiddleName>Иванович</MiddleName>
      <Declension>dative</Declension>
      <Format>full</Format>
    </tns:Incline>
  </soap:Body>
</soap:Envelope>`
	r := doReq(t, "POST", base+"/soap", map[string]string{
		"Authorization": authHdr, "Content-Type": "text/xml", "SOAPAction": "urn:Incline",
	}, xmlReq)
	want := "<?xml version=\"1.0\" encoding=\"utf-8\"?><soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\"  xmlns:tns=\"urn:LvovichService\"><soap:Body><tns:InclineResponse><InclineResponse><LastName>Иванову</LastName><FirstName>Ивану</FirstName><MiddleName>Ивановичу</MiddleName><Gender>male</Gender></InclineResponse></tns:InclineResponse></soap:Body></soap:Envelope>"
	if r.status != 200 || r.body != want {
		t.Errorf("status=%d\nbody=%s\nwant=%s", r.status, r.body, want)
	}
	if r.ct != "text/xml" {
		t.Errorf("ct=%q", r.ct)
	}
	if r.etag != "" {
		t.Errorf("soap-etag=%q (should be empty)", r.etag)
	}
}

func TestSoapInclineInitials(t *testing.T) {
	base := startServer(t, testCfg())
	inner := `<LastName>Иванов</LastName><FirstName>Иван</FirstName><MiddleName>Иванович</MiddleName><Declension>dative</Declension><Format>initials</Format>`
	req := strings.Replace(soapEnvelope, "%s", "Incline", 3)
	req = strings.Replace(req, `<tns:Incline>`+inner+`</tns:Incline>`, `<tns:Incline>`+inner+`</tns:Incline>`, 1)
	// просто соберём напрямую
	req = `<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="urn:LvovichService"><soap:Body><tns:Incline>` + inner + `</tns:Incline></soap:Body></soap:Envelope>`
	r := doReq(t, "POST", base+"/soap", map[string]string{
		"Authorization": authHdr, "Content-Type": "text/xml", "SOAPAction": "urn:Incline",
	}, req)
	want := "<?xml version=\"1.0\" encoding=\"utf-8\"?><soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\"  xmlns:tns=\"urn:LvovichService\"><soap:Body><tns:InclineResponse><InclineResponse><LastName>Иванову</LastName><Initials>И.И.</Initials></InclineResponse></tns:InclineResponse></soap:Body></soap:Envelope>"
	if r.body != want {
		t.Errorf("body=%s", r.body)
	}
}

func TestSoapGetGender(t *testing.T) {
	base := startServer(t, testCfg())
	req := `<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="urn:LvovichService"><soap:Body><tns:GetGender><LastName>Смирнова</LastName><FirstName>Анна</FirstName><MiddleName></MiddleName></tns:GetGender></soap:Body></soap:Envelope>`
	r := doReq(t, "POST", base+"/soap", map[string]string{
		"Authorization": authHdr, "Content-Type": "text/xml", "SOAPAction": "urn:GetGender",
	}, req)
	want := "<?xml version=\"1.0\" encoding=\"utf-8\"?><soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\"  xmlns:tns=\"urn:LvovichService\"><soap:Body><tns:GetGenderResponse><GetGenderResponse><Gender>female</Gender></GetGenderResponse></tns:GetGenderResponse></soap:Body></soap:Envelope>"
	if r.body != want {
		t.Errorf("body=%s", r.body)
	}
}

func TestSoapCities(t *testing.T) {
	base := startServer(t, testCfg())
	cases := []struct {
		op, act, inner, want string
	}{
		{"CityIn", "urn:CityIn", "<Name>Москва</Name><Gender>female</Gender>",
			"<tns:CityInResponse><CityInResponse><Name>Москве</Name></CityInResponse></tns:CityInResponse>"},
		{"CityFrom", "urn:CityFrom", "<Name>Орел</Name><Gender>female</Gender>",
			"<tns:CityFromResponse><CityFromResponse><Name>Орла</Name></CityFromResponse></tns:CityFromResponse>"},
		{"CityTo", "urn:CityTo", "<Name>Москва</Name>",
			"<tns:CityToResponse><CityToResponse><Name>Москву</Name></CityToResponse></tns:CityToResponse>"},
	}
	for _, c := range cases {
		req := `<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="urn:LvovichService"><soap:Body><tns:` + c.op + `>` + c.inner + `</tns:` + c.op + `></soap:Body></soap:Envelope>`
		r := doReq(t, "POST", base+"/soap", map[string]string{
			"Authorization": authHdr, "Content-Type": "text/xml", "SOAPAction": c.act,
		}, req)
		if !strings.Contains(r.body, c.want) {
			t.Errorf("%s: body=%s", c.op, r.body)
		}
	}
}

func TestSoapCityFaultNoName(t *testing.T) {
	base := startServer(t, testCfg())
	req := `<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="urn:LvovichService"><soap:Body><tns:CityIn></tns:CityIn></soap:Body></soap:Envelope>`
	r := doReq(t, "POST", base+"/soap", map[string]string{
		"Authorization": authHdr, "Content-Type": "text/xml", "SOAPAction": "urn:CityIn",
	}, req)
	want := "<Fault>Cannot read properties of null (reading &apos;Name&apos;)</Fault>"
	if r.status != 200 || !strings.Contains(r.body, want) {
		t.Errorf("status=%d body=%s", r.status, r.body)
	}
}

func TestSoapOrgs(t *testing.T) {
	base := startServer(t, testCfg())
	cases := []struct {
		op, act, inner, want string
	}{
		{"OrgIn", "urn:OrgIn", "<Name>ООО «Ромашка»</Name>",
			"<tns:OrgInResponse><OrgInResponse><Name>ООО «Ромашке»</Name></OrgInResponse></tns:OrgInResponse>"},
		{"OrgFrom", "urn:OrgFrom", "<Name>ООО «Ромашка»</Name>",
			"<tns:OrgFromResponse><OrgFromResponse><Name>ООО «Ромашки»</Name></OrgFromResponse></tns:OrgFromResponse>"},
		{"OrgTo", "urn:OrgTo", "<Name>ООО «Ромашка»</Name>",
			"<tns:OrgToResponse><OrgToResponse><Name>ООО «Ромашку»</Name></OrgToResponse></tns:OrgToResponse>"},
	}
	for _, c := range cases {
		req := `<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="urn:LvovichService"><soap:Body><tns:` + c.op + `>` + c.inner + `</tns:` + c.op + `></soap:Body></soap:Envelope>`
		r := doReq(t, "POST", base+"/soap", map[string]string{
			"Authorization": authHdr, "Content-Type": "text/xml", "SOAPAction": c.act,
		}, req)
		if !strings.Contains(r.body, c.want) {
			t.Errorf("%s: body=%s", c.op, r.body)
		}
	}
}

func TestSoapNoActionAndBadAction(t *testing.T) {
	base := startServer(t, testCfg())
	inner := `<LastName>Иванов</LastName>`
	req := `<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="urn:LvovichService"><soap:Body><tns:Incline>` + inner + `</tns:Incline></soap:Body></soap:Envelope>`
	// движок без SOAPAction (по телу)
	r := doReq(t, "POST", base+"/soap", map[string]string{
		"Authorization": authHdr, "Content-Type": "text/xml",
	}, req)
	if r.status != 200 || !strings.Contains(r.body, "LastName>Иванова</LastName") {
		t.Errorf("noaction: status=%d body=%s", r.status, r.body)
	}
	// неизвестный SOAPAction -> 500 'style'
	r = doReq(t, "POST", base+"/soap", map[string]string{
		"Authorization": authHdr, "Content-Type": "text/xml", "SOAPAction": "urn:LvovichService",
	}, req)
	if r.status != 500 || !strings.Contains(r.body, "reading &apos;style&apos;") {
		t.Errorf("badaction: status=%d body=%s", r.status, r.body)
	}
}

func TestSoapUnknownOp(t *testing.T) {
	base := startServer(t, testCfg())
	req := `<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="urn:LvovichService"><soap:Body><tns:NoSuchOp><Name>X</Name></tns:NoSuchOp></soap:Body></soap:Envelope>`
	r := doReq(t, "POST", base+"/soap", map[string]string{
		"Authorization": authHdr, "Content-Type": "text/xml", "SOAPAction": "urn:Incline",
	}, req)
	if r.status != 500 || !strings.Contains(r.body, "reading 'description'") {
		t.Errorf("status=%d body=%s", r.status, r.body)
	}
}

func TestSwaggerUI(t *testing.T) {
	base := startServer(t, testCfg())
	r := doReq(t, "GET", base+"/api-docs", map[string]string{"Authorization": authHdr}, "")
	if r.status != 200 || r.ct != "text/html; charset=utf-8" {
		t.Fatalf("apidocs: status=%d ct=%q", r.status, r.ct)
	}
	if data, _ := swaggerStatic.ReadFile("swagger/static/index.html"); string(data) != r.body {
		t.Errorf("index mismatch (len %d vs %d)", len(data), len(r.body))
	}
	r = doReq(t, "GET", base+"/api-docs/swagger-ui-init.js", map[string]string{"Authorization": authHdr}, "")
	if r.ct != "application/javascript; charset=utf-8" {
		t.Errorf("initjs ct=%q", r.ct)
	}
	if data, _ := swaggerStatic.ReadFile("swagger/static/swagger-ui-init.js"); string(data) != r.body {
		t.Errorf("initjs mismatch")
	}
	r = doReq(t, "GET", base+"/api-docs/swagger-ui.css", map[string]string{"Authorization": authHdr}, "")
	if r.ct != "text/css; charset=utf-8" || r.headers.Get("Cache-Control") != "public, max-age=0" {
		t.Errorf("css headers: ct=%q cc=%q", r.ct, r.headers.Get("Cache-Control"))
	}
	if r.etag != `W/"2bb21-19f678cb4ba"` {
		t.Errorf("css etag=%q", r.etag)
	}
	// 304 при If-None-Match
	r2 := doReq(t, "GET", base+"/api-docs/swagger-ui.css", map[string]string{
		"Authorization": authHdr, "If-None-Match": `W/"2bb21-19f678cb4ba"`,
	}, "")
	if r2.status != 304 {
		t.Errorf("css 304 status=%d", r2.status)
	}
}

func TestSwaggerDisabled(t *testing.T) {
	cfg := testCfg()
	cfg.Swagger = false
	base := startServer(t, cfg)
	r := doReq(t, "GET", base+"/api-docs", map[string]string{"Authorization": authHdr}, "")
	if r.status != 404 {
		t.Errorf("apidocs disabled status=%d", r.status)
	}
}

// TestLoggingDisabledNoLogFile проверяет, что при Logging=false запросы
// обрабатываются корректно, а файл лога вообще не создаётся (модуль логирования
// не задействуется в горячем пути).
func TestLoggingDisabledNoLogFile(t *testing.T) {
	cfg := testCfg()
	cfg.Logging = false
	dir := t.TempDir()
	logPath := filepath.Join(dir, "server.log")

	h := NewServer(cfg, logPath)
	defer h.Close()

	req, err := http.NewRequest(http.MethodPost, "/api/gender", strings.NewReader(`{"SurName":"Смирнова","FirstName":"Анна"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.RequestURI = "/api/gender"
	req.RemoteAddr = "127.0.0.1:0"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHdr)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}

	h.Close()
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Fatalf("при отключённом логировании файл лога не должен создаваться: %v", err)
	}
}