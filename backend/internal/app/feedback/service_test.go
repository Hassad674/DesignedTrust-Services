package feedback

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "marketplace-backend/internal/domain/feedback"
	portrepo "marketplace-backend/internal/port/repository"
)

const testSalt = "test-anonymization-salt-32-bytes-min"

func validSubmit() SubmitReportInput {
	return SubmitReportInput{
		Type:        string(domain.TypeBug),
		Title:       "Login button does nothing",
		Description: "Clicking the login button on the homepage does nothing at all.",
	}
}

func newServiceWithRepo(repo *mockFeedbackRepo) *Service {
	return NewService(ServiceDeps{
		Reports:           repo,
		Storage:           &mockStorage{},
		AnonymizationSalt: testSalt,
	})
}

func TestService_SubmitReport_AnonymousText_OK(t *testing.T) {
	var captured *domain.Report
	repo := &mockFeedbackRepo{
		createReportFn: func(_ context.Context, report *domain.Report, attachments []*domain.Attachment) error {
			captured = report
			assert.Empty(t, attachments, "anonymous text report has no attachments")
			return nil
		},
	}
	svc := newServiceWithRepo(repo)

	in := validSubmit()
	in.RawIP = "203.0.113.7"
	report, err := svc.SubmitReport(context.Background(), in)
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.True(t, report.IsAnonymous())
	assert.Equal(t, domain.StatusNew, report.Status)

	// IP must be hashed with the salt, never stored raw.
	wantHash := sha256.Sum256([]byte("203.0.113.7" + testSalt))
	assert.Equal(t, hex.EncodeToString(wantHash[:]), captured.IPHash)
	assert.NotContains(t, captured.IPHash, "203.0.113.7")
}

func TestService_SubmitReport_EmptyIP_LeavesHashEmpty(t *testing.T) {
	repo := &mockFeedbackRepo{
		createReportFn: func(_ context.Context, report *domain.Report, _ []*domain.Attachment) error {
			assert.Empty(t, report.IPHash)
			return nil
		},
	}
	svc := newServiceWithRepo(repo)
	_, err := svc.SubmitReport(context.Background(), validSubmit())
	require.NoError(t, err)
}

func TestService_SubmitReport_AnonymousWithAttachments_Rejected(t *testing.T) {
	repo := &mockFeedbackRepo{
		createReportFn: func(_ context.Context, _ *domain.Report, _ []*domain.Attachment) error {
			t.Fatal("repo.CreateReport must NOT be called when an anonymous submit carries attachments")
			return nil
		},
	}
	svc := newServiceWithRepo(repo)

	in := validSubmit()
	in.Attachments = []AttachmentRef{
		{Kind: "image", ObjectKey: "reports/a/b.png", ContentType: "image/png", SizeBytes: 1024},
	}
	report, err := svc.SubmitReport(context.Background(), in)
	assert.Nil(t, report)
	assert.ErrorIs(t, err, domain.ErrAttachmentNotAllowed)
}

func TestService_SubmitReport_LoggedInWithAttachments_OK(t *testing.T) {
	reporterID := uuid.New()
	var capturedAttachments []*domain.Attachment
	repo := &mockFeedbackRepo{
		createReportFn: func(_ context.Context, report *domain.Report, attachments []*domain.Attachment) error {
			capturedAttachments = attachments
			assert.False(t, report.IsAnonymous())
			return nil
		},
	}
	svc := newServiceWithRepo(repo)

	in := validSubmit()
	in.Type = string(domain.TypeSecurity)
	in.ReporterID = &reporterID
	in.Attachments = []AttachmentRef{
		{Kind: "image", ObjectKey: "reports/a/b.png", ContentType: "image/png", SizeBytes: 1024},
		{Kind: "video", ObjectKey: "reports/a/c.mp4", ContentType: "video/mp4", SizeBytes: 5 << 20},
	}
	report, err := svc.SubmitReport(context.Background(), in)
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Len(t, capturedAttachments, 2)
	assert.Equal(t, domain.AttachmentImage, capturedAttachments[0].Kind)
	assert.Equal(t, domain.AttachmentVideo, capturedAttachments[1].Kind)
	assert.Equal(t, report.ID, capturedAttachments[0].ReportID)
}

