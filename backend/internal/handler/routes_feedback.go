package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"marketplace-backend/internal/handler/middleware"
)

// Feedback rate-limit policies. Tighter than the global throttle because
// the public submit endpoint is anonymous and abuse-prone, and the
// presign endpoint mints upload URLs. Windows are 15 minutes per the
// product brief.
var (
	// feedbackSubmitPolicy throttles anonymous + authenticated submits
	// per IP: ~5 reports / 15 min. Keyed by IP because the endpoint is
	// public (no user id for anonymous callers).
	feedbackSubmitPolicy = middleware.RateLimitPolicy{
		Class:  "feedback_submit",
		Limit:  5,
		Window: 15 * time.Minute,
	}
	// feedbackPresignPolicy throttles authenticated presign requests per
	// user: ~10 / 15 min. Keyed by user id (the endpoint is auth-gated).
	feedbackPresignPolicy = middleware.RateLimitPolicy{
		Class:  "feedback_presign",
		Limit:  10,
		Window: 15 * time.Minute,
	}
)

// mountFeedbackRoutes wires the public platform-feedback surface:
//
//	POST /api/v1/feedback                       — anonymous allowed
//	                                              (optional auth recognises
//	                                              logged-in reporters);
//	                                              IP-throttled.
//	POST /api/v1/feedback/attachments/presign   — AUTH REQUIRED (media is
//	                                              logged-in only);
//	                                              user-throttled.
//
// nil handler = feature disabled, mounting is a no-op (modularity rule).
func mountFeedbackRoutes(
	r chi.Router,
	deps RouterDeps,
	auth func(http.Handler) http.Handler,
	optionalAuth func(http.Handler) http.Handler,
) {
	if deps.Feedback == nil {
		return
	}
	r.Route("/feedback", func(r chi.Router) {
		r.Use(middleware.NoCache)

		// Public submit: anonymous allowed. optionalAuth stamps identity
		// for logged-in reporters (so they may attach media) but never
		// rejects an anonymous caller. The submit limiter keys by IP.
		r.Group(func(r chi.Router) {
			r.Use(optionalAuth)
			if deps.RateLimiter != nil {
				r.Use(deps.RateLimiter.Middleware(feedbackSubmitPolicy, deps.RateLimiter.IPKey()))
			}
			r.Post("/", deps.Feedback.Submit)
		})

		// Presign: AUTH REQUIRED — media is logged-in only. Anonymous
		// callers get 401 from the auth middleware. The presign limiter
		// keys by user id.
		r.Group(func(r chi.Router) {
			r.Use(auth)
			if deps.RateLimiter != nil {
				r.Use(deps.RateLimiter.Middleware(feedbackPresignPolicy, middleware.UserKey()))
			}
			r.Post("/attachments/presign", deps.Feedback.PresignAttachment)
		})
	})
}
