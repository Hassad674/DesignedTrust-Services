/**
 * @vitest-environment node
 *
 * Verifies the SSR / non-browser fallbacks of the context capture
 * helpers. In a node environment there is no `window` or `navigator`,
 * so every browser-derived field must degrade to an empty string
 * instead of throwing (the helper runs during React's server pass).
 */
import { describe, it, expect } from "vitest"
import { captureFeedbackContext, currentPageUrl } from "../capture-context"

describe("capture-context (SSR / node environment)", () => {
  it("returns empty browser-derived fields without throwing", () => {
    const ctx = captureFeedbackContext({ locale: "fr", role: "provider" })
    // Caller-supplied values survive.
    expect(ctx.locale).toBe("fr")
    expect(ctx.role).toBe("provider")
    expect(ctx.platform).toBe("web")
    // Browser-derived values fall back to "".
    expect(ctx.viewport).toBe("")
    expect(ctx.user_agent).toBe("")
  })

  it("returns an empty page URL outside a browser", () => {
    expect(currentPageUrl()).toBe("")
  })
})
