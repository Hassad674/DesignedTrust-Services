package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	feedbackapp "marketplace-backend/internal/app/feedback"
	feedbackdomain "marketplace-backend/internal/domain/feedback"
	"marketplace-backend/internal/handler/middleware"
	portrepo "marketplace-backend/internal/port/repository"
	portservice "marketplace-backend/internal/port/service"
)

// --- stubs -----------------------------------------------------------

// stubFeedbackRepo is a configurable in-memory FeedbackRepository for
// the handler tests. Closures default to benign behaviour so a test
// only overrides what it exercises.
type stubFeedbackRepo struct {
	createReportFn    func(ctx context.Context, r *feedbackdomain.Report, a []*feedbackdomain.Attachment) error
	listReportsFn     func(ctx context.Context, f portrepo.ReportFilter, c string, l int) ([]*portrepo.ReportSummary, string, error)
	getReportFn       func(ctx context.Context, id uuid.UUID) (*feedbackdomain.Report, error)
	updateReportFn    func(ctx context.Context, r *feedbackdomain.Report) error
	listAttachmentsFn func(ctx context.Context, id uuid.UUID) ([]*feedbackdomain.Attachment, error)
	addNoteFn         func(ctx context.Context, n *feedbackdomain.Note) error
	listNotesFn       func(ctx context.Context, id uuid.UUID) ([]*feedbackdomain.Note, error)
}

func (s *stubFeedbackRepo) CreateReport(ctx context.Context, r *feedbackdomain.Report, a []*feedbackdomain.Attachment) error {
	if s.createReportFn != nil {
		return s.createReportFn(ctx, r, a)
	}
	return nil
}
func (s *stubFeedbackRepo) ListReports(ctx context.Context, f portrepo.ReportFilter, c string, l int) ([]*portrepo.ReportSummary, string, error) {
	if s.listReportsFn != nil {
		return s.listReportsFn(ctx, f, c, l)
	}
	return []*portrepo.ReportSummary{}, "", nil
}
func (s *stubFeedbackRepo) GetReport(ctx context.Context, id uuid.UUID) (*feedbackdomain.Report, error) {
	if s.getReportFn != nil {
		return s.getReportFn(ctx, id)
	}
	return nil, feedbackdomain.ErrNotFound
}
func (s *stubFeedbackRepo) UpdateReport(ctx context.Context, r *feedbackdomain.Report) error {
	if s.updateReportFn != nil {
		return s.updateReportFn(ctx, r)
	}
	return nil
}
func (s *stubFeedbackRepo) ListAttachments(ctx context.Context, id uuid.UUID) ([]*feedbackdomain.Attachment, error) {
	if s.listAttachmentsFn != nil {
		return s.listAttachmentsFn(ctx, id)
	}
	return nil, nil
}
func (s *stubFeedbackRepo) AddNote(ctx context.Context, n *feedbackdomain.Note) error {
	if s.addNoteFn != nil {
		return s.addNoteFn(ctx, n)
	}
	return nil
}
func (s *stubFeedbackRepo) ListNotes(ctx context.Context, id uuid.UUID) ([]*feedbackdomain.Note, error) {
	if s.listNotesFn != nil {
		return s.listNotesFn(ctx, id)
	}
	return nil, nil
}

var _ portrepo.FeedbackRepository = (*stubFeedbackRepo)(nil)

// stubFeedbackStorage stubs the storage port for presign tests.
type stubFeedbackStorage struct {
	uploadURL string
}

func (s *stubFeedbackStorage) Upload(context.Context, string, io.Reader, string, int64) (string, error) {
	return "", nil
}
func (s *stubFeedbackStorage) Delete(context.Context, string) error { return nil }
func (s *stubFeedbackStorage) BulkDelete(context.Context, []string) ([]portservice.BulkDeleteResult, error) {
	return nil, nil
}
func (s *stubFeedbackStorage) GetPublicURL(key string) string { return "https://public/" + key }
func (s *stubFeedbackStorage) GetPresignedUploadURL(_ context.Context, key, _ string, _ time.Duration) (string, error) {
	if s.uploadURL != "" {
		return s.uploadURL, nil
	}
	return "https://r2/put/" + key, nil
}
func (s *stubFeedbackStorage) GetPresignedDownloadURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://signed/" + key, nil
}
func (s *stubFeedbackStorage) GetPresignedDownloadURLAsAttachment(context.Context, string, string, time.Duration) (string, error) {
	return "", nil
}
func (s *stubFeedbackStorage) Download(context.Context, string) ([]byte, error) { return nil, nil }

var _ portservice.StorageService = (*stubFeedbackStorage)(nil)

