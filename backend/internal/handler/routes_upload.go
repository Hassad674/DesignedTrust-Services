package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"marketplace-backend/internal/domain/organization"
	"marketplace-backend/internal/handler/middleware"
)

// mountUploadRoutes wires the /upload endpoint group. Stacks the
// SEC-11 upload-class limiter on top of the global IP throttle so
// every upload endpoint shares the same per-user quota.
func mountUploadRoutes(r chi.Router, deps RouterDeps, auth func(http.Handler) http.Handler) {
	r.Route("/upload", func(r chi.Router) {
		r.Use(auth)
		r.Use(middleware.NoCache)
		// SEC-11: upload-class limiter on top of the global IP
		// throttle. Stacked here on the whole subtree so every upload
		// endpoint shares the same quota. RATE-LIMIT-PROD bumped the
		// default to 30/min/user and made it env-overridable via
		// RATE_LIMIT_UPLOAD_PER_MINUTE — see UploadRateLimitPolicy.
		if deps.RateLimiter != nil {
			r.Use(deps.RateLimiter.Middleware(UploadRateLimitPolicy(deps.Config), middleware.UserKey()))
		}
		// Profile-related uploads require org profile edit permission
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(organization.PermOrgProfileEdit))
			r.Post("/photo", deps.Upload.UploadPhoto)
			r.Post("/video", deps.Upload.UploadVideo)
			r.Delete("/video", deps.Upload.DeleteVideo)
			r.Post("/referrer-video", deps.Upload.UploadReferrerVideo)
			r.Delete("/referrer-video", deps.Upload.DeleteReferrerVideo)
			r.Post("/portfolio-image", deps.Upload.UploadPortfolioImage)
			r.Post("/portfolio-video", deps.Upload.UploadPortfolioVideo)
			// DIRECT-to-R2 presigned video flow (bypasses the Vercel
			// proxy 4.5 MB body cap). presign issues a short-lived PUT
			// URL; complete persists the URL + fires the SAME moderation
			// pipeline as the multipart endpoints above. Same
			// permission gate as the multipart counterparts.
			r.Post("/video/presign", deps.Upload.PresignVideo)
			r.Post("/video/complete", deps.Upload.CompleteVideo)
			r.Post("/referrer-video/presign", deps.Upload.PresignReferrerVideo)
			r.Post("/referrer-video/complete", deps.Upload.CompleteReferrerVideo)
			r.Post("/portfolio-video/presign", deps.Upload.PresignPortfolioVideo)
			r.Post("/portfolio-video/complete", deps.Upload.CompletePortfolioVideo)
		})
		// Review video upload requires review permission
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(organization.PermReviewsRespond))
			r.Post("/review-video", deps.Upload.UploadReviewVideo)
			r.Post("/review-video/presign", deps.Upload.PresignReviewVideo)
			r.Post("/review-video/complete", deps.Upload.CompleteReviewVideo)
		})
	})
}
