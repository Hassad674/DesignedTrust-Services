package middleware

import (
	"context"
	"net/http"
	"strings"
)

// OptionalAuthFromDeps is an identity-aware-but-never-rejecting variant
// of AuthFromDeps. It is for PUBLIC endpoints that behave differently
// for an authenticated caller without requiring authentication — e.g.
// the platform feedback submit, where anonymous visitors may post text
// and logged-in reporters may additionally attach media.
//
// Behaviour:
//   - A valid session cookie / Bearer token stamps the full auth
//     context (user_id, role, is_admin, org, permissions) exactly like
//     AuthFromDeps, so handlers read identity via middleware.GetUserID.
//   - ANY problem — no credential, malformed header, expired/revoked
//     token, lookup failure — is swallowed and the request proceeds
//     ANONYMOUSLY (no error written). The handler then sees no user id.
//
// This deliberately never returns 401/403/503: the contract of the
// endpoints it guards is "anonymous is allowed". Routes that must
// reject anonymous callers use AuthFromDeps instead.
func OptionalAuthFromDeps(deps AuthDeps) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ctx := optionalCookieAuth(r, deps); ctx != nil {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			if ctx := optionalBearerAuth(r, deps); ctx != nil {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			// No usable credential — proceed anonymously.
			next.ServeHTTP(w, r)
		})
	}
}

// optionalCookieAuth returns a stamped context when the session cookie
// resolves to a valid, non-revoked session — otherwise nil. It never
// writes a response: a revoked / missing session simply yields nil so
// the caller falls through to the anonymous path.
func optionalCookieAuth(r *http.Request, deps AuthDeps) context.Context {
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		return nil
	}
	session, err := deps.SessionService.Get(r.Context(), cookie.Value)
	if err != nil {
		return nil
	}
	if verifySessionVersion(r.Context(), deps.SessionVersions, session.UserID, session.SessionVersion) != sessionVersionMatch {
		return nil
	}
	return stampAuthContext(r.Context(), authStamp{
		UserID:      session.UserID,
		Role:        session.Role,
		IsAdmin:     session.IsAdmin,
		OrgID:       session.OrganizationID,
		OrgRole:     session.OrgRole,
		Permissions: session.Permissions,
	}, deps.OrgOverrides)
}

// optionalBearerAuth mirrors optionalCookieAuth for a Bearer token.
func optionalBearerAuth(r *http.Request, deps AuthDeps) context.Context {
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil
	}
	claims, err := deps.TokenService.ValidateAccessToken(parts[1])
	if err != nil {
		return nil
	}
	if verifySessionVersion(r.Context(), deps.SessionVersions, claims.UserID, claims.SessionVersion) != sessionVersionMatch {
		return nil
	}
	return stampAuthContext(r.Context(), authStamp{
		UserID:      claims.UserID,
		Role:        claims.Role,
		IsAdmin:     claims.IsAdmin,
		OrgID:       claims.OrganizationID,
		OrgRole:     claims.OrgRole,
		Permissions: claims.Permissions,
	}, deps.OrgOverrides)
}
