import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { apiClient, ApiError } from "../api-client"

// The signup-OTP access gate: a 403 whose code === "email_not_verified"
// must funnel the caller to /verify-email. Every OTHER 403 (forbidden,
// ownership, …) and every other status must be left untouched and
// surface to the caller as a normal ApiError.

const mockFetch = vi.fn()
const assignMock = vi.fn()

// Swap window.location for a mock we fully control. jsdom's real
// location.assign is a no-op that emits a "Not implemented" warning, so
// a stub is both quieter and assertable. `pathname` is settable per-test
// to exercise the locale-prefix + loop-guard branches.
function setLocation(pathname: string) {
  Object.defineProperty(window, "location", {
    configurable: true,
    value: { pathname, assign: assignMock },
  })
}

function fail403(code: string) {
  mockFetch.mockResolvedValueOnce({
    ok: false,
    status: 403,
    json: () => Promise.resolve({ error: code, message: "denied" }),
  })
}

beforeEach(() => {
  vi.stubGlobal("fetch", mockFetch)
  setLocation("/dashboard")
})

afterEach(() => {
  mockFetch.mockReset()
  assignMock.mockReset()
  vi.unstubAllGlobals()
})

describe("apiClient — email_not_verified gate", () => {
  it("redirects to /verify-email on a 403 email_not_verified", async () => {
    fail403("email_not_verified")
    await expect(apiClient("/api/v1/profile")).rejects.toThrow(ApiError)
    expect(assignMock).toHaveBeenCalledWith("/verify-email")
  })

  it("still throws the ApiError (caller's catch must run) on the gated 403", async () => {
    fail403("email_not_verified")
    try {
      await apiClient("/api/v1/profile")
      expect.fail("should have thrown")
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      expect((err as ApiError).status).toBe(403)
      expect((err as ApiError).code).toBe("email_not_verified")
    }
  })

  it("preserves the active locale prefix in the redirect target", async () => {
    setLocation("/fr/messages")
    fail403("email_not_verified")
    await expect(apiClient("/api/v1/messages")).rejects.toThrow(ApiError)
    expect(assignMock).toHaveBeenCalledWith("/fr/verify-email")
  })

  it("does NOT redirect when already on the verify-email screen (loop guard)", async () => {
    setLocation("/verify-email")
    fail403("email_not_verified")
    await expect(apiClient("/api/v1/profile")).rejects.toThrow(ApiError)
    expect(assignMock).not.toHaveBeenCalled()
  })

  it("does NOT redirect when already on the locale-prefixed verify-email screen", async () => {
    setLocation("/en/verify-email")
    fail403("email_not_verified")
    await expect(apiClient("/api/v1/profile")).rejects.toThrow(ApiError)
    expect(assignMock).not.toHaveBeenCalled()
  })

  it("does NOT redirect on a plain 'forbidden' 403 (RBAC / ownership)", async () => {
    fail403("forbidden")
    await expect(apiClient("/api/v1/missions/other")).rejects.toThrow(ApiError)
    expect(assignMock).not.toHaveBeenCalled()
  })

  it("does NOT redirect on a 401 unauthorized", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ error: "unauthorized", message: "no" }),
    })
    await expect(apiClient("/api/v1/profile")).rejects.toThrow(ApiError)
    expect(assignMock).not.toHaveBeenCalled()
  })

  it("does NOT redirect when email_not_verified arrives with a non-403 status", async () => {
    // Defensive: the code key alone must not trigger the funnel — it is
    // gated on status===403 AND code===email_not_verified.
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: () =>
        Promise.resolve({ error: "email_not_verified", message: "weird" }),
    })
    await expect(apiClient("/api/v1/profile")).rejects.toThrow(ApiError)
    expect(assignMock).not.toHaveBeenCalled()
  })
})
