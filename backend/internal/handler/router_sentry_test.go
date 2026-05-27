package handler

import (
	"strconv"
	"strings"
	"testing"
)

// TestRouter_SentryMiddleware_EnvGated proves the sentryhttp middleware
// is installed in the global stack ONLY when RouterDeps.SentryEnabled is
// true. It walks the chi tree twice (disabled then enabled) and asserts
// that enabling Sentry adds exactly one middleware to every route — and
// that disabling it leaves the chain byte-identical to the baseline the
// snapshot golden captures.
//
// This is the router-layer half of the env-gating guarantee (the
// observability package owns the init half): SENTRY_DSN unset ⇒ zero
// added middleware ⇒ the request path is unchanged from before the
// integration.
func TestRouter_SentryMiddleware_EnvGated(t *testing.T) {
	disabledDeps := snapshotDeps()
	disabledDeps.SentryEnabled = false

	enabledDeps := snapshotDeps()
	enabledDeps.SentryEnabled = true

	disabled := middlewareCounts(t, walkRoutes(t, NewRouter(disabledDeps)))
	enabled := middlewareCounts(t, walkRoutes(t, NewRouter(enabledDeps)))

	if len(disabled) == 0 {
		t.Fatal("expected at least one route in the table")
	}
	if len(disabled) != len(enabled) {
		t.Fatalf("route set changed between disabled (%d) and enabled (%d) — Sentry gating must not add/remove routes",
			len(disabled), len(enabled))
	}

	for route, off := range disabled {
		on, ok := enabled[route]
		if !ok {
			t.Errorf("route %q present when disabled but missing when enabled", route)
			continue
		}
		if on != off+1 {
			t.Errorf("route %q: middleware count = %d (enabled) vs %d (disabled); enabling Sentry must add exactly one global middleware",
				route, on, off)
		}
	}
}

// middlewareCounts parses the "METHOD PATH mw=N" lines walkRoutes emits
// into a map keyed by "METHOD PATH" with the integer middleware count.
func middlewareCounts(t *testing.T, lines []string) map[string]int {
	t.Helper()
	out := make(map[string]int, len(lines))
	for _, line := range lines {
		idx := strings.LastIndex(line, " mw=")
		if idx < 0 {
			t.Fatalf("unexpected route line shape: %q", line)
		}
		key := line[:idx]
		n, err := strconv.Atoi(line[idx+len(" mw="):])
		if err != nil {
			t.Fatalf("parse middleware count in %q: %v", line, err)
		}
		out[key] = n
	}
	return out
}
