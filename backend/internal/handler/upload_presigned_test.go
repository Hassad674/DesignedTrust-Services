package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mediadomain "marketplace-backend/internal/domain/media"
	"marketplace-backend/internal/domain/profile"
	"marketplace-backend/internal/handler/middleware"
)

// ---------------------------------------------------------------------------
// Shared-helper unit tests
// ---------------------------------------------------------------------------

func TestResolveVideoExt(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantExt     string
		wantOK      bool
	}{
		{"mp4", "video/mp4", "mp4", true},
		{"webm", "video/webm", "webm", true},
		{"quicktime mov", "video/quicktime", "mov", true},
		{"matroska", "video/x-matroska", "mkv", true},
		{"mp4 with codecs param", "video/mp4; codecs=avc1.42E01E", "mp4", true},
		{"uppercase", "VIDEO/MP4", "mp4", true},
		{"image rejected", "image/png", "", false},
		{"html rejected", "text/html", "", false},
		{"svg rejected", "image/svg+xml", "", false},
		{"empty rejected", "", "", false},
		{"octet-stream rejected", "application/octet-stream", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, ok := resolveVideoExt(tt.contentType)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantExt, ext)
		})
	}
}

func TestVerifyKeyNamespace(t *testing.T) {
	orgPrefix := "profiles/" + uuid.New().String() + "/video"
	tests := []struct {
		name    string
		fileKey string
		prefix  string
		want    bool
	}{
		{"valid mp4 in namespace", orgPrefix + "/" + uuid.New().String() + ".mp4", orgPrefix, true},
		{"valid webm in namespace", orgPrefix + "/" + uuid.New().String() + ".webm", orgPrefix, true},
		{"valid mov in namespace", orgPrefix + "/" + uuid.New().String() + ".mov", orgPrefix, true},
		{"foreign org prefix", "profiles/" + uuid.New().String() + "/video/x.mp4", orgPrefix, false},
		{"disallowed extension html", orgPrefix + "/x.html", orgPrefix, false},
		{"no extension", orgPrefix + "/x", orgPrefix, false},
		{"path traversal", orgPrefix + "/../secret.mp4", orgPrefix, false},
		{"nested segment", orgPrefix + "/a/b.mp4", orgPrefix, false},
		{"empty key", "", orgPrefix, false},
		{"empty prefix", orgPrefix + "/x.mp4", "", false},
		{"prefix-as-key no file", orgPrefix, orgPrefix, false},
		{"sibling prefix collision", orgPrefix + "_evil/x.mp4", orgPrefix, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, verifyKeyNamespace(tt.fileKey, tt.prefix))
		})
	}
}

// ---------------------------------------------------------------------------
// Request body helpers
// ---------------------------------------------------------------------------

