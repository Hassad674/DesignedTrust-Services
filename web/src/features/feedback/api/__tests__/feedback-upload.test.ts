import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { uploadFeedbackAttachment } from "../feedback-upload"

const mockPresign = vi.fn()
vi.mock("../feedback-api", () => ({
  presignFeedbackAttachment: (...a: unknown[]) => mockPresign(...a),
}))

function makeFile(name = "shot.png", type = "image/png", size = 4096): File {
  const f = new File([new Uint8Array(8)], name, { type })
  Object.defineProperty(f, "size", { value: size })
  return f
}

beforeEach(() => {
  vi.clearAllMocks()
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("uploadFeedbackAttachment", () => {
  it("presigns then PUTs the bytes to R2 and returns the attachment ref", async () => {
    const calls: string[] = []
    mockPresign.mockImplementation((body: { content_type: string }) => {
      calls.push("presign")
      expect(body).toEqual({
        kind: "image",
        content_type: "image/png",
        size_bytes: 4096,
        filename: "shot.png",
      })
      return Promise.resolve({
        url: "https://r2/put/abc?sig=1",
        object_key: "feedback/abc.png",
        kind: "image",
      })
    })
    const fetchMock = vi.fn().mockImplementation((url: string) => {
      calls.push(`fetch:${url}`)
      return Promise.resolve({ ok: true, status: 200 })
    })
    vi.stubGlobal("fetch", fetchMock)

    const ref = await uploadFeedbackAttachment(makeFile(), "image")

    expect(ref).toEqual({
      kind: "image",
      object_key: "feedback/abc.png",
      content_type: "image/png",
      size_bytes: 4096,
    })
    // Order: presign first, then the direct R2 PUT.
    expect(calls).toEqual(["presign", "fetch:https://r2/put/abc?sig=1"])
  })

  it("PUTs the raw File with the content-type header to the presigned URL", async () => {
    mockPresign.mockResolvedValueOnce({
      url: "https://r2/put/xyz",
      object_key: "feedback/xyz.mp4",
      kind: "video",
    })
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200 })
    vi.stubGlobal("fetch", fetchMock)

    const file = makeFile("clip.mp4", "video/mp4", 999)
    await uploadFeedbackAttachment(file, "video")

    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe("https://r2/put/xyz")
    expect(init.method).toBe("PUT")
    expect(init.body).toBe(file)
    expect(init.headers["Content-Type"]).toBe("video/mp4")
  })

  it("throws when the R2 PUT fails (no silent success)", async () => {
    mockPresign.mockResolvedValueOnce({
      url: "https://r2/put",
      object_key: "k",
      kind: "image",
    })
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 403 }))

    await expect(uploadFeedbackAttachment(makeFile(), "image")).rejects.toThrow(
      /upload failed: 403/,
    )
  })

  it("propagates a presign error without PUTting anything", async () => {
    mockPresign.mockRejectedValueOnce(new Error("presign 401"))
    const fetchMock = vi.fn()
    vi.stubGlobal("fetch", fetchMock)

    await expect(
      uploadFeedbackAttachment(makeFile(), "image"),
    ).rejects.toThrow("presign 401")
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it("falls back to application/octet-stream when the File has no type", async () => {
    mockPresign.mockResolvedValueOnce({
      url: "https://r2/put",
      object_key: "k",
      kind: "image",
    })
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 200 }))

    const typeless = new File([new Uint8Array(4)], "noext", { type: "" })
    Object.defineProperty(typeless, "size", { value: 10 })
    await uploadFeedbackAttachment(typeless, "image")

    expect(mockPresign.mock.calls[0][0].content_type).toBe(
      "application/octet-stream",
    )
  })
})
