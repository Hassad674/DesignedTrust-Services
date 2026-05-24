package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/domain/profile"
	"marketplace-backend/internal/handler/middleware"
)

// stubProfileCacheInvalidator records every Invalidate call so a test
// can assert the upload handler busts the cached legacy profile read
// after a successful media write. The optional err lets a test verify
// a flaky cache flush is swallowed (the upload must still succeed).
type stubProfileCacheInvalidator struct {
	calls []uuid.UUID
	err   error
}

func (s *stubProfileCacheInvalidator) Invalidate(_ context.Context, orgID uuid.UUID) error {
	s.calls = append(s.calls, orgID)
	return s.err
}

// withOrgContext attaches the authenticated user + org IDs the
// profile-scoped upload handlers read from request context.
func withOrgContext(req *http.Request, userID, orgID uuid.UUID) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, middleware.ContextKeyOrganizationID, orgID)
	return req.WithContext(ctx)
}

// TestUploadHandler_PhotoUpload_InvalidatesProfileCache pins the
// avatar/photo-persistence fix: after a successful agency photo upload
// the handler MUST bust the Redis-backed public profile cache for the
// owning org. Without this the cached GET /api/v1/profile (which backs
// the hero, sidebar and navbar avatar) keeps serving the stale empty
// photo_url until the TTL expires, so the uploaded photo never shows.
func TestUploadHandler_PhotoUpload_InvalidatesProfileCache(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	storage := &mockStorageService{
		uploadFn: func(_ context.Context, key string, _ io.Reader, _ string, _ int64) (string, error) {
			return "https://storage.example.com/" + key, nil
		},
	}
	profiles := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
			return testProfile(orgID), nil
		},
	}
	inv := &stubProfileCacheInvalidator{}

	h := NewUploadHandler(storage, profiles, nil).WithProfileCacheInvalidator(inv)

	req := buildMultipartRequest(
		http.MethodPost, "/api/v1/upload/photo",
		"file", "photo.jpg", "image/jpeg", validJPEG(),
	)
	req = withOrgContext(req, userID, orgID)
	rec := httptest.NewRecorder()

	h.UploadPhoto(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	require.Len(t, inv.calls, 1, "photo upload must invalidate the profile cache exactly once")
	assert.Equal(t, orgID, inv.calls[0], "cache must be invalidated for the owning org")
}

// TestUploadHandler_VideoUpload_InvalidatesProfileCache mirrors the
// photo assertion for the legacy multipart presentation-video path.
func TestUploadHandler_VideoUpload_InvalidatesProfileCache(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	storage := &mockStorageService{
		uploadFn: func(_ context.Context, key string, _ io.Reader, _ string, _ int64) (string, error) {
			return "https://storage.example.com/" + key, nil
		},
	}
	profiles := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
			return testProfile(orgID), nil
		},
	}
	inv := &stubProfileCacheInvalidator{}

	h := NewUploadHandler(storage, profiles, nil).WithProfileCacheInvalidator(inv)

	req := buildMultipartRequest(
		http.MethodPost, "/api/v1/upload/video",
		"file", "intro.mp4", "video/mp4", validMP4(),
	)
	req = withOrgContext(req, userID, orgID)
	rec := httptest.NewRecorder()

	h.UploadVideo(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	require.Len(t, inv.calls, 1, "video upload must invalidate the profile cache exactly once")
	assert.Equal(t, orgID, inv.calls[0])
}

// TestUploadHandler_DeleteVideo_InvalidatesProfileCache — clearing the
// video must also flush the cache so the empty state shows immediately.
func TestUploadHandler_DeleteVideo_InvalidatesProfileCache(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	profiles := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
			return testProfile(orgID), nil
		},
	}
	inv := &stubProfileCacheInvalidator{}

	h := NewUploadHandler(&mockStorageService{}, profiles, nil).
		WithProfileCacheInvalidator(inv)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/upload/video", nil)
	req = withOrgContext(req, userID, orgID)
	rec := httptest.NewRecorder()

	h.DeleteVideo(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	require.Len(t, inv.calls, 1)
	assert.Equal(t, orgID, inv.calls[0])
}

