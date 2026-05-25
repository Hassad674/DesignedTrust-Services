package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	feedbackapp "marketplace-backend/internal/app/feedback"
	feedbackdomain "marketplace-backend/internal/domain/feedback"
	"marketplace-backend/internal/handler/middleware"
	portrepo "marketplace-backend/internal/port/repository"
)

func newAdminFeedbackHandlerForTest(repo portrepo.FeedbackRepository) *AdminFeedbackHandler {
	svc := feedbackapp.NewService(feedbackapp.ServiceDeps{
		Reports:           repo,
		Storage:           &stubFeedbackStorage{},
		AnonymizationSalt: "handler-test-salt",
	})
	return NewAdminFeedbackHandler(svc)
}

// adminCtx returns a request carrying an admin identity (is_admin=true
// + a user id) so the per-handler defensive admin check passes.
func adminCtx(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextKeyIsAdmin, true)
	ctx = context.WithValue(ctx, middleware.ContextKeyUserID, uuid.New())
	return r.WithContext(ctx)
}

// withFeedbackChiID attaches a chi route param {id} to the request context.
func withFeedbackChiID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func sampleReport(id uuid.UUID) *feedbackdomain.Report {
	now := time.Now().UTC()
	return &feedbackdomain.Report{
		ID:        id,
		Type:      feedbackdomain.TypeBug,
		Title:     "Something broke",
		Status:    feedbackdomain.StatusNew,
		Context:   map[string]any{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// --- List ------------------------------------------------------------

func TestAdminFeedbackHandler_List_200(t *testing.T) {
	report := sampleReport(uuid.New())
	repo := &stubFeedbackRepo{
		listReportsFn: func(_ context.Context, f portrepo.ReportFilter, _ string, _ int) ([]*portrepo.ReportSummary, string, error) {
			assert.Equal(t, "bug", f.Type)
			assert.Equal(t, "new", f.Status)
			return []*portrepo.ReportSummary{{Report: report, AttachmentCount: 2, NoteCount: 1}}, "", nil
		},
	}
	h := newAdminFeedbackHandlerForTest(repo)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/admin/feedback?type=bug&status=new", nil)
	r = adminCtx(r)
	w := httptest.NewRecorder()
	h.List(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, _ := resp["data"].([]any)
	require.Len(t, data, 1)
	row, _ := data[0].(map[string]any)
	assert.Equal(t, float64(2), row["attachment_count"])
	assert.Equal(t, float64(1), row["note_count"])
}

func TestAdminFeedbackHandler_List_403_NonAdmin(t *testing.T) {
	h := newAdminFeedbackHandlerForTest(&stubFeedbackRepo{})
	r := httptest.NewRequest(http.MethodGet, "/api/v1/admin/feedback", nil) // no admin ctx
	w := httptest.NewRecorder()
	h.List(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- Get -------------------------------------------------------------

func TestAdminFeedbackHandler_Get_200_WithAttachmentsAndNotes(t *testing.T) {
	id := uuid.New()
	repo := &stubFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*feedbackdomain.Report, error) { return sampleReport(id), nil },
		listAttachmentsFn: func(_ context.Context, _ uuid.UUID) ([]*feedbackdomain.Attachment, error) {
			return []*feedbackdomain.Attachment{{ID: uuid.New(), ReportID: id, Kind: feedbackdomain.AttachmentImage, ObjectKey: "reports/x/y.png"}}, nil
		},
		listNotesFn: func(_ context.Context, _ uuid.UUID) ([]*feedbackdomain.Note, error) {
			return []*feedbackdomain.Note{{ID: uuid.New(), ReportID: id, AdminUserID: uuid.New(), Body: "triaging"}}, nil
		},
	}
	h := newAdminFeedbackHandlerForTest(repo)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/admin/feedback/"+id.String(), nil)
	r = withFeedbackChiID(adminCtx(r), id.String())
	w := httptest.NewRecorder()
	h.Get(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	attachments, _ := resp["attachments"].([]any)
	require.Len(t, attachments, 1)
	att, _ := attachments[0].(map[string]any)
	assert.Equal(t, "https://signed/reports/x/y.png", att["url"], "attachment must carry a presigned GET url")
	notes, _ := resp["notes"].([]any)
	require.Len(t, notes, 1)
}

func TestAdminFeedbackHandler_Get_404(t *testing.T) {
	id := uuid.New()
	repo := &stubFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*feedbackdomain.Report, error) {
			return nil, feedbackdomain.ErrNotFound
		},
	}
	h := newAdminFeedbackHandlerForTest(repo)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/admin/feedback/"+id.String(), nil)
	r = withFeedbackChiID(adminCtx(r), id.String())
	w := httptest.NewRecorder()
	h.Get(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminFeedbackHandler_Get_400_BadID(t *testing.T) {
	h := newAdminFeedbackHandlerForTest(&stubFeedbackRepo{})
	r := httptest.NewRequest(http.MethodGet, "/api/v1/admin/feedback/not-a-uuid", nil)
	r = withFeedbackChiID(adminCtx(r), "not-a-uuid")
	w := httptest.NewRecorder()
	h.Get(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminFeedbackHandler_Get_403_NonAdmin(t *testing.T) {
	h := newAdminFeedbackHandlerForTest(&stubFeedbackRepo{})
	id := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/admin/feedback/"+id.String(), nil)
	r = withFeedbackChiID(r, id.String()) // no admin ctx
	w := httptest.NewRecorder()
	h.Get(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- Update ----------------------------------------------------------

func TestAdminFeedbackHandler_Update_200(t *testing.T) {
	id := uuid.New()
	var persisted *feedbackdomain.Report
	repo := &stubFeedbackRepo{
		getReportFn:    func(_ context.Context, _ uuid.UUID) (*feedbackdomain.Report, error) { return sampleReport(id), nil },
		updateReportFn: func(_ context.Context, rr *feedbackdomain.Report) error { persisted = rr; return nil },
	}
	h := newAdminFeedbackHandlerForTest(repo)
	body := mustJSON(t, map[string]any{"status": "resolved", "severity": "high"})
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/feedback/"+id.String(), bytes.NewReader(body))
	r = withFeedbackChiID(adminCtx(r), id.String())
	w := httptest.NewRecorder()
	h.Update(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, persisted)
	assert.Equal(t, feedbackdomain.StatusResolved, persisted.Status)
	assert.Equal(t, feedbackdomain.SeverityHigh, persisted.Severity)
	require.NotNil(t, persisted.ResolvedAt)
	require.NotNil(t, persisted.ResolvedBy)
}

func TestAdminFeedbackHandler_Update_400_InvalidStatus(t *testing.T) {
	id := uuid.New()
	repo := &stubFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*feedbackdomain.Report, error) { return sampleReport(id), nil },
	}
	h := newAdminFeedbackHandlerForTest(repo)
	body := mustJSON(t, map[string]any{"status": "closed"})
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/feedback/"+id.String(), bytes.NewReader(body))
	r = withFeedbackChiID(adminCtx(r), id.String())
	w := httptest.NewRecorder()
	h.Update(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminFeedbackHandler_Update_404(t *testing.T) {
	id := uuid.New()
	repo := &stubFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*feedbackdomain.Report, error) {
			return nil, feedbackdomain.ErrNotFound
		},
	}
	h := newAdminFeedbackHandlerForTest(repo)
	body := mustJSON(t, map[string]any{"status": "triaged"})
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/feedback/"+id.String(), bytes.NewReader(body))
	r = withFeedbackChiID(adminCtx(r), id.String())
	w := httptest.NewRecorder()
	h.Update(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminFeedbackHandler_Update_403_NonAdmin(t *testing.T) {
	h := newAdminFeedbackHandlerForTest(&stubFeedbackRepo{})
	id := uuid.New()
	body := mustJSON(t, map[string]any{"status": "triaged"})
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/feedback/"+id.String(), bytes.NewReader(body))
	r = withFeedbackChiID(r, id.String()) // no admin ctx
	w := httptest.NewRecorder()
	h.Update(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- AddNote ---------------------------------------------------------

func TestAdminFeedbackHandler_AddNote_201(t *testing.T) {
	id := uuid.New()
	var persisted *feedbackdomain.Note
	repo := &stubFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*feedbackdomain.Report, error) { return sampleReport(id), nil },
		addNoteFn:   func(_ context.Context, n *feedbackdomain.Note) error { persisted = n; return nil },
	}
	h := newAdminFeedbackHandlerForTest(repo)
	body := mustJSON(t, map[string]any{"body": "Reproduced on staging, escalating."})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/admin/feedback/"+id.String()+"/notes", bytes.NewReader(body))
	r = withFeedbackChiID(adminCtx(r), id.String())
	w := httptest.NewRecorder()
	h.AddNote(w, r)
	require.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, persisted)
	assert.Equal(t, "Reproduced on staging, escalating.", persisted.Body)
}

func TestAdminFeedbackHandler_AddNote_400_EmptyBody(t *testing.T) {
	id := uuid.New()
	repo := &stubFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*feedbackdomain.Report, error) { return sampleReport(id), nil },
	}
	h := newAdminFeedbackHandlerForTest(repo)
	body := mustJSON(t, map[string]any{"body": "   "})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/admin/feedback/"+id.String()+"/notes", bytes.NewReader(body))
	r = withFeedbackChiID(adminCtx(r), id.String())
	w := httptest.NewRecorder()
	h.AddNote(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminFeedbackHandler_AddNote_404(t *testing.T) {
	id := uuid.New()
	repo := &stubFeedbackRepo{
		getReportFn: func(_ context.Context, _ uuid.UUID) (*feedbackdomain.Report, error) {
			return nil, feedbackdomain.ErrNotFound
		},
	}
	h := newAdminFeedbackHandlerForTest(repo)
	body := mustJSON(t, map[string]any{"body": "note body"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/admin/feedback/"+id.String()+"/notes", bytes.NewReader(body))
	r = withFeedbackChiID(adminCtx(r), id.String())
	w := httptest.NewRecorder()
	h.AddNote(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminFeedbackHandler_Update_400_MalformedBody(t *testing.T) {
	id := uuid.New()
	h := newAdminFeedbackHandlerForTest(&stubFeedbackRepo{})
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/feedback/"+id.String(), bytes.NewReader([]byte("{bad")))
	r = withFeedbackChiID(adminCtx(r), id.String())
	w := httptest.NewRecorder()
	h.Update(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminFeedbackHandler_Update_400_BadID(t *testing.T) {
	h := newAdminFeedbackHandlerForTest(&stubFeedbackRepo{})
	body := mustJSON(t, map[string]any{"status": "triaged"})
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/feedback/not-a-uuid", bytes.NewReader(body))
	r = withFeedbackChiID(adminCtx(r), "not-a-uuid")
	w := httptest.NewRecorder()
	h.Update(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminFeedbackHandler_AddNote_400_MalformedBody(t *testing.T) {
	id := uuid.New()
	h := newAdminFeedbackHandlerForTest(&stubFeedbackRepo{})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/admin/feedback/"+id.String()+"/notes", bytes.NewReader([]byte("{bad")))
	r = withFeedbackChiID(adminCtx(r), id.String())
	w := httptest.NewRecorder()
	h.AddNote(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminFeedbackHandler_AddNote_403_NonAdmin(t *testing.T) {
	h := newAdminFeedbackHandlerForTest(&stubFeedbackRepo{})
	id := uuid.New()
	body := mustJSON(t, map[string]any{"body": "note body"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/admin/feedback/"+id.String()+"/notes", bytes.NewReader(body))
	r = withFeedbackChiID(r, id.String()) // no admin ctx
	w := httptest.NewRecorder()
	h.AddNote(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
