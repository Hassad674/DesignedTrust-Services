import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import {
  deleteReferrerVideo,
  deleteVideo,
  uploadPhoto,
  uploadReferrerVideo,
  uploadVideo,
} from "../upload-api"

// The photo + delete helpers call window.fetch directly (photo is
// multipart; delete is a JSON-less DELETE). The VIDEO helpers now use
// the DIRECT-to-R2 presigned flow (presign + complete via apiClient,
// PUT to R2 via fetch) so the prod 413 (Vercel proxy body cap) no
// longer truncates large videos — apiClient is mocked for those.
type FetchArgs = [RequestInfo | URL, RequestInit | undefined]

const mockApiClient = vi.fn()
vi.mock("@/shared/lib/api-client", () => ({
  apiClient: (...a: unknown[]) => mockApiClient(...a),
  API_BASE_URL: "",
}))

describe("shared/lib/upload-api", () => {
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal("fetch", fetchMock)
    mockApiClient.mockReset()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  // ---------------------------------------------------------------------
  // uploadFile shared behaviour exercised through the public helpers
  // ---------------------------------------------------------------------

  describe("uploadPhoto", () => {
    it("posts the file as multipart/form-data with credentials and returns the parsed response", async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ url: "https://cdn/test.jpg" }),
      })

      const file = new File(["x"], "avatar.png", { type: "image/png" })
      const result = await uploadPhoto(file)

      expect(result).toEqual({ url: "https://cdn/test.jpg" })

      const [url, init] = fetchMock.mock.calls[0] as FetchArgs
      expect(String(url)).toContain("/api/v1/upload/photo")
      expect(init?.method).toBe("POST")
      expect(init?.credentials).toBe("include")
      expect(init?.body).toBeInstanceOf(FormData)

      const form = init!.body as FormData
      expect(form.get("file")).toBe(file)
    })

    it("throws an Error with the backend-provided message when the server replies with !ok", async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: async () => ({ message: "File too large" }),
      })

      await expect(
        uploadPhoto(new File(["x"], "big.png", { type: "image/png" })),
      ).rejects.toThrow("File too large")
    })

    it("throws a generic Error when the error body cannot be parsed as JSON", async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        // simulate JSON parse failure
        json: async () => {
          throw new Error("invalid json")
        },
      })

      await expect(
        uploadPhoto(new File(["x"], "x.png", { type: "image/png" })),
      ).rejects.toThrow("Upload failed")
    })

    it("falls back to the generic message when the parsed body has no message field", async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: async () => ({}),
      })

      await expect(
        uploadPhoto(new File(["x"], "x.png", { type: "image/png" })),
      ).rejects.toThrow("Upload failed")
    })
  })

  describe("uploadVideo (direct-to-R2 presigned)", () => {
    it("presigns /upload/video, PUTs to R2, then completes", async () => {
      mockApiClient
        .mockResolvedValueOnce({
          upload_url: "https://pub-x.r2.dev/put/i?sig=1",
          file_key: "profiles/org/video/i.mp4",
          public_url: "https://cdn/intro.mp4",
        })
        .mockResolvedValueOnce({ url: "https://cdn/intro.mp4" })
      fetchMock.mockResolvedValueOnce({ ok: true, status: 200 })

      const file = new File(["v"], "intro.mp4", { type: "video/mp4" })
      const result = await uploadVideo(file)

      expect(result.url).toBe("https://cdn/intro.mp4")
      expect(mockApiClient.mock.calls[0][0]).toBe("/api/v1/upload/video/presign")
      expect(mockApiClient.mock.calls[1][0]).toBe("/api/v1/upload/video/complete")
      // bytes PUT directly to the R2 origin, not the proxy
      const [putUrl, putInit] = fetchMock.mock.calls[0] as FetchArgs
      expect(String(putUrl)).toBe("https://pub-x.r2.dev/put/i?sig=1")
      expect(putInit?.method).toBe("PUT")
    })
  })

  describe("uploadReferrerVideo (direct-to-R2 presigned)", () => {
    it("targets the referrer-specific presign + complete endpoints", async () => {
      mockApiClient
        .mockResolvedValueOnce({
          upload_url: "https://pub-x.r2.dev/put/r?sig=1",
          file_key: "profiles/org/referrer_video/r.mp4",
          public_url: "https://cdn/ref.mp4",
        })
        .mockResolvedValueOnce({ url: "https://cdn/ref.mp4" })
      fetchMock.mockResolvedValueOnce({ ok: true, status: 200 })

      const file = new File(["v"], "ref.mp4", { type: "video/mp4" })
      const result = await uploadReferrerVideo(file)

      expect(result.url).toBe("https://cdn/ref.mp4")
      expect(mockApiClient.mock.calls[0][0]).toBe("/api/v1/upload/referrer-video/presign")
      expect(mockApiClient.mock.calls[1][0]).toBe("/api/v1/upload/referrer-video/complete")
    })
  })

  describe("deleteVideo", () => {
    it("issues a DELETE on /api/v1/upload/video with credentials and resolves on ok", async () => {
      fetchMock.mockResolvedValueOnce({ ok: true })

      await deleteVideo()

      const [url, init] = fetchMock.mock.calls[0] as FetchArgs
      expect(String(url)).toContain("/api/v1/upload/video")
      expect(init?.method).toBe("DELETE")
      expect(init?.credentials).toBe("include")
    })

    it("throws when the backend rejects the deletion", async () => {
      fetchMock.mockResolvedValueOnce({ ok: false })

      await expect(deleteVideo()).rejects.toThrow("Failed to delete video")
    })
  })

  describe("deleteReferrerVideo", () => {
    it("issues a DELETE on the referrer endpoint", async () => {
      fetchMock.mockResolvedValueOnce({ ok: true })

      await deleteReferrerVideo()

      const [url, init] = fetchMock.mock.calls[0] as FetchArgs
      expect(String(url)).toContain("/api/v1/upload/referrer-video")
      expect(init?.method).toBe("DELETE")
    })

    it("throws a referrer-specific error message on failure", async () => {
      fetchMock.mockResolvedValueOnce({ ok: false })

      await expect(deleteReferrerVideo()).rejects.toThrow(
        "Failed to delete referrer video",
      )
    })
  })
})
