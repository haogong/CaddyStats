package naivestats

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	startTime = time.Now()
	caddy.RegisterModule(NaiveStat{})
	caddy.RegisterAdminHTTPHandler("/naive_stats", http.HandlerFunc(handleStats))
	httpcaddyfile.RegisterHandlerDirective("naive_stat", parseCaddyfile)
}

// NaiveStat implements a traffic counter for NaiveProxy requests.
type NaiveStat struct{}

// CaddyModule returns the Caddy module information.
func (NaiveStat) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.naive_stat",
		New: func() caddy.Module { return new(NaiveStat) },
	}
}

// Provision sets up the module.
func (s *NaiveStat) Provision(ctx caddy.Context) error {
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (s *NaiveStat) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return nil
}

// parseCaddyfile sets up the handler from Caddyfile tokens.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var s NaiveStat
	err := s.UnmarshalCaddyfile(h.Dispenser)
	return &s, err
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (s *NaiveStat) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	atomic.AddUint64(&requestCount, 1)

	r.Body = &countReader{r: r.Body}
	cw := &countWriter{w: w}

	return next.ServeHTTP(cw, r)
}

// Interface guards
var (
	_ caddy.Provisioner           = (*NaiveStat)(nil)
	_ caddyhttp.MiddlewareHandler = (*NaiveStat)(nil)
	_ caddyfile.Unmarshaler       = (*NaiveStat)(nil)
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
	h, ok := cw.w.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	conn, _, err := h.Hijack()
	if err != nil {
		return nil, nil, err
	}
	cc := &countConn{Conn: conn}
	brw := bufio.NewReadWriter(bufio.NewReader(cc), bufio.NewWriter(cc))
	return cc, brw, nil
}

// countConn wraps net.Conn to count bytes on hijacked connections (H1 CONNECT).
type countConn struct {
	net.Conn
}

func (cc *countConn) Read(p []byte) (int, error) {
	n, err := cc.Conn.Read(p)
	if n > 0 {
		atomic.AddUint64(&upstreamBytes, uint64(n))
	}
	return n, err
}

func (cc *countConn) Write(p []byte) (int, error) {
	n, err := cc.Conn.Write(p)
	if n > 0 {
		atomic.AddUint64(&downstreamBytes, uint64(n))
	}
	return n, err
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
