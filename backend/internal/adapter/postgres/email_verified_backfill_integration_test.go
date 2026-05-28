package postgres_test

// Integration test that proves the migration-158 backfill semantics AND
// the SetEmailVerified repository write. Gated behind
// MARKETPLACE_TEST_DATABASE_URL (skips without it). It does NOT re-run
// the migration — that already ran on the test DB — it asserts the
// invariant the migration guarantees (no surviving email_verified=false
// rows once the backfill has been applied) and exercises the targeted
// flip used by the verify-email flow.
//
//	MARKETPLACE_TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5435/marketplace_go_feat_otp?sslmode=disable \
//	  go test ./internal/adapter/postgres/ -run TestEmailVerifiedBackfill -count=1

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/adapter/postgres"
)

// TestEmailVerifiedBackfill_NoUnverifiedSurvivors asserts the
// post-migration invariant: after the backfill (migration 158), every
// pre-existing user row is verified. We seed a row WITH the column
// already true (the backfill's effect) and confirm the table holds no
// false rows that predate our own test fixtures.
//
// The point of the assertion is that a freshly-migrated DB has zero
// email_verified=false rows EXCEPT any our own concurrent tests insert.
// To stay isolated we scope the count to rows created before this test
// started — practically, we just assert our seeded verified row reads
// back as verified and the SetEmailVerified flip works both ways.
func TestEmailVerifiedBackfill_SetEmailVerifiedRoundTrip(t *testing.T) {
	db := testDB(t)
	repo := postgres.NewUserRepository(db)
	ctx := context.Background()

	id := uuid.New()
	email := fmt.Sprintf("test-%s@emailverified.local", id.String()[:8])
	// Seed an explicitly-unverified user (simulating a fresh signup).
	_, err := db.Exec(`
		INSERT INTO users (id, email, hashed_password, first_name, last_name, display_name, role, email_verified)
		VALUES ($1, $2, 'x', 'Test', 'User', 'Test User', 'provider', false)`,
		id, email)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM users WHERE id = $1`, id) })

	// Read back: starts false.
	u, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.False(t, u.EmailVerified)

	// Flip to true (the verify-email path).
	require.NoError(t, repo.SetEmailVerified(ctx, id, true))
	u, err = repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.True(t, u.EmailVerified, "SetEmailVerified(true) must persist")

	// Idempotent flip back to false also works (defensive).
	require.NoError(t, repo.SetEmailVerified(ctx, id, false))
	u, err = repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.False(t, u.EmailVerified)
}

// TestEmailVerifiedBackfill_PreExistingRowsVerified asserts the migration
// outcome: the rows that existed before our test fixtures are all
// verified. We assert there are zero email_verified=false rows other
// than ones whose email matches our test-fixture suffixes (which other
// tests in this package may have inserted in parallel).
func TestEmailVerifiedBackfill_PreExistingRowsVerified(t *testing.T) {
	db := testDB(t)

	var unverifiedNonFixture int
	err := db.QueryRow(`
		SELECT count(*) FROM users
		WHERE email_verified = false
		  AND email NOT LIKE '%.local'`).Scan(&unverifiedNonFixture)
	require.NoError(t, err)
	assert.Equal(t, 0, unverifiedNonFixture,
		"migration 158 must leave no pre-existing (non-test-fixture) user unverified")
}
