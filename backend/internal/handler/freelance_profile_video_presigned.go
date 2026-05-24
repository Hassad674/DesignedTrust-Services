package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"marketplace-backend/internal/domain/freelanceprofile"
	mediadomain "marketplace-backend/internal/domain/media"
	res "marketplace-backend/pkg/response"
)

// freelance_profile_video_presigned.go adds the DIRECT-to-R2 (presigned
// PUT) counterparts of the multipart freelance video endpoints in
// freelance_profile_video_handler.go. The bytes bypass the Vercel proxy
// body cap by uploading straight to R2; the backend issues the URL and,
// on completion, persists video_url via UpdateVideo AND fires the SAME
// detached RecordUpload moderation goroutine the multipart Upload uses.

// freelanceVideoKeyPrefix is the storage namespace for freelance persona
// videos. Matches the prefix segment of buildPersonaVideoKey
// ("profiles/<orgID>/video") so the moderation pipeline's
// extractStorageKey resolves the object identically.
func freelanceVideoKeyPrefix(orgID uuid.UUID) string {
	return fmt.Sprintf("profiles/%s/video", orgID.String())
}

// PresignVideo handles POST /api/v1/freelance-profile/video/presign.
func (h *FreelanceProfileVideoHandler) PresignVideo(w http.ResponseWriter, r *http.Request) {
	_, orgID, ok := readVideoAuthContext(w, r)
	if !ok {
		return
	}
	req, ok := decodePresignVideoRequest(w, r)
	if !ok {
		return
	}
	uploadURL, fileKey, publicURL, err := issuePresignedVideo(
		r.Context(), h.storage, freelanceVideoKeyPrefix(orgID), req.ContentType)
	if err != nil {
		slog.Error("freelance video presign failed", "error", err)
		res.Error(w, http.StatusInternalServerError, "presign_failed", "failed to create upload URL")
		return
	}
	writePresignResponse(w, uploadURL, fileKey, publicURL)
}

// CompleteVideo handles POST /api/v1/freelance-profile/video/complete.
// Mirrors the multipart Upload's persistence + moderation exactly: it
// deletes the previous object, writes the new video_url (UpdateVideo)
// and fires the detached RecordUpload goroutine.
func (h *FreelanceProfileVideoHandler) CompleteVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, orgID, ok := readVideoAuthContext(w, r)
	if !ok {
		return
	}
	req, ok := decodeCompleteVideoRequest(w, r)
	if !ok {
		return
	}
	if !verifyKeyNamespace(req.FileKey, freelanceVideoKeyPrefix(orgID)) {
		res.Error(w, http.StatusForbidden, "invalid_file_key",
			"file_key does not belong to this upload namespace")
		return
	}
	url := h.storage.GetPublicURL(req.FileKey)

	// Best-effort delete of the previous object before stamping the new
	// URL — identical to the multipart Upload path.
	h.deletePreviousObject(ctx, orgID, userID)

	if err := h.profiles.UpdateVideo(ctx, orgID, url); err != nil {
		slog.Error("freelance profile update video failed", "error", err, "user_id", userID)
		if errors.Is(err, freelanceprofile.ErrProfileNotFound) {
			res.Error(w, http.StatusNotFound, "freelance_profile_not_found", "freelance profile not found")
			return
		}
		res.Error(w, http.StatusInternalServerError, "update_failed", "failed to update profile")
		return
	}

	if h.recorder != nil {
		// Detach from the request lifetime so the moderation pipeline
		// survives the response — identical to the multipart Upload.
		bgCtx := context.WithoutCancel(r.Context())
		go h.recorder.RecordUpload( // #nosec G118 -- detached after request lifetime; RecordUpload applies its own 60s timeout
			bgCtx,
			userID, url, req.Filename, req.ContentType, req.FileSize,
			mediadomain.ContextProfileVideo,
		)
	}
	res.JSON(w, http.StatusOK, map[string]string{"video_url": url})
}
