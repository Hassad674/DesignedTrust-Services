package middleware

import (
	"net/http"

	"marketplace-backend/pkg/response"
)

// RequireEmailVerified is the signup-OTP access gate. It rejects any
// authenticated request whose caller has not yet verified their email
// with a 403 `email_not_verified`. It MUST be chained AFTER Auth — it
// reads the email-verified state from the request context that Auth
// stamps from the JWT `ev` claim (bearer) or the session record
// (cookie), so the check costs zero database round-trips on the hot
// path.
//
// Design notes:
//
//   - Login stays UNGATED on purpose. An unverified user can still log
//     in and obtain tokens; this middleware is what controls access to
//     the gated route groups. Gating login itself would be a hard
//     lockout with no recovery surface — the user could never reach
//     /auth/verify-email to fix it.
//
//   - The allowlist (verify-email, resend-verification, logout, refresh,
//     GET /me) is NOT encoded here. It is expressed at the router by
//     simply NOT chaining this middleware onto those routes — they use
//     the bare Auth middleware. Keeping the allowlist as "absence of the
//     gate on a route" rather than a path string set here means a new
//     allowlisted route is a visible router diff, not a hidden constant.
//
//   - A missing email-verified stamp (GetEmailVerified ok=false) means
//     the request was not authenticated, which should be impossible
//     behind Auth. We treat it as 401 to make the misconfiguration loud
//     rather than silently allowing the request through.
//
//   - The legacy-token fail-safe lives in the JWT/session decoders
//     (an absent `ev` claim decodes to verified=true), so an in-flight
//     access token minted just before the deploy is not gated. By the
//     time it reaches this middleware the value is already resolved.
func RequireEmailVerified(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verified, ok := GetEmailVerified(r.Context())
		if !ok {
			// Gate mounted without Auth ahead of it, or an unauthenticated
			// request slipped through. Either way the caller is not
			// authenticated as far as this layer can tell.
			response.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
			return
		}
		if !verified {
			response.Error(w, http.StatusForbidden, "email_not_verified",
				"verify your email address to continue")
			return
		}
		next.ServeHTTP(w, r)
	})
}
