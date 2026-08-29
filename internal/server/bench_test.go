package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// benchServer создаёт сервер с лог-файлом во временной папке (как в проде —
// запись на диск на каждый запрос является частью замера).
func benchServer(b *testing.B) *Server {
	b.Helper()
	return benchServerCfg(b, testCfg())
}

// benchServerCfg создаёт сервер с заданным конфигом (mode sync/async).
func benchServerCfg(b *testing.B, cfg Config) *Server {
	b.Helper()
	h := NewServer(cfg, filepath.Join(b.TempDir(), "server.log"))
	b.Cleanup(h.Close)
	return h
}

// benchReq прогоняет один запрос через handler сервера.
// Запрос строится заново на каждой итерации (тело Reader нельзя переиспользовать).
// RequestURI и RemoteAddr выставляются вручную: при прямом вызове ServeHTTP без
// реального клиента http.NewRequest не заполняет их, а rawPath()/clientIP() на них опираются.
func benchReq(b *testing.B, h http.Handler, method, target, contentType, body string) {
	b.Helper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest(method, target, strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		req.RequestURI = target
		req.RemoteAddr = "127.0.0.1:0"
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Authorization", authHdr)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		res := rr.Result()
		if res.StatusCode != http.StatusOK {
			b.Fatalf("status %d", res.StatusCode)
		}
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}
}

func BenchmarkRestInclineFull(b *testing.B) {
	benchReq(b, benchServer(b), http.MethodPost, "/api/incline", "application/json",
		`{"SurName":"Иванов","FirstName":"Иван","SecondName":"Иванович","declension":"dative"}`)
}

func BenchmarkRestInclineInitials(b *testing.B) {
	benchReq(b, benchServer(b), http.MethodPost, "/api/incline", "application/json",
		`{"SurName":"Петров","FirstName":"Пётр","SecondName":"Петрович","declension":"genitive","format":"initials"}`)
}

func BenchmarkRestGender(b *testing.B) {
	benchReq(b, benchServer(b), http.MethodPost, "/api/gender", "application/json",
		`{"SurName":"Смирнова","FirstName":"Анна"}`)
}

func BenchmarkRestCityIn(b *testing.B) {
	benchReq(b, benchServer(b), http.MethodPost, "/api/city/in", "application/json",
		`{"name":"Москва","gender":"female"}`)
}

func BenchmarkRestCityFrom(b *testing.B) {
	benchReq(b, benchServer(b), http.MethodPost, "/api/city/from", "application/json",
		`{"name":"Москва","gender":"female"}`)
}

func BenchmarkRestCityTo(b *testing.B) {
	benchReq(b, benchServer(b), http.MethodPost, "/api/city/to", "application/json",
		`{"name":"Москва"}`)
}

func BenchmarkSoapIncline(b *testing.B) {
	const soapBody = `<?xml version="1.0" encoding="UTF-8"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><Incline xmlns="urn:LvovichService"><LastName>Иванов</LastName><FirstName>Иван</FirstName><MiddleName>Иванович</MiddleName><Declension>dative</Declension></Incline></soap:Body></soap:Envelope>`
	benchReq(b, benchServer(b), http.MethodPost, "/soap", "text/xml", soapBody)
}

// syncCfg — конфиг с режимом лога "sync" (прямая запись на диск на каждый вызов),
// для сравнения с асинхронным флашем.
func syncCfg() Config {
	c := testCfg()
	c.LogMode = "sync"
	return c
}

// noLogCfg — конфиг с полностью отключённым логированием (Logging=false).
func noLogCfg() Config {
	c := testCfg()
	c.Logging = false
	return c
}

func BenchmarkSyncRestInclineFull(b *testing.B) {
	benchReq(b, benchServerCfg(b, syncCfg()), http.MethodPost, "/api/incline", "application/json",
		`{"SurName":"Иванов","FirstName":"Иван","SecondName":"Иванович","declension":"dative"}`)
}

func BenchmarkSyncRestGender(b *testing.B) {
	benchReq(b, benchServerCfg(b, syncCfg()), http.MethodPost, "/api/gender", "application/json",
		`{"SurName":"Смирнова","FirstName":"Анна"}`)
}

func BenchmarkSyncSoapIncline(b *testing.B) {
	const soapBody = `<?xml version="1.0" encoding="UTF-8"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><Incline xmlns="urn:LvovichService"><LastName>Иванов</LastName><FirstName>Иван</FirstName><MiddleName>Иванович</MiddleName><Declension>dative</Declension></Incline></soap:Body></soap:Envelope>`
	benchReq(b, benchServerCfg(b, syncCfg()), http.MethodPost, "/soap", "text/xml", soapBody)
}

