package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// okHandler is a sentinel next-handler that records whether the gate let
// the request through.
func okHandler(passed *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*passed = true
		w.WriteHeader(http.StatusOK)
	})
}

func reqWithEmailVerified(value bool, present bool) *http.Request {
	ctx := context.Background()
	if present {
		ctx = context.WithValue(ctx, ContextKeyEmailVerified, value)
	}
	return httptest.NewRequest(http.MethodGet, "/protected", nil).WithContext(ctx)
}

func TestRequireEmailVerified_AllowsVerified(t *testing.T) {
	var passed bool
	rec := httptest.NewRecorder()
	RequireEmailVerified(okHandler(&passed)).ServeHTTP(rec, reqWithEmailVerified(true, true))

	assert.True(t, passed, "a verified caller must pass the gate")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireEmailVerified_BlocksUnverified(t *testing.T) {
	var passed bool
	rec := httptest.NewRecorder()
	RequireEmailVerified(okHandler(&passed)).ServeHTTP(rec, reqWithEmailVerified(false, true))

	assert.False(t, passed, "an unverified caller must NOT reach the handler")
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "email_not_verified")
}

func TestRequireEmailVerified_RejectsUnauthenticated(t *testing.T) {
	// No stamp on the context at all → the gate was mounted without Auth
	// ahead of it, or an unauthenticated request slipped through. The
	// gate makes this loud with a 401 rather than silently allowing it.
	var passed bool
	rec := httptest.NewRecorder()
	RequireEmailVerified(okHandler(&passed)).ServeHTTP(rec, reqWithEmailVerified(false, false))

	assert.False(t, passed)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestGetEmailVerified covers the context getter's presence semantics.
func TestGetEmailVerified(t *testing.T) {
	t.Run("present true", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ContextKeyEmailVerified, true)
		v, ok := GetEmailVerified(ctx)
		assert.True(t, ok)
		assert.True(t, v)
	})
	t.Run("present false", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ContextKeyEmailVerified, false)
		v, ok := GetEmailVerified(ctx)
		assert.True(t, ok)
		assert.False(t, v)
	})
	t.Run("absent", func(t *testing.T) {
		v, ok := GetEmailVerified(context.Background())
		assert.False(t, ok)
		assert.False(t, v)
	})
}
