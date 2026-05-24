package handler

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	"marketplace-backend/internal/handler/dto/request"
	"marketplace-backend/internal/handler/dto/response"
	portservice "marketplace-backend/internal/port/service"
	res "marketplace-backend/pkg/response"
	"marketplace-backend/pkg/validator"
)

// upload_presigned.go hosts the shared, transport-level helpers for the
// DIRECT-to-R2 video upload flow.
//
// WHY THIS EXISTS — the same-origin upload path
// (services.designedtrust.com/api/v1/upload/...) is proxied to the Go
// backend by Next.js `rewrites()` through Vercel, whose serverless
// proxy caps request bodies at ~4.5 MB. Videos > 4.5 MB therefore got a
// 413 from the edge before ever reaching the backend (which allows 50–
// 100 MB). The fix mirrors the messaging attachment flow exactly: the
// browser asks the backend for a short-lived presigned PUT URL (a tiny
// JSON request, no body cap) then PUTs the bytes DIRECTLY to the R2
// origin (no Vercel proxy, no cap). The backend then "completes" the
// upload by persisting the URL AND triggering the SAME moderation
// pipeline that the multipart handlers fire after storage.Upload.
//
// The presign step is the single security choke-point for the direct
// flow: it derives the storage key (extension included) from a
// server-side video allowlist, so a hostile client cannot choose the
// stored extension or path. Ownership is re-verified at the complete
// step by matching the returned file_key against the caller's namespace
// prefix — a client that tampered with the key it PUTs to cannot make
// the backend persist or moderate an object outside its own namespace.

// presignedVideoExpiry bounds the lifetime of an issued PUT URL. Short
// enough to limit replay, long enough for a 50–100 MB upload on a slow
// uplink. Mirrors the messaging attachment expiry.
const presignedVideoExpiry = 15 * time.Minute

// videoContentTypeExt maps the allowlisted video MIME types to the safe
// file extension the SERVER stamps onto the storage key. The client's
// declared content type is validated against this map; anything else is
// rejected at the presign boundary. Mirrors the ScopeVideo allowlist in
// detectMimeFromBytes (mp4 / webm / quicktime) so the direct flow can
// never store an extension the multipart flow would have refused.
var videoContentTypeExt = map[string]string{
	"video/mp4":        "mp4",
	"video/webm":       "webm",
	"video/quicktime":  "mov",
	"video/x-matroska": "mkv",
}

// resolveVideoExt returns the server-controlled extension for an
// allowlisted video content type, plus ok=false when the type is not a
// permitted video type. The lookup is case-insensitive on the MIME and
// trims any "; codecs=..." parameter the browser may append.
func resolveVideoExt(contentType string) (string, bool) {
	base := strings.TrimSpace(strings.ToLower(contentType))
	if idx := strings.IndexByte(base, ';'); idx >= 0 {
		base = strings.TrimSpace(base[:idx])
	}
	ext, ok := videoContentTypeExt[base]
	if !ok {
		return "", false
	}
	return ext, true
}

// issuePresignedVideo builds a namespaced, randomized storage key for a
// video and asks the storage port for a short-lived presigned PUT URL.
// keyPrefix is the caller's namespace (e.g. "profiles/<orgID>/video");
// the final key is "<keyPrefix>/<uuid>.<extFromContentType>". Returns
// the upload URL (PUT target), the storage key, and the public URL the
// object will be served from once uploaded.
//
// The function is the direct-flow analogue of validateAndBuildKey: it
// is the single place the storage key is minted, so the served
// extension is always derived from the validated content type — never
// from a client-supplied filename.
func issuePresignedVideo(
	ctx context.Context,
	storage portservice.StorageService,
	keyPrefix string,
	contentType string,
) (uploadURL, fileKey, publicURL string, err error) {
	ext, ok := resolveVideoExt(contentType)
	if !ok {
		return "", "", "", errUnsupportedVideoType
	}
	fileKey = fmt.Sprintf("%s/%s.%s", keyPrefix, uuid.New().String(), ext)
	uploadURL, err = storage.GetPresignedUploadURL(ctx, fileKey, contentType, presignedVideoExpiry)
	if err != nil {
		return "", "", "", fmt.Errorf("presign video upload: %w", err)
	}
	publicURL = storage.GetPublicURL(fileKey)
	return uploadURL, fileKey, publicURL, nil
}

