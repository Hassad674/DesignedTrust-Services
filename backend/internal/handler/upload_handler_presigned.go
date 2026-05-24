package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	mediadomain "marketplace-backend/internal/domain/media"
	"marketplace-backend/internal/handler/dto/request"
	"marketplace-backend/internal/handler/middleware"
	res "marketplace-backend/pkg/response"
)

// upload_handler_presigned.go adds the DIRECT-to-R2 (presigned PUT)
// counterparts of the legacy multipart video endpoints hosted in
// upload_handler_more.go. The bytes bypass the Vercel proxy (which caps
// bodies at ~4.5 MB) by uploading straight to the R2 origin; the
// backend only issues the URL and, on completion, persists the
// video_url AND fires the SAME moderation pipeline (trackUpload →
// RecordUpload) the multipart path uses. See upload_presigned.go for
// the rationale and the shared helpers.
//
// The multipart endpoints in upload_handler_more.go are kept intact for
// backward compatibility (agency photos, small files, non-web clients).

// PresignVideo issues a presigned PUT URL for the legacy agency
// presentation video (namespace profiles/<orgID>/video). Mirrors the
// key prefix used by UploadVideo so the moderation pipeline's
// extractStorageKey resolves the object identically.
func (h *UploadHandler) PresignVideo(w http.ResponseWriter, r *http.Request) {
	_, orgID, ok := h.presignVideoContext(w, r)
	if !ok {
		return
	}
	prefix := fmt.Sprintf("profiles/%s/video", orgID.String())
	h.issueAndRespond(w, r, prefix)
}

// CompleteVideo persists the legacy agency presentation video URL after
// the client PUT the bytes to R2, then triggers moderation. Reuses the
// exact persistence path of the multipart UploadVideo (GetByOrganizationID
// → Update) plus the shared trackUpload goroutine.
func (h *UploadHandler) CompleteVideo(w http.ResponseWriter, r *http.Request) {
	userID, orgID, ok := h.presignVideoContext(w, r)
	if !ok {
		return
	}
	req, url, ok := h.completeVideoCommon(w, r, fmt.Sprintf("profiles/%s/video", orgID.String()))
	if !ok {
		return
	}

	profile, err := h.profiles.GetByOrganizationID(r.Context(), orgID)
	if err != nil {
		slog.Error("get profile failed", "error", err, "user_id", userID)
		res.Error(w, http.StatusInternalServerError, "profile_error", "failed to get profile")
		return
	}
	profile.PresentationVideoURL = url
	if err := h.profiles.Update(r.Context(), profile); err != nil {
		slog.Error("update profile video failed", "error", err, "user_id", userID)
		res.Error(w, http.StatusInternalServerError, "update_failed", "failed to update profile")
		return
	}
	h.invalidateProfileCache(r.Context(), orgID)

	h.trackUpload(r.Context(), trackUploadInput{
		UploaderID: userID,
		FileURL:    url,
		FileName:   req.Filename,
		FileType:   req.ContentType,
		FileSize:   req.FileSize,
		MediaCtx:   mediadomain.ContextProfileVideo,
	})
	res.JSON(w, http.StatusOK, map[string]string{"url": url})
}

// PresignReferrerVideo issues a presigned PUT URL for the legacy agency
// referrer video (namespace profiles/<orgID>/referrer_video).
func (h *UploadHandler) PresignReferrerVideo(w http.ResponseWriter, r *http.Request) {
	_, orgID, ok := h.presignVideoContext(w, r)
	if !ok {
		return
	}
	prefix := fmt.Sprintf("profiles/%s/referrer_video", orgID.String())
	h.issueAndRespond(w, r, prefix)
}

// CompleteReferrerVideo persists the legacy agency referrer video URL
// then triggers moderation. Mirrors the multipart UploadReferrerVideo.
func (h *UploadHandler) CompleteReferrerVideo(w http.ResponseWriter, r *http.Request) {
	userID, orgID, ok := h.presignVideoContext(w, r)
	if !ok {
		return
	}
	req, url, ok := h.completeVideoCommon(w, r, fmt.Sprintf("profiles/%s/referrer_video", orgID.String()))
	if !ok {
		return
	}

	profile, err := h.profiles.GetByOrganizationID(r.Context(), orgID)
	if err != nil {
		slog.Error("get profile failed", "error", err, "user_id", userID)
		res.Error(w, http.StatusInternalServerError, "profile_error", "failed to get profile")
		return
	}
	profile.ReferrerVideoURL = url
	if err := h.profiles.Update(r.Context(), profile); err != nil {
		slog.Error("update profile referrer video failed", "error", err, "user_id", userID)
		res.Error(w, http.StatusInternalServerError, "update_failed", "failed to update profile")
		return
	}
	h.invalidateProfileCache(r.Context(), orgID)

	h.trackUpload(r.Context(), trackUploadInput{
		UploaderID: userID,
		FileURL:    url,
		FileName:   req.Filename,
		FileType:   req.ContentType,
		FileSize:   req.FileSize,
		MediaCtx:   mediadomain.ContextReferrerVideo,
	})
	res.JSON(w, http.StatusOK, map[string]string{"url": url})
}

