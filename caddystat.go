package caddystat

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	startTime = time.Now()
	caddy.RegisterModule(Stat{})
	caddy.RegisterAdminHTTPHandler("/stats", http.HandlerFunc(handleStats))
}

// Stat implements a simple traffic counter for HTTP requests.
type Stat struct{}

// CaddyModule returns the Caddy module information.
func (Stat) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.stat",
		New: func() caddy.Module { return new(Stat) },
	}
}

// Provision sets up the module.
func (s *Stat) Provision(ctx caddy.Context) error {
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (s *Stat) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	atomic.AddUint64(&requestCount, 1)

	r.Body = &countReader{r: r.Body}
	cw := &countWriter{w: w}

	return next.ServeHTTP(cw, r)
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Stat)(nil)
	_ caddyhttp.MiddlewareHandler = (*Stat)(nil)
)

var (
	startTime       time.Time
	upstreamBytes   uint64
	downstreamBytes uint64
	requestCount    uint64
)

type countReader struct {
	r io.ReadCloser
}

func (cr *countReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	if n > 0 {
		atomic.AddUint64(&upstreamBytes, uint64(n))
	}
	return n, err
}

func (cr *countReader) Close() error {
	return cr.r.Close()
}

type countWriter struct {
	w http.ResponseWriter
}

func (cw *countWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *countWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	if n > 0 {
		atomic.AddUint64(&downstreamBytes, uint64(n))
	}
	return n, err
}

func (cw *countWriter) WriteHeader(statusCode int) {
	cw.w.WriteHeader(statusCode)
}

func (cw *countWriter) Flush() {
	if f, ok := cw.w.(http.Flusher); ok {
		f.Flush()
	}
}

func (cw *countWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := cw.w.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := map[string]interface{}{
		"upstream_bytes":   atomic.LoadUint64(&upstreamBytes),
		"downstream_bytes": atomic.LoadUint64(&downstreamBytes),
		"request_count":    atomic.LoadUint64(&requestCount),
		"uptime_seconds":   time.Since(startTime).Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
