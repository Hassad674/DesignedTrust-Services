package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	feedbackapp "marketplace-backend/internal/app/feedback"
	feedbackdomain "marketplace-backend/internal/domain/feedback"
	"marketplace-backend/internal/handler/dto/request"
	"marketplace-backend/internal/handler/dto/response"
	"marketplace-backend/internal/handler/middleware"
	jsondec "marketplace-backend/pkg/decode"
	res "marketplace-backend/pkg/response"
	"marketplace-backend/pkg/sanitize"
)

// FeedbackHandler exposes the public feedback surface: submit a bug /
// security report (anonymous or authenticated) and presign a media
// upload (authenticated only). Admin triage lives in
// AdminFeedbackHandler so each surface stays independently removable.
type FeedbackHandler struct {
	svc *feedbackapp.Service
}

// NewFeedbackHandler wires the public feedback handler.
func NewFeedbackHandler(svc *feedbackapp.Service) *FeedbackHandler {
	return &FeedbackHandler{svc: svc}
}

// submitFeedbackMaxBody bounds the submit JSON body. Generous enough for
// the 5 000-char description + a handful of attachment references, tight
// enough to short-circuit a DoS via an unbounded body.
const submitFeedbackMaxBody = 32 << 10

// presignFeedbackMaxBody bounds the presign JSON body — a tiny envelope
// (kind + content type + size + filename), never file bytes.
const presignFeedbackMaxBody = 2 << 10