// PresignPortfolioVideo issues a presigned PUT URL for a portfolio
// video (namespace portfolios/<userID>/video — portfolio is per-user,
// matching the multipart UploadPortfolioVideo).
func (h *UploadHandler) PresignPortfolioVideo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "user not found")
		return
	}
	prefix := fmt.Sprintf("portfolios/%s/video", userID.String())
	h.issueAndRespond(w, r, prefix)
}

// CompletePortfolioVideo triggers moderation for a portfolio video. No
// handler-side DB write: the caller persists the returned URL onto the
// portfolio item it owns (mirrors the multipart handler, which also
// only returns the URL + fires trackUpload).
func (h *UploadHandler) CompletePortfolioVideo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "user not found")
		return
	}
	req, url, ok := h.completeVideoCommon(w, r, fmt.Sprintf("portfolios/%s/video", userID.String()))
	if !ok {
		return
	}
	h.trackUpload(r.Context(), trackUploadInput{
		UploaderID: userID,
		FileURL:    url,
		FileName:   req.Filename,
		FileType:   req.ContentType,
		FileSize:   req.FileSize,
		MediaCtx:   mediadomain.ContextPortfolioVideo,
	})
	res.JSON(w, http.StatusOK, map[string]string{"url": url})
}

// PresignReviewVideo issues a presigned PUT URL for a review video
// (namespace reviews/<userID>/video — review video is per-user,
// matching the multipart UploadReviewVideo).
func (h *UploadHandler) PresignReviewVideo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "user not found")
		return
	}
	prefix := fmt.Sprintf("reviews/%s/video", userID.String())
	h.issueAndRespond(w, r, prefix)
}

// CompleteReviewVideo triggers moderation for a review video. No
// handler-side DB write: the returned URL is carried into the review
// create payload (video_url), mirroring the multipart UploadReviewVideo.
func (h *UploadHandler) CompleteReviewVideo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "user not found")
		return
	}
	req, url, ok := h.completeVideoCommon(w, r, fmt.Sprintf("reviews/%s/video", userID.String()))
	if !ok {
		return
	}
	h.trackUpload(r.Context(), trackUploadInput{
		UploaderID: userID,
		FileURL:    url,
		FileName:   req.Filename,
		FileType:   req.ContentType,
		FileSize:   req.FileSize,
		MediaCtx:   mediadomain.ContextReviewVideo,
	})
	res.JSON(w, http.StatusOK, map[string]string{"url": url})
}

// presignVideoContext reads the user + org IDs (both required for the
// profile-scoped video surfaces) and writes a 401 on failure.
func (h *UploadHandler) presignVideoContext(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "user not found")
		return uuid.Nil, uuid.Nil, false
	}
	orgID, ok := middleware.GetOrganizationID(r.Context())
	if !ok {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "organization not found in context")
		return uuid.Nil, uuid.Nil, false
	}
	return userID, orgID, true
}

// issueAndRespond decodes the presign request, mints the namespaced key,
// asks storage for a presigned PUT URL, and writes the standard
// envelope. Shared by every UploadHandler presign endpoint.
func (h *UploadHandler) issueAndRespond(w http.ResponseWriter, r *http.Request, keyPrefix string) {
	req, ok := decodePresignVideoRequest(w, r)
	if !ok {
		return
	}
	uploadURL, fileKey, publicURL, err := issuePresignedVideo(r.Context(), h.storage, keyPrefix, req.ContentType)
	if err != nil {
		slog.Error("presign video upload failed", "error", err)
		res.Error(w, http.StatusInternalServerError, "presign_failed", "failed to create upload URL")
		return
	}
	writePresignResponse(w, uploadURL, fileKey, publicURL)
}

// completeVideoCommon decodes the complete request, re-verifies the
// file_key sits under the caller's namespace (ownership guard), and
// resolves the public URL the object is served from. Returns the
// request + public URL on success. Writes the error response and
// returns ok=false on any failure.
func (h *UploadHandler) completeVideoCommon(
	w http.ResponseWriter,
	r *http.Request,
	expectedPrefix string,
) (request.CompleteVideoUploadRequest, string, bool) {
	req, ok := decodeCompleteVideoRequest(w, r)
	if !ok {
		return req, "", false
	}
	if !verifyKeyNamespace(req.FileKey, expectedPrefix) {
		res.Error(w, http.StatusForbidden, "invalid_file_key",
			"file_key does not belong to this upload namespace")
		return req, "", false
	}
	return req, h.storage.GetPublicURL(req.FileKey), true
}
