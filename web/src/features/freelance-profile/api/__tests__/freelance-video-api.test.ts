import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import {
  uploadFreelanceVideo,
  deleteFreelanceVideo,
} from "../freelance-video-api"

// Verifies the freelance video api hits the per-persona endpoints. Since
// the prod 413 fix, the upload follows the DIRECT-to-R2 presigned flow
// (presign via apiClient → PUT to R2 → complete via apiClient), so the
// upload tests mock apiClient + fetch; the delete path is still a plain
// fetch DELETE.

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

describe("freelance-video-api", () => {
  it("uploadFreelanceVideo runs the presigned direct-to-R2 flow", async () => {
    mockApiClient
      .mockResolvedValueOnce({
        upload_url: "https://pub-x.r2.dev/put/abc?sig=1",
        file_key: "profiles/org/video/abc.mp4",
        public_url: "https://pub-x.r2.dev/profiles/org/video/abc.mp4",
      })
      .mockResolvedValueOnce({ video_url: "https://pub-x.r2.dev/profiles/org/video/abc.mp4" })

    const file = new File(["dummy"], "intro.mp4", { type: "video/mp4" })
    const result = await uploadFreelanceVideo(file)

    // presign + complete via apiClient
    expect(mockApiClient.mock.calls[0][0]).toBe("/api/v1/freelance-profile/video/presign")
    expect(mockApiClient.mock.calls[1][0]).toBe("/api/v1/freelance-profile/video/complete")
    // bytes PUT DIRECTLY to the R2 upload_url (bypasses the proxy)
    expect(calls).toHaveLength(1)
    expect(calls[0].url).toBe("https://pub-x.r2.dev/put/abc?sig=1")
    expect(calls[0].init?.method).toBe("PUT")
    expect(result.video_url).toBe("https://pub-x.r2.dev/profiles/org/video/abc.mp4")
  })

  it("deleteFreelanceVideo DELETEs against /api/v1/freelance-profile/video", async () => {
    await deleteFreelanceVideo()

    expect(calls).toHaveLength(1)
    expect(calls[0].url).toContain("/api/v1/freelance-profile/video")
    expect(calls[0].init?.method).toBe("DELETE")
    expect(calls[0].init?.credentials).toBe("include")
  })

  it("uploadFreelanceVideo throws when the R2 PUT fails", async () => {
    mockApiClient.mockResolvedValueOnce({
      upload_url: "https://pub-x.r2.dev/put/abc",
      file_key: "k",
      public_url: "https://pub-x.r2.dev/k",
    })
    globalThis.fetch = vi.fn(async () => new Response(null, { status: 500 })) as typeof fetch

    const file = new File(["dummy"], "intro.mp4", { type: "video/mp4" })
    await expect(uploadFreelanceVideo(file)).rejects.toThrow(/upload failed: 500/)
  })
})