func newFeedbackHandlerForTest(repo portrepo.FeedbackRepository, storage portservice.StorageService) *FeedbackHandler {
	svc := feedbackapp.NewService(feedbackapp.ServiceDeps{
		Reports:           repo,
		Storage:           storage,
		AnonymizationSalt: "handler-test-salt",
	})
	return NewFeedbackHandler(svc)
}

func withUserCtx(r *http.Request, uid uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextKeyUserID, uid)
	return r.WithContext(ctx)
}

// --- Submit ----------------------------------------------------------

func TestFeedbackHandler_Submit_AnonymousText_201(t *testing.T) {
	var created *feedbackdomain.Report
	repo := &stubFeedbackRepo{
		createReportFn: func(_ context.Context, r *feedbackdomain.Report, a []*feedbackdomain.Attachment) error {
			created = r
			assert.Empty(t, a)
			return nil
		},
	}
	h := newFeedbackHandlerForTest(repo, &stubFeedbackStorage{})

	body := mustJSON(t, map[string]any{
		"type":        "bug",
		"title":       "Login button does nothing",
		"description": "Clicking the login button on the homepage does nothing.",
		"page_url":    "https://app.example.com/login",
		"context":     map[string]any{"role": "agency", "locale": "fr"},
	})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", bytes.NewReader(body))
	r.RemoteAddr = "203.0.113.5:9999"
	w := httptest.NewRecorder()

	h.Submit(w, r)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, created)
	assert.True(t, created.IsAnonymous())
	assert.NotEmpty(t, created.IPHash)
	assert.NotContains(t, created.IPHash, "203.0.113.5")
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "bug", resp["type"])
	assert.Equal(t, "new", resp["status"])
}

func TestFeedbackHandler_Submit_StripsHTML(t *testing.T) {
	var created *feedbackdomain.Report
	repo := &stubFeedbackRepo{
		createReportFn: func(_ context.Context, r *feedbackdomain.Report, _ []*feedbackdomain.Attachment) error {
			created = r
			return nil
		},
	}
	h := newFeedbackHandlerForTest(repo, &stubFeedbackStorage{})
	body := mustJSON(t, map[string]any{
		"type":        "bug",
		"title":       "Bug <script>alert(1)</script> here",
		"description": "Body with <b>bold</b> and <img src=x onerror=alert(1)> tags inside.",
	})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Submit(w, r)
	require.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, created)
	assert.NotContains(t, created.Title, "<script>")
	assert.NotContains(t, created.Description, "<img")
	assert.NotContains(t, created.Description, "<b>")
}

func TestFeedbackHandler_Submit_LoggedInWithAttachments_201(t *testing.T) {
	reporterID := uuid.New()
	var capturedAttachments []*feedbackdomain.Attachment
	repo := &stubFeedbackRepo{
		createReportFn: func(_ context.Context, r *feedbackdomain.Report, a []*feedbackdomain.Attachment) error {
			capturedAttachments = a
			assert.False(t, r.IsAnonymous())
			return nil
		},
	}
	h := newFeedbackHandlerForTest(repo, &stubFeedbackStorage{})
	body := mustJSON(t, map[string]any{
		"type":        "security",
		"title":       "XSS in profile bio",
		"description": "The profile bio field renders unescaped HTML to other users.",
		"attachment_keys": []map[string]any{
			{"kind": "image", "object_key": "reports/a/b.png", "content_type": "image/png", "size_bytes": 2048},
		},
	})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", bytes.NewReader(body))
	r = withUserCtx(r, reporterID)
	w := httptest.NewRecorder()

	h.Submit(w, r)
	require.Equal(t, http.StatusCreated, w.Code)
	require.Len(t, capturedAttachments, 1)
	assert.Equal(t, feedbackdomain.AttachmentImage, capturedAttachments[0].Kind)
}

