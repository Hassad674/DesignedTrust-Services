/**
 * direct-video-upload.test.ts
 *
 * Unit tests for the shared DIRECT-to-R2 video upload helper. The flow
 * is presign (JSON via apiClient) → PUT bytes to R2 (raw fetch to the
 * absolute upload_url, bypassing the Vercel proxy) → complete (JSON via
 * apiClient). apiClient is mocked so we never hit a real backend; the
 * global fetch is mocked so we can assert the direct PUT targets the R2
 * URL with the file as the body.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { uploadVideoDirect } from "../direct-video-upload"

const mockApiClient = vi.fn()
vi.mock("@/shared/lib/api-client", () => ({
  apiClient: (...a: unknown[]) => mockApiClient(...a),
}))

const endpoints = {
  presign: "/api/v1/upload/video/presign",
  complete: "/api/v1/upload/video/complete",
}

function makeFile(name = "intro.mp4", type = "video/mp4", size = 12_000_000): File {
  const f = new File([new Uint8Array(8)], name, { type })
  // jsdom File.size reflects the blob parts (8 bytes); override so we can
  // assert the size forwarded to /complete is the real file size.
  Object.defineProperty(f, "size", { value: size })
  return f
}

beforeEach(() => {
  vi.clearAllMocks()
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("uploadVideoDirect", () => {
  it("runs presign → PUT-to-R2 → complete in order", async () => {
    const calls: string[] = []
    mockApiClient.mockImplementation((path: string) => {
      calls.push(`apiClient:${path}`)
      if (path === endpoints.presign) {
        return Promise.resolve({
          upload_url: "https://pub-x.r2.dev/put/abc?sig=1",
          file_key: "profiles/org/video/abc.mp4",
          public_url: "https://pub-x.r2.dev/profiles/org/video/abc.mp4",
        })
      }
      return Promise.resolve({ url: "https://pub-x.r2.dev/profiles/org/video/abc.mp4" })
    })
    const fetchMock = vi.fn().mockImplementation((url: string) => {
      calls.push(`fetch:${url}`)
      return Promise.resolve({ ok: true, status: 200 })
    })
    vi.stubGlobal("fetch", fetchMock)

    const file = makeFile()
    const res = await uploadVideoDirect<{ url: string }>(endpoints, file)

    expect(res.url).toBe("https://pub-x.r2.dev/profiles/org/video/abc.mp4")
    // Order: presign first, then the direct R2 PUT, then complete.
    expect(calls).toEqual([
      "apiClient:/api/v1/upload/video/presign",
      "fetch:https://pub-x.r2.dev/put/abc?sig=1",
      "apiClient:/api/v1/upload/video/complete",
    ])
  })

  it("sends filename + content_type in the presign body", async () => {
    mockApiClient.mockResolvedValueOnce({
      upload_url: "https://r2/put",
      file_key: "k",
      public_url: "https://r2/k",
    })
    mockApiClient.mockResolvedValueOnce({ url: "https://r2/k" })
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 200 }))

    await uploadVideoDirect(endpoints, makeFile("clip.webm", "video/webm", 999))

    const presignCall = mockApiClient.mock.calls[0]
    expect(presignCall[0]).toBe(endpoints.presign)
    expect(presignCall[1].body).toEqual({ filename: "clip.webm", content_type: "video/webm" })
  })

  it("PUTs the raw File to the presigned URL with the content type header", async () => {
    mockApiClient.mockResolvedValueOnce({
      upload_url: "https://r2/put/xyz",
      file_key: "k",
      public_url: "https://r2/k",
    })
    mockApiClient.mockResolvedValueOnce({ url: "https://r2/k" })
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200 })
    vi.stubGlobal("fetch", fetchMock)

    const file = makeFile("a.mp4", "video/mp4", 42)
    await uploadVideoDirect(endpoints, file)

    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe("https://r2/put/xyz")
    expect(init.method).toBe("PUT")
    expect(init.body).toBe(file)
    expect(init.headers["Content-Type"]).toBe("video/mp4")
  })

  it("forwards file_key + size + content_type to the complete endpoint", async () => {
    mockApiClient.mockResolvedValueOnce({
      upload_url: "https://r2/put",
      file_key: "profiles/org/video/k.mp4",
      public_url: "https://r2/k",
    })
    mockApiClient.mockResolvedValueOnce({ url: "https://r2/k" })
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 200 }))

    await uploadVideoDirect(endpoints, makeFile("a.mp4", "video/mp4", 7777))

    const completeCall = mockApiClient.mock.calls[1]
    expect(completeCall[0]).toBe(endpoints.complete)
    expect(completeCall[1].body).toEqual({
      file_key: "profiles/org/video/k.mp4",
      filename: "a.mp4",
      content_type: "video/mp4",
      file_size: 7777,
    })
  })

  it("throws and does NOT call complete when the R2 PUT fails", async () => {
    mockApiClient.mockResolvedValueOnce({
      upload_url: "https://r2/put",
      file_key: "k",
      public_url: "https://r2/k",
    })
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 403 }))

    await expect(uploadVideoDirect(endpoints, makeFile())).rejects.toThrow(/upload failed: 403/)
    // presign was called, complete must NOT have been reached.
    expect(mockApiClient).toHaveBeenCalledTimes(1)
  })

  it("propagates a presign error without PUTting anything", async () => {
    mockApiClient.mockRejectedValueOnce(new Error("presign boom"))
    const fetchMock = vi.fn()
    vi.stubGlobal("fetch", fetchMock)

    await expect(uploadVideoDirect(endpoints, makeFile())).rejects.toThrow("presign boom")
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it("falls back to application/octet-stream when the File has no type", async () => {
    mockApiClient.mockResolvedValueOnce({
      upload_url: "https://r2/put",
      file_key: "k",
      public_url: "https://r2/k",
    })
    mockApiClient.mockResolvedValueOnce({ url: "https://r2/k" })
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 200 }))

    const typeless = new File([new Uint8Array(4)], "novideo", { type: "" })
    await uploadVideoDirect(endpoints, typeless)

    expect(mockApiClient.mock.calls[0][1].body.content_type).toBe("application/octet-stream")
  })
})