// errUnsupportedVideoType is returned by issuePresignedVideo when the
// client-declared content type is not an allowlisted video type.
var errUnsupportedVideoType = fmt.Errorf("unsupported video content type")

// decodePresignVideoRequest decodes + validates the presign request
// body, returning the (filename, contentType) pair. Writes the error
// response and returns ok=false on any failure so the caller can early
// return. The filename is accepted only for the human-facing media row
// (FileName) — it never influences the storage key.
func decodePresignVideoRequest(w http.ResponseWriter, r *http.Request) (req request.PresignVideoRequest, ok bool) {
	if err := decodeVideoJSON(r, &req); err != nil {
		res.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return req, false
	}
	if _, valid := resolveVideoExt(req.ContentType); !valid {
		res.Error(w, http.StatusUnsupportedMediaType, "invalid_type",
			"content_type must be one of video/mp4, video/webm, video/quicktime, video/x-matroska")
		return req, false
	}
	return req, true
}

// decodeCompleteVideoRequest decodes + validates the complete request
// body. Writes the error response and returns ok=false on failure.
func decodeCompleteVideoRequest(w http.ResponseWriter, r *http.Request) (req request.CompleteVideoUploadRequest, ok bool) {
	if err := decodeVideoJSON(r, &req); err != nil {
		res.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return req, false
	}
	if _, valid := resolveVideoExt(req.ContentType); !valid {
		res.Error(w, http.StatusUnsupportedMediaType, "invalid_type",
			"content_type must be one of video/mp4, video/webm, video/quicktime, video/x-matroska")
		return req, false
	}
	return req, true
}

// verifyKeyNamespace re-checks at the complete step that file_key sits
// under the caller's expected namespace prefix AND carries an
// allowlisted video extension. This is the ownership guard for the
// direct flow: a client that PUT to a tampered key (e.g. another org's
// namespace, or a `.html` suffix) cannot make the backend persist or
// moderate that object. Returns false on any mismatch.
func verifyKeyNamespace(fileKey, expectedPrefix string) bool {
	if fileKey == "" || expectedPrefix == "" {
		return false
	}
	if !strings.HasPrefix(fileKey, expectedPrefix+"/") {
		return false
	}
	// Reject path traversal and nested segments beyond the single
	// "<prefix>/<file>" shape the presign step mints.
	rest := strings.TrimPrefix(fileKey, expectedPrefix+"/")
	if rest == "" || strings.Contains(rest, "/") || strings.Contains(rest, "..") {
		return false
	}
	ext := strings.TrimPrefix(strings.ToLower(path.Ext(rest)), ".")
	for _, allowed := range videoContentTypeExt {
		if ext == allowed {
			return true
		}
	}
	return false
}

// writePresignResponse emits the standard presign envelope shared by all
// video surfaces, mirroring the messaging PresignedURLResponse shape.
func writePresignResponse(w http.ResponseWriter, uploadURL, fileKey, publicURL string) {
	res.JSON(w, http.StatusOK, response.PresignedURLResponse{
		UploadURL: uploadURL,
		FileKey:   fileKey,
		PublicURL: publicURL,
	})
}

// decodeVideoJSON caps the presign/complete bodies at a tight 2 KiB —
// these are small JSON envelopes (filename + content type + size), never
// file bytes. Rejects unknown fields (DisallowUnknownFields) then runs
// the struct `validate:` tags. Uses validator.DecodeJSONWithCap so the
// decode_sweep_test guardrail stays satisfied (no raw json.NewDecoder).
func decodeVideoJSON(r *http.Request, dst any) error {
	if err := validator.DecodeJSONWithCap(nil, r, dst, 2<<10); err != nil {
		return err
	}
	return validator.Validate(dst)
}
