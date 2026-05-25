package feedback

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "marketplace-backend/internal/domain/feedback"
)

func baseReport(id uuid.UUID) *domain.Report {
	now := time.Now().UTC()
	return &domain.Report{
		ID:        id,
		Type:      domain.TypeBug,
		Title:     "Something broke",
		Status:    domain.StatusNew,
		Context:   map[string]any{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestService_GetReport_WithAttachmentsAndNotes(t *testing.T) {
	reportID := uuid.New()
	report := baseReport(reportID)
	attachment := &domain.Attachment{ID: uuid.New(), ReportID: reportID, Kind: domain.AttachmentImage, ObjectKey: "reports/x/y.png"}
	note := &domain.Note{ID: uuid.New(), ReportID: reportID, AdminUserID: uuid.New(), Body: "looking into it"}

	repo := &mockFeedbackRepo{
		getReportFn: func(_ context.Context, id uuid.UUID) (*domain.Report, error) {
			assert.Equal(t, reportID, id)
			return report, nil
		},
		listAttachmentsFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Attachment, error) {
			return []*domain.Attachment{attachment}, nil
		},
		listNotesFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Note, error) { return []*domain.Note{note}, nil },
	}
	var signedKey string
	storage := &mockStorage{
		presignDownloadFn: func(_ context.Context, key string, _ time.Duration) (string, error) {
			signedKey = key
			return "https://signed/" + key, nil
		},
	}
	svc := NewService(ServiceDeps{Reports: repo, Storage: storage, AnonymizationSalt: testSalt})

	detail, err := svc.GetReport(context.Background(), reportID)
	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Equal(t, report, detail.Report)
	require.Len(t, detail.Attachments, 1)
	assert.Equal(t, "reports/x/y.png", signedKey)
	assert.Equal(t, "https://signed/reports/x/y.png", detail.Attachments[0].PresignedURL)
	require.Len(t, detail.Notes, 1)
	assert.Equal(t, note, detail.Notes[0])
}

func TestService_GetReport_NotFound(t *testing.T) {
	repo := &mockFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return nil, domain.ErrNotFound },
	}
	svc := newServiceWithRepo(repo)
	detail, err := svc.GetReport(context.Background(), uuid.New())
	assert.Nil(t, detail)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestService_GetReport_PresignFailure_LeavesURLEmpty(t *testing.T) {
	reportID := uuid.New()
	repo := &mockFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return baseReport(reportID), nil },
		listAttachmentsFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Attachment, error) {
			return []*domain.Attachment{{ID: uuid.New(), ReportID: reportID, Kind: domain.AttachmentImage, ObjectKey: "reports/x/y.png"}}, nil
		},
		listNotesFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Note, error) { return nil, nil },
	}
	storage := &mockStorage{
		presignDownloadFn: func(_ context.Context, _ string, _ time.Duration) (string, error) {
			return "", errors.New("sign failed")
		},
	}
	svc := NewService(ServiceDeps{Reports: repo, Storage: storage, AnonymizationSalt: testSalt})

	detail, err := svc.GetReport(context.Background(), reportID)
	require.NoError(t, err, "a single failed presign must not blank the whole detail")
	require.Len(t, detail.Attachments, 1)
	assert.Empty(t, detail.Attachments[0].PresignedURL)
}

func TestService_UpdateReport_StatusToResolved_StampsAndAudits(t *testing.T) {
	reportID := uuid.New()
	adminID := uuid.New()
	report := baseReport(reportID)
	report.Status = domain.StatusInProgress

	var persisted *domain.Report
	repo := &mockFeedbackRepo{
		getReportFn:    func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return report, nil },
		updateReportFn: func(_ context.Context, r *domain.Report) error { persisted = r; return nil },
	}
	auditRec := &mockAuditRecorder{}
	svc := NewService(ServiceDeps{Reports: repo, Storage: &mockStorage{}, Audit: auditRec, AnonymizationSalt: testSalt})

	resolved := domain.StatusResolved
	updated, err := svc.UpdateReport(context.Background(), reportID, adminID, domain.Update{Status: &resolved}, "203.0.113.9")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, domain.StatusResolved, persisted.Status)
	require.NotNil(t, persisted.ResolvedBy)
	assert.Equal(t, adminID, *persisted.ResolvedBy)
	require.NotNil(t, persisted.ResolvedAt)

	// Exactly one audit event of the update kind.
	require.Len(t, auditRec.events, 1)
	ev := auditRec.events[0]
	assert.Equal(t, AdminActionUpdate, ev.Action)
	assert.Equal(t, adminID, ev.AdminUserID)
	assert.Equal(t, reportID, ev.ReportID)
	assert.Equal(t, "203.0.113.9", ev.IPAddress)
	assert.Equal(t, "resolved", ev.Metadata["status"])
}

func TestService_UpdateReport_SeverityOnly(t *testing.T) {
	reportID := uuid.New()
	report := baseReport(reportID)
	repo := &mockFeedbackRepo{
		getReportFn:    func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return report, nil },
		updateReportFn: func(_ context.Context, _ *domain.Report) error { return nil },
	}
	svc := newServiceWithRepo(repo)

	high := domain.SeverityHigh
	updated, err := svc.UpdateReport(context.Background(), reportID, uuid.New(), domain.Update{Severity: &high}, "")
	require.NoError(t, err)
	assert.Equal(t, domain.SeverityHigh, updated.Severity)
	assert.Equal(t, domain.StatusNew, updated.Status, "status untouched")
}

