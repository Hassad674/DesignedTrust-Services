import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import {
  verifyEmail,
  resendVerification,
  VerifyEmailError,
} from "../verify-email-api"

// API_BASE_URL is stubbed to "" so the fetch URL is the bare path.
vi.mock("@/shared/lib/api-client", () => ({
  API_BASE_URL: "",
}))

beforeEach(() => {
  vi.clearAllMocks()
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("verify-email-api / verifyEmail", () => {
  it("POSTs the code with credentials and resolves void on success", async () => {
    const fetchMock = vi.fn(async () => ({ ok: true, json: async () => ({}) }))
    vi.stubGlobal("fetch", fetchMock)

    await expect(verifyEmail("123456")).resolves.toBeUndefined()
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/auth/verify-email",
      expect.objectContaining({
        method: "POST",
        credentials: "include",
        body: JSON.stringify({ code: "123456" }),
      }),
    )
  })

  it("throws VerifyEmailError with the backend code on a 4xx", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: false,
      json: async () => ({ error: "invalid_code", message: "nope" }),
    }))
    vi.stubGlobal("fetch", fetchMock)

    await expect(verifyEmail("000000")).rejects.toMatchObject({
      code: "invalid_code",
    })
  })

  it("maps the canonical envelope shape { error: { code } } too", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: false,
      json: async () => ({ error: { code: "challenge_expired" } }),
    }))
    vi.stubGlobal("fetch", fetchMock)

    try {
      await verifyEmail("000000")
      throw new Error("should have thrown")
    } catch (err) {
      expect(err).toBeInstanceOf(VerifyEmailError)
      expect((err as VerifyEmailError).code).toBe("challenge_expired")
    }
  })

  it("falls back to 'unknown' when the body is not JSON", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: false,
      json: async () => {
        throw new Error("not json")
      },
    }))
    vi.stubGlobal("fetch", fetchMock)

    await expect(verifyEmail("000000")).rejects.toMatchObject({
      code: "unknown",
    })
  })
})

describe("verify-email-api / resendVerification", () => {
  it("POSTs with credentials and returns the payload on success", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      json: async () => ({ email_verified: false, message: "sent" }),
    }))
    vi.stubGlobal("fetch", fetchMock)

    const resp = await resendVerification()
    expect(resp).toEqual({ email_verified: false, message: "sent" })
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/auth/resend-verification",
      expect.objectContaining({ method: "POST", credentials: "include" }),
    )
  })

  it("surfaces the already-verified no-op branch", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      json: async () => ({ email_verified: true, message: "already" }),
    }))
    vi.stubGlobal("fetch", fetchMock)

    const resp = await resendVerification()
    expect(resp.email_verified).toBe(true)
  })

  it("throws VerifyEmailError on 429 (rate-limited)", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: false,
      json: async () => ({ error: "too_many_attempts", message: "slow down" }),
    }))
    vi.stubGlobal("fetch", fetchMock)

    await expect(resendVerification()).rejects.toMatchObject({
      code: "too_many_attempts",
    })
  })
})
