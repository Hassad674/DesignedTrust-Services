package handler

import (
	"net/http"

	"marketplace-backend/internal/handler/middleware"
	res "marketplace-backend/pkg/response"
	"marketplace-backend/pkg/validator"
)

// VerifyEmail completes the signup email-verification flow. Body shape:
// { code }. The authenticated caller submits the 6-digit code emailed at
// registration; on success the account is marked verified and a FRESH
// token pair (carrying email_verified=true) is issued so the client is
// not stuck behind the gate on its stale claim. The response mirrors the
// regular login shape (web mode sets a new session cookie, token mode
// returns the bearer pair).
//
// Mounted on the signup-OTP allowlist (bare auth, NOT authVerified) so
// an unverified user can actually reach it.
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "user not found in context")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := validator.DecodeJSON(r, &req); err != nil {
		res.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if errs := validator.ValidateRequired(map[string]string{"code": req.Code}); errs != nil {
		res.ValidationError(w, errs)
		return
	}

	output, err := h.authService.VerifyEmail(r.Context(), userID, req.Code, h.sessionFingerprint(r))
	if err != nil {
		// Reuse the 2FA error mapper — the underlying twofactor sentinels
		// (invalid_code, challenge_expired, too_many_attempts, no_challenge)
		// are exactly the failure modes here too.
		handleTwoFactorError(w, err)
		return
	}

	h.sendAuthResponse(w, r, http.StatusOK, output)
}

// ResendVerification issues a fresh email_verification OTP for the
// authenticated caller. Returns 200 in every success branch — including
// the already-verified no-op — so a redundant tap is harmless. The
// per-user rate limiter on the route is the inbox-bombing guard.
//
// Mounted on the signup-OTP allowlist (bare auth).
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		res.Error(w, http.StatusUnauthorized, "unauthorized", "user not found in context")
		return
	}

	alreadyVerified, err := h.authService.ResendVerification(r.Context(), userID, h.sessionFingerprint(r))
	if err != nil {
		handleAuthError(w, err)
		return
	}

	if alreadyVerified {
		res.JSON(w, http.StatusOK, map[string]any{
			"email_verified": true,
			"message":        "Your email is already verified.",
		})
		return
	}
	res.JSON(w, http.StatusOK, map[string]any{
		"email_verified": false,
		"message":        "A new verification code has been sent to your email.",
	})
}