func BenchmarkSyncParallelRestGender(b *testing.B) {
	benchReqParallel(b, benchServerCfg(b, syncCfg()), http.MethodPost, "/api/gender", "application/json",
		`{"SurName":"Смирнова","FirstName":"Анна"}`)
}

func BenchmarkSyncParallelSoapIncline(b *testing.B) {
	const soapBody = `<?xml version="1.0" encoding="UTF-8"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><Incline xmlns="urn:LvovichService"><LastName>Иванов</LastName><FirstName>Иван</FirstName><MiddleName>Иванович</MiddleName><Declension>dative</Declension></Incline></soap:Body></soap:Envelope>`
	benchReqParallel(b, benchServerCfg(b, syncCfg()), http.MethodPost, "/soap", "text/xml", soapBody)
}

// benchReqParallel гоняет запросы из нескольких горутин (b.RunParallel),
// имитируя конкурентные вызовы реального сервера. Каждая горутина строит
// собственный запрос.
func benchReqParallel(b *testing.B, h http.Handler, method, target, contentType, body string) {
	b.Helper()
	hdr := map[string]string{"Content-Type": contentType, "Authorization": authHdr}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, err := http.NewRequest(method, target, strings.NewReader(body))
			if err != nil {
				b.Fatal(err)
			}
			req.RequestURI = target
			req.RemoteAddr = "127.0.0.1:0"
			for k, v := range hdr {
				req.Header.Set(k, v)
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			res := rr.Result()
			if res.StatusCode != http.StatusOK {
				b.Fatalf("status %d", res.StatusCode)
			}
			_, _ = io.Copy(io.Discard, res.Body)
			_ = res.Body.Close()
		}
	})
}

func BenchmarkParallelRestInclineFull(b *testing.B) {
	benchReqParallel(b, benchServer(b), http.MethodPost, "/api/incline", "application/json",
		`{"SurName":"Иванов","FirstName":"Иван","SecondName":"Иванович","declension":"dative"}`)
}

func BenchmarkParallelRestInclineInitials(b *testing.B) {
	benchReqParallel(b, benchServer(b), http.MethodPost, "/api/incline", "application/json",
		`{"SurName":"Петров","FirstName":"Пётр","SecondName":"Петрович","declension":"genitive","format":"initials"}`)
}

func BenchmarkParallelRestGender(b *testing.B) {
	benchReqParallel(b, benchServer(b), http.MethodPost, "/api/gender", "application/json",
		`{"SurName":"Смирнова","FirstName":"Анна"}`)
}

func BenchmarkParallelRestCityIn(b *testing.B) {
	benchReqParallel(b, benchServer(b), http.MethodPost, "/api/city/in", "application/json",
		`{"name":"Москва","gender":"female"}`)
}

func BenchmarkParallelRestCityFrom(b *testing.B) {
	benchReqParallel(b, benchServer(b), http.MethodPost, "/api/city/from", "application/json",
		`{"name":"Москва","gender":"female"}`)
}

func BenchmarkParallelRestCityTo(b *testing.B) {
	benchReqParallel(b, benchServer(b), http.MethodPost, "/api/city/to", "application/json",
		`{"name":"Москва"}`)
}

func BenchmarkParallelSoapIncline(b *testing.B) {
	const soapBody = `<?xml version="1.0" encoding="UTF-8"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><Incline xmlns="urn:LvovichService"><LastName>Иванов</LastName><FirstName>Иван</FirstName><MiddleName>Иванович</MiddleName><Declension>dative</Declension></Incline></soap:Body></soap:Envelope>`
	benchReqParallel(b, benchServer(b), http.MethodPost, "/soap", "text/xml", soapBody)
}

// Бенчмарки с полностью отключённым логированием (Logging=false).
// Показывают чистое ядро склонения без затрат на сборку и буферизацию лога.

func BenchmarkNoLogRestInclineFull(b *testing.B) {
	benchReq(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/incline", "application/json",
		`{"SurName":"Иванов","FirstName":"Иван","SecondName":"Иванович","declension":"dative"}`)
}

func BenchmarkNoLogRestInclineInitials(b *testing.B) {
	benchReq(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/incline", "application/json",
		`{"SurName":"Петров","FirstName":"Пётр","SecondName":"Петрович","declension":"genitive","format":"initials"}`)
}

func BenchmarkNoLogRestGender(b *testing.B) {
	benchReq(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/gender", "application/json",
		`{"SurName":"Смирнова","FirstName":"Анна"}`)
}

