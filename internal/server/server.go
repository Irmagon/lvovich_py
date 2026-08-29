package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// Server — HTTP-сервер, повторяющий поведение scripts/soap/server.ts.
type Server struct {
	cfg  Config
	log  *Logger
	core *Core
}

// NewServer создаёт сервер с конфигурацией и логгером.
// Если cfg.Logging=false, логгер создаётся отключённым (запись в файл не ведётся).
func NewServer(cfg Config, logPath string) *Server {
	return &Server{cfg: cfg, log: NewLogger(logPath, cfg.LogMode, cfg.FlushMs, cfg.BufferKB, cfg.Logging), core: NewCore()}
}

// Log возвращает логгер (для сообщения при старте).
func (s *Server) Log() *Logger { return s.log }

// Close останавливает флашер логгера и сбрасывает остаток буфера на диск.
func (s *Server) Close() { s.log.Close() }

// pathMatches — соответствие express-префиксу: '/api' или '/api/...'.
func pathMatches(p, prefix string) bool {
	return p == prefix || strings.HasPrefix(p, prefix+"/")
}

// clientIP — IP клиента (без порта), как req.ip без trust proxy.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *Server) ipAllowed(ip string) bool {
	if len(s.cfg.AllowedIPs) == 0 {
		return true
	}
	for _, a := range s.cfg.AllowedIPs {
		if a == ip {
			return true
		}
	}
	return false
}

// rawPath — путь запроса без query (как Express req.path/originalUrl без query).
func rawPath(r *http.Request) string {
	raw := r.RequestURI
	if i := strings.IndexByte(raw, '?'); i >= 0 {
		raw = raw[:i]
	}
	return raw
}

// logAccess пишет строку лога доступа. При отключённом логировании
// (Enabled()==false) строка вообще не формируется и модуль логирования
// не вызывается.
func (s *Server) logAccess(ip string, r *http.Request, prefix ...string) {
	if !s.log.Enabled() {
		return
	}
	pre := ""
	if len(prefix) > 0 {
		pre = prefix[0] + " "
	}
	s.log.Log(ip, fmt.Sprintf("%s%s %s - %s", pre, r.Method, r.RequestURI, bodyLogText(r)))
}

// ServeHTTP разбирает, логирует, применяет whitelist+auth и маршрутизирует.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bi, handled := s.parseBody(w, r)
	if handled {
		return
	}
	if bi != nil {
		r = withBodyInfo(r, bi)
	}

	path := rawPath(r)
	ip := clientIP(r)
	s.logAccess(ip, r)

	if pathMatches(path, "/api") || pathMatches(path, "/soap") {
		if !s.ipAllowed(ip) {
			s.logAccess(ip, r, "blocked IP "+ip)
			sendJSON(w, r, http.StatusForbidden, `{"error":"Forbidden"}`)
			return
		}
		if s.cfg.Token != "" {
			hdr := r.Header.Get("Authorization")
			tok := ""
			if strings.HasPrefix(hdr, "Bearer ") {
				tok = hdr[len("Bearer "):]
			}
			if tok != s.cfg.Token {
				sendJSON(w, r, http.StatusUnauthorized, `{"error":"Unauthorized"}`)
				return
			}
		}
	}

	switch {
	case path == "/wsdl" && r.Method == http.MethodGet:
		s.serveWsdl(w, r)
		return
	case path == "/soap" && r.Method == http.MethodGet:
		if _, has := r.URL.Query()["wsdl"]; has {
			s.serveWsdlSoap(w, r)
		} else {
			sendNoBody(w, http.StatusOK)
		}
		return
	case path == "/soap" && r.Method == http.MethodPost:
		s.handleSoap(w, r)
		return
	case path == "/api/incline" && r.Method == http.MethodPost:
		s.handleRestIncline(w, r)
		return
	case path == "/api/gender" && r.Method == http.MethodPost:
		s.handleRestGender(w, r)
		return
	case path == "/api/city/in" && r.Method == http.MethodPost:
		s.handleRestCity(w, r, "in")
		return
	case path == "/api/city/from" && r.Method == http.MethodPost:
		s.handleRestCity(w, r, "from")
		return
	case path == "/api/city/to" && r.Method == http.MethodPost:
		s.handleRestCity(w, r, "to")
		return
	case path == "/api/org/in" && r.Method == http.MethodPost:
		s.handleRestOrg(w, r, "in")
		return
	case path == "/api/org/from" && r.Method == http.MethodPost:
		s.handleRestOrg(w, r, "from")
		return
	case path == "/api/org/to" && r.Method == http.MethodPost:
		s.handleRestOrg(w, r, "to")
		return
	}

	if s.cfg.Swagger && r.Method == http.MethodGet {
		if path == "/api-docs" || path == "/api-docs/" {
			s.serveAsset(w, r, "index.html")
			return
		}
		if strings.HasPrefix(path, "/api-docs/") && s.serveAsset(w, r, path[len("/api-docs/"):]) {
			return
		}
	}

	s.notFound(w, r)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	errPage(w, http.StatusNotFound, "Error", "Cannot "+r.Method+" "+rawPath(r))
}

// serveWsdl — GET /wsdl: WSDL как text/xml (как express-роут).
func (s *Server) serveWsdl(w http.ResponseWriter, r *http.Request) {
	sendBody(w, r, contentHeader("text/xml; charset=utf-8"), http.StatusOK, string(wsdlData))
}

// serveWsdlSoap — GET /soap?wsdl: WSDL как application/xml (node-soap), без etag.
func (s *Server) serveWsdlSoap(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("Content-Type", "application/xml")
	h.Set("X-Powered-By", "Express")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(wsdlData)
}

// contentHeader строит заголовки для sendBody.
func contentHeader(ct string) http.Header {
	h := http.Header{}
	h.Set("Content-Type", ct)
	return h
}
