import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import {
  buildSentryInitOptions,
  initSentry,
  isSentryEnabled,
  resolveSentryDsn,
  resolveSentryEnvironment,
  scrubSentryEvent,
  SENTRY_TRACES_SAMPLE_RATE,
  type SentryInitOptions,
} from "../sentry"

const ORIGINAL_ENV = { ...process.env }
const VALID_DSN = "https://abc123@o42.ingest.us.sentry.io/4509"

// A tiny fake SDK surface so we never load the real @sentry/nextjs in
// unit tests but can still assert exactly when (and with what)
// Sentry.init was invoked. The mock is typed to the init signature so
// it satisfies SentryInitClient without an `as` cast.
function makeClient(): { init: ReturnType<typeof vi.fn<(o: SentryInitOptions) => unknown>> } {
  return { init: vi.fn<(o: SentryInitOptions) => unknown>() }
}

beforeEach(() => {
  delete process.env.NEXT_PUBLIC_SENTRY_DSN
  delete process.env.SENTRY_DSN
  delete process.env.VERCEL_ENV
})

afterEach(() => {
  process.env = { ...ORIGINAL_ENV }
})

describe("resolveSentryDsn", () => {
  it("returns undefined when no DSN env is set", () => {
    expect(resolveSentryDsn()).toBeUndefined()
    expect(resolveSentryDsn(true)).toBeUndefined()
  })

  it("client-only reads NEXT_PUBLIC_SENTRY_DSN", () => {
    process.env.NEXT_PUBLIC_SENTRY_DSN = VALID_DSN
    expect(resolveSentryDsn(true)).toBe(VALID_DSN)
  })

  it("client-only ignores the server-only SENTRY_DSN", () => {
    process.env.SENTRY_DSN = VALID_DSN
    expect(resolveSentryDsn(true)).toBeUndefined()
  })

  it("server prefers SENTRY_DSN over the public fallback", () => {
    process.env.SENTRY_DSN = "https://server@o1.ingest.de.sentry.io/1"
    process.env.NEXT_PUBLIC_SENTRY_DSN = VALID_DSN
    expect(resolveSentryDsn()).toBe("https://server@o1.ingest.de.sentry.io/1")
  })

  it("server falls back to NEXT_PUBLIC_SENTRY_DSN when SENTRY_DSN is absent", () => {
    process.env.NEXT_PUBLIC_SENTRY_DSN = VALID_DSN
    expect(resolveSentryDsn()).toBe(VALID_DSN)
  })

  it("treats an empty / whitespace DSN as unset", () => {
    process.env.NEXT_PUBLIC_SENTRY_DSN = "   "
    process.env.SENTRY_DSN = ""
    expect(resolveSentryDsn()).toBeUndefined()
    expect(resolveSentryDsn(true)).toBeUndefined()
  })
})

describe("isSentryEnabled", () => {
  it("is false with no DSN configured", () => {
    expect(isSentryEnabled()).toBe(false)
    expect(isSentryEnabled(true)).toBe(false)
  })

  it("is true once a DSN is present", () => {
    process.env.NEXT_PUBLIC_SENTRY_DSN = VALID_DSN
    expect(isSentryEnabled()).toBe(true)
    expect(isSentryEnabled(true)).toBe(true)
  })
})

describe("initSentry — env gating (CRITICAL: inert without DSN)", () => {
  it("does NOT call Sentry.init when no DSN is set", () => {
    const client = makeClient()
    const invoked = initSentry(client)
    expect(invoked).toBe(false)
    expect(client.init).not.toHaveBeenCalled()
  })

  it("does NOT call Sentry.init on the client when only SENTRY_DSN is set", () => {
    process.env.SENTRY_DSN = VALID_DSN
    const client = makeClient()
    const invoked = initSentry(client, true)
    expect(invoked).toBe(false)
    expect(client.init).not.toHaveBeenCalled()
  })

  it("calls Sentry.init once when NEXT_PUBLIC_SENTRY_DSN is set", () => {
    process.env.NEXT_PUBLIC_SENTRY_DSN = VALID_DSN
    const client = makeClient()
    const invoked = initSentry(client, true)
    expect(invoked).toBe(true)
    expect(client.init).toHaveBeenCalledTimes(1)
    expect(client.init).toHaveBeenCalledWith(
      expect.objectContaining({ dsn: VALID_DSN }),
    )
  })

  it("calls Sentry.init on the server when SENTRY_DSN is set", () => {
    process.env.SENTRY_DSN = VALID_DSN
    const client = makeClient()
    const invoked = initSentry(client)
    expect(invoked).toBe(true)
    expect(client.init).toHaveBeenCalledTimes(1)
  })
})

describe("buildSentryInitOptions", () => {
  it("sets sendDefaultPii=false and the 0.1 trace sample rate (RGPD + budget)", () => {
    const options = buildSentryInitOptions(VALID_DSN)
    expect(options.sendDefaultPii).toBe(false)
    expect(options.tracesSampleRate).toBe(SENTRY_TRACES_SAMPLE_RATE)
    expect(options.tracesSampleRate).toBe(0.1)
    expect(options.dsn).toBe(VALID_DSN)
  })

  it("derives environment from VERCEL_ENV when present", () => {
    process.env.VERCEL_ENV = "preview"
    expect(buildSentryInitOptions(VALID_DSN).environment).toBe("preview")
  })

  it("falls back to NODE_ENV for the environment", () => {
    expect(buildSentryInitOptions(VALID_DSN).environment).toBe(
      process.env.NODE_ENV ?? "development",
    )
  })

  it("wires beforeSend to the PII scrubber", () => {
    const options = buildSentryInitOptions(VALID_DSN)
    expect(options.beforeSend).toBe(scrubSentryEvent)
  })
})

describe("resolveSentryEnvironment", () => {
  it("prefers VERCEL_ENV", () => {
    process.env.VERCEL_ENV = "production"
    expect(resolveSentryEnvironment()).toBe("production")
  })

  it("returns a non-empty string even with nothing set", () => {
    delete process.env.VERCEL_ENV
    expect(resolveSentryEnvironment().length).toBeGreaterThan(0)
  })
})

describe("scrubSentryEvent — drops auth-bearing headers (RGPD)", () => {
  it("removes cookie / authorization / csrf headers regardless of case", () => {
    const event = {
      request: {
        headers: {
          Cookie: "session_id=secret",
          Authorization: "Bearer token",
          "X-CSRF-Token": "csrf",
          "user-agent": "vitest",
        },
      },
    }
    const scrubbed = scrubSentryEvent(event)
    const headers = scrubbed.request?.headers as Record<string, unknown>
    expect(headers).not.toHaveProperty("Cookie")
    expect(headers).not.toHaveProperty("Authorization")
    expect(headers).not.toHaveProperty("X-CSRF-Token")
    // Non-sensitive headers are preserved for debugging.
    expect(headers).toHaveProperty("user-agent", "vitest")
  })

  it("is a safe no-op when there is no request on the event", () => {
    expect(() => scrubSentryEvent({})).not.toThrow()
    expect(scrubSentryEvent({})).toEqual({})
  })
})
