package postgres_test

// Integration tests for the migration-157 purpose scoping on the 2FA
// challenge repository. Gated behind MARKETPLACE_TEST_DATABASE_URL (same
// convention as the rest of the postgres suite) so a fresh checkout
// without Docker simply skips.
//
//	MARKETPLACE_TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5435/marketplace_go_feat_otp?sslmode=disable \
//	  go test ./internal/adapter/postgres/ -run TestTwoFactorChallengePurpose -count=1
//
// The critical security property under test: a pending challenge of one
// purpose must NEVER be returned by a FindLatestPendingForUser call for
// the other purpose. Each test creates its own user and cleans it up.

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/adapter/postgres"
	"marketplace-backend/internal/domain/twofactor"
	"marketplace-backend/internal/port/repository"
)

// insertVerifyUser creates a minimal user row whose id can anchor the
// challenge FK, and cleans it up (cascading the challenges) afterwards.
func insertVerifyUser(t *testing.T) (uuid.UUID, *postgres.TwoFactorChallengeRepository) {
	t.Helper()
	db := testDB(t)
	id := uuid.New()
	email := fmt.Sprintf("test-%s@2fapurpose.local", id.String()[:8])
	_, err := db.Exec(`
		INSERT INTO users (id, email, hashed_password, first_name, last_name, display_name, role)
		VALUES ($1, $2, 'x', 'Test', 'User', 'Test User', 'provider')`,
		id, email)
	require.NoError(t, err, "insert test user")
	t.Cleanup(func() {
		// two_factor_challenges cascades on the user FK.
		_, _ = db.Exec(`DELETE FROM users WHERE id = $1`, id)
	})
	return id, postgres.NewTwoFactorChallengeRepository(db)
}

func mustCreateChallenge(t *testing.T, repo *postgres.TwoFactorChallengeRepository, uid uuid.UUID, purpose twofactor.Purpose) *twofactor.Challenge {
	t.Helper()
	c, err := twofactor.New(twofactor.NewChallengeInput{
		UserID:   uid,
		CodeHash: "hashed-code",
		Purpose:  purpose,
	})
	require.NoError(t, err)
	require.NoError(t, repo.Create(context.Background(), c))
	return c
}

// TestTwoFactorChallengePurpose_Isolation is the core security assertion:
// an email_verification challenge cannot be found via a login_2fa lookup
// and vice-versa, even for the same user with both pending at once.
func TestTwoFactorChallengePurpose_Isolation(t *testing.T) {
	uid, repo := insertVerifyUser(t)
	ctx := context.Background()

	verifyChal := mustCreateChallenge(t, repo, uid, twofactor.PurposeEmailVerification)
	loginChal := mustCreateChallenge(t, repo, uid, twofactor.PurposeLogin2FA)

	// Lookup scoped to email_verification returns the verification row.
	gotVerify, err := repo.FindLatestPendingForUser(ctx, uid, twofactor.PurposeEmailVerification)
	require.NoError(t, err)
	assert.Equal(t, verifyChal.ID, gotVerify.ID)
	assert.Equal(t, twofactor.PurposeEmailVerification, gotVerify.Purpose)

	// Lookup scoped to login_2fa returns the login row — NOT the
	// verification row, proving the flows are isolated.
	gotLogin, err := repo.FindLatestPendingForUser(ctx, uid, twofactor.PurposeLogin2FA)
	require.NoError(t, err)
	assert.Equal(t, loginChal.ID, gotLogin.ID)
	assert.Equal(t, twofactor.PurposeLogin2FA, gotLogin.Purpose)
}

// TestTwoFactorChallengePurpose_NoCrossPurposeMatch asserts that when a
// user only has an email_verification challenge, a login_2fa lookup finds
// nothing (and vice-versa).
func TestTwoFactorChallengePurpose_NoCrossPurposeMatch(t *testing.T) {
	uid, repo := insertVerifyUser(t)
	ctx := context.Background()

	mustCreateChallenge(t, repo, uid, twofactor.PurposeEmailVerification)

	_, err := repo.FindLatestPendingForUser(ctx, uid, twofactor.PurposeLogin2FA)
	assert.ErrorIs(t, err, repository.ErrTwoFactorChallengeNotFound,
		"a login_2fa lookup must not match a pending email_verification challenge")
}
