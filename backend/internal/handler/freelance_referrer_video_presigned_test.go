package handler

import (
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
)

// ---------------------------------------------------------------------------
// Freelance persona — presigned video flow
// ---------------------------------------------------------------------------

func TestFreelanceProfileVideo_Presign_Success(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	var key, ct string
	storage := &mockStorageService{
		getPresignedUploadFn: func(_ context.Context, k, c string, expiry time.Duration) (string, error) {
			key, ct = k, c
			assert.Equal(t, presignedVideoExpiry, expiry)
			return "https://r2/put/" + k, nil
		},
		getPublicURLFn: func(k string) string { return "https://cdn/" + k },
	}
	h := NewFreelanceProfileVideoHandler(storage, &mockFreelanceProfileRepo{}, nil)

	req := jsonRequest(http.MethodPost, "/api/v1/freelance-profile/video/presign",
		presignBody(t, "intro.mov", "video/quicktime"))
	req = withVideoCtx(req, uid, orgID)
	rec := httptest.NewRecorder()
	h.PresignVideo(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, strings.HasPrefix(key, "profiles/"+orgID.String()+"/video/"),
		"freelance key must use the profiles/<orgID>/video namespace, got %q", key)
	assert.True(t, strings.HasSuffix(key, ".mov"), "extension from content-type, got %q", key)
	assert.Equal(t, "video/quicktime", ct)
}

