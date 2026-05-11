package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSanitizeMetadata_AllSensitiveKeysCovered locks down the full
// allowlist so a contributor who adds a new key to the map cannot
// silently drop one of the existing entries. Pinning the exact set
// keeps the privacy contract auditable in a single place.
func TestSanitizeMetadata_AllSensitiveKeysCovered(t *testing.T) {
	t.Parallel()

	wantKeys := map[string]struct{}{
		"email":      {},
		"to_email":   {},
		"from_email": {},
		"recipient":  {},
		"phone":      {},
		"iban":       {},
	}

	// Run a sample value through every documented sensitive key; the
	// output must be hashed (not equal to the input) for each.
	for key := range wantKeys {
		t.Run(key, func(t *testing.T) {
			got := SanitizeMetadata(map[string]any{key: "redact-me"})
			require.Contains(t, got, key)
			assert.NotEqual(t, "redact-me", got[key],
				"key %q must hash its value", key)
			assert.Len(t, got[key].(string), sanitizedHashLength,
				"hashed value must be exactly sanitizedHashLength chars")
		})
	}

	// Pin the constant — a refactor that shrinks the prefix would
	// reduce entropy and possibly cause cross-row collisions.
	assert.Equal(t, 16, sanitizedHashLength,
		"sanitizedHashLength documented at 16 hex chars (64 bits entropy)")
}

// TestSanitizeMetadata_NonSensitiveKeysWithMapValuesAreWalked asserts
// that a non-sensitive key whose value is a nested map still has its
// nested sensitive keys redacted. This is a regression: an earlier
// implementation only walked sensitive keys recursively, missing
// nested "reason" → map → "email" patterns produced by the auth
// failure path.
func TestSanitizeMetadata_NonSensitiveKeysWithMapValuesAreWalked(t *testing.T) {
	t.Parallel()

	in := map[string]any{
		"context": map[string]any{
			"email":  "deep@example.com",
			"reason": "ok",
		},
	}
	got := SanitizeMetadata(in)
	ctx, ok := got["context"].(map[string]any)
	require.True(t, ok, "nested map must be preserved")
	assert.NotEqual(t, "deep@example.com", ctx["email"],
		"sensitive key inside a non-sensitive parent must still be redacted")
	assert.Equal(t, "ok", ctx["reason"], "non-sensitive sibling must pass through")
}

// TestSanitizeMetadata_DoubleSanitizeIsStable asserts the function is
// safe to call twice in a row — the second pass must not mangle the
// already-hashed value (it gets re-hashed but stays a string with the
// same length). The contract: "no panic, output is always a 16-hex
// string for any sensitive key, even if the input is already a hash".
func TestSanitizeMetadata_DoubleSanitizeIsStable(t *testing.T) {
	t.Parallel()

	once := SanitizeMetadata(map[string]any{"email": "alice@example.com"})
	twice := SanitizeMetadata(once)

	require.Contains(t, twice, "email")
	got, ok := twice["email"].(string)
	require.True(t, ok)
	assert.Len(t, got, sanitizedHashLength,
		"double-sanitize must still produce a 16-char hash string")
}
