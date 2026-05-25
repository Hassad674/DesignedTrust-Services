package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/port/service"
)

// captureIdentity is a terminal handler that records the user id the
// upstream middleware stamped (or its absence) so a test can assert
// whether OptionalAuth recognised the caller.
func captureIdentity(seen *uuid.UUID, anon *bool) http.Handler {
	return http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if uid, ok := GetUserID(r.Context()); ok {
			*seen = uid
			return
		}
		*anon = true
	})
}

func TestOptionalAuthFromDeps_NoCredential_ProceedsAnonymous(t *testing.T) {
	deps := AuthDeps{
		TokenService:   &mockTokenService{validateAccessFn: func(string) (*service.TokenClaims, error) { return nil, errors.New("nope") }},
		SessionService: &mockSessionService{getFn: func(context.Context, string) (*service.Session, error) { return nil, errors.New("nope") }},
	}
	var seen uuid.UUID
	var anon bool
	h := OptionalAuthFromDeps(deps)(captureIdentity(&seen, &anon))

	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.True(t, anon, "no credential must proceed anonymously")
	assert.Equal(t, http.StatusOK, w.Code, "optional auth must never reject")
}

func TestOptionalAuthFromDeps_ValidBearer_StampsIdentity(t *testing.T) {
	userID := uuid.New()
	deps := AuthDeps{
		TokenService: &mockTokenService{
			validateAccessFn: func(token string) (*service.TokenClaims, error) {
				assert.Equal(t, "good-token", token)
				return &service.TokenClaims{UserID: userID, Role: "agency", SessionVersion: 1}, nil
			},
		},
		SessionService:  &mockSessionService{getFn: func(context.Context, string) (*service.Session, error) { return nil, errors.New("no cookie") }},
		SessionVersions: &mockSessionVersionChecker{getFn: func(context.Context, uuid.UUID) (int, error) { return 1, nil }},
	}
	var seen uuid.UUID
	var anon bool
	h := OptionalAuthFromDeps(deps)(captureIdentity(&seen, &anon))

	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", nil)
	r.Header.Set("Authorization", "Bearer good-token")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	require.False(t, anon, "a valid bearer must stamp identity")
	assert.Equal(t, userID, seen)
}

func TestOptionalAuthFromDeps_RevokedSession_ProceedsAnonymous(t *testing.T) {
	userID := uuid.New()
	deps := AuthDeps{
		TokenService: &mockTokenService{
			validateAccessFn: func(string) (*service.TokenClaims, error) {
				// Token carries version 1, but the live version is 2 →
				// revoked. OptionalAuth must NOT reject — it falls through
				// to anonymous.
				return &service.TokenClaims{UserID: userID, Role: "agency", SessionVersion: 1}, nil
			},
		},
		SessionService:  &mockSessionService{getFn: func(context.Context, string) (*service.Session, error) { return nil, errors.New("no cookie") }},
		SessionVersions: &mockSessionVersionChecker{getFn: func(context.Context, uuid.UUID) (int, error) { return 2, nil }},
	}
	var seen uuid.UUID
	var anon bool
	h := OptionalAuthFromDeps(deps)(captureIdentity(&seen, &anon))

	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", nil)
	r.Header.Set("Authorization", "Bearer revoked-token")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.True(t, anon, "a revoked token must degrade to anonymous, never reject")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuthFromDeps_MalformedBearer_ProceedsAnonymous(t *testing.T) {
	deps := AuthDeps{
		TokenService:   &mockTokenService{validateAccessFn: func(string) (*service.TokenClaims, error) { return nil, errors.New("nope") }},
		SessionService: &mockSessionService{getFn: func(context.Context, string) (*service.Session, error) { return nil, errors.New("nope") }},
	}
	var seen uuid.UUID
	var anon bool
	h := OptionalAuthFromDeps(deps)(captureIdentity(&seen, &anon))

	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", nil)
	r.Header.Set("Authorization", "Basic Zm9vOmJhcg==") // not a Bearer
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	assert.True(t, anon)
}

func TestOptionalAuthFromDeps_ValidCookie_StampsIdentity(t *testing.T) {
	userID := uuid.New()
	deps := AuthDeps{
		TokenService: &mockTokenService{validateAccessFn: func(string) (*service.TokenClaims, error) { return nil, errors.New("no bearer") }},
		SessionService: &mockSessionService{
			getFn: func(_ context.Context, sessionID string) (*service.Session, error) {
				assert.Equal(t, "sess-123", sessionID)
				return &service.Session{UserID: userID, Role: "enterprise", SessionVersion: 3}, nil
			},
		},
		SessionVersions: &mockSessionVersionChecker{getFn: func(context.Context, uuid.UUID) (int, error) { return 3, nil }},
	}
	var seen uuid.UUID
	var anon bool
	h := OptionalAuthFromDeps(deps)(captureIdentity(&seen, &anon))

	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", nil)
	r.AddCookie(&http.Cookie{Name: "session_id", Value: "sess-123"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	require.False(t, anon, "a valid session cookie must stamp identity")
	assert.Equal(t, userID, seen)
}
