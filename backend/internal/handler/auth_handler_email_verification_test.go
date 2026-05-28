package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/app/auth"
	"marketplace-backend/internal/domain/twofactor"
	"marketplace-backend/internal/domain/user"
	"marketplace-backend/internal/handler/middleware"
	"marketplace-backend/internal/port/service"
)

// stubEmailVerifyGate is a controllable auth.TwoFactorGate for the
// handler-level email-verification tests.
type stubEmailVerifyGate struct {
	mu                 sync.Mutex
	requestCalls       int
	lastRequestPurpose twofactor.Purpose
	verifyPurposeErr   error
	verifyPurposeCalls int
}

func (g *stubEmailVerifyGate) IsEnabledForUser(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (g *stubEmailVerifyGate) RequestChallenge(_ context.Context, in auth.TwoFactorChallengeRequest) (uuid.UUID, error) {
	g.mu.Lock()
	g.requestCalls++
	g.lastRequestPurpose = in.Purpose
	g.mu.Unlock()
	return uuid.New(), nil
}
func (g *stubEmailVerifyGate) VerifyChallenge(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (g *stubEmailVerifyGate) VerifyChallengeWithPurpose(_ context.Context, _ uuid.UUID, _ string, _ twofactor.Purpose) error {
	g.mu.Lock()
	g.verifyPurposeCalls++
	g.mu.Unlock()
	return g.verifyPurposeErr
}

// newEmailVerifyHandler wires an AuthHandler whose auth service has the
// given gate, and whose user repo returns the supplied user from GetByID.
func newEmailVerifyHandler(t *testing.T, gate auth.TwoFactorGate, u *user.User) (*AuthHandler, *mockUserRepo) {
	t.Helper()
	repo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*user.User, error) {
			cp := *u
			return &cp, nil
		},
	}
	authSvc := auth.NewServiceWithDeps(auth.ServiceDeps{
		Users:       repo,
		Hasher:      &mockHasher{},
		Tokens:      &mockTokenService{},
		Email:       &mockEmailService{},
		Sessions:    &mockSessionService{},
		FrontendURL: "https://example.com",
	})
	authSvc.SetTwoFactorGate(gate)
	h := NewAuthHandler(authSvc, nil, &mockSessionService{}, testCookieConfig())
	return h, repo
}

func ctxWithVerifyUser(userID uuid.UUID) context.Context {
	return context.WithValue(context.Background(), middleware.ContextKeyUserID, userID)
}

func TestVerifyEmailHandler_Success(t *testing.T) {
	uid := uuid.New()
	gate := &stubEmailVerifyGate{}
	u := &user.User{ID: uid, Email: "n@example.com", Role: user.RoleProvider, Status: user.StatusActive, EmailVerified: false}
	h, _ := newEmailVerifyHandler(t, gate, u)

	body, _ := json.Marshal(map[string]string{"code": "123456"})
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewReader(body)).WithContext(ctxWithVerifyUser(uid))
	req.Header.Set("X-Auth-Mode", "token")
	rec := httptest.NewRecorder()

	h.VerifyEmail(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	// The purpose-scoped verify ran (proves email_verification scoping).
	assert.Equal(t, 1, gate.verifyPurposeCalls)
	// Token-mode response carries fresh tokens.
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["access_token"])
}

func TestVerifyEmailHandler_WrongCode(t *testing.T) {
	uid := uuid.New()
	gate := &stubEmailVerifyGate{verifyPurposeErr: twofactor.ErrCodeMismatch}
	u := &user.User{ID: uid, Email: "n@example.com", Role: user.RoleProvider, Status: user.StatusActive, EmailVerified: false}
	h, _ := newEmailVerifyHandler(t, gate, u)

	body, _ := json.Marshal(map[string]string{"code": "000000"})
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewReader(body)).WithContext(ctxWithVerifyUser(uid))
	rec := httptest.NewRecorder()

	h.VerifyEmail(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid_code")
}

