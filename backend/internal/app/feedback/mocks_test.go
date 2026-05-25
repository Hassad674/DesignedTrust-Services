package feedback

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"

	domain "marketplace-backend/internal/domain/feedback"
	portrepo "marketplace-backend/internal/port/repository"
	portservice "marketplace-backend/internal/port/service"
)

// mockFeedbackRepo is a hand-written mock of
// repository.FeedbackRepository. Each method delegates to an injectable
// closure so a test can stub exactly the behaviour it needs; an unset
// closure panics loudly so a test never silently exercises a default.
type mockFeedbackRepo struct {
	createReportFn    func(ctx context.Context, report *domain.Report, attachments []*domain.Attachment) error
	listReportsFn     func(ctx context.Context, filter portrepo.ReportFilter, cursor string, limit int) ([]*portrepo.ReportSummary, string, error)
	getReportFn       func(ctx context.Context, id uuid.UUID) (*domain.Report, error)
	updateReportFn    func(ctx context.Context, report *domain.Report) error
	listAttachmentsFn func(ctx context.Context, reportID uuid.UUID) ([]*domain.Attachment, error)
	addNoteFn         func(ctx context.Context, note *domain.Note) error
	listNotesFn       func(ctx context.Context, reportID uuid.UUID) ([]*domain.Note, error)
}

func (m *mockFeedbackRepo) CreateReport(ctx context.Context, report *domain.Report, attachments []*domain.Attachment) error {
	return m.createReportFn(ctx, report, attachments)
}

func (m *mockFeedbackRepo) ListReports(ctx context.Context, filter portrepo.ReportFilter, cursor string, limit int) ([]*portrepo.ReportSummary, string, error) {
	return m.listReportsFn(ctx, filter, cursor, limit)
}

func (m *mockFeedbackRepo) GetReport(ctx context.Context, id uuid.UUID) (*domain.Report, error) {
	return m.getReportFn(ctx, id)
}

func (m *mockFeedbackRepo) UpdateReport(ctx context.Context, report *domain.Report) error {
	return m.updateReportFn(ctx, report)
}

func (m *mockFeedbackRepo) ListAttachments(ctx context.Context, reportID uuid.UUID) ([]*domain.Attachment, error) {
	return m.listAttachmentsFn(ctx, reportID)
}

func (m *mockFeedbackRepo) AddNote(ctx context.Context, note *domain.Note) error {
	return m.addNoteFn(ctx, note)
}

func (m *mockFeedbackRepo) ListNotes(ctx context.Context, reportID uuid.UUID) ([]*domain.Note, error) {
	return m.listNotesFn(ctx, reportID)
}

// Compile-time assertion that the mock satisfies the port.
var _ portrepo.FeedbackRepository = (*mockFeedbackRepo)(nil)

// mockStorage is a hand-written mock of service.StorageService. Only the
// two methods the feedback service touches (GetPresignedUploadURL,
// GetPresignedDownloadURL) carry injectable closures; the rest return
// zero values because the feedback service never calls them.
type mockStorage struct {
	presignUploadFn   func(ctx context.Context, key, contentType string, expiry time.Duration) (string, error)
	presignDownloadFn func(ctx context.Context, key string, expiry time.Duration) (string, error)
}

func (m *mockStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (string, error) {
	return "", nil
}

func (m *mockStorage) Delete(ctx context.Context, key string) error { return nil }

func (m *mockStorage) BulkDelete(ctx context.Context, keys []string) ([]portservice.BulkDeleteResult, error) {
	return nil, nil
}

func (m *mockStorage) GetPublicURL(key string) string { return "https://public/" + key }

func (m *mockStorage) GetPresignedUploadURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	return m.presignUploadFn(ctx, key, contentType, expiry)
}

func (m *mockStorage) GetPresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if m.presignDownloadFn == nil {
		return "https://signed/" + key, nil
	}
	return m.presignDownloadFn(ctx, key, expiry)
}

func (m *mockStorage) GetPresignedDownloadURLAsAttachment(ctx context.Context, key, filename string, expiry time.Duration) (string, error) {
	return "", nil
}

func (m *mockStorage) Download(ctx context.Context, key string) ([]byte, error) { return nil, nil }

var _ portservice.StorageService = (*mockStorage)(nil)

// mockAuditRecorder captures the events the service emits so tests can
// assert that an admin mutation produced exactly one audit row of the
// expected shape.
type mockAuditRecorder struct {
	events []AuditEvent
}

func (m *mockAuditRecorder) RecordAdminAction(ctx context.Context, event AuditEvent) {
	m.events = append(m.events, event)
}

var _ auditRecorder = (*mockAuditRecorder)(nil)
