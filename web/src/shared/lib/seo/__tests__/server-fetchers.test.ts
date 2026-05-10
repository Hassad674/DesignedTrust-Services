import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"

import {
  fetchPublicReviews,
  fetchPublicAverageRating,
} from "../server-fetchers"

const originalFetch = global.fetch

beforeEach(() => {
  global.fetch = vi.fn() as unknown as typeof fetch
})

afterEach(() => {
  global.fetch = originalFetch
  vi.restoreAllMocks()
})

function mockFetch(response: { ok: boolean; data?: unknown }) {
  ;(global.fetch as unknown as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
    ok: response.ok,
    json: async () => response.data,
  })
}

describe("fetchPublicReviews", () => {
  it("returns the data array when the endpoint responds 200", async () => {
    mockFetch({
      ok: true,
      data: { data: [{ id: "rev-1" }, { id: "rev-2" }] },
    })
    const out = await fetchPublicReviews("org-1", 5)
    expect(out).toHaveLength(2)
  })

  it("returns null when the endpoint responds non-2xx", async () => {
    mockFetch({ ok: false })
    expect(await fetchPublicReviews("org-1")).toBeNull()
  })

  it("returns null when fetch throws (transient backend hiccup)", async () => {
    ;(global.fetch as unknown as ReturnType<typeof vi.fn>).mockRejectedValueOnce(
      new Error("ECONNREFUSED"),
    )
    expect(await fetchPublicReviews("org-1")).toBeNull()
  })

  it("returns null when the response shape is unexpected", async () => {
    mockFetch({ ok: true, data: { wrong: "shape" } })
    expect(await fetchPublicReviews("org-1")).toBeNull()
  })
})

describe("fetchPublicAverageRating", () => {
  it("returns the AverageRating envelope on success", async () => {
    mockFetch({ ok: true, data: { data: { average: 4.5, count: 12 } } })
    const out = await fetchPublicAverageRating("org-1")
    expect(out).toEqual({ average: 4.5, count: 12 })
  })

  it("returns null on non-2xx", async () => {
    mockFetch({ ok: false })
    expect(await fetchPublicAverageRating("org-1")).toBeNull()
  })
})

