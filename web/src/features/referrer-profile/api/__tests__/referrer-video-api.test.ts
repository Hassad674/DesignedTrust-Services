import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import {
  uploadReferrerVideo,
  deleteReferrerVideo,
} from "../referrer-video-api"

// Since the prod 413 fix, uploadReferrerVideo follows the DIRECT-to-R2
// presigned flow (presign via apiClient → PUT to R2 → complete via
// apiClient). The delete path is still a plain fetch DELETE.

const mockApiClient = vi.fn()
vi.mock("@/shared/lib/api-client", () => ({
  apiClient: (...a: unknown[]) => mockApiClient(...a),
  API_BASE_URL: "",
}))

interface FetchCall {
  url: string
  init?: RequestInit
}

const calls: FetchCall[] = []
const originalFetch = globalThis.fetch

beforeEach(() => {
  calls.length = 0
  mockApiClient.mockReset()
  globalThis.fetch = vi.fn(async (url: RequestInfo | URL, init?: RequestInit) => {
    calls.push({ url: String(url), init })
    return new Response(null, { status: 200 })
  }) as typeof fetch
})

afterEach(() => {
  globalThis.fetch = originalFetch
})

describe("referrer-video-api", () => {
  it("uploadReferrerVideo runs the presigned direct-to-R2 flow", async () => {
    mockApiClient
      .mockResolvedValueOnce({
        upload_url: "https://pub-x.r2.dev/put/r?sig=1",
        file_key: "profiles/org/referrer_video/r.mp4",
        public_url: "https://pub-x.r2.dev/profiles/org/referrer_video/r.mp4",
      })
      .mockResolvedValueOnce({ video_url: "https://pub-x.r2.dev/profiles/org/referrer_video/r.mp4" })

    const file = new File(["dummy"], "referrer.mp4", { type: "video/mp4" })
    const result = await uploadReferrerVideo(file)

    expect(mockApiClient.mock.calls[0][0]).toBe("/api/v1/referrer-profile/video/presign")
    expect(mockApiClient.mock.calls[1][0]).toBe("/api/v1/referrer-profile/video/complete")
    expect(calls).toHaveLength(1)
    expect(calls[0].url).toBe("https://pub-x.r2.dev/put/r?sig=1")
    expect(calls[0].init?.method).toBe("PUT")
    expect(result.video_url).toBe("https://pub-x.r2.dev/profiles/org/referrer_video/r.mp4")
  })

  it("deleteReferrerVideo DELETEs /api/v1/referrer-profile/video", async () => {
    await deleteReferrerVideo()

    expect(calls).toHaveLength(1)
    expect(calls[0].url).toContain("/api/v1/referrer-profile/video")
    expect(calls[0].init?.method).toBe("DELETE")
    expect(calls[0].init?.credentials).toBe("include")
  })

  it("uploadReferrerVideo throws when the R2 PUT fails", async () => {
    mockApiClient.mockResolvedValueOnce({
      upload_url: "https://pub-x.r2.dev/put/r",
      file_key: "k",
      public_url: "https://pub-x.r2.dev/k",
    })
    globalThis.fetch = vi.fn(async () => new Response(null, { status: 500 })) as typeof fetch

    const file = new File(["dummy"], "referrer.mp4", { type: "video/mp4" })
    await expect(uploadReferrerVideo(file)).rejects.toThrow(/upload failed: 500/)
  })
})
