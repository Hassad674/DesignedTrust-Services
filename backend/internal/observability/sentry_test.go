package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
)

// resetSentryHub unbinds the global Sentry client so an enabled-path
// test does not leak its bound client into sibling tests. The Sentry
// SDK keeps a single process-global hub; tests that call InitSentry
// with a real DSN must register this cleanup.
func resetSentryHub(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		sentry.CurrentHub().BindClient(nil)
	})
}

// A syntactically-valid DSN. sentry.Init parses + binds a client but
// never dials the network at init time (the HTTP transport buffers and
// flushes asynchronously), so this is safe to use in a unit test with
// no outbound connectivity.
const fakeDSN = "https://publickey@o0.ingest.sentry.io/0"

func TestInitSentry_Disabled_WhenDSNEmpty(t *testing.T) {
	enabled, flush := InitSentry(SentryConfig{DSN: ""})

	if enabled {
		t.Fatal("InitSentry must report disabled when DSN is empty")
	}
	if flush == nil {
		t.Fatal("flush closure must never be nil — callers defer it unconditionally")
	}
	// The no-op flush must be safe to call (it is invoked in the
	// graceful-shutdown path regardless of enablement).
	flush(2 * time.Second)
	flush(0) // zero timeout must also be safe
}

func TestInitSentry_Disabled_WhenDSNWhitespace(t *testing.T) {
	enabled, flush := InitSentry(SentryConfig{DSN: "   \t  "})

	if enabled {
		t.Fatal("InitSentry must treat a whitespace-only DSN as empty (disabled)")
	}
	if flush == nil {
		t.Fatal("flush closure must never be nil")
	}
	flush(time.Second)
}

func TestInitSentry_Enabled_WhenDSNSet(t *testing.T) {
	resetSentryHub(t)

	enabled, flush := InitSentry(SentryConfig{
		DSN:         fakeDSN,
		Environment: "production",
		Release:     "test-release",
	})

	if !enabled {
		t.Fatal("InitSentry must report enabled when a valid DSN is set")
	}
	if flush == nil {
		t.Fatal("flush closure must never be nil")
	}
	// The real flush is bounded by the timeout; with no events buffered
	// it returns quickly. We must not block the test suite on it.
	flush(time.Second)

	// The global hub must now carry a bound client.
	if sentry.CurrentHub().Client() == nil {
		t.Fatal("expected a bound Sentry client after enabled InitSentry")
	}
}

func TestInitSentry_Disabled_WhenDSNMalformed(t *testing.T) {
	resetSentryHub(t)

	// A malformed DSN makes sentry.Init return an error. InitSentry must
	// swallow it (log WARN) and report disabled — a bad DSN must NEVER
	// block the API from booting.
	enabled, flush := InitSentry(SentryConfig{DSN: "not-a-valid-dsn"})

	if enabled {
		t.Fatal("InitSentry must report disabled when sentry.Init fails on a malformed DSN")
	}
	if flush == nil {
		t.Fatal("flush closure must never be nil even on init failure")
	}
	flush(time.Second)
}

func TestLoadSentryFromEnv_EnvironmentPrecedence(t *testing.T) {
	tests := []struct {
		name       string
		sentryEnv  string
		env        string
		railwayEnv string
		want       string
	}{
		{
			name:      "SENTRY_ENVIRONMENT wins over all",
			sentryEnv: "sentry-env",
			env:       "app-env",
			want:      "sentry-env",
		},
		{
			name: "ENV used when SENTRY_ENVIRONMENT empty",
			env:  "staging",
			want: "staging",
		},
		{
			name:       "RAILWAY_ENVIRONMENT used when SENTRY_ENVIRONMENT + ENV empty",
			railwayEnv: "railway-prod",
			want:       "railway-prod",
		},
		{
			name: "defaults to production when all empty",
			want: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Setenv handles save + restore per subtest.
			t.Setenv("SENTRY_ENVIRONMENT", tt.sentryEnv)
			t.Setenv("ENV", tt.env)
			t.Setenv("RAILWAY_ENVIRONMENT", tt.railwayEnv)
			t.Setenv("SENTRY_DSN", "")

			cfg := LoadSentryFromEnv()
			if cfg.Environment != tt.want {
				t.Fatalf("Environment = %q, want %q", cfg.Environment, tt.want)
			}
		})
	}
}

func TestLoadSentryFromEnv_ReadsDSNAndRelease(t *testing.T) {
	t.Setenv("SENTRY_DSN", "  "+fakeDSN+"  ")
	t.Setenv("SENTRY_RELEASE", "  v1.2.3  ")
	t.Setenv("SENTRY_ENVIRONMENT", "")
	t.Setenv("ENV", "")
	t.Setenv("RAILWAY_ENVIRONMENT", "")

	cfg := LoadSentryFromEnv()

	if cfg.DSN != fakeDSN {
		t.Fatalf("DSN = %q, want trimmed %q", cfg.DSN, fakeDSN)
	}
	if cfg.Release != "v1.2.3" {
		t.Fatalf("Release = %q, want trimmed %q", cfg.Release, "v1.2.3")
	}
}

