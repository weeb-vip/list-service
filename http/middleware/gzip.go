package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GzipMiddleware compresses HTTP responses using gzip if the client supports it
func GzipMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the client accepts gzip encoding
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Set the content encoding header
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			// Create gzip writer
			gz := gzip.NewWriter(w)
			defer gz.Close()

			// Wrap the response writer
			gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
			next.ServeHTTP(gzw, r)
		})
	}
}