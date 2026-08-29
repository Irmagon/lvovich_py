package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestParallelStatuses(t *testing.T) {
	h := NewServer(testCfg(), filepath.Join(t.TempDir(), "server.log"))
	t.Cleanup(h.Close)
	const soapBody = `<?xml version="1.0" encoding="UTF-8"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><Incline xmlns="urn:LvovichService"><LastName>Иванов</LastName><FirstName>Иван</FirstName><MiddleName>Иванович</MiddleName><Declension>dative</Declension></Incline></soap:Body></soap:Envelope>`

	run := func(makeReq func() *http.Request) map[int]int {
		var mu sync.Mutex
		statuses := map[int]int{}
		var wg sync.WaitGroup
		for i := 0; i < 16; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					rr := httptest.NewRecorder()
					h.ServeHTTP(rr, makeReq())
					code := rr.Code
					rr.Result().Body.Close()
					mu.Lock()
					statuses[code]++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
		return statuses
	}

	soapReq := func() *http.Request {
		r, err := http.NewRequest(http.MethodPost, "/soap", strings.NewReader(soapBody))
		if err != nil {
			t.Fatal(err)
		}
		r.RequestURI = "/soap"
		r.RemoteAddr = "127.0.0.1:0"
		r.Header.Set("Content-Type", "application/xml")
		r.Header.Set("Authorization", authHdr)
		return r
	}
	t.Logf("SOAP statuses: %v", run(soapReq))

	restReq := func() *http.Request {
		r, err := http.NewRequest(http.MethodPost, "/api/incline",
			strings.NewReader(`{"SurName":"Иванов","FirstName":"Иван","SecondName":"Иванович","declension":"dative"}`))
		if err != nil {
			t.Fatal(err)
		}
		r.RequestURI = "/api/incline"
		r.RemoteAddr = "127.0.0.1:0"
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", authHdr)
		return r
	}
	t.Logf("REST statuses: %v", run(restReq))
}