func TestFreelanceProfileVideo_Presign_Unauthorized(t *testing.T) {
	h := NewFreelanceProfileVideoHandler(&mockStorageService{}, &mockFreelanceProfileRepo{}, nil)
	req := jsonRequest(http.MethodPost, "/api/v1/freelance-profile/video/presign",
		presignBody(t, "x.mp4", "video/mp4"))
	rec := httptest.NewRecorder()
	h.PresignVideo(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestFreelanceProfileVideo_Complete_PersistsAndModerates(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	fileKey := "profiles/" + orgID.String() + "/video/" + uuid.New().String() + ".mp4"

	var savedURL string
	storage := &mockStorageService{getPublicURLFn: func(k string) string { return "https://cdn/" + k }}
	repo := &mockFreelanceProfileRepo{
		updateVideoFn: func(_ context.Context, _ uuid.UUID, url string) error { savedURL = url; return nil },
		getVideoFn:    func(_ context.Context, _ uuid.UUID) (string, error) { return "", nil },
	}
	h := NewFreelanceProfileVideoHandler(storage, repo, nil).withRecorder(newFakeRecorder())
	rec := h.recorder.(*fakeRecorder)

	req := jsonRequest(http.MethodPost, "/api/v1/freelance-profile/video/complete",
		completeBody(t, fileKey, "intro.mp4", "video/mp4", 30_000_000))
	req = withVideoCtx(req, uid, orgID)
	w := httptest.NewRecorder()
	h.CompleteVideo(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	wantURL := "https://cdn/" + fileKey
	assert.Equal(t, wantURL, resp["video_url"])
	assert.Equal(t, wantURL, savedURL, "video URL persisted via UpdateVideo")

	require.Eventually(t, func() bool {
		rec.mu.Lock()
		defer rec.mu.Unlock()
		return len(rec.calls) == 1
	}, time.Second, 5*time.Millisecond, "moderation RecordUpload must fire exactly once")
	// The goroutine must inherit request values (WithoutCancel), not a
	// bare context.Background() — request-scoped logging relies on it.
	// lastCtx() takes the recorder lock itself, so call it BEFORE we
	// grab the lock below (it would otherwise self-deadlock).
	assert.NotNil(t, rec.lastCtx())
	rec.mu.Lock()
	defer rec.mu.Unlock()
	assert.Equal(t, mediadomain.ContextProfileVideo, rec.calls[0].MediaCtx)
	assert.Equal(t, "intro.mp4", rec.calls[0].FileName)
}

func TestFreelanceProfileVideo_Complete_RejectsForeignKey(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	foreign := "profiles/" + uuid.New().String() + "/video/" + uuid.New().String() + ".mp4"

	updateCalled := false
	repo := &mockFreelanceProfileRepo{
		updateVideoFn: func(_ context.Context, _ uuid.UUID, _ string) error { updateCalled = true; return nil },
	}
	h := NewFreelanceProfileVideoHandler(&mockStorageService{}, repo, nil).withRecorder(newFakeRecorder())
	rec := h.recorder.(*fakeRecorder)

	req := jsonRequest(http.MethodPost, "/api/v1/freelance-profile/video/complete",
		completeBody(t, foreign, "x.mp4", "video/mp4", 5000))
	req = withVideoCtx(req, uid, orgID)
	w := httptest.NewRecorder()
	h.CompleteVideo(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, updateCalled, "must not persist a foreign-namespace key")
	// Give any (erroneously spawned) goroutine a beat — there should be none.
	time.Sleep(20 * time.Millisecond)
	rec.mu.Lock()
	defer rec.mu.Unlock()
	assert.Empty(t, rec.calls, "moderation must not run for a rejected key")
}

// ---------------------------------------------------------------------------
// Referrer persona — presigned video flow
// ---------------------------------------------------------------------------

func TestReferrerProfileVideo_Presign_Success(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	var key string
	storage := &mockStorageService{
		getPresignedUploadFn: func(_ context.Context, k, _ string, _ time.Duration) (string, error) {
			key = k
			return "https://r2/put/" + k, nil
		},
		getPublicURLFn: func(k string) string { return "https://cdn/" + k },
	}
	h := NewReferrerProfileVideoHandler(storage, &mockReferrerProfileRepo{}, nil)

	req := jsonRequest(http.MethodPost, "/api/v1/referrer-profile/video/presign",
		presignBody(t, "ref.webm", "video/webm"))
	req = withVideoCtx(req, uid, orgID)
	rec := httptest.NewRecorder()
	h.PresignVideo(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, strings.HasPrefix(key, "profiles/"+orgID.String()+"/referrer_video/"),
		"referrer key must use the profiles/<orgID>/referrer_video namespace, got %q", key)
	assert.True(t, strings.HasSuffix(key, ".webm"))
}

func TestReferrerProfileVideo_Complete_PersistsAndModerates(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	fileKey := "profiles/" + orgID.String() + "/referrer_video/" + uuid.New().String() + ".mp4"

	var savedURL string
	storage := &mockStorageService{getPublicURLFn: func(k string) string { return "https://cdn/" + k }}
	repo := &mockReferrerProfileRepo{
		updateVideoFn: func(_ context.Context, _ uuid.UUID, url string) error { savedURL = url; return nil },
		getVideoFn:    func(_ context.Context, _ uuid.UUID) (string, error) { return "", nil },
	}
	h := NewReferrerProfileVideoHandler(storage, repo, nil).withRecorder(newFakeRecorder())
	rec := h.recorder.(*fakeRecorder)

	req := jsonRequest(http.MethodPost, "/api/v1/referrer-profile/video/complete",
		completeBody(t, fileKey, "ref.mp4", "video/mp4", 15_000_000))
	req = withVideoCtx(req, uid, orgID)
	w := httptest.NewRecorder()
	h.CompleteVideo(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://cdn/"+fileKey, savedURL)

	require.Eventually(t, func() bool {
		rec.mu.Lock()
		defer rec.mu.Unlock()
		return len(rec.calls) == 1
	}, time.Second, 5*time.Millisecond)
	rec.mu.Lock()
	defer rec.mu.Unlock()
	assert.Equal(t, mediadomain.ContextReferrerVideo, rec.calls[0].MediaCtx)
}

func TestReferrerProfileVideo_Complete_RejectsForeignKey(t *testing.T) {
	uid, orgID := uuid.New(), uuid.New()
	foreign := "profiles/" + orgID.String() + "/video/" + uuid.New().String() + ".mp4" // freelance namespace, not referrer
	h := NewReferrerProfileVideoHandler(&mockStorageService{}, &mockReferrerProfileRepo{}, nil)

	req := jsonRequest(http.MethodPost, "/api/v1/referrer-profile/video/complete",
		completeBody(t, foreign, "x.mp4", "video/mp4", 5000))
	req = withVideoCtx(req, uid, orgID)
	w := httptest.NewRecorder()
	h.CompleteVideo(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