// TestUploadHandler_CompleteVideo_InvalidatesProfileCache covers the
// DIRECT-to-R2 presigned completion path (the flow the web app uses for
// agency videos > 4.5 MB) — it persists via the same profiles.Update
// and so must bust the cache too.
func TestUploadHandler_CompleteVideo_InvalidatesProfileCache(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	// The completion handler re-verifies file_key sits under the
	// caller's namespace before persisting, so the key must match.
	fileKey := fmt.Sprintf("profiles/%s/video/abc.mp4", orgID.String())

	storage := &mockStorageService{
		getPublicURLFn: func(key string) string {
			return "https://storage.example.com/" + key
		},
	}
	profiles := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
			return testProfile(orgID), nil
		},
	}
	inv := &stubProfileCacheInvalidator{}

	h := NewUploadHandler(storage, profiles, nil).WithProfileCacheInvalidator(inv)

	body, _ := json.Marshal(map[string]any{
		"file_key":     fileKey,
		"filename":     "intro.mp4",
		"content_type": "video/mp4",
		"file_size":    1024,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload/video/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withOrgContext(req, userID, orgID)
	rec := httptest.NewRecorder()

	h.CompleteVideo(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	require.Len(t, inv.calls, 1, "presigned video completion must invalidate the profile cache")
	assert.Equal(t, orgID, inv.calls[0])
}

// TestUploadHandler_FailedWrite_DoesNotInvalidate guards cache-aside
// ordering: a failed persistence MUST NOT bust the cache, otherwise a
// concurrent read would re-populate from the stale row right after we
// cleared it. We assert via the profile-update failure path.
func TestUploadHandler_FailedWrite_DoesNotInvalidate(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	storage := &mockStorageService{
		uploadFn: func(_ context.Context, key string, _ io.Reader, _ string, _ int64) (string, error) {
			return "https://storage.example.com/" + key, nil
		},
	}
	profiles := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
			return testProfile(orgID), nil
		},
		updateFn: func(_ context.Context, _ *profile.Profile) error {
			return errors.New("db down")
		},
	}
	inv := &stubProfileCacheInvalidator{}

	h := NewUploadHandler(storage, profiles, nil).WithProfileCacheInvalidator(inv)

	req := buildMultipartRequest(
		http.MethodPost, "/api/v1/upload/photo",
		"file", "photo.jpg", "image/jpeg", validJPEG(),
	)
	req = withOrgContext(req, userID, orgID)
	rec := httptest.NewRecorder()

	h.UploadPhoto(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Empty(t, inv.calls, "a failed write must not invalidate the cache")
}

// TestUploadHandler_FlakyCacheFlush_StillSucceeds — an Invalidate error
// must be swallowed (logged) so a transient Redis blip never fails an
// otherwise-successful upload. The bytes are stored, the row updated;
// the stale entry simply expires on its TTL.
func TestUploadHandler_FlakyCacheFlush_StillSucceeds(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	storage := &mockStorageService{
		uploadFn: func(_ context.Context, key string, _ io.Reader, _ string, _ int64) (string, error) {
			return "https://storage.example.com/" + key, nil
		},
	}
	profiles := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
			return testProfile(orgID), nil
		},
	}
	inv := &stubProfileCacheInvalidator{err: errors.New("redis timeout")}

	h := NewUploadHandler(storage, profiles, nil).WithProfileCacheInvalidator(inv)

	req := buildMultipartRequest(
		http.MethodPost, "/api/v1/upload/photo",
		"file", "photo.jpg", "image/jpeg", validJPEG(),
	)
	req = withOrgContext(req, userID, orgID)
	rec := httptest.NewRecorder()

	h.UploadPhoto(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "a flaky cache flush must not fail the upload")
	require.Len(t, inv.calls, 1)
}

// TestUploadHandler_NilInvalidator_NoPanic — the invalidator is optional
// (tests + any build without the cache). A nil invalidator must be a
// safe no-op, never a nil-deref.
func TestUploadHandler_NilInvalidator_NoPanic(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	storage := &mockStorageService{
		uploadFn: func(_ context.Context, key string, _ io.Reader, _ string, _ int64) (string, error) {
			return "https://storage.example.com/" + key, nil
		},
	}
	profiles := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
			return testProfile(orgID), nil
		},
	}

	// No WithProfileCacheInvalidator → field stays nil.
	h := NewUploadHandler(storage, profiles, nil)

	req := buildMultipartRequest(
		http.MethodPost, "/api/v1/upload/photo",
		"file", "photo.jpg", "image/jpeg", validJPEG(),
	)
	req = withOrgContext(req, userID, orgID)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() { h.UploadPhoto(rec, req) })
	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
}
