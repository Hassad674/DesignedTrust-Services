package observability

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

// Sentry error monitoring for the Go API.
//
// Zero-overhead, fully env-gated default
//
//	When SENTRY_DSN is empty (the default in dev / CI, and in production
//	until the operator sets the DSN on Railway) InitSentry does NOTHING:
//	no client is initialised, no goroutine is started, no network is
//	dialled, and SentryEnabled reports false so the router never installs
//	the sentryhttp middleware. The API behaves byte-for-byte as it did
//	before this package existed. This lets the integration merge + deploy
//	safely ahead of the DSN being provisioned — it activates only once
//	SENTRY_DSN exists in the environment.
//
// Environment variables
//
//	SENTRY_DSN           — the project DSN. Empty disables Sentry entirely.
//	SENTRY_ENVIRONMENT   — overrides the environment tag (falls back to
//	                       ENV, then RAILWAY_ENVIRONMENT, then "production").
//	SENTRY_RELEASE       — optional release identifier (git sha / tag).
//
// PII safety (RGPD)
//
//	SendDefaultPII is hard-set to false so the SDK never attaches cookies,
//	the client IP, or sensitive headers (Authorization, Cookie, …). The
//	beforeSend hook below is a defense-in-depth second pass: it strips the
//	Authorization / Cookie headers, clears the query string (it can carry
//	tokens), and blanks the user identity (email / ip / username) on every
//	outbound event — even if a future code path flips SendDefaultPII or
//	builds an event by hand.
//
// Tracing is disabled (EnableTracing=false): distributed tracing is the
// OTel pipeline's job (see otel.go). Sentry here is strictly error +
// panic capture so the two observability surfaces do not double-bill.

// sentryFlushTimeout bounds the buffered-event flush during graceful
// shutdown. Matches the otel flush budget (2s) wired in wire_serve.go.
const sentryFlushTimeout = 2 * time.Second

// SentryConfig captures the boot-time inputs for the Sentry client.
// The zero value (empty DSN) selects the inert path.
type SentryConfig struct {
	// DSN is the Sentry project DSN. Empty selects the inert path —
	// InitSentry returns enabled=false and does not touch the SDK.
	DSN string

	// Environment is stamped on every event (e.g. "production").
	Environment string

	// Release is an optional release identifier (git sha / tag).
	Release string
}

// LoadSentryFromEnv reads the SENTRY_* environment variables. The
// environment falls back, in order, to ENV then RAILWAY_ENVIRONMENT
// then "production" so a Railway deployment that only sets SENTRY_DSN
// still tags events with a sensible environment.
func LoadSentryFromEnv() SentryConfig {
	env := firstNonEmpty(
		os.Getenv("SENTRY_ENVIRONMENT"),
		os.Getenv("ENV"),
		os.Getenv("RAILWAY_ENVIRONMENT"),
		"production",
	)
	return SentryConfig{
		DSN:         strings.TrimSpace(os.Getenv("SENTRY_DSN")),
		Environment: env,
		Release:     strings.TrimSpace(os.Getenv("SENTRY_RELEASE")),
	}
}

// SentryFlushFunc flushes buffered events to Sentry within the given
// timeout. Always non-nil — callers may defer it unconditionally; the
// inert path returns a no-op closure.
type SentryFlushFunc func(timeout time.Duration)

// noopSentryFlush is returned when Sentry is disabled. Always safe to
// call so the graceful-shutdown path can defer it unconditionally.
func noopSentryFlush(time.Duration) {}

// InitSentry initialises the global Sentry client when cfg.DSN is set.
//
// Returns (enabled, flush). When cfg.DSN is empty the function does
// nothing and returns (false, noop) — zero overhead, the inert default.
// When the DSN is set it calls sentry.Init and returns (true, flush)
// where flush drains buffered events during graceful shutdown.
//
// An sentry.Init error is NEVER fatal: it is logged at WARN and the
// function returns (false, noop) so a malformed DSN cannot block the
// API from booting.
func InitSentry(cfg SentryConfig) (bool, SentryFlushFunc) {
	if strings.TrimSpace(cfg.DSN) == "" {
		// Inert path — no init, no middleware, no overhead.
		return false, noopSentryFlush
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.DSN,
		Environment: cfg.Environment,
		Release:     cfg.Release,
		// Tracing is the OTel pipeline's responsibility (otel.go).
		EnableTracing: false,
		// RGPD: never attach cookies, client IP, or sensitive headers.
		SendDefaultPII: false,
		// Defense-in-depth PII scrub on every outbound event.
		BeforeSend: scrubEvent,
	})
	if err != nil {
		// Init failure must never block boot — log and continue inert.
		slog.Warn("sentry init failed, continuing without error monitoring", "error", err)
		return false, noopSentryFlush
	}

	slog.Info("sentry enabled", "environment", cfg.Environment)

	return true, func(timeout time.Duration) {
		if timeout <= 0 {
			timeout = sentryFlushTimeout
		}
		sentry.Flush(timeout)
	}
}

// SentryHTTPMiddleware returns the sentryhttp middleware that captures
// handler panics as Sentry events.
//
// Repanic is true so the existing middleware.Recovery handler still runs
// after Sentry records the panic — the panic is re-raised and bubbles up
// to Recovery, which logs it and writes the project's standard 500
// envelope. Without Repanic, sentryhttp would swallow the panic and the
// client would get an empty/garbage response.
//
// Only mount this when InitSentry returned enabled=true. Mounting it
// without an initialised client would wrap every request in a handler
// whose hub has no transport — harmless but pointless overhead.
func SentryHTTPMiddleware() func(http.Handler) http.Handler {
	handler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})
	return handler.Handle
}

// scrubEvent is the BeforeSend hook. It is a defense-in-depth pass on
// top of SendDefaultPII=false: it removes the Authorization / Cookie
// request headers, clears the query string (it can embed tokens), and
// blanks the user identity (email / ip / username) so no PII reaches
// Sentry even if an event was assembled outside the SDK's request path.
// Returning the (mutated) event keeps the panic/error report; returning
// nil would drop the event entirely.
func scrubEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	if event == nil {
		return nil
	}

	if event.Request != nil {
		event.Request.Cookies = ""
		event.Request.QueryString = ""
		if event.Request.Headers != nil {
			for k := range event.Request.Headers {
				if isSensitiveRequestHeader(k) {
					delete(event.Request.Headers, k)
				}
			}
		}
	}

	// Blank obvious user PII. Keeping the user object empty still lets
	// Sentry group the event; it just won't carry an email / IP / name.
	event.User = sentry.User{}

	return event
}

// isSensitiveRequestHeader reports whether a request header name must be
// scrubbed from a Sentry event. Case-insensitive — HTTP header keys are
// canonicalised inconsistently across the stack.
func isSensitiveRequestHeader(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "authorization", "cookie", "set-cookie", "proxy-authorization", "x-api-key":
		return true
	default:
		return false
	}
}

// firstNonEmpty returns the first argument that is non-empty after
// trimming surrounding whitespace, or "" when all are empty.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
