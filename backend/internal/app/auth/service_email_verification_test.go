package auth

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/domain/twofactor"
	"marketplace-backend/internal/domain/user"
)

// mockTwoFactorGate is a controllable auth.TwoFactorGate. It records the
// purpose threaded into RequestChallenge / VerifyChallengeWithPurpose so
// the signup-OTP tests can assert the email_verification scoping without
// reaching into the twofactor service.
type mockTwoFactorGate struct {
	mu sync.Mutex

	enabled bool

	requestCalls       int
	lastRequestPurpose twofactor.Purpose
	requestErr         error

	verifyPurposeCalls   int
	lastVerifyPurpose    twofactor.Purpose
	verifyPurposeErr     error
	verifyChallengeErr   error
	verifyChallengeCalls int
}

func (m *mockTwoFactorGate) IsEnabledForUser(_ context.Context, _ uuid.UUID) (bool, error) {
	return m.enabled, nil
}

func (m *mockTwoFactorGate) RequestChallenge(_ context.Context, in TwoFactorChallengeRequest) (uuid.UUID, error) {
	m.mu.Lock()
	m.requestCalls++
	m.lastRequestPurpose = in.Purpose
	m.mu.Unlock()
	if m.requestErr != nil {
		return uuid.Nil, m.requestErr
	}
	return uuid.New(), nil
}

func (m *mockTwoFactorGate) VerifyChallenge(_ context.Context, _ uuid.UUID, _ string) error {
	m.mu.Lock()
	m.verifyChallengeCalls++
	m.mu.Unlock()
	return m.verifyChallengeErr
}

func (m *mockTwoFactorGate) VerifyChallengeWithPurpose(_ context.Context, _ uuid.UUID, _ string, purpose twofactor.Purpose) error {
	m.mu.Lock()
	m.verifyPurposeCalls++
	m.lastVerifyPurpose = purpose
	m.mu.Unlock()
	return m.verifyPurposeErr
}

// newEmailVerifyService wires an auth.Service with the given gate and a
// user repo seeded with one user. Returns the service, the gate, the
// repo, and the seeded user id.
func newEmailVerifyService(t *testing.T, gate TwoFactorGate, seed *user.User) (*Service, *mockUserRepo) {
	t.Helper()
	repo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*user.User, error) {
			if seed == nil {
				return nil, user.ErrUserNotFound
			}
			// Return a copy so the service mutating EmailVerified does not
			// retroactively change what a later GetByID call observes
			// unless the test wants it.
			cp := *seed
			return &cp, nil
		},
	}
	svc := NewServiceWithDeps(ServiceDeps{
		Users:  repo,
		Hasher: &mockHasher{},
		Tokens: &mockTokenService{},
		Email:  &mockEmailService{},
	})
	svc.SetTwoFactorGate(gate)
	return svc, repo
}

func seededUser(verified bool) *user.User {
	return &user.User{
		ID:            uuid.New(),
		Email:         "newuser@example.com",
		Role:          user.RoleProvider,
		Status:        user.StatusActive,
		EmailVerified: verified,
	}
}