func presignBody(t *testing.T, filename, contentType string) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(map[string]any{"filename": filename, "content_type": contentType})
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func completeBody(t *testing.T, fileKey, filename, contentType string, size int64) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(map[string]any{
		"file_key": fileKey, "filename": filename, "content_type": contentType, "file_size": size,
	})
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func jsonRequest(method, url string, body *bytes.Buffer) *http.Request {
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// ---------------------------------------------------------------------------
// UploadHandler.PresignVideo — happy path + validation
// ---------------------------------------------------------------------------

func TestUploadHandler_PresignVideo_Success(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()

	var presignedKey, presignedCT string
	storage := &mockStorageService{
		getPresignedUploadFn: func(_ context.Context, key, ct string, expiry time.Duration) (string, error) {
			presignedKey, presignedCT = key, ct
			assert.Equal(t, presignedVideoExpiry, expiry)
			return "https://r2.example.com/put/" + key + "?sig=abc", nil
		},
		getPublicURLFn: func(key string) string { return "https://cdn.example.com/" + key },
	}
	h := NewUploadHandler(storage, &mockProfileRepo{}, nil)

	req := jsonRequest(http.MethodPost, "/api/v1/upload/video/presign", presignBody(t, "intro.mov", "video/mp4"))
	req = withVideoCtx(req, uid, orgID)
	rec := httptest.NewRecorder()
	h.PresignVideo(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	// Key is server-built, namespaced by org, extension from content-type
	// (mp4) — NOT from the client filename (.mov).
	assert.True(t, strings.HasPrefix(presignedKey, "profiles/"+orgID.String()+"/video/"),
		"key must be namespaced under the org video prefix, got %q", presignedKey)
	assert.True(t, strings.HasSuffix(presignedKey, ".mp4"),
		"extension must derive from content-type, got %q", presignedKey)
	assert.Equal(t, "video/mp4", presignedCT)
	assert.Contains(t, resp["upload_url"], "https://r2.example.com/put/")
	assert.Equal(t, presignedKey, resp["file_key"])
	assert.Equal(t, "https://cdn.example.com/"+presignedKey, resp["public_url"])
}

func TestUploadHandler_PresignVideo_RejectsNonVideoContentType(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	h := NewUploadHandler(&mockStorageService{}, &mockProfileRepo{}, nil)

	req := jsonRequest(http.MethodPost, "/api/v1/upload/video/presign", presignBody(t, "x.png", "image/png"))
	req = withVideoCtx(req, uid, orgID)
	rec := httptest.NewRecorder()
	h.PresignVideo(rec, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "invalid_type", resp["error"])
}

func TestUploadHandler_PresignVideo_Unauthorized(t *testing.T) {
	h := NewUploadHandler(&mockStorageService{}, &mockProfileRepo{}, nil)
	req := jsonRequest(http.MethodPost, "/api/v1/upload/video/presign", presignBody(t, "x.mp4", "video/mp4"))
	rec := httptest.NewRecorder()
	h.PresignVideo(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUploadHandler_PresignVideo_RejectsUnknownField(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	h := NewUploadHandler(&mockStorageService{}, &mockProfileRepo{}, nil)

	body := bytes.NewBufferString(`{"filename":"x.mp4","content_type":"video/mp4","evil":"y"}`)
	req := jsonRequest(http.MethodPost, "/api/v1/upload/video/presign", body)
	req = withVideoCtx(req, uid, orgID)
	rec := httptest.NewRecorder()
	h.PresignVideo(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ---------------------------------------------------------------------------
// UploadHandler.CompleteVideo — persistence + moderation + ownership
// ---------------------------------------------------------------------------

func TestUploadHandler_CompleteVideo_PersistsAndModerates(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	fileKey := "profiles/" + orgID.String() + "/video/" + uuid.New().String() + ".mp4"

	var savedVideoURL string
	storage := &mockStorageService{
		getPublicURLFn: func(key string) string { return "https://cdn.example.com/" + key },
	}
	repo := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, id uuid.UUID) (*profile.Profile, error) {
			return &profile.Profile{OrganizationID: id}, nil
		},
		updateFn: func(_ context.Context, p *profile.Profile) error {
			savedVideoURL = p.PresentationVideoURL
			return nil
		},
	}
	h := NewUploadHandler(storage, repo, nil)
	rec := newFakeRecorder()
	h.recorder = rec
	defer func() { _ = h.Stop(context.Background()) }()

	req := jsonRequest(http.MethodPost, "/api/v1/upload/video/complete",
		completeBody(t, fileKey, "intro.mp4", "video/mp4", 12_000_000))
	req = withVideoCtx(req, uid, orgID)
	w := httptest.NewRecorder()
	h.CompleteVideo(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	wantURL := "https://cdn.example.com/" + fileKey
	assert.Equal(t, wantURL, resp["url"])
	assert.Equal(t, wantURL, savedVideoURL, "video URL must be persisted to the profile")

	// Moderation pipeline must have fired exactly once with the right context.
	require.NoError(t, h.Stop(context.Background()))
	rec.mu.Lock()
	defer rec.mu.Unlock()
	require.Len(t, rec.calls, 1, "RecordUpload must be invoked exactly once (moderation preserved)")
	assert.Equal(t, uid, rec.calls[0].UploaderID)
	assert.Equal(t, mediadomain.ContextProfileVideo, rec.calls[0].MediaCtx)
	assert.Equal(t, "intro.mp4", rec.calls[0].FileName,
		"FileName must be non-empty so mediadomain.NewMedia does not reject the row")
}

func TestUploadHandler_CompleteVideo_RejectsForeignKeyNamespace(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	foreignKey := "profiles/" + uuid.New().String() + "/video/" + uuid.New().String() + ".mp4"

	updateCalled := false
	repo := &mockProfileRepo{
		updateFn: func(_ context.Context, _ *profile.Profile) error { updateCalled = true; return nil },
	}
	h := NewUploadHandler(&mockStorageService{}, repo, nil)
	rec := newFakeRecorder()
	h.recorder = rec
	defer func() { _ = h.Stop(context.Background()) }()

	req := jsonRequest(http.MethodPost, "/api/v1/upload/video/complete",
		completeBody(t, foreignKey, "intro.mp4", "video/mp4", 5000))
	req = withVideoCtx(req, uid, orgID)
	w := httptest.NewRecorder()
	h.CompleteVideo(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "invalid_file_key", resp["error"])
	assert.False(t, updateCalled, "must NOT persist a foreign-namespace key")

	require.NoError(t, h.Stop(context.Background()))
	rec.mu.Lock()
	defer rec.mu.Unlock()
	assert.Empty(t, rec.calls, "moderation must NOT run for a rejected key")
}

func TestUploadHandler_CompleteVideo_RejectsTamperedExtension(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	// Correct namespace but a non-video extension the presign step would
	// never have minted — the ownership guard must reject it.
	badKey := "profiles/" + orgID.String() + "/video/" + uuid.New().String() + ".html"
	h := NewUploadHandler(&mockStorageService{}, &mockProfileRepo{}, nil)

	req := jsonRequest(http.MethodPost, "/api/v1/upload/video/complete",
		completeBody(t, badKey, "x.html", "video/mp4", 5000))
	req = withVideoCtx(req, uid, orgID)
	w := httptest.NewRecorder()
	h.CompleteVideo(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---------------------------------------------------------------------------
// Review + portfolio video — user-namespaced, no DB write, moderation fires
// ---------------------------------------------------------------------------

func TestUploadHandler_PresignReviewVideo_UserNamespaced(t *testing.T) {
	uid := uuid.New()
	var key string
	storage := &mockStorageService{
		getPresignedUploadFn: func(_ context.Context, k, _ string, _ time.Duration) (string, error) {
			key = k
			return "https://r2/put/" + k, nil
		},
	}
	h := NewUploadHandler(storage, &mockProfileRepo{}, nil)
	req := jsonRequest(http.MethodPost, "/api/v1/upload/review-video/presign", presignBody(t, "r.mp4", "video/mp4"))
	req = withVideoCtx(req, uid, uuid.New())
	rec := httptest.NewRecorder()
	h.PresignReviewVideo(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, strings.HasPrefix(key, "reviews/"+uid.String()+"/video/"),
		"review video key must be user-namespaced, got %q", key)
}

func TestUploadHandler_CompleteReviewVideo_ModeratesNoDBWrite(t *testing.T) {
	uid := uuid.New()
	fileKey := "reviews/" + uid.String() + "/video/" + uuid.New().String() + ".webm"
	storage := &mockStorageService{
		getPublicURLFn: func(k string) string { return "https://cdn/" + k },
	}
	h := NewUploadHandler(storage, &mockProfileRepo{}, nil)
	rec := newFakeRecorder()
	h.recorder = rec
	defer func() { _ = h.Stop(context.Background()) }()

	req := jsonRequest(http.MethodPost, "/api/v1/upload/review-video/complete",
		completeBody(t, fileKey, "r.webm", "video/webm", 9000))
	req = withVideoCtx(req, uid, uuid.New())
	w := httptest.NewRecorder()
	h.CompleteReviewVideo(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "https://cdn/"+fileKey, resp["url"])

	require.NoError(t, h.Stop(context.Background()))
	rec.mu.Lock()
	defer rec.mu.Unlock()
	require.Len(t, rec.calls, 1)
	assert.Equal(t, mediadomain.ContextReviewVideo, rec.calls[0].MediaCtx)
}

func TestUploadHandler_CompletePortfolioVideo_Moderates(t *testing.T) {
	uid := uuid.New()
	fileKey := "portfolios/" + uid.String() + "/video/" + uuid.New().String() + ".mp4"
	storage := &mockStorageService{getPublicURLFn: func(k string) string { return "https://cdn/" + k }}
	h := NewUploadHandler(storage, &mockProfileRepo{}, nil)
	rec := newFakeRecorder()
	h.recorder = rec
	defer func() { _ = h.Stop(context.Background()) }()

	req := jsonRequest(http.MethodPost, "/api/v1/upload/portfolio-video/complete",
		completeBody(t, fileKey, "p.mp4", "video/mp4", 20_000_000))
	req = withVideoCtx(req, uid, uuid.New())
	w := httptest.NewRecorder()
	h.CompletePortfolioVideo(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, h.Stop(context.Background()))
	rec.mu.Lock()
	defer rec.mu.Unlock()
	require.Len(t, rec.calls, 1)
	assert.Equal(t, mediadomain.ContextPortfolioVideo, rec.calls[0].MediaCtx)
}

func TestUploadHandler_CompleteReferrerVideo_PersistsAndModerates(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	fileKey := "profiles/" + orgID.String() + "/referrer_video/" + uuid.New().String() + ".mp4"

	var savedURL string
	storage := &mockStorageService{getPublicURLFn: func(k string) string { return "https://cdn/" + k }}
	repo := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, id uuid.UUID) (*profile.Profile, error) {
			return &profile.Profile{OrganizationID: id}, nil
		},
		updateFn: func(_ context.Context, p *profile.Profile) error { savedURL = p.ReferrerVideoURL; return nil },
	}
	h := NewUploadHandler(storage, repo, nil)
	rec := newFakeRecorder()
	h.recorder = rec
	defer func() { _ = h.Stop(context.Background()) }()

	req := jsonRequest(http.MethodPost, "/api/v1/upload/referrer-video/complete",
		completeBody(t, fileKey, "ref.mp4", "video/mp4", 7000))
	req = withVideoCtx(req, uid, orgID)
	w := httptest.NewRecorder()
	h.CompleteReferrerVideo(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://cdn/"+fileKey, savedURL)
	require.NoError(t, h.Stop(context.Background()))
	rec.mu.Lock()
	defer rec.mu.Unlock()
	require.Len(t, rec.calls, 1)
	assert.Equal(t, mediadomain.ContextReferrerVideo, rec.calls[0].MediaCtx)
}

// ---------------------------------------------------------------------------
// Auth context helper (UploadHandler video surfaces need both user + org).
// ---------------------------------------------------------------------------

func TestUploadHandler_CompleteVideo_Unauthorized(t *testing.T) {
	h := NewUploadHandler(&mockStorageService{}, &mockProfileRepo{}, nil)
	req := jsonRequest(http.MethodPost, "/api/v1/upload/video/complete",
		completeBody(t, "profiles/x/video/y.mp4", "y.mp4", "video/mp4", 1))
	// no org in context
	req = req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, uuid.New()))
	rec := httptest.NewRecorder()
	h.CompleteVideo(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