func TestFeedbackHandler_Submit_AnonymousWithAttachments_400(t *testing.T) {
	repo := &stubFeedbackRepo{
		createReportFn: func(_ context.Context, _ *feedbackdomain.Report, _ []*feedbackdomain.Attachment) error {
			t.Fatal("must not persist an anonymous report carrying attachments")
			return nil
		},
	}
	h := newFeedbackHandlerForTest(repo, &stubFeedbackStorage{})
	body := mustJSON(t, map[string]any{
		"type":        "bug",
		"title":       "Has a screenshot",
		"description": "Anonymous tried to attach a screenshot which is not allowed.",
		"attachment_keys": []map[string]any{
			{"kind": "image", "object_key": "reports/a/b.png", "content_type": "image/png", "size_bytes": 2048},
		},
	})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Submit(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFeedbackHandler_Submit_Validation_400(t *testing.T) {
	h := newFeedbackHandlerForTest(&stubFeedbackRepo{}, &stubFeedbackStorage{})
	tests := []struct {
		name string
		body map[string]any
	}{
		{"invalid type", map[string]any{"type": "feature", "title": "valid title here", "description": "valid description here ok"}},
		{"title too short", map[string]any{"type": "bug", "title": "ab", "description": "valid description here ok"}},
		{"description too short", map[string]any{"type": "bug", "title": "valid title here", "description": "short"}},
		{"unknown field rejected", map[string]any{"type": "bug", "title": "valid title", "description": "valid description here", "evil": "x"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", bytes.NewReader(mustJSON(t, tt.body)))
			w := httptest.NewRecorder()
			h.Submit(w, r)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestFeedbackHandler_Submit_Honeypot_200_Dropped(t *testing.T) {
	repo := &stubFeedbackRepo{
		createReportFn: func(_ context.Context, _ *feedbackdomain.Report, _ []*feedbackdomain.Attachment) error {
			t.Fatal("honeypot submission must NOT be persisted")
			return nil
		},
	}
	h := newFeedbackHandlerForTest(repo, &stubFeedbackStorage{})
	body := mustJSON(t, map[string]any{
		"type":        "bug",
		"title":       "valid title here",
		"description": "valid description here ok",
		"hp":          "i-am-a-bot",
	})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Submit(w, r)
	// Looks like success to the bot, but nothing was persisted.
	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Presign ---------------------------------------------------------

func TestFeedbackHandler_Presign_Anonymous_401(t *testing.T) {
	h := newFeedbackHandlerForTest(&stubFeedbackRepo{}, &stubFeedbackStorage{})
	body := mustJSON(t, map[string]any{"kind": "image", "content_type": "image/png", "size_bytes": 1024})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/attachments/presign", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.PresignAttachment(w, r) // no user in context
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFeedbackHandler_Presign_Authenticated_200(t *testing.T) {
	h := newFeedbackHandlerForTest(&stubFeedbackRepo{}, &stubFeedbackStorage{})
	body := mustJSON(t, map[string]any{"kind": "video", "content_type": "video/mp4", "size_bytes": 5 << 20, "filename": "clip.mp4"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/attachments/presign", bytes.NewReader(body))
	r = withUserCtx(r, uuid.New())
	w := httptest.NewRecorder()
	h.PresignAttachment(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "video", resp["kind"])
	objectKey, _ := resp["object_key"].(string)
	assert.Contains(t, objectKey, "reports/")
	assert.Contains(t, objectKey, ".mp4")
	assert.NotContains(t, objectKey, "clip.mp4", "client filename must never appear in the storage key")
}

func TestFeedbackHandler_Submit_MalformedBody_400(t *testing.T) {
	h := newFeedbackHandlerForTest(&stubFeedbackRepo{}, &stubFeedbackStorage{})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", bytes.NewReader([]byte("{not json")))
	w := httptest.NewRecorder()
	h.Submit(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFeedbackHandler_Submit_RepoError_500(t *testing.T) {
	repo := &stubFeedbackRepo{
		createReportFn: func(context.Context, *feedbackdomain.Report, []*feedbackdomain.Attachment) error {
			return assertAnError{}
		},
	}
	h := newFeedbackHandlerForTest(repo, &stubFeedbackStorage{})
	body := mustJSON(t, map[string]any{"type": "bug", "title": "valid title here", "description": "valid description here ok"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Submit(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFeedbackHandler_Presign_MalformedBody_400(t *testing.T) {
	h := newFeedbackHandlerForTest(&stubFeedbackRepo{}, &stubFeedbackStorage{})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/attachments/presign", bytes.NewReader([]byte("{bad")))
	r = withUserCtx(r, uuid.New())
	w := httptest.NewRecorder()
	h.PresignAttachment(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// assertAnError is a non-domain error used to exercise the 500 path.
type assertAnError struct{}

func (assertAnError) Error() string { return "infra failure" }

func TestFeedbackHandler_Presign_BadInput_400(t *testing.T) {
	h := newFeedbackHandlerForTest(&stubFeedbackRepo{}, &stubFeedbackStorage{})
	tests := []struct {
		name string
		body map[string]any
	}{
		{"disallowed mime", map[string]any{"kind": "image", "content_type": "image/gif", "size_bytes": 1024}},
		{"oversized video", map[string]any{"kind": "video", "content_type": "video/mp4", "size_bytes": 60 << 20}},
		{"invalid kind", map[string]any{"kind": "audio", "content_type": "image/png", "size_bytes": 1024}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/attachments/presign", bytes.NewReader(mustJSON(t, tt.body)))
			r = withUserCtx(r, uuid.New())
			w := httptest.NewRecorder()
			h.PresignAttachment(w, r)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