func TestVerifyEmailHandler_MissingCode(t *testing.T) {
	uid := uuid.New()
	u := &user.User{ID: uid, Email: "n@example.com", Role: user.RoleProvider, Status: user.StatusActive}
	h, _ := newEmailVerifyHandler(t, &stubEmailVerifyGate{}, u)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewReader(body)).WithContext(ctxWithVerifyUser(uid))
	rec := httptest.NewRecorder()

	h.VerifyEmail(rec, req)
	// ValidateRequired surfaces a 422 Unprocessable Entity (the shared
	// validation-error envelope), not a 400 — the code field is missing.
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestResendVerificationHandler_SendsForUnverified(t *testing.T) {
	uid := uuid.New()
	gate := &stubEmailVerifyGate{}
	u := &user.User{ID: uid, Email: "n@example.com", Role: user.RoleProvider, Status: user.StatusActive, EmailVerified: false}
	h, _ := newEmailVerifyHandler(t, gate, u)

	req := httptest.NewRequest(http.MethodPost, "/auth/resend-verification", nil).WithContext(ctxWithVerifyUser(uid))
	rec := httptest.NewRecorder()

	h.ResendVerification(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, 1, gate.requestCalls)
	assert.Equal(t, twofactor.PurposeEmailVerification, gate.lastRequestPurpose)
}

func TestResendVerificationHandler_NoopForVerified(t *testing.T) {
	uid := uuid.New()
	gate := &stubEmailVerifyGate{}
	u := &user.User{ID: uid, Email: "n@example.com", Role: user.RoleProvider, Status: user.StatusActive, EmailVerified: true}
	h, _ := newEmailVerifyHandler(t, gate, u)

	req := httptest.NewRequest(http.MethodPost, "/auth/resend-verification", nil).WithContext(ctxWithVerifyUser(uid))
	rec := httptest.NewRecorder()

	h.ResendVerification(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 0, gate.requestCalls)
	assert.Contains(t, rec.Body.String(), "already verified")
}

// TestEmailVerifiedGate_EndToEnd wires the REAL auth middleware + the
// RequireEmailVerified gate behind a chi router and drives a bearer
// token whose `ev` claim controls access. Proves the full chain:
// JWT claim → auth-context stamp → gate decision.
func TestEmailVerifiedGate_EndToEnd(t *testing.T) {
	uid := uuid.New()

	// Token service that decodes a bearer string of the form
	// "verified" / "unverified" into a claims struct with the matching
	// email_verified value. SessionVersion 0 matches the nil checker.
	tokens := &mockTokenService{
		validateAccessFn: func(token string) (*service.TokenClaims, error) {
			switch token {
			case "verified":
				return &service.TokenClaims{UserID: uid, Role: "provider", EmailVerified: true}, nil
			case "unverified":
				return &service.TokenClaims{UserID: uid, Role: "provider", EmailVerified: false}, nil
			}
			return nil, user.ErrUnauthorized
		},
	}

	authMW := middleware.AuthFromDeps(middleware.AuthDeps{
		TokenService:   tokens,
		SessionService: &mockSessionService{}, // cookie path no-ops (Get errors)
	})
	gated := func(next http.Handler) http.Handler {
		return authMW(middleware.RequireEmailVerified(next))
	}

	r := chi.NewRouter()
	r.With(authMW).Get("/me", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })         // allowlist (bare auth)
	r.With(gated).Get("/feature", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })     // gated

	call := func(path, token string) int {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec.Code
	}

	// Gated route: verified passes, unverified is 403.
	assert.Equal(t, http.StatusOK, call("/feature", "verified"))
	assert.Equal(t, http.StatusForbidden, call("/feature", "unverified"))
	// Allowlisted route (/me): reachable even while unverified.
	assert.Equal(t, http.StatusOK, call("/me", "unverified"))
}
