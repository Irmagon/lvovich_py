package server

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// WeakETag — Express-совместимый слабый ETag для строкового тела:
// W/"<длина-hex>-<base64(sha1, без padding)>" (как в пакете etag для Express).
func WeakETag(body string) string {
	h := sha1.Sum([]byte(body))
	enc := base64.StdEncoding.EncodeToString(h[:])
	enc = strings.TrimRight(enc, "=")
	return fmt.Sprintf("W/\"%x-%s\"", len(body), enc)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// sendJSON пишет JSON-ответ в express-стиле (Content-Type, ETag, X-Powered-By).
func sendJSON(w http.ResponseWriter, r *http.Request, status int, body string) {
	h := w.Header()
	h.Set("Content-Type", "application/json; charset=utf-8")
	sendBody(w, r, h, status, body)
}

// sendBody — общая отправка тела с ETag, 304 и X-Powered-By.
func sendBody(w http.ResponseWriter, r *http.Request, h http.Header, status int, body string) {
	h.Set("X-Powered-By", "Express")
	etag := WeakETag(body)
	h.Set("ETag", etag)
	if status == http.StatusOK {
		if inm := r.Header.Get("If-None-Match"); inm == etag || inm == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// serveXML пишет XML-ответ в express/soap-стиле.
func serveXML(w http.ResponseWriter, r *http.Request, status int, body string) {
	h := w.Header()
	h.Set("Content-Type", "text/xml")
	sendBody(w, r, h, status, body)
}

// sendNoBody — ответ без тела и без Content-Type (GET /soap).
func sendNoBody(w http.ResponseWriter, status int) {
	h := w.Header()
	h.Set("X-Powered-By", "Express")
	w.WriteHeader(status)
}

// errPage — страница ошибки Express finalhandler (404/400).
func errPage(w http.ResponseWriter, status int, title, pre string) {
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Content-Security-Policy", "default-src 'none'")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Powered-By", "Express")
	body := "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<title>" + title + "</title>\n</head>\n<body>\n<pre>" + pre + "</pre>\n</body>\n</html>\n"
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}