func TestScrubEvent_StripsPII(t *testing.T) {
	event := sentry.NewEvent()
	event.Request = &sentry.Request{
		URL:         "https://api.example.com/api/v1/auth/login",
		Method:      "POST",
		QueryString: "token=secret-jwt&page=2",
		Cookies:     "session=abc123",
		Headers: map[string]string{
			"Authorization": "Bearer secret-token",
			"Cookie":        "session=abc123",
			"Content-Type":  "application/json",
			"User-Agent":    "test-agent",
		},
	}
	event.User = sentry.User{
		ID:        "user-123",
		Email:     "victim@example.com",
		IPAddress: "203.0.113.7",
		Username:  "victim",
	}

	got := scrubEvent(event, nil)

	if got == nil {
		t.Fatal("scrubEvent must return the (scrubbed) event, not nil — dropping the event loses the error report")
	}
	if got.Request.QueryString != "" {
		t.Errorf("QueryString not cleared: %q", got.Request.QueryString)
	}
	if got.Request.Cookies != "" {
		t.Errorf("Cookies not cleared: %q", got.Request.Cookies)
	}
	if _, ok := got.Request.Headers["Authorization"]; ok {
		t.Error("Authorization header must be scrubbed")
	}
	if _, ok := got.Request.Headers["Cookie"]; ok {
		t.Error("Cookie header must be scrubbed")
	}
	// Non-sensitive headers must survive so the event keeps debugging value.
	if got.Request.Headers["Content-Type"] != "application/json" {
		t.Error("Content-Type header must be preserved")
	}
	if got.Request.Headers["User-Agent"] != "test-agent" {
		t.Error("User-Agent header must be preserved")
	}
	// User identity must be fully blanked (RGPD).
	if got.User.ID != "" || got.User.Email != "" || got.User.IPAddress != "" || got.User.Username != "" {
		t.Errorf("user PII not blanked: %+v", got.User)
	}
}

func TestScrubEvent_NilEvent(t *testing.T) {
	if got := scrubEvent(nil, nil); got != nil {
		t.Fatal("scrubEvent(nil) must return nil")
	}
}

func TestScrubEvent_NilRequest(t *testing.T) {
	event := sentry.NewEvent()
	event.Request = nil
	event.User = sentry.User{Email: "x@example.com"}

	got := scrubEvent(event, nil)
	if got == nil {
		t.Fatal("scrubEvent must not drop an event that has no Request")
	}
	if got.User.Email != "" {
		t.Error("user PII must still be blanked when Request is nil")
	}
}

func TestIsSensitiveRequestHeader(t *testing.T) {
	tests := []struct {
		header string
		want   bool
	}{
		{"Authorization", true},
		{"authorization", true},
		{"  Authorization  ", true},
		{"Cookie", true},
		{"Set-Cookie", true},
		{"Proxy-Authorization", true},
		{"X-API-Key", true},
		{"x-api-key", true},
		{"Content-Type", false},
		{"User-Agent", false},
		{"Accept", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			if got := isSensitiveRequestHeader(tt.header); got != tt.want {
				t.Fatalf("isSensitiveRequestHeader(%q) = %v, want %v", tt.header, got, tt.want)
			}
		})
	}
}

// TestSentryHTTPMiddleware_RepanicsSoRecoveryStillRuns proves the
// middleware is configured with Repanic=true: a handler panic must be
// re-raised after Sentry observes it so the project's existing Recovery
// middleware (registered as the outer handler) can still produce the
// 500 envelope. We simulate "outer Recovery" with a deferred recover in
// the test and assert the panic reaches it.
func TestSentryHTTPMiddleware_RepanicsSoRecoveryStillRuns(t *testing.T) {
	resetSentryHub(t)
	// A bound client is required for the sentryhttp hub to be live; use
	// the inert no-network fake DSN.
	if enabled, _ := InitSentry(SentryConfig{DSN: fakeDSN}); !enabled {
		t.Fatal("precondition: Sentry must be enabled for this test")
	}

	mw := SentryHTTPMiddleware()
	if mw == nil {
		t.Fatal("SentryHTTPMiddleware must return a non-nil middleware")
	}

	panicky := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})
	wrapped := mw(panicky)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)

	repanicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				repanicked = true
			}
		}()
		wrapped.ServeHTTP(rec, req)
	}()

	if !repanicked {
		t.Fatal("sentryhttp must re-panic (Repanic=true) so the outer Recovery middleware still runs")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   string
	}{
		{"first non-empty", []string{"a", "b"}, "a"},
		{"skips empty", []string{"", "b"}, "b"},
		{"skips whitespace", []string{"  ", "\t", "c"}, "c"},
		{"all empty returns empty", []string{"", "  "}, ""},
		{"no args returns empty", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstNonEmpty(tt.values...); got != tt.want {
				t.Fatalf("firstNonEmpty(%v) = %q, want %q", tt.values, got, tt.want)
			}
		})
	}
}
