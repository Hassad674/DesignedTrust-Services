import { describe, it, expect, beforeEach, afterEach, vi } from "vitest"
import { captureFeedbackContext, currentPageUrl } from "../capture-context"

describe("captureFeedbackContext", () => {
  beforeEach(() => {
    // jsdom provides window/navigator; pin deterministic values.
    window.innerWidth = 1280
    window.innerHeight = 720
    Object.defineProperty(window.navigator, "userAgent", {
      value: "VitestUA/1.0",
      configurable: true,
    })
  })

  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it("captures locale, role, platform, viewport and user agent", () => {
    const ctx = captureFeedbackContext({ locale: "fr", role: "agency" })
    expect(ctx.locale).toBe("fr")
    expect(ctx.role).toBe("agency")
    expect(ctx.platform).toBe("web")
    expect(ctx.viewport).toBe("1280x720")
    expect(ctx.user_agent).toBe("VitestUA/1.0")
  })

  it("uses an empty role string when the reporter is anonymous", () => {
    const ctx = captureFeedbackContext({ locale: "en" })
    expect(ctx.role).toBe("")
  })

  it("reads the app version from the public env var when present", () => {
    vi.stubEnv("NEXT_PUBLIC_APP_VERSION", "1.4.2")
    expect(captureFeedbackContext({ locale: "fr" }).app_version).toBe("1.4.2")
  })

  it("degrades app version to empty string when the env var is absent", () => {
    vi.stubEnv("NEXT_PUBLIC_APP_VERSION", "")
    expect(captureFeedbackContext({ locale: "fr" }).app_version).toBe("")
  })
})

describe("currentPageUrl", () => {
  it("returns the current location href in a browser", () => {
    expect(currentPageUrl()).toBe(window.location.href)
    expect(currentPageUrl().length).toBeGreaterThan(0)
  })
})
