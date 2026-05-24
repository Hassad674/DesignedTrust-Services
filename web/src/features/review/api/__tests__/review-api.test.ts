import { describe, it, expect, vi, beforeEach } from "vitest"
import {
  fetchReviewsByUser,
  fetchAverageRating,
  fetchCanReview,
  createReview,
  uploadReviewVideo,
} from "../review-api"

const mockApiClient = vi.fn()

vi.mock("@/shared/lib/api-client", () => ({
  apiClient: (...a: unknown[]) => mockApiClient(...a),
  API_BASE_URL: "",
}))

beforeEach(() => {
  vi.clearAllMocks()
  mockApiClient.mockResolvedValue({})
})

describe("review-api / fetchReviewsByUser", () => {
  it("calls without cursor", async () => {
    await fetchReviewsByUser("u-1")
    expect(mockApiClient).toHaveBeenCalledWith("/api/v1/reviews/user/u-1")
  })

  it("appends cursor when provided", async () => {
    await fetchReviewsByUser("u-1", "abc")
    expect(mockApiClient).toHaveBeenCalledWith(
      "/api/v1/reviews/user/u-1?cursor=abc",
    )
  })
})

describe("review-api / fetchAverageRating", () => {
  it("calls /average/:id", () => {
    fetchAverageRating("u-1")
    expect(mockApiClient).toHaveBeenCalledWith(
      "/api/v1/reviews/average/u-1",
    )
  })
})

describe("review-api / fetchCanReview", () => {
  it("calls /can-review/:id", () => {
    fetchCanReview("p-1")
    expect(mockApiClient).toHaveBeenCalledWith(
      "/api/v1/reviews/can-review/p-1",
    )
  })
})

describe("review-api / createReview", () => {
  it("POSTs the payload to /reviews", () => {
    createReview({
      proposal_id: "p-1",
      global_rating: 5,
      comment: "great",
    })
    expect(mockApiClient).toHaveBeenCalledWith("/api/v1/reviews", {
      method: "POST",
      body: expect.objectContaining({
        proposal_id: "p-1",
        global_rating: 5,
      }),
    })
  })

  it("supports optional sub-rating fields", () => {
    createReview({
      proposal_id: "p-1",
      global_rating: 4,
      timeliness: 5,
      communication: 4,
      quality: 4,
    })
    const body = (mockApiClient.mock.calls[0][1] as { body: { timeliness: number } }).body
    expect(body.timeliness).toBe(5)
  })

  it("supports title_visible toggle", () => {
    createReview({
      proposal_id: "p-1",
      global_rating: 3,
      title_visible: true,
    })
    const body = (mockApiClient.mock.calls[0][1] as { body: { title_visible: boolean } }).body
    expect(body.title_visible).toBe(true)
  })
})

// uploadReviewVideo now uses the DIRECT-to-R2 presigned flow (presign +
// complete via apiClient, PUT to R2 via fetch) so videos bypass the
// Vercel proxy body cap that 413'd large videos in production.
describe("review-api / uploadReviewVideo (direct-to-R2 presigned)", () => {
  it("presigns, PUTs to R2, completes, and returns the url", async () => {
    mockApiClient
      .mockResolvedValueOnce({
        upload_url: "https://pub-x.r2.dev/put/v?sig=1",
        file_key: "reviews/u/video/v.mp4",
        public_url: "https://cdn/v.mp4",
      })
      .mockResolvedValueOnce({ url: "https://cdn/v.mp4" })
    const mockFetch = vi.fn(async () => ({ ok: true, status: 200 }))
    vi.stubGlobal("fetch", mockFetch)

    const file = new File(["x"], "v.mp4", { type: "video/mp4" })
    const url = await uploadReviewVideo(file)

    expect(url).toBe("https://cdn/v.mp4")
    expect(mockApiClient.mock.calls[0][0]).toBe("/api/v1/upload/review-video/presign")
    expect(mockApiClient.mock.calls[1][0]).toBe("/api/v1/upload/review-video/complete")
    expect(mockFetch).toHaveBeenCalledWith(
      "https://pub-x.r2.dev/put/v?sig=1",
      expect.objectContaining({ method: "PUT" }),
    )
    vi.unstubAllGlobals()
  })

  it("throws when the R2 PUT fails (non-2xx)", async () => {
    mockApiClient.mockResolvedValueOnce({
      upload_url: "https://pub-x.r2.dev/put/v",
      file_key: "reviews/u/video/v.mp4",
      public_url: "https://cdn/v.mp4",
    })
    const mockFetch = vi.fn(async () => ({ ok: false, status: 413 }))
    vi.stubGlobal("fetch", mockFetch)

    const file = new File(["x"], "v.mp4", { type: "video/mp4" })
    await expect(uploadReviewVideo(file)).rejects.toThrow(/upload failed: 413/)
    vi.unstubAllGlobals()
  })

  it("propagates a presign error before any PUT", async () => {
    mockApiClient.mockRejectedValueOnce(new Error("presign denied"))
    const mockFetch = vi.fn()
    vi.stubGlobal("fetch", mockFetch)

    const file = new File(["x"], "v.mp4", { type: "video/mp4" })
    await expect(uploadReviewVideo(file)).rejects.toThrow("presign denied")
    expect(mockFetch).not.toHaveBeenCalled()
    vi.unstubAllGlobals()
  })
})
