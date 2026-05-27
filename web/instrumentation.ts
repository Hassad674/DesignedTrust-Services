// Next.js 16 server instrumentation entry point. `register()` runs once
// per server runtime (nodejs + edge) at boot; we lazily import the
// matching Sentry config so each runtime only pulls its own SDK build.
// Both configs are env-gated, so when no DSN is set this loads modules
// that do nothing — the server stays inert before the DSN is
// configured on Vercel.
import * as Sentry from "@sentry/nextjs"

export async function register(): Promise<void> {
  if (process.env.NEXT_RUNTIME === "nodejs") {
    await import("./sentry.server.config")
  }
  if (process.env.NEXT_RUNTIME === "edge") {
    await import("./sentry.edge.config")
  }
}

// Forwards React Server Component + route handler errors to Sentry.
// Inert when Sentry never initialised (no DSN) — captureRequestError
// has no client to send through — so it is safe to export before the
// DSN is configured on Vercel.
export const onRequestError = Sentry.captureRequestError
