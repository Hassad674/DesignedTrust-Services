package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	mediadomain "marketplace-backend/internal/domain/media"
	"marketplace-backend/internal/domain/referrerprofile"
	res "marketplace-backend/pkg/response"
)

// referrer_profile_video_presigned.go adds the DIRECT-to-R2 (presigned
// PUT) counterparts of the multipart referrer video endpoints in
// referrer_profile_video_handler.go. Mirrors the freelance presigned
// flow one-for-one, writing to referrer_profiles.video_url.

// referrerVideoKeyPrefix is the storage namespace for referrer persona
// videos. Matches the prefix segment of buildPersonaVideoKey for the
// referrer prefix ("profiles/<orgID>/referrer_video") so the moderation
// pipeline's extractStorageKey resolves the object identically.
func referrerVideoKeyPrefix(orgID uuid.UUID) string {
	return fmt.Sprintf("profiles/%s/referrer_video", orgID.String())
}

// PresignVideo handles POST /api/v1/referrer-profile/video/presign.
func (h *ReferrerProfileVideoHandler) PresignVideo(w http.ResponseWriter, r *http.Request) {
	_, orgID, ok := readVideoAuthContext(w, r)
	if !ok {
		return
	}
	req, ok := decodePresignVideoRequest(w, r)
	if !ok {
		return
	}
	uploadURL, fileKey, publicURL, err := issuePresignedVideo(
		r.Context(), h.storage, referrerVideoKeyPrefix(orgID), req.ContentType)
	if err != nil {
		slog.Error("referrer video presign failed", "error", err)
		res.Error(w, http.StatusInternalServerError, "presign_failed", "failed to create upload URL")
		return
	}
	writePresignResponse(w, uploadURL, fileKey, publicURL)
}

// CompleteVideo handles POST /api/v1/referrer-profile/video/complete.
// Mirrors the multipart Upload's persistence + moderation exactly.
func (h *ReferrerProfileVideoHandler) CompleteVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, orgID, ok := readVideoAuthContext(w, r)
	if !ok {
		return
	}
	req, ok := decodeCompleteVideoRequest(w, r)
	if !ok {
		return
	}
	if !verifyKeyNamespace(req.FileKey, referrerVideoKeyPrefix(orgID)) {
		res.Error(w, http.StatusForbidden, "invalid_file_key",
			"file_key does not belong to this upload namespace")
		return
	}
	url := h.storage.GetPublicURL(req.FileKey)

	h.deletePreviousObject(ctx, orgID, userID)

	if err := h.profiles.UpdateVideo(ctx, orgID, url); err != nil {
		slog.Error("referrer profile update video failed", "error", err, "user_id", userID)
		if errors.Is(err, referrerprofile.ErrProfileNotFound) {
			res.Error(w, http.StatusNotFound, "referrer_profile_not_found", "referrer profile not found")
			return
		}
		res.Error(w, http.StatusInternalServerError, "update_failed", "failed to update profile")
		return
	}

	if h.recorder != nil {
		bgCtx := context.WithoutCancel(r.Context())
		go h.recorder.RecordUpload( // #nosec G118 -- detached after request lifetime; RecordUpload applies its own 60s timeout
			bgCtx,
			userID, url, req.Filename, req.ContentType, req.FileSize,
			mediadomain.ContextReferrerVideo,
		)
	}
	res.JSON(w, http.StatusOK, map[string]string{"video_url": url})
}