func BenchmarkNoLogRestCityIn(b *testing.B) {
	benchReq(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/city/in", "application/json",
		`{"name":"Москва","gender":"female"}`)
}

func BenchmarkNoLogRestCityFrom(b *testing.B) {
	benchReq(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/city/from", "application/json",
		`{"name":"Москва","gender":"female"}`)
}

func BenchmarkNoLogRestCityTo(b *testing.B) {
	benchReq(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/city/to", "application/json",
		`{"name":"Москва"}`)
}

func BenchmarkNoLogSoapIncline(b *testing.B) {
	const soapBody = `<?xml version="1.0" encoding="UTF-8"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><Incline xmlns="urn:LvovichService"><LastName>Иванов</LastName><FirstName>Иван</FirstName><MiddleName>Иванович</MiddleName><Declension>dative</Declension></Incline></soap:Body></soap:Envelope>`
	benchReq(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/soap", "text/xml", soapBody)
}

func BenchmarkNoLogParallelRestInclineFull(b *testing.B) {
	benchReqParallel(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/incline", "application/json",
		`{"SurName":"Иванов","FirstName":"Иван","SecondName":"Иванович","declension":"dative"}`)
}

func BenchmarkNoLogParallelRestInclineInitials(b *testing.B) {
	benchReqParallel(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/incline", "application/json",
		`{"SurName":"Петров","FirstName":"Пётр","SecondName":"Петрович","declension":"genitive","format":"initials"}`)
}

func BenchmarkNoLogParallelRestGender(b *testing.B) {
	benchReqParallel(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/gender", "application/json",
		`{"SurName":"Смирнова","FirstName":"Анна"}`)
}

func BenchmarkNoLogParallelRestCityIn(b *testing.B) {
	benchReqParallel(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/city/in", "application/json",
		`{"name":"Москва","gender":"female"}`)
}

func BenchmarkNoLogParallelRestCityFrom(b *testing.B) {
	benchReqParallel(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/city/from", "application/json",
		`{"name":"Москва","gender":"female"}`)
}

func BenchmarkNoLogParallelRestCityTo(b *testing.B) {
	benchReqParallel(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/city/to", "application/json",
		`{"name":"Москва"}`)
}

func BenchmarkNoLogParallelSoapIncline(b *testing.B) {
	const soapBody = `<?xml version="1.0" encoding="UTF-8"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><Incline xmlns="urn:LvovichService"><LastName>Иванов</LastName><FirstName>Иван</FirstName><MiddleName>Иванович</MiddleName><Declension>dative</Declension></Incline></soap:Body></soap:Envelope>`
	benchReqParallel(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/soap", "text/xml", soapBody)
}

// Бенчмарки склонения организаций (REST).

func BenchmarkRestOrgIn(b *testing.B) {
	benchReq(b, benchServer(b), http.MethodPost, "/api/org/in", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}

func BenchmarkRestOrgFrom(b *testing.B) {
	benchReq(b, benchServer(b), http.MethodPost, "/api/org/from", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}

func BenchmarkRestOrgTo(b *testing.B) {
	benchReq(b, benchServer(b), http.MethodPost, "/api/org/to", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}

func BenchmarkParallelRestOrgIn(b *testing.B) {
	benchReqParallel(b, benchServer(b), http.MethodPost, "/api/org/in", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}

func BenchmarkParallelRestOrgFrom(b *testing.B) {
	benchReqParallel(b, benchServer(b), http.MethodPost, "/api/org/from", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}

func BenchmarkParallelRestOrgTo(b *testing.B) {
	benchReqParallel(b, benchServer(b), http.MethodPost, "/api/org/to", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}

func BenchmarkNoLogRestOrgIn(b *testing.B) {
	benchReq(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/org/in", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}

func BenchmarkNoLogRestOrgFrom(b *testing.B) {
	benchReq(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/org/from", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}

func BenchmarkNoLogRestOrgTo(b *testing.B) {
	benchReq(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/org/to", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}

func BenchmarkNoLogParallelRestOrgIn(b *testing.B) {
	benchReqParallel(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/org/in", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}

func BenchmarkNoLogParallelRestOrgFrom(b *testing.B) {
	benchReqParallel(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/org/from", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}

func BenchmarkNoLogParallelRestOrgTo(b *testing.B) {
	benchReqParallel(b, benchServerCfg(b, noLogCfg()), http.MethodPost, "/api/org/to", "application/json",
		`{"name":"ООО «Ромашка»"}`)
}