// TestRegister_AutoSendsEmailVerificationOTP asserts that a successful
// registration fires exactly one email_verification challenge (purpose
// scoped) AND still returns tokens (contract unchanged).
func TestRegister_AutoSendsEmailVerificationOTP(t *testing.T) {
	gate := &mockTwoFactorGate{}
	repo := &mockUserRepo{
		existsByEmailFn: func(_ context.Context, _ string) (bool, error) { return false, nil },
		createFn:        func(_ context.Context, _ *user.User) error { return nil },
	}
	svc := NewServiceWithDeps(ServiceDeps{
		Users:  repo,
		Hasher: &mockHasher{},
		Tokens: &mockTokenService{},
		Email:  &mockEmailService{},
	})
	svc.SetTwoFactorGate(gate)

	out, err := svc.Register(context.Background(), RegisterInput{
		Email:       "newuser@example.com",
		Password:    "Sup3rSecret!1",
		FirstName:   "New",
		LastName:    "User",
		DisplayName: "New User",
		Role:        user.RoleProvider,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	// Tokens MUST still be returned — the register contract is unchanged.
	assert.NotEmpty(t, out.AccessToken)
	assert.NotEmpty(t, out.RefreshToken)
	assert.NotNil(t, out.User)
	// Exactly one challenge, scoped to email_verification.
	assert.Equal(t, 1, gate.requestCalls)
	assert.Equal(t, twofactor.PurposeEmailVerification, gate.lastRequestPurpose)
}

// TestRegister_OTPFailureDoesNotFailRegistration asserts the auto-send is
// best-effort: a challenge-issue error is swallowed and the registration
// still succeeds with tokens.
func TestRegister_OTPFailureDoesNotFailRegistration(t *testing.T) {
	gate := &mockTwoFactorGate{requestErr: errors.New("email backend down")}
	repo := &mockUserRepo{
		existsByEmailFn: func(_ context.Context, _ string) (bool, error) { return false, nil },
		createFn:        func(_ context.Context, _ *user.User) error { return nil },
	}
	svc := NewServiceWithDeps(ServiceDeps{
		Users:  repo,
		Hasher: &mockHasher{},
		Tokens: &mockTokenService{},
		Email:  &mockEmailService{},
	})
	svc.SetTwoFactorGate(gate)

	out, err := svc.Register(context.Background(), RegisterInput{
		Email: "newuser@example.com", Password: "Sup3rSecret!1",
		FirstName: "New", LastName: "User", DisplayName: "New User",
		Role: user.RoleProvider,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.NotEmpty(t, out.AccessToken)
}

// TestVerifyEmail_SetsVerifiedAndReissuesTokens asserts the happy path:
// a correct code flips email_verified, the verify is purpose-scoped, and
// a fresh token pair is returned.
func TestVerifyEmail_SetsVerifiedAndReissuesTokens(t *testing.T) {
	gate := &mockTwoFactorGate{}
	seed := seededUser(false)
	svc, repo := newEmailVerifyService(t, gate, seed)

	var verifiedSet bool
	repo.setEmailVerifiedFn = func(_ context.Context, _ uuid.UUID, verified bool) error {
		verifiedSet = verified
		return nil
	}

	out, err := svc.VerifyEmail(context.Background(), seed.ID, "123456", SessionFingerprint{})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.True(t, verifiedSet, "SetEmailVerified must be called with true")
	assert.Equal(t, twofactor.PurposeEmailVerification, gate.lastVerifyPurpose,
		"verify must be scoped to email_verification")
	assert.Equal(t, 1, gate.verifyPurposeCalls)
	// Fresh tokens returned, and the returned user is now verified.
	assert.NotEmpty(t, out.AccessToken)
	assert.NotEmpty(t, out.RefreshToken)
	require.NotNil(t, out.User)
	assert.True(t, out.User.EmailVerified)
}

// TestVerifyEmail_WrongCodePropagatesSentinel asserts a bad code returns
// the twofactor sentinel and does NOT flip the flag.
func TestVerifyEmail_WrongCodePropagatesSentinel(t *testing.T) {
	gate := &mockTwoFactorGate{verifyPurposeErr: twofactor.ErrCodeMismatch}
	seed := seededUser(false)
	svc, repo := newEmailVerifyService(t, gate, seed)

	var setCalled bool
	repo.setEmailVerifiedFn = func(_ context.Context, _ uuid.UUID, _ bool) error {
		setCalled = true
		return nil
	}

	_, err := svc.VerifyEmail(context.Background(), seed.ID, "000000", SessionFingerprint{})
	assert.ErrorIs(t, err, twofactor.ErrCodeMismatch)
	assert.False(t, setCalled, "the flag must NOT flip on a wrong code")
}

// TestVerifyEmail_AlreadyVerifiedIsIdempotent asserts an already-verified
// account re-issues tokens WITHOUT requiring a code (no verify call).
func TestVerifyEmail_AlreadyVerifiedIsIdempotent(t *testing.T) {
	gate := &mockTwoFactorGate{}
	seed := seededUser(true) // already verified
	svc, _ := newEmailVerifyService(t, gate, seed)

	out, err := svc.VerifyEmail(context.Background(), seed.ID, "", SessionFingerprint{})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.NotEmpty(t, out.AccessToken)
	assert.Equal(t, 0, gate.verifyPurposeCalls, "no code check for an already-verified account")
}

// TestResendVerification_SendsWhenUnverified asserts a fresh
// email_verification challenge is issued for an unverified user.
func TestResendVerification_SendsWhenUnverified(t *testing.T) {
	gate := &mockTwoFactorGate{}
	seed := seededUser(false)
	svc, _ := newEmailVerifyService(t, gate, seed)

	already, err := svc.ResendVerification(context.Background(), seed.ID, SessionFingerprint{})
	require.NoError(t, err)
	assert.False(t, already)
	assert.Equal(t, 1, gate.requestCalls)
	assert.Equal(t, twofactor.PurposeEmailVerification, gate.lastRequestPurpose)
}

// TestResendVerification_NoopWhenVerified asserts the call short-circuits
// to a no-op (no challenge) for an already-verified account.
func TestResendVerification_NoopWhenVerified(t *testing.T) {
	gate := &mockTwoFactorGate{}
	seed := seededUser(true)
	svc, _ := newEmailVerifyService(t, gate, seed)

	already, err := svc.ResendVerification(context.Background(), seed.ID, SessionFingerprint{})
	require.NoError(t, err)
	assert.True(t, already)
	assert.Equal(t, 0, gate.requestCalls, "no challenge issued for a verified account")
}
