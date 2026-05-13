package observability

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HTTPSpanNameFormatter produces the span name for an inbound request.
// The default implementation falls back to "<METHOD> <URL.Path>" for
// the simple span name when no chi route pattern is available.
type HTTPSpanNameFormatter func(operation string, r *http.Request) string

// HTTPMiddleware wraps the handler with otelhttp.NewHandler so every
// inbound request is captured as a span, with trace context
// propagation honoured both in (extracted from incoming W3C headers)
// and out (injected into the response). When OTel is disabled the
// global tracer is the no-op SDK so the wrap reduces to a thin pass-
// through with no allocations.
//
// Span attributes set by otelhttp:
//   - http.method, http.scheme, http.target, http.status_code
//   - net.peer.ip, user_agent.original (truncated)
//   - net.protocol.name, net.protocol.version
//
// Attributes deliberately NOT set (PII concerns):
//   - request body, query string values, cookies
//   - Authorization header (auth tokens)
//
// Use spanNameFn to compose a low-cardinality name from the chi route
// pattern when available. Empty operation falls back to the request
// method + path; production should pass spanNameFn = ChiRouteName so
// /api/v1/users/123 collapses to /api/v1/users/{id}.
func HTTPMiddleware(operation string, opts ...otelhttp.Option) func(http.Handler) http.Handler {
	if operation == "" {
		operation = "http.server"
	}
	// Always emit the standard W3C trace headers on the response so
	// downstream consumers (frontend, mobile, internal services) can
	// stitch their own spans into the same trace.
	allOpts := append([]otelhttp.Option{
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
		otelhttp.WithFilter(filterPathExclusions),
	}, opts...)
	return func(next http.Handler) http.Handler {
		instrumented := otelhttp.NewHandler(next, operation, allOpts...)
		// /api/v1/ws bypasses the otelhttp.NewHandler wrap entirely
		// (not just the span — the ResponseWriter wrap itself). The
		// otelhttp.WrapResponseWriter calls WriteHeader after the WS
		// hijack, writing garbage bytes onto the hijacked TCP socket.
		// Strict browser clients (Chromium/Firefox/Edge) reject the
		// upgrade with ERR_INVALID_HTTP_RESPONSE; Node `ws` tolerates
		// the garbage. otelhttp.WithFilter() alone is not sufficient
		// because it only skips the *span*, not the writer wrap.
		// See filterPathExclusions for the full repro narrative.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isHijackPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			instrumented.ServeHTTP(w, r)
		})
	}
}

// isHijackPath returns true for paths whose handler hijacks the
// underlying TCP connection (currently only the WebSocket endpoint).
// Hijacking paths must NOT be wrapped by otelhttp.NewHandler.
func isHijackPath(p string) bool {
	return p == "/api/v1/ws"
}

// filterPathExclusions skips spans for noisy infra endpoints that
// would otherwise dominate the trace volume without adding signal.
// Health probes fire every 5 seconds per pod; metrics scrapes every
// 15 seconds per Prometheus instance — none of those are user
// traffic and recording them costs disk + ingest budget.
//
// /api/v1/ws is excluded for a different, critical reason:
// otelhttp.NewHandler wraps the http.ResponseWriter with its own
// instrumentation wrapper. When the WebSocket handshake hijacks the
// underlying TCP connection (via http.Hijacker), the otelhttp
// wrapper's later WriteHeader call writes spurious bytes onto the
// hijacked socket. Permissive clients (Node `ws`) tolerate the
// garbage; strict browser clients (Chromium / Firefox / Edge)
// reject the upgrade with ERR_INVALID_HTTP_RESPONSE and the WS
// never opens — taking realtime messaging down with it. Skipping
// the wrap for /api/v1/ws keeps the hijack path clean. Repro
// confirmed by Playwright e2e/realtime-bug.spec.ts and 7371
// "WriteHeader on hijacked connection" warnings in production-like
// backend logs.
func filterPathExclusions(r *http.Request) bool {
	switch r.URL.Path {
	case "/health", "/ready", "/metrics", "/api/v1/ws":
		return false
	}
	return true
}

// HTTPClientTransport returns an http.RoundTripper that emits a span
// per outbound request, with W3C trace context injected so downstream
// services (Stripe, Resend, Typesense, OpenAI, …) can stitch their
// own spans onto the same trace. When OTel is disabled the wrap is a
// thin pass-through.
//
// Use it on every http.Client used to call external services:
//
//	httpClient := &http.Client{
//	    Timeout:   10 * time.Second,
//	    Transport: observability.HTTPClientTransport(http.DefaultTransport, "stripe"),
//	}
func HTTPClientTransport(base http.RoundTripper, peerName string) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	opts := []otelhttp.Option{
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
	}
	if peerName != "" {
		opts = append(opts, otelhttp.WithSpanOptions(
			trace.WithAttributes(attribute.String("peer.service", peerName)),
		))
	}
	return otelhttp.NewTransport(base, opts...)
}

// SpanFromContext returns the span associated with ctx, or a non-
// recording stub if none is set. Re-exported for test ergonomics.
var SpanFromContext = trace.SpanFromContext

// Propagator is exposed so callers building their own headers can
// reuse the same propagator without importing otel directly.
var Propagator = func() propagation.TextMapPropagator { return otel.GetTextMapPropagator() }
