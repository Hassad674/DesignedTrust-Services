package middleware

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
	// hijacked is set to true once Hijack() succeeds. After hijack,
	// the underlying ResponseWriter no longer owns the TCP socket —
	// the WebSocket / SSE upgrader does. Any further WriteHeader /
	// Write call on the underlying writer would emit spurious bytes
	// onto the hijacked socket, which Chromium / Firefox / Edge
	// reject with ERR_INVALID_HTTP_RESPONSE during the WS handshake
	// (taking realtime messaging down with it). The Go HTTP server
	// also logs a noisy "http: response.WriteHeader on hijacked
	// connection" warning. We short-circuit both paths below.
	hijacked bool
}

func (w *wrappedResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	if w.hijacked {
		return
	}
	w.ResponseWriter.WriteHeader(code)
}

// Write mirrors WriteHeader's hijack guard so a stray Write call
// after a WS upgrade cannot corrupt the hijacked socket either.
func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	if w.hijacked {
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

// Hijack implements http.Hijacker so that WebSocket upgrades work
// through the logging middleware.
func (w *wrappedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}
	conn, brw, err := hj.Hijack()
	if err == nil {
		w.hijacked = true
	}
	return conn, brw, err
}

// Flush implements http.Flusher for streaming responses.
func (w *wrappedResponseWriter) Flush() {
	if fl, ok := w.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &wrappedResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", GetRequestID(r.Context()),
			"remote_addr", r.RemoteAddr,
		)
	})
}
