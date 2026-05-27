// Sentry SDK glue. Centralised so the three Next.js runtime entry
// points (instrumentation-client.ts for the browser, instrumentation.ts
// for nodejs + edge) reach a single, env-gated init helper instead of
// each duplicating the gating + options.
//
// Two responsibilities:
//   1. Read the env and decide whether Sentry is enabled at all. The SDK
//      MUST stay completely inert when no DSN is configured so the app
//      builds + runs identically to today on a deploy that predates the
//      DSN being set on Vercel. Client reads NEXT_PUBLIC_SENTRY_DSN;
//      server/edge reads SENTRY_DSN with a NEXT_PUBLIC_SENTRY_DSN
//      fallback (the public var is injected into every runtime).
//   2. Provide buildSentryInitOptions() — a pure, testable function that
//      returns the exact options object passed to Sentry.init, and
//      initSentry() which calls Sentry.init ONLY when a DSN is present.
//      When the DSN is absent initSentry() is a no-op: it does not call
//      Sentry.init, emits no console noise, and adds zero runtime cost.
//
// RGPD posture: sendDefaultPii is false (Sentry never attaches IPs,
// cookies, or request bodies by default) and a beforeSend hook strips
// the few headers that could carry a session token before an event
// leaves the browser/server.

/** Structural slice of a Sentry event the scrubber touches. Declared
 * locally (not pulled from the SDK's wide `Event` type) so the file
 * stays SDK-decoupled, reviewable, and free of a circular type
 * reference. The SDK's real `ErrorEvent` is a subtype of this (it adds
 * many optional fields), so the generic scrubber below accepts it. */
export interface SentryScrubbableEvent {
  request?: {
    headers?: Record<string, unknown>
  }
}

/** The shape we pass to Sentry.init. Declared locally (rather than
 * importing the SDK's wide option types) so the file has a stable,
 * reviewable contract and the unit test can assert on exact values.
 * `beforeSend` is generic so it slots into the SDK's
 * `(event: ErrorEvent) => ErrorEvent` contract without coupling this
 * module to the SDK's event type. */
export interface SentryInitOptions {
  dsn: string
  environment: string
  tracesSampleRate: number
  sendDefaultPii: false
  beforeSend: <E extends SentryScrubbableEvent>(event: E) => E
}

/** The minimal slice of the Sentry SDK initSentry() depends on. Keeps
 * the module unit-testable with a tiny fake instead of the full SDK.
 * The real `@sentry/nextjs` namespace satisfies this structurally — its
 * `init(options)` accepts a superset of `SentryInitOptions`. */
export interface SentryInitClient {
  init: (options: SentryInitOptions) => unknown
}

/** Default trace sampling rate. 10% keeps performance overhead and
 * Sentry quota low while still surfacing latency regressions. */
export const SENTRY_TRACES_SAMPLE_RATE = 0.1

// Request/response headers that may carry an authenticated session and
// must never reach Sentry. The httpOnly session_id cookie is the
// project's auth primitive (see web/CLAUDE.md "Auth and Middleware"),
// so cookie + authorization are scrubbed defensively even though
// sendDefaultPii:false already keeps Sentry from attaching them.
const SENSITIVE_HEADERS = ["cookie", "authorization", "x-csrf-token"] as const

/**
 * Resolve the active DSN for a given runtime.
 *
 * - Browser bundles can only read `NEXT_PUBLIC_*` vars, so the client
 *   passes `clientOnly: true` and we look at NEXT_PUBLIC_SENTRY_DSN.
 * - Server/edge can read either; SENTRY_DSN wins, with the public var
 *   as a fallback so a single Vercel var lights up every runtime.
 *
 * Returns `undefined` when no DSN is configured — the signal that
 * Sentry must stay inert.
 */
export function resolveSentryDsn(clientOnly = false): string | undefined {
  const publicDsn = trimToUndefined(process.env.NEXT_PUBLIC_SENTRY_DSN)
  if (clientOnly) return publicDsn
  return trimToUndefined(process.env.SENTRY_DSN) ?? publicDsn
}

/** Whether Sentry should initialise for this runtime. */
export function isSentryEnabled(clientOnly = false): boolean {
  return resolveSentryDsn(clientOnly) !== undefined
}

/**
 * Resolve the reporting environment. Prefer Vercel's per-deployment
 * `VERCEL_ENV` (production / preview / development) and fall back to
 * NODE_ENV for self-hosted / local runs.
 */
export function resolveSentryEnvironment(): string {
  return (
    trimToUndefined(process.env.VERCEL_ENV) ??
    trimToUndefined(process.env.NODE_ENV) ??
    "development"
  )
}

/**
 * Strip headers that could carry an auth token from an outgoing Sentry
 * event. Pure + side-effect-free on the input is not required — Sentry
 * passes us a throwaway event object — but we mutate defensively only
 * the request headers branch.
 */
export function scrubSentryEvent<E extends SentryScrubbableEvent>(
  event: E,
): E {
  const headers = event?.request?.headers
  if (headers && typeof headers === "object") {
    for (const key of Object.keys(headers)) {
      if (SENSITIVE_HEADERS.includes(key.toLowerCase() as never)) {
        delete (headers as Record<string, unknown>)[key]
      }
    }
  }
  return event
}

/**
 * Build the options object handed to Sentry.init. Pure function: no
 * SDK call, no side effects — unit-testable in isolation. Callers must
 * guarantee `dsn` is a non-empty string (initSentry checks first).
 */
export function buildSentryInitOptions(dsn: string): SentryInitOptions {
  return {
    dsn,
    environment: resolveSentryEnvironment(),
    tracesSampleRate: SENTRY_TRACES_SAMPLE_RATE,
    // RGPD: never let Sentry attach IPs, cookies, headers, or request
    // bodies. Combined with the beforeSend scrub this keeps events free
    // of personal data.
    sendDefaultPii: false,
    beforeSend: scrubSentryEvent,
  }
}

/**
 * Initialise Sentry for the current runtime — but ONLY when a DSN is
 * configured. With no DSN this returns `false` without calling
 * Sentry.init, so the SDK is completely inert (no transport, no
 * console output, no overhead). This is what lets us merge + deploy
 * before the DSN exists on Vercel.
 *
 * @param client - the SDK surface (`@sentry/nextjs`), injected so the
 *   runtime entry files share this gate and tests can pass a fake.
 * @param clientOnly - true from the browser entry (reads only the
 *   NEXT_PUBLIC_ DSN); false/omitted on server + edge.
 * @returns whether Sentry.init was invoked.
 */
export function initSentry(client: SentryInitClient, clientOnly = false): boolean {
  const dsn = resolveSentryDsn(clientOnly)
  if (!dsn) return false
  client.init(buildSentryInitOptions(dsn))
  return true
}

/** Trim a possibly-undefined env value, collapsing "" to undefined. */
function trimToUndefined(value: string | undefined): string | undefined {
  if (value === undefined) return undefined
  const trimmed = value.trim()
  return trimmed === "" ? undefined : trimmed
}