func TestService_UpdateReport_NoFields_Rejected(t *testing.T) {
	reportID := uuid.New()
	repo := &mockFeedbackRepo{
		getReportFn:    func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return baseReport(reportID), nil },
		updateReportFn: func(_ context.Context, _ *domain.Report) error { t.Fatal("must not persist"); return nil },
	}
	svc := newServiceWithRepo(repo)
	_, err := svc.UpdateReport(context.Background(), reportID, uuid.New(), domain.Update{}, "")
	assert.ErrorIs(t, err, domain.ErrNoUpdateFields)
}

func TestService_UpdateReport_InvalidStatus_Rejected(t *testing.T) {
	reportID := uuid.New()
	repo := &mockFeedbackRepo{
		getReportFn:    func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return baseReport(reportID), nil },
		updateReportFn: func(_ context.Context, _ *domain.Report) error { t.Fatal("must not persist"); return nil },
	}
	svc := newServiceWithRepo(repo)
	bad := domain.ReportStatus("closed")
	_, err := svc.UpdateReport(context.Background(), reportID, uuid.New(), domain.Update{Status: &bad}, "")
	assert.ErrorIs(t, err, domain.ErrInvalidStatus)
}

func TestService_UpdateReport_NotFound(t *testing.T) {
	repo := &mockFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return nil, domain.ErrNotFound },
	}
	svc := newServiceWithRepo(repo)
	triaged := domain.StatusTriaged
	_, err := svc.UpdateReport(context.Background(), uuid.New(), uuid.New(), domain.Update{Status: &triaged}, "")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestService_AddNote_OK_Audits(t *testing.T) {
	reportID := uuid.New()
	adminID := uuid.New()
	var persisted *domain.Note
	repo := &mockFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return baseReport(reportID), nil },
		addNoteFn:   func(_ context.Context, n *domain.Note) error { persisted = n; return nil },
	}
	auditRec := &mockAuditRecorder{}
	svc := NewService(ServiceDeps{Reports: repo, Storage: &mockStorage{}, Audit: auditRec, AnonymizationSalt: testSalt})

	note, err := svc.AddNote(context.Background(), reportID, adminID, "Reproduced on staging", "10.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, note)
	assert.Equal(t, "Reproduced on staging", persisted.Body)
	assert.Equal(t, adminID, persisted.AdminUserID)
	require.Len(t, auditRec.events, 1)
	assert.Equal(t, AdminActionNote, auditRec.events[0].Action)
}

func TestService_AddNote_ReportNotFound(t *testing.T) {
	repo := &mockFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return nil, domain.ErrNotFound },
		addNoteFn: func(_ context.Context, _ *domain.Note) error {
			t.Fatal("must not persist a note for a missing report")
			return nil
		},
	}
	svc := newServiceWithRepo(repo)
	_, err := svc.AddNote(context.Background(), uuid.New(), uuid.New(), "x", "")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestService_AddNote_EmptyBody_Rejected(t *testing.T) {
	reportID := uuid.New()
	repo := &mockFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return baseReport(reportID), nil },
		addNoteFn:   func(_ context.Context, _ *domain.Note) error { t.Fatal("must not persist"); return nil },
	}
	svc := newServiceWithRepo(repo)
	_, err := svc.AddNote(context.Background(), reportID, uuid.New(), "   ", "")
	assert.ErrorIs(t, err, domain.ErrNoteBodyEmpty)
}

func TestService_GetReport_ListAttachmentsError(t *testing.T) {
	reportID := uuid.New()
	repo := &mockFeedbackRepo{
		getReportFn:       func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return baseReport(reportID), nil },
		listAttachmentsFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Attachment, error) { return nil, errors.New("boom") },
	}
	svc := newServiceWithRepo(repo)
	_, err := svc.GetReport(context.Background(), reportID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list attachments")
}

func TestService_GetReport_ListNotesError(t *testing.T) {
	reportID := uuid.New()
	repo := &mockFeedbackRepo{
		getReportFn:       func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return baseReport(reportID), nil },
		listAttachmentsFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Attachment, error) { return nil, nil },
		listNotesFn:       func(_ context.Context, _ uuid.UUID) ([]*domain.Note, error) { return nil, errors.New("boom") },
	}
	svc := newServiceWithRepo(repo)
	_, err := svc.GetReport(context.Background(), reportID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list notes")
}

func TestService_UpdateReport_PersistError_Wrapped(t *testing.T) {
	reportID := uuid.New()
	repo := &mockFeedbackRepo{
		getReportFn:    func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return baseReport(reportID), nil },
		updateReportFn: func(_ context.Context, _ *domain.Report) error { return errors.New("db down") },
	}
	svc := newServiceWithRepo(repo)
	triaged := domain.StatusTriaged
	_, err := svc.UpdateReport(context.Background(), reportID, uuid.New(), domain.Update{Status: &triaged}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "persist")
}

func TestService_AddNote_PersistError_Wrapped(t *testing.T) {
	reportID := uuid.New()
	repo := &mockFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return baseReport(reportID), nil },
		addNoteFn:   func(_ context.Context, _ *domain.Note) error { return errors.New("db down") },
	}
	svc := newServiceWithRepo(repo)
	_, err := svc.AddNote(context.Background(), reportID, uuid.New(), "valid note body", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "persist")
}

func TestService_AddNote_NilAudit_NoPanic(t *testing.T) {
	reportID := uuid.New()
	repo := &mockFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*domain.Report, error) { return baseReport(reportID), nil },
		addNoteFn:   func(_ context.Context, _ *domain.Note) error { return nil },
	}
	svc := newServiceWithRepo(repo) // no Audit wired
	_, err := svc.AddNote(context.Background(), reportID, uuid.New(), "note", "")
	require.NoError(t, err)
}
