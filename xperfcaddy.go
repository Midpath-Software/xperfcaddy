package xperfcaddy

import (
	"fmt"
	"net/http"
	"time"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

type responseWrapper struct {
	http.ResponseWriter
	start    time.Time
	written  bool
	logger   *zap.Logger
}

func (rw *responseWrapper) WriteHeader(status int) {
	if !rw.written {
		duration := time.Since(rw.start)
		rw.Header().Set("X-Perf-Caddy", fmt.Sprintf("%v", duration))
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWrapper) Write(b []byte) (int, error) {
	if !rw.written {
		duration := time.Since(rw.start)
		rw.Header().Set("X-Perf-Caddy", fmt.Sprintf("%v", duration))
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("xperfcaddy", parseCaddyfile)
}

// Middleware implements an HTTP handler that adds X-Perf-Caddy header
// with request timing information.
type Middleware struct {
	logger *zap.Logger
}

// Provision implements caddy.Provisioner.
func (m *Middleware) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger()
	return nil
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.xperfcaddy",
		New: func() caddy.Module { return new(Middleware) },
	}
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	
	wrapper := &responseWrapper{
		ResponseWriter: w,
		start:         time.Now(),
		logger:        m.logger,
	}

	err := next.ServeHTTP(wrapper, r)

	if !wrapper.written {
		duration := time.Since(wrapper.start)
		w.Header().Set("X-Perf-Caddy", fmt.Sprintf("%v", duration))
	}

	return err
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler. Syntax:
//
//	xperfcaddy
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name
	if d.NextArg() {
		return d.ArgErr()
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
