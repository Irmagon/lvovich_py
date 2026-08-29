package server

import (
	"embed"
	"net/http"
)

//go:embed swagger/static
var swaggerStatic embed.FS

// asset — метаданные статики Swagger UI (совпадают с ответами swagger-ui-express).
type asset struct {
	contentType  string
	etag         string
	lastModified string
	cacheControl string
	acceptRanges bool
}

var assetTable = map[string]asset{
	"index.html": {
		contentType: "text/html; charset=utf-8",
		etag:        `W/"be7-9xiVWK/dq2a+vJHYNAuNjnOELik"`,
	},
	"swagger-ui-init.js": {
		contentType: "application/javascript; charset=utf-8",
		etag:        `W/"2526-aL99pre8PBlfulGj11htsbtr9zw"`,
	},
	"swagger-ui.css": {
		contentType:  "text/css; charset=utf-8",
		etag:         `W/"2bb21-19f678cb4ba"`,
		lastModified: "Wed, 15 Jul 2026 20:51:42 GMT",
		cacheControl: "public, max-age=0",
		acceptRanges: true,
	},
	"swagger-ui-bundle.js": {
		contentType:  "text/javascript; charset=utf-8",
		etag:         `W/"177639-19f678cb4f9"`,
		lastModified: "Wed, 15 Jul 2026 20:51:42 GMT",
		cacheControl: "public, max-age=0",
		acceptRanges: true,
	},
	"swagger-ui-standalone-preset.js": {
		contentType:  "text/javascript; charset=utf-8",
		etag:         `W/"3eab6-19f678cb525"`,
		lastModified: "Wed, 15 Jul 2026 20:51:42 GMT",
		cacheControl: "public, max-age=0",
		acceptRanges: true,
	},
	"favicon-32x32.png": {
		contentType:  "image/png",
		etag:         `W/"274-19f678cb56f"`,
		lastModified: "Wed, 15 Jul 2026 20:51:42 GMT",
		cacheControl: "public, max-age=0",
		acceptRanges: true,
	},
	"favicon-16x16.png": {
		contentType:  "image/png",
		etag:         `W/"299-19f678cb579"`,
		lastModified: "Wed, 15 Jul 2026 20:51:42 GMT",
		cacheControl: "public, max-age=0",
		acceptRanges: true,
	},
}

// serveAsset отдаёт ассет Swagger UI (или false, если имя не из таблицы).
func (s *Server) serveAsset(w http.ResponseWriter, r *http.Request, name string) bool {
	as, ok := assetTable[name]
	if !ok {
		return false
	}
	data, err := swaggerStatic.ReadFile("swagger/static/" + name)
	if err != nil {
		return false
	}
	if as.etag != "" && r.Header.Get("If-None-Match") == as.etag {
		h := w.Header()
		h.Set("ETag", as.etag)
		h.Set("X-Powered-By", "Express")
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	h := w.Header()
	h.Set("Content-Type", as.contentType)
	h.Set("X-Powered-By", "Express")
	if as.etag != "" {
		h.Set("ETag", as.etag)
	}
	if as.lastModified != "" {
		h.Set("Last-Modified", as.lastModified)
	}
	if as.cacheControl != "" {
		h.Set("Cache-Control", as.cacheControl)
	}
	if as.acceptRanges {
		h.Set("Accept-Ranges", "bytes")
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
	return true
}
