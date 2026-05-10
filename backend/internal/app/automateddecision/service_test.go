package automateddecision_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	app "marketplace-backend/internal/app/automateddecision"
	"marketplace-backend/internal/domain/automateddecision"
)

// fakeRepo is the inline mock for the AutomatedDecisionAppealRepository
// port. Tests inject the createFn closure to assert the persisted
// appeal or to inject an error path.
type fakeRepo struct {
	createFn func(ctx context.Context, appeal *automateddecision.Appeal) error
	created  *automateddecision.Appeal
}

func (f *fakeRepo) Create(ctx context.Context, appeal *automateddecision.Appeal) error {
	f.created = appeal
	if f.createFn != nil {
		return f.createFn(ctx, appeal)
	}
	return nil
}

func TestFileAppeal_Persists_AndNotifiesAdmin(t *testing.T) {
	repo := &fakeRepo{}
	email := newEmailFake()
	svc := app.NewService(app.ServiceDeps{
		Repo:       repo,
		Email:      email,
		AdminEmail: "rgpd@marketplace.test",
	})

	uid := uuid.New()
	appeal, err := svc.FileAppeal(context.Background(), app.FileAppealInput{
		UserID:       uid,
		DecisionType: "moderation",
		ReferenceID:  "moderation-result-id-123",
		Reason:       "Mon contenu est conforme — merci de revoir.",
	})

	require.NoError(t, err)
	require.NotNil(t, appeal)
	assert.Equal(t, automateddecision.DecisionMod, appeal.DecisionType)
	assert.Equal(t, automateddecision.StatusPending, appeal.Status)
	assert.Equal(t, uid, appeal.UserID)
	require.NotNil(t, repo.created)
	assert.Equal(t, appeal.ID, repo.created.ID)

	assert.Equal(t, 1, email.calls())
	assert.Equal(t, "rgpd@marketplace.test", email.lastTo())
	assert.Contains(t, email.lastSub(), "moderation")
	assert.Contains(t, email.lastBody(), appeal.ID.String())
	assert.Contains(t, email.lastBody(), "moderation-result-id-123")
}

func TestFileAppeal_InvalidDecisionType_DoesNotPersist(t *testing.T) {
	repo := &fakeRepo{}
	email := newEmailFake()
	svc := app.NewService(app.ServiceDeps{
		Repo:       repo,
		Email:      email,
		AdminEmail: "rgpd@marketplace.test",
	})

	_, err := svc.FileAppeal(context.Background(), app.FileAppealInput{
		UserID:       uuid.New(),
		DecisionType: "search-ranking", // not the canonical "ranking"
		ReferenceID:  "ref",
		Reason:       "reason",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, automateddecision.ErrInvalidDecisionType)
	assert.Nil(t, repo.created)
	assert.Equal(t, 0, email.calls())
}

func TestFileAppeal_PersistFailure_BubblesUp(t *testing.T) {
	boom := errors.New("db down")
	repo := &fakeRepo{createFn: func(ctx context.Context, _ *automateddecision.Appeal) error {
		return boom
	}}
	email := newEmailFake()
	svc := app.NewService(app.ServiceDeps{
		Repo:       repo,
		Email:      email,
		AdminEmail: "rgpd@marketplace.test",
	})

	_, err := svc.FileAppeal(context.Background(), app.FileAppealInput{
		UserID:       uuid.New(),
		DecisionType: "ranking",
		ReferenceID:  "trace",
		Reason:       "Pourquoi je n'apparais plus en recherche ?",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
	assert.Equal(t, 0, email.calls(), "email must not fire when persistence failed")
}

func TestFileAppeal_EmailFailure_DoesNotAbort(t *testing.T) {
	repo := &fakeRepo{}
	email := newEmailFake()
	email.f.failWith = errors.New("smtp down")
	svc := app.NewService(app.ServiceDeps{
		Repo:       repo,
		Email:      email,
		AdminEmail: "rgpd@marketplace.test",
	})

	appeal, err := svc.FileAppeal(context.Background(), app.FileAppealInput{
		UserID:       uuid.New(),
		DecisionType: "payment",
		ReferenceID:  "pi_xyz",
		Reason:       "Stripe a bloqué mon paiement à tort.",
	})

	require.NoError(t, err)
	require.NotNil(t, appeal)
	require.NotNil(t, repo.created)
	assert.Equal(t, 1, email.calls())
}

func TestFileAppeal_NoAdminEmail_SkipsNotification(t *testing.T) {
	repo := &fakeRepo{}
	email := newEmailFake()
	svc := app.NewService(app.ServiceDeps{
		Repo:       repo,
		Email:      email,
		AdminEmail: "", // not configured — production safe-skip
	})

	_, err := svc.FileAppeal(context.Background(), app.FileAppealInput{
		UserID:       uuid.New(),
		DecisionType: "moderation",
		ReferenceID:  "ref",
		Reason:       "reason",
	})

	require.NoError(t, err)
	require.NotNil(t, repo.created)
	assert.Equal(t, 0, email.calls())
}
