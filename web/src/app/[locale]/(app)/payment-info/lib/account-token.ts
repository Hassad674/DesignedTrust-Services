/**
 * Client-side Stripe **account token** generation for the inline embedded
 * KYC onboarding.
 *
 * Why this exists: Stripe FORBIDS creating account tokens from the
 * application server in live mode — they MUST be created in the browser
 * with the PUBLISHABLE key. The backend's
 * `POST /api/v1/payment-info/account-session` accepts an OPTIONAL
 * `account_token`; when present on the first call (no connected account
 * yet) the account is created as a platform-collection Custom account and
 * the embedded onboarding runs fully INLINE (no Stripe-hosted auth popup).
 * When the token is absent the backend falls back gracefully to the
 * Stripe-collection flow (which shows an auth step). Generating a fresh
 * token per call is cheap and safe — the backend only consumes it when it
 * actually creates the account.
 *
 * Contract of `createAccountTokenSafe`: it NEVER throws and NEVER blocks
 * the KYC page. On any failure (missing key, Stripe.js load error,
 * `createToken` error, unexpected exception) it logs and resolves to
 * `null`, letting the caller send the account-session request WITHOUT
 * `account_token`.
 */

import { loadStripe, type Stripe } from "@stripe/stripe-js"

// Memoise the Stripe.js loader per publishable key so we run `loadStripe`
// at most once per key per page-load instead of on every account-session
// call (the embedded flow may invoke the session fetch multiple times).
const loaderCache = new Map<string, Promise<Stripe | null>>()

function getStripe(publishableKey: string): Promise<Stripe | null> {
  const cached = loaderCache.get(publishableKey)
  if (cached) return cached
  const promise = loadStripe(publishableKey)
  loaderCache.set(publishableKey, promise)
  return promise
}

/**
 * Generates a Stripe account token in the browser carrying
 * `tos_shown_and_accepted`. Returns the token id (e.g. `ct_...`) on
 * success, or `null` if the token could not be created for any reason.
 * Failures are logged, never thrown.
 */
export async function createAccountTokenSafe(
  publishableKey: string,
): Promise<string | null> {
  if (!publishableKey) return null

  try {
    const stripe = await getStripe(publishableKey)
    if (!stripe) {
      // loadStripe resolves to null when the key is invalid or Stripe.js
      // failed to load — degrade to the popup fallback rather than block.
      console.error("createAccountTokenSafe: Stripe.js failed to load")
      return null
    }

    // The typed `createToken('account', data)` overload (Stripe.js v9)
    // takes the account params directly (TokenCreateParams.Account), with
    // `tos_shown_and_accepted` at the top level — not nested under `account`.
    const { token, error } = await stripe.createToken("account", {
      tos_shown_and_accepted: true,
    })

    if (error || !token) {
      console.error(
        "createAccountTokenSafe: createToken failed",
        error?.message ?? "no token returned",
      )
      return null
    }

    return token.id
  } catch (err) {
    console.error("createAccountTokenSafe: unexpected error", err)
    return null
  }
}
