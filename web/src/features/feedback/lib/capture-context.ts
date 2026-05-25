import type { FeedbackContext } from "../types"

// Auto-captured client context attached to every feedback submission so
// triage has the page, locale, viewport and user-agent without asking
// the reporter. All values are best-effort: in a non-browser (SSR /
// test) environment the browser-derived fields fall back to empty
// strings, which the backend omits from the stored JSONB.

const PLATFORM_WEB = "web"

/** Args the caller already holds in React context (avoids re-reading). */
type CaptureContextArgs = {
  /** Active locale from next-intl. */
  locale: string
  /** Authenticated user's role, or undefined when anonymous. */
  role?: string
}

/**
 * captureFeedbackContext assembles the structured context object. The
 * page URL is returned separately by `currentPageUrl` because the
 * submit body carries it as a top-level `page_url`, not inside context.
 */
export function captureFeedbackContext({
  locale,
  role,
}: CaptureContextArgs): FeedbackContext {
  return {
    role: role ?? "",
    locale,
    platform: PLATFORM_WEB,
    app_version: appVersion(),
    viewport: currentViewport(),
    user_agent: currentUserAgent(),
  }
}

/** The current page URL (origin + path + query), or "" outside a browser. */
export function currentPageUrl(): string {
  if (typeof window === "undefined") return ""
  return window.location.href
}

/** "WIDTHxHEIGHT" of the current viewport, or "" outside a browser. */
function currentViewport(): string {
  if (typeof window === "undefined") return ""
  return `${window.innerWidth}x${window.innerHeight}`
}

/** navigator.userAgent, or "" outside a browser. */
function currentUserAgent(): string {
  if (typeof navigator === "undefined") return ""
  return navigator.userAgent
}

/**
 * appVersion reads the build version exposed via the public env var when
 * present. Kept optional so a missing var degrades to "" rather than
 * failing the capture — the field is informational triage metadata.
 */
function appVersion(): string {
  return process.env.NEXT_PUBLIC_APP_VERSION ?? ""
}
