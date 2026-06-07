import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"

import { loadStripe } from "@stripe/stripe-js"
import { createAccountTokenSafe } from "../account-token"

// Mock Stripe.js — every test controls what loadStripe/createToken return.
vi.mock("@stripe/stripe-js", () => ({
  loadStripe: vi.fn(),
}))

const loadStripeMock = vi.mocked(loadStripe)

// A minimal Stripe stub exposing only createToken (the surface we use).
function stripeWithCreateToken(createToken: ReturnType<typeof vi.fn>) {
  // The helper only calls `createToken`; cast through unknown so we don't
  // have to stub the entire Stripe interface in a unit test.
  return { createToken } as unknown as Awaited<ReturnType<typeof loadStripe>>
}

describe("createAccountTokenSafe", () => {
  beforeEach(() => {
    loadStripeMock.mockReset()
    vi.spyOn(console, "error").mockImplementation(() => {})
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it("returns the token id and requests tos_shown_and_accepted when createToken succeeds", async () => {
    const createToken = vi.fn().mockResolvedValue({
      token: { id: "ct_test_123" },
      error: undefined,
    })
    loadStripeMock.mockResolvedValue(stripeWithCreateToken(createToken))

    const id = await createAccountTokenSafe("pk_test_alpha")

    expect(id).toBe("ct_test_123")
    expect(createToken).toHaveBeenCalledWith("account", {
      tos_shown_and_accepted: true,
    })
  })

  it("returns null when createToken yields a Stripe error (caller proceeds without account_token)", async () => {
    const createToken = vi.fn().mockResolvedValue({
      token: undefined,
      error: { message: "tokenization failed" },
    })
    loadStripeMock.mockResolvedValue(stripeWithCreateToken(createToken))

    const id = await createAccountTokenSafe("pk_test_beta")

    expect(id).toBeNull()
    expect(createToken).toHaveBeenCalledTimes(1)
  })

  it("returns null when createToken resolves without a token and without an error", async () => {
    const createToken = vi.fn().mockResolvedValue({
      token: undefined,
      error: undefined,
    })
    loadStripeMock.mockResolvedValue(stripeWithCreateToken(createToken))

    const id = await createAccountTokenSafe("pk_test_gamma")

    expect(id).toBeNull()
  })

  it("returns null without calling loadStripe when the publishable key is empty", async () => {
    const id = await createAccountTokenSafe("")

    expect(id).toBeNull()
    expect(loadStripeMock).not.toHaveBeenCalled()
  })

  it("returns null when Stripe.js fails to load (loadStripe resolves null)", async () => {
    loadStripeMock.mockResolvedValue(null)

    const id = await createAccountTokenSafe("pk_test_delta")

    expect(id).toBeNull()
  })

  it("returns null and never throws when createToken throws", async () => {
    const createToken = vi.fn().mockRejectedValue(new Error("network down"))
    loadStripeMock.mockResolvedValue(stripeWithCreateToken(createToken))

    await expect(createAccountTokenSafe("pk_test_epsilon")).resolves.toBeNull()
  })

  it("memoises the Stripe loader per publishable key across calls", async () => {
    const createToken = vi.fn().mockResolvedValue({
      token: { id: "ct_test_memo" },
      error: undefined,
    })
    loadStripeMock.mockResolvedValue(stripeWithCreateToken(createToken))

    const key = "pk_test_memoised_unique"
    await createAccountTokenSafe(key)
    await createAccountTokenSafe(key)

    // Two token generations, but loadStripe runs at most once for this key.
    expect(createToken).toHaveBeenCalledTimes(2)
    expect(loadStripeMock).toHaveBeenCalledTimes(1)
  })
})
