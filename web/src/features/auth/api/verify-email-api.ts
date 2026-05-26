// Auth-side email-verification API surface — the signup OTP flow.
//
// After /auth/register the new account exists but is `email_verified:
// false`; the backend auto-emailed a 6-digit code. These two endpoints
// drive the verify-email screen:
//
//   POST /auth/verify-email        { code } — verifies the code, flips
//     email_verified, and RE-ISSUES the session (web: a fresh Set-Cookie
//     carrying the verified state; mobile: a bearer pair). We never read
//     the success body — the caller invalidates the ["session"] query and
//     lets /auth/me refetch.
//   POST /auth/resend-verification — issues a fresh code (rate-limited).
//     Returns { email_verified, message }; email_verified=true is the
//     already-verified no-op branch.
//
// Both routes are on the backend's signup-OTP allowlist (bare auth, NOT
// the email-verified gate) so an unverified caller can reach them. Both
// rely on the httpOnly session cookie set at register — hence the raw
// fetch with credentials:"include", mirroring two-factor-api.ts.
import { API_BASE_URL } from "@/shared/lib/api-client"

const API_URL = API_BASE_URL

export type ResendVerificationResponse = {
  email_verified: boolean
  message: string
}

// VerifyEmailError carries the backend error `code` (the 2FA sentinel
// set — invalid_code / challenge_expired / too_many_attempts /
// no_challenge / session_invalid) so the form maps it to a localized
// message the same way the 2FA verify form does.
export class VerifyEmailError extends Error {
  code: string

  constructor(code: string, message: string) {
    super(message)
    this.name = "VerifyEmailError"
    this.code = code
  }
}

async function readErrorCode(res: Response): Promise<VerifyEmailError> {
  const body = await res
    .json()
    .catch(() => ({ error: "unknown", message: "verification failed" }))
  // Backend speaks the legacy flat shape here ({ error: "code", message })
  // — the canonical { error: { code } } envelope is only used by the
  // apiClient wrapper, which these raw-fetch routes bypass.
  const code =
    typeof body?.error === "string" ? body.error : (body?.error?.code ?? "unknown")
  const message = body?.message || "verification failed"
  return new VerifyEmailError(code, message)
}

/**
 * Submits the 6-digit signup code. On success the backend re-sets the
 * session cookie with email_verified=true; we resolve void and let the
 * caller bust the ["session"] cache + navigate to the dashboard. On
 * failure we throw VerifyEmailError so the form can branch on `code`.
 */
export async function verifyEmail(code: string): Promise<void> {
  const res = await fetch(`${API_URL}/api/v1/auth/verify-email`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ code }),
  })
  if (!res.ok) {
    throw await readErrorCode(res)
  }
}

/**
 * Requests a fresh verification code. Resolves with the backend payload
 * so the caller can short-circuit the already-verified branch
 * (email_verified === true). Throws VerifyEmailError on a non-2xx
 * (e.g. 429 rate-limited, email outage) so the form surfaces a retry
 * message.
 */
export async function resendVerification(): Promise<ResendVerificationResponse> {
  const res = await fetch(`${API_URL}/api/v1/auth/resend-verification`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
  })
  if (!res.ok) {
    throw await readErrorCode(res)
  }
  return res.json()
}