func TestService_SubmitReport_LoggedInBadAttachment_Rejected(t *testing.T) {
	reporterID := uuid.New()
	tests := []struct {
		name    string
		ref     AttachmentRef
		wantErr error
	}{
		{
			name:    "disallowed content type",
			ref:     AttachmentRef{Kind: "image", ObjectKey: "reports/a/b.gif", ContentType: "image/gif", SizeBytes: 1024},
			wantErr: domain.ErrUnsupportedContentType,
		},
		{
			name:    "oversized image",
			ref:     AttachmentRef{Kind: "image", ObjectKey: "reports/a/b.png", ContentType: "image/png", SizeBytes: domain.MaxImageSizeBytes + 1},
			wantErr: domain.ErrAttachmentTooLarge,
		},
		{
			name:    "tampered key extension",
			ref:     AttachmentRef{Kind: "image", ObjectKey: "reports/a/b.html", ContentType: "image/png", SizeBytes: 1024},
			wantErr: domain.ErrUnsupportedContentType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFeedbackRepo{
				createReportFn: func(_ context.Context, _ *domain.Report, _ []*domain.Attachment) error {
					t.Fatal("repo must not be called when an attachment is invalid")
					return nil
				},
			}
			svc := newServiceWithRepo(repo)
			in := validSubmit()
			in.ReporterID = &reporterID
			in.Attachments = []AttachmentRef{tt.ref}
			_, err := svc.SubmitReport(context.Background(), in)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestService_SubmitReport_InvalidDomain_Rejected(t *testing.T) {
	repo := &mockFeedbackRepo{
		createReportFn: func(_ context.Context, _ *domain.Report, _ []*domain.Attachment) error {
			t.Fatal("repo must not be called on invalid input")
			return nil
		},
	}
	svc := newServiceWithRepo(repo)

	in := validSubmit()
	in.Type = "feature" // invalid
	_, err := svc.SubmitReport(context.Background(), in)
	assert.ErrorIs(t, err, domain.ErrInvalidType)
}

func TestService_SubmitReport_RepoError_Wrapped(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &mockFeedbackRepo{
		createReportFn: func(_ context.Context, _ *domain.Report, _ []*domain.Attachment) error {
			return repoErr
		},
	}
	svc := newServiceWithRepo(repo)
	_, err := svc.SubmitReport(context.Background(), validSubmit())
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}

func TestService_PresignAttachment(t *testing.T) {
	tests := []struct {
		name        string
		in          PresignInput
		wantErr     error
		wantKind    string
		wantExtPart string
	}{
		{
			name:        "image ok",
			in:          PresignInput{Kind: "image", ContentType: "image/png", SizeBytes: 1024},
			wantKind:    "image",
			wantExtPart: ".png",
		},
		{
			name:        "video ok",
			in:          PresignInput{Kind: "video", ContentType: "video/mp4", SizeBytes: 5 << 20},
			wantKind:    "video",
			wantExtPart: ".mp4",
		},
		{
			name:    "disallowed mime",
			in:      PresignInput{Kind: "image", ContentType: "application/pdf", SizeBytes: 1024},
			wantErr: domain.ErrUnsupportedContentType,
		},
		{
			name:    "oversized video",
			in:      PresignInput{Kind: "video", ContentType: "video/mp4", SizeBytes: domain.MaxVideoSizeBytes + 1},
			wantErr: domain.ErrAttachmentTooLarge,
		},
		{
			name:    "invalid kind",
			in:      PresignInput{Kind: "audio", ContentType: "image/png", SizeBytes: 1024},
			wantErr: domain.ErrInvalidAttachmentKind,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var presignedKey string
			storage := &mockStorage{
				presignUploadFn: func(_ context.Context, key, _ string, _ time.Duration) (string, error) {
					presignedKey = key
					return "https://r2/put/" + key, nil
				},
			}
			svc := NewService(ServiceDeps{Reports: &mockFeedbackRepo{}, Storage: storage, AnonymizationSalt: testSalt})

			res, err := svc.PresignAttachment(context.Background(), tt.in)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
				assert.Empty(t, presignedKey, "storage must not be called on validation failure")
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, tt.wantKind, res.Kind)
			// Key must be server-minted under reports/ and carry the
			// server-controlled extension — never a client filename.
			assert.Contains(t, res.ObjectKey, "reports/")
			assert.Contains(t, res.ObjectKey, tt.wantExtPart)
			assert.Equal(t, presignedKey, res.ObjectKey)
			assert.Contains(t, res.UploadURL, "https://r2/put/")
		})
	}
}

func TestService_PresignAttachment_StorageError_Wrapped(t *testing.T) {
	storage := &mockStorage{
		presignUploadFn: func(_ context.Context, _, _ string, _ time.Duration) (string, error) {
			return "", errors.New("r2 unavailable")
		},
	}
	svc := NewService(ServiceDeps{Reports: &mockFeedbackRepo{}, Storage: storage, AnonymizationSalt: testSalt})
	res, err := svc.PresignAttachment(context.Background(), PresignInput{Kind: "image", ContentType: "image/png", SizeBytes: 1024})
	assert.Nil(t, res)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "presign attachment")
}

func TestService_ListReports_ClampsLimitAndPassesFilter(t *testing.T) {
	var gotFilter portrepo.ReportFilter
	var gotLimit int
	repo := &mockFeedbackRepo{
		listReportsFn: func(_ context.Context, filter portrepo.ReportFilter, _ string, limit int) ([]*portrepo.ReportSummary, string, error) {
			gotFilter = filter
			gotLimit = limit
			return []*portrepo.ReportSummary{}, "", nil
		},
	}
	svc := newServiceWithRepo(repo)
	filter := portrepo.ReportFilter{Type: "bug", Status: "new", Severity: "high", Search: "login"}

	_, _, err := svc.ListReports(context.Background(), filter, "", 0) // 0 -> clamp to 20
	require.NoError(t, err)
	assert.Equal(t, 20, gotLimit)
	assert.Equal(t, filter, gotFilter)

	_, _, err = svc.ListReports(context.Background(), filter, "", 500) // >100 -> clamp
	require.NoError(t, err)
	assert.Equal(t, 20, gotLimit)

	_, _, err = svc.ListReports(context.Background(), filter, "", 50) // in-range preserved
	require.NoError(t, err)
	assert.Equal(t, 50, gotLimit)
}
