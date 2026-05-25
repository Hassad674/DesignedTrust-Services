package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/handler/middleware"
)

func TestFeedbackRoutes_SubmitRateLimit_429(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	deps := snapshotDeps()
	deps.RateLimiter = middleware.NewRateLimiter(client, nil)
	// Wire a functional feedback handler so the first requests succeed
	// (201) and only the throttled one returns 429 — proving the limiter
	// sits in front of a working handler rather than masking a panic.
	deps.Feedback = newFeedbackHandlerForTest(&stubFeedbackRepo{}, &stubFeedbackStorage{})
	router := NewRouter(deps)

	body := func() *bytes.Reader {
		return bytes.NewReader(mustJSON(t, map[string]any{
			"type":        "bug",
			"title":       "valid title here",
			"description": "valid description here ok for the rate limit probe",
		}))
	}

	// The submit policy is 5 / 15 min per IP. The first 5 succeed; the
	// 6th must be throttled with 429 + Retry-After.
	var got429 bool
	var sawSuccess bool
	for i := 0; i < 7; i++ {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/", body())
		r.RemoteAddr = "198.51.100.7:5555"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code == http.StatusCreated {
			sawSuccess = true
		}
		if w.Code == http.StatusTooManyRequests {
			got429 = true
			assert.NotEmpty(t, w.Header().Get("Retry-After"), "429 must carry Retry-After")
			break
		}
	}
	assert.True(t, sawSuccess, "early submits must succeed (201) before the cap")
	assert.True(t, got429, "submit endpoint must throttle after the per-IP cap")
}

func TestFeedbackRoutes_Presign_Anonymous_401(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	deps := snapshotDeps()
	deps.RateLimiter = middleware.NewRateLimiter(client, nil)
	router := NewRouter(deps)

	body := bytes.NewReader(mustJSON(t, map[string]any{
		"kind": "image", "content_type": "image/png", "size_bytes": 1024,
	}))
	r := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/attachments/presign", body)
	r.RemoteAddr = "198.51.100.8:5555"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	// The auth middleware rejects the anonymous caller before the
	// handler runs — media is logged-in only.
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
