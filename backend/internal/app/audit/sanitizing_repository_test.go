package audit_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appaudit "marketplace-backend/internal/app/audit"
	"marketplace-backend/internal/domain/audit"
)

// fakeAuditRepo is a minimal stub of `repository.AuditRepository`
// that captures the last entry written so the sanitization
// integration test can assert against the sanitized payload that
// would actually hit the DB. Inlined here per the project's
// no-mocks-package convention (see backend/CLAUDE.md).
type fakeAuditRepo struct {
	logged   *audit.Entry
	logErr   error
	listEntries []*audit.Entry
	listCursor  string
	listErr     error
}

func (r *fakeAuditRepo) Log(_ context.Context, entry *audit.Entry) error {
	r.logged = entry
	return r.logErr
}

func (r *fakeAuditRepo) ListByResource(_ context.Context, _ audit.ResourceType, _ uuid.UUID, _ string, _ int) ([]*audit.Entry, string, error) {
	return r.listEntries, r.listCursor, r.listErr
}

func (r *fakeAuditRepo) ListByUser(_ context.Context, _ uuid.UUID, _ string, _ int) ([]*audit.Entry, string, error) {
	return r.listEntries, r.listCursor, r.listErr
}

func hashOf(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:16]
}

func TestSanitizingRepository_LogRedactsEmailBeforePersist(t *testing.T) {
	t.Parallel()

	inner := &fakeAuditRepo{}
	repo := appaudit.NewSanitizingRepository(inner)

	entry, err := audit.NewEntry(audit.NewEntryInput{
		Action:       audit.ActionLoginFailure,
		ResourceType: audit.ResourceTypeUser,
		Metadata: map[string]any{
			"email":  "victim@example.com",
			"reason": "invalid_password",
		},
	})
	require.NoError(t, err)

	require.NoError(t, repo.Log(context.Background(), entry))

	require.NotNil(t, inner.logged, "Log must forward to the inner repo")
	assert.Equal(t, hashOf("victim@example.com"), inner.logged.Metadata["email"], "email must be hashed before persist")
	assert.Equal(t, "invalid_password", inner.logged.Metadata["reason"], "non-sensitive metadata must be untouched")
}

func TestSanitizingRepository_LogPropagatesInnerError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("db down")
	inner := &fakeAuditRepo{logErr: wantErr}
	repo := appaudit.NewSanitizingRepository(inner)

	entry, err := audit.NewEntry(audit.NewEntryInput{
		Action:       audit.ActionLogout,
		ResourceType: audit.ResourceTypeUser,
		Metadata:     map[string]any{"email": "alice@example.com"},
	})
	require.NoError(t, err)

	gotErr := repo.Log(context.Background(), entry)
	assert.ErrorIs(t, gotErr, wantErr, "errors from the wrapped repo must surface unchanged")
}

func TestSanitizingRepository_LogNilEntryPropagates(t *testing.T) {
	t.Parallel()

	inner := &fakeAuditRepo{}
	repo := appaudit.NewSanitizingRepository(inner)

	// Defensive: the contract is "do not crash on nil" — the inner
	// repo will return nil-safely or fail loud, the wrapper must
	// not panic before that point.
	require.NotPanics(t, func() {
		_ = repo.Log(context.Background(), nil)
	})
}

func TestSanitizingRepository_LogPreservesNonSensitiveKeys(t *testing.T) {
	t.Parallel()

	inner := &fakeAuditRepo{}
	repo := appaudit.NewSanitizingRepository(inner)

	entry, err := audit.NewEntry(audit.NewEntryInput{
		Action:       audit.ActionMemberRoleChanged,
		ResourceType: audit.ResourceTypeMember,
		Metadata: map[string]any{
			"old_role": "member",
			"new_role": "admin",
		},
	})
	require.NoError(t, err)

	require.NoError(t, repo.Log(context.Background(), entry))

	assert.Equal(t, "member", inner.logged.Metadata["old_role"])
	assert.Equal(t, "admin", inner.logged.Metadata["new_role"])
}

func TestSanitizingRepository_ListByResourceForwardsVerbatim(t *testing.T) {
	t.Parallel()

	want := []*audit.Entry{{ID: uuid.New(), Action: audit.ActionLoginSuccess}}
	inner := &fakeAuditRepo{listEntries: want, listCursor: "next"}
	repo := appaudit.NewSanitizingRepository(inner)

	got, cursor, err := repo.ListByResource(context.Background(), audit.ResourceTypeUser, uuid.New(), "", 50)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, "next", cursor)
}

func TestSanitizingRepository_ListByUserForwardsVerbatim(t *testing.T) {
	t.Parallel()

	want := []*audit.Entry{{ID: uuid.New(), Action: audit.ActionLogout}}
	inner := &fakeAuditRepo{listEntries: want}
	repo := appaudit.NewSanitizingRepository(inner)

	got, _, err := repo.ListByUser(context.Background(), uuid.New(), "", 25)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