// Submit handles POST /api/v1/feedback. Anonymous is allowed; a
// logged-in caller is recognised via the optional-auth middleware and
// may additionally attach media. The submitter IP is hashed before
// storage (RGPD) inside the service.
//
//	201 -> report accepted (envelope with id/type/status)
//	200 -> honeypot tripped (silently dropped, looks like success)
//	400 -> validation / attachment-not-allowed / bad attachment
func (h *FeedbackHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req request.SubmitFeedbackRequest
	if err := jsondec.DecodeBody(w, r, &req, submitFeedbackMaxBody); err != nil {
		res.Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	// Honeypot: a hidden field a human never fills. A bot that auto-fills
	// every input trips it. Return 200 with a synthetic id so the bot
	// cannot distinguish a drop from a real accept.
	if req.HP != "" {
		res.JSON(w, http.StatusOK, response.SubmitFeedbackResponse{
			ID:     uuid.New().String(),
			Type:   req.Type,
			Status: string(feedbackdomain.StatusNew),
		})
		return
	}

	// Recognise the optional authenticated identity. Anonymous → nil.
	var reporterID *uuid.UUID
	if uid, ok := middleware.GetUserID(r.Context()); ok && uid != uuid.Nil {
		copied := uid
		reporterID = &copied
	}

	in := feedbackapp.SubmitReportInput{
		ReporterID: reporterID,
		Type:       req.Type,
		// Strip HTML from the free-text fields at the boundary to prevent
		// stored XSS before the domain enforces the length bounds.
		Title:         sanitize.StripHTML(req.Title),
		Description:   sanitize.StripHTML(req.Description),
		PageURL:       req.PageURL,
		Context:       buildFeedbackContext(req.Context),
		ReporterEmail: req.ReporterEmail,
		RawIP:         remoteIPFromRequest(r),
		Attachments:   toAttachmentRefs(reporterID, req.AttachmentKeys),
	}

	report, err := h.svc.SubmitReport(r.Context(), in)
	if err != nil {
		handleFeedbackError(w, err, "submit feedback")
		return
	}
	res.JSON(w, http.StatusCreated, response.SubmitFeedbackResponseFrom(report))
}

// PresignAttachment handles POST /api/v1/feedback/attachments/presign.
// AUTH REQUIRED (mounted behind AuthFromDeps) — media is logged-in only,
// so an anonymous caller is rejected with 401 by the middleware before
// reaching here. Validates the content-type allowlist + size cap and
// returns a short-lived PUT URL plus the server-minted object key.
func (h *FeedbackHandler) PresignAttachment(w http.ResponseWriter, r *http.Request) {
	// Defensive identity check: the route is auth-gated, but never trust
	// the wiring alone for a media-issuing endpoint.
	if _, ok := middleware.GetUserID(r.Context()); !ok {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req request.PresignFeedbackAttachmentRequest
	if err := jsondec.DecodeBody(w, r, &req, presignFeedbackMaxBody); err != nil {
		res.Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	result, err := h.svc.PresignAttachment(r.Context(), feedbackapp.PresignInput{
		Kind:        req.Kind,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
	})
	if err != nil {
		handleFeedbackError(w, err, "presign feedback attachment")
		return
	}
	res.JSON(w, http.StatusOK, response.PresignFeedbackAttachmentResponse{
		UploadURL: result.UploadURL,
		ObjectKey: result.ObjectKey,
		Kind:      result.Kind,
	})
}

// buildFeedbackContext converts the typed request context into the
// free-form map the domain carries. Returns an empty map when the
// client sent no context. Empty fields are omitted so the stored JSONB
// stays compact.
func buildFeedbackContext(c *request.FeedbackContext) map[string]any {
	out := map[string]any{}
	if c == nil {
		return out
	}
	putIfSet(out, "role", c.Role)
	putIfSet(out, "locale", c.Locale)
	putIfSet(out, "platform", c.Platform)
	putIfSet(out, "app_version", c.AppVersion)
	putIfSet(out, "viewport", c.Viewport)
	putIfSet(out, "user_agent", c.UserAgent)
	return out
}

func putIfSet(m map[string]any, key, value string) {
	if value != "" {
		m[key] = value
	}
}

// toAttachmentRefs maps the request attachment list to service refs.
// Returns nil for an anonymous caller so the service's logged-in-only
// rule fires cleanly (an anonymous payload with attachments is rejected
// by the service rather than silently honoured here).
func toAttachmentRefs(reporterID *uuid.UUID, in []request.SubmitFeedbackAttachment) []feedbackapp.AttachmentRef {
	if len(in) == 0 {
		return nil
	}
	out := make([]feedbackapp.AttachmentRef, 0, len(in))
	for _, a := range in {
		out = append(out, feedbackapp.AttachmentRef{
			Kind:        a.Kind,
			ObjectKey:   a.ObjectKey,
			ContentType: a.ContentType,
			SizeBytes:   a.SizeBytes,
		})
	}
	return out
}

// handleFeedbackError maps feedback domain errors to HTTP status codes.
// Validation errors → 400, not-found → 404, attachment-not-allowed →
// 400; anything else is logged and returned as a generic 500.
func handleFeedbackError(w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, feedbackdomain.ErrNotFound):
		res.Error(w, http.StatusNotFound, "not_found", "report not found")
	case isFeedbackValidationError(err):
		res.Error(w, http.StatusBadRequest, "validation_error", err.Error())
	default:
		slog.Error(op, "error", err)
		res.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

// isFeedbackValidationError reports whether err is one of the feedback
// domain's client-fault sentinels (4xx) as opposed to an infrastructure
// failure (5xx).
func isFeedbackValidationError(err error) bool {
	for _, sentinel := range []error{
		feedbackdomain.ErrInvalidType,
		feedbackdomain.ErrTitleTooShort,
		feedbackdomain.ErrTitleTooLong,
		feedbackdomain.ErrDescriptionTooShort,
		feedbackdomain.ErrDescriptionTooLong,
		feedbackdomain.ErrPageURLTooLong,
		feedbackdomain.ErrReporterEmailTooLong,
		feedbackdomain.ErrInvalidStatus,
		feedbackdomain.ErrInvalidSeverity,
		feedbackdomain.ErrNoUpdateFields,
		feedbackdomain.ErrInvalidAttachmentKind,
		feedbackdomain.ErrInvalidAttachmentSize,
		feedbackdomain.ErrAttachmentNotAllowed,
		feedbackdomain.ErrUnsupportedContentType,
		feedbackdomain.ErrAttachmentTooLarge,
		feedbackdomain.ErrNoteBodyEmpty,
		feedbackdomain.ErrNoteBodyTooLong,
		feedbackdomain.ErrMissingReport,
		feedbackdomain.ErrMissingObjectKey,
		feedbackdomain.ErrMissingAdmin,
	} {
		if errors.Is(err, sentinel) {
			return true
		}
	}
	return false
}
