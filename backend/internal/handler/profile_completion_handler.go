package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"marketplace-backend/internal/app/profilecompletion"
	"marketplace-backend/internal/handler/middleware"
	res "marketplace-backend/pkg/response"
)

// ProfileCompletionService is the narrow contract the handler depends
// on. Defined locally so the handler does not pull a concrete service
// type into its surface — any value (concrete service, decorator, mock)
// that returns a Report for (userID, orgID) satisfies it.
type ProfileCompletionService interface {
	Compute(ctx context.Context, userID, orgID uuid.UUID) (*profilecompletion.Report, error)
}

// ProfileCompletionHandler exposes GET /api/v1/me/profile/completion.
// Auth-required: the caller is identified through the JWT context, the
// org is read from the same context, and the report is computed from
// the existing readers (no mutation, no background work).
//
// Cache-Control is private + max-age=30 so a freshly-loaded sidebar
// avoids hammering the endpoint when the user navigates between
// pages, while still picking up updates within half a minute after a
// section save (the frontend additionally invalidates the query when
// it detects a relevant write — that is the fast path).
type ProfileCompletionHandler struct {
	svc ProfileCompletionService
}

// NewProfileCompletionHandler wires the handler with its single
// dependency. Returning a pointer keeps the symmetry with every other
// handler in this package; the field is otherwise immutable.
func NewProfileCompletionHandler(svc ProfileCompletionService) *ProfileCompletionHandler {
	return &ProfileCompletionHandler{svc: svc}
}

// GetMyCompletion handles the GET /api/v1/me/profile/completion route.
// The response is the JSON-serialized Report — every field is
// non-optional so the frontend never has to write defensive code.
func (h *ProfileCompletionHandler) GetMyCompletion(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == uuid.Nil {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "user not found in context")
		return
	}
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok || orgID == uuid.Nil {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "organization not found in context")
		return
	}

	report, err := h.svc.Compute(r.Context(), userID, orgID)
	if err != nil {
		res.Error(w, http.StatusInternalServerError, "internal_error", "failed to compute profile completion")
		return
	}

	w.Header().Set("Cache-Control", "private, max-age=30")
	res.JSON(w, http.StatusOK, report)
}
