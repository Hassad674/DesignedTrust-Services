package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	feedbackapp "marketplace-backend/internal/app/feedback"
	feedbackdomain "marketplace-backend/internal/domain/feedback"
	"marketplace-backend/internal/handler/dto/request"
	"marketplace-backend/internal/handler/dto/response"
	"marketplace-backend/internal/handler/middleware"
	"marketplace-backend/internal/port/repository"
	jsondec "marketplace-backend/pkg/decode"
	res "marketplace-backend/pkg/response"
)

// AdminFeedbackHandler exposes the admin triage surface for platform
// feedback: list / detail / update / add-note. Gated by the /admin
// route group's RequireAdmin middleware AND a defensive per-handler
// admin check. Kept separate from FeedbackHandler so the public submit
// surface and the admin surface remain independently removable.
type AdminFeedbackHandler struct {
	svc *feedbackapp.Service
}

// NewAdminFeedbackHandler wires the admin feedback handler.
func NewAdminFeedbackHandler(svc *feedbackapp.Service) *AdminFeedbackHandler {
	return &AdminFeedbackHandler{svc: svc}
}

// adminFeedbackBodyCap bounds the PATCH / note JSON bodies. Notes are
// capped at 5 000 chars in the domain; 16 KiB at the transport layer
// leaves headroom while bounding a DoS.
const adminFeedbackBodyCap = 16 << 10

// List handles GET /api/v1/admin/feedback. Filters: type, status,
// severity; search on title; cursor pagination.
func (h *AdminFeedbackHandler) List(w http.ResponseWriter, r *http.Request) {
	if !middleware.GetIsAdmin(r.Context()) {
		res.Error(w, http.StatusForbidden, "forbidden", "admin access required")
		return
	}
	q := r.URL.Query()
	filter := repository.ReportFilter{
		Type:     q.Get("type"),
		Status:   q.Get("status"),
		Severity: q.Get("severity"),
		Search:   q.Get("search"),
	}
	cursor := q.Get("cursor")
	limit := parseLimit(q.Get("limit"), 20)

	summaries, nextCursor, err := h.svc.ListReports(r.Context(), filter, cursor, limit)
	if err != nil {
		handleFeedbackError(w, err, "admin list feedback")
		return
	}

	data := make([]response.AdminFeedbackReportResponse, 0, len(summaries))
	for _, s := range summaries {
		data = append(data, response.AdminFeedbackReportFrom(s.Report, s.AttachmentCount, s.NoteCount))
	}
	res.JSON(w, http.StatusOK, map[string]any{
		"data":        data,
		"next_cursor": nextCursor,
		"has_more":    nextCursor != "",
	})
}

// Get handles GET /api/v1/admin/feedback/{id}. Returns the full report
// with attachments (each carrying a presigned GET URL) and notes
// (newest first).
func (h *AdminFeedbackHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !middleware.GetIsAdmin(r.Context()) {
		res.Error(w, http.StatusForbidden, "forbidden", "admin access required")
		return
	}
	id, ok := parseFeedbackID(w, r)
	if !ok {
		return
	}
	detail, err := h.svc.GetReport(r.Context(), id)
	if err != nil {
		handleFeedbackError(w, err, "admin get feedback")
		return
	}
	res.JSON(w, http.StatusOK, buildReportDetailResponse(detail))
}

// Update handles PATCH /api/v1/admin/feedback/{id}. Updates status
// and/or severity; resolving stamps resolved_at/resolved_by.
func (h *AdminFeedbackHandler) Update(w http.ResponseWriter, r *http.Request) {
	adminID, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	id, ok := parseFeedbackID(w, r)
	if !ok {
		return
	}

	var req request.UpdateFeedbackReportRequest
	if err := jsondec.DecodeBody(w, r, &req, adminFeedbackBodyCap); err != nil {
		res.Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	update := feedbackdomain.Update{}
	if req.Status != nil {
		status := feedbackdomain.ReportStatus(*req.Status)
		update.Status = &status
	}
	if req.Severity != nil {
		severity := feedbackdomain.Severity(*req.Severity)
		update.Severity = &severity
	}

	report, err := h.svc.UpdateReport(r.Context(), id, adminID, update, remoteIPFromRequest(r))
	if err != nil {
		handleFeedbackError(w, err, "admin update feedback")
		return
	}
	res.JSON(w, http.StatusOK, response.AdminFeedbackReportFrom(report, 0, 0))
}

// AddNote handles POST /api/v1/admin/feedback/{id}/notes.
func (h *AdminFeedbackHandler) AddNote(w http.ResponseWriter, r *http.Request) {
	adminID, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	id, ok := parseFeedbackID(w, r)
	if !ok {
		return
	}

	var req request.AddFeedbackNoteRequest
	if err := jsondec.DecodeBody(w, r, &req, adminFeedbackBodyCap); err != nil {
		res.Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	note, err := h.svc.AddNote(r.Context(), id, adminID, req.Body, remoteIPFromRequest(r))
	if err != nil {
		handleFeedbackError(w, err, "admin add feedback note")
		return
	}
	res.JSON(w, http.StatusCreated, response.FeedbackNoteFrom(note))
}

// requireAdmin enforces the defensive per-handler admin check and
// returns the admin's user id. The /admin route group already gates on
// RequireAdmin; this is defense-in-depth for the mutation endpoints.
func (h *AdminFeedbackHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	if !middleware.GetIsAdmin(r.Context()) {
		res.Error(w, http.StatusForbidden, "forbidden", "admin access required")
		return uuid.Nil, false
	}
	adminID, ok := middleware.GetUserID(r.Context())
	if !ok || adminID == uuid.Nil {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return uuid.Nil, false
	}
	return adminID, true
}

// parseFeedbackID extracts and validates the {id} URL param.
func parseFeedbackID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		res.Error(w, http.StatusBadRequest, "invalid_id", "id must be a valid UUID")
		return uuid.Nil, false
	}
	return id, true
}

// buildReportDetailResponse assembles the admin detail envelope from the
// service's ReportDetail.
func buildReportDetailResponse(detail *feedbackapp.ReportDetail) response.AdminFeedbackReportDetailResponse {
	out := response.AdminFeedbackReportDetailResponse{
		AdminFeedbackReportResponse: response.AdminFeedbackReportFrom(
			detail.Report, len(detail.Attachments), len(detail.Notes)),
		Attachments: make([]response.FeedbackAttachmentResponse, 0, len(detail.Attachments)),
		Notes:       make([]response.FeedbackNoteResponse, 0, len(detail.Notes)),
	}
	for _, a := range detail.Attachments {
		out.Attachments = append(out.Attachments, response.FeedbackAttachmentResponse{
			ID:           a.Attachment.ID.String(),
			Kind:         string(a.Attachment.Kind),
			ContentType:  a.Attachment.ContentType,
			SizeBytes:    a.Attachment.SizeBytes,
			PresignedURL: a.PresignedURL,
			CreatedAt:    a.Attachment.CreatedAt,
		})
	}
	for _, n := range detail.Notes {
		out.Notes = append(out.Notes, response.FeedbackNoteFrom(n))
	}
	return out
}
