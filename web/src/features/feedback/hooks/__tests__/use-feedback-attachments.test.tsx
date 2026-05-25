import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { renderHook, act, waitFor } from "@testing-library/react"
import { useFeedbackAttachments } from "../use-feedback-attachments"

const mockUpload = vi.fn()
vi.mock("../../api/feedback-upload", () => ({
  uploadFeedbackAttachment: (...a: unknown[]) => mockUpload(...a),
}))

function fileOf(name: string, type: string, size: number): File {
  const f = new File([new Uint8Array(1)], name, { type })
  Object.defineProperty(f, "size", { value: size })
  return f
}

beforeEach(() => {
  vi.clearAllMocks()
  // jsdom lacks object-URL APIs — stub them for the preview lifecycle.
  URL.createObjectURL = vi.fn(() => "blob:stub")
  URL.revokeObjectURL = vi.fn()
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe("useFeedbackAttachments", () => {
  it("uploads a valid file and exposes its ref once complete", async () => {
    const ref = {
      kind: "image",
      object_key: "feedback/abc.png",
      content_type: "image/png",
      size_bytes: 1024,
    }
    mockUpload.mockResolvedValueOnce(ref)

    const { result } = renderHook(() => useFeedbackAttachments())

    act(() => {
      result.current.addFiles([fileOf("shot.png", "image/png", 1024)])
    })

    // Immediately in the uploading state, no ref yet.
    expect(result.current.attachments).toHaveLength(1)
    expect(result.current.isUploading).toBe(true)
    expect(result.current.uploadedRefs).toEqual([])

    await waitFor(() => {
      expect(result.current.attachments[0].status).toBe("uploaded")
    })
    expect(result.current.uploadedRefs).toEqual([ref])
    expect(result.current.isUploading).toBe(false)
    expect(mockUpload).toHaveBeenCalledWith(expect.any(File), "image")
  })

  it("rejects an unsupported file without calling the uploader", () => {
    const { result } = renderHook(() => useFeedbackAttachments())
    act(() => {
      result.current.addFiles([fileOf("doc.pdf", "application/pdf", 10)])
    })
    expect(result.current.attachments[0].status).toBe("error")
    expect(result.current.attachments[0].rejection).toBe("unsupported_type")
    expect(mockUpload).not.toHaveBeenCalled()
  })

  it("rejects an oversized image without calling the uploader", () => {
    const { result } = renderHook(() => useFeedbackAttachments())
    act(() => {
      result.current.addFiles([fileOf("big.png", "image/png", 11 * 1024 * 1024)])
    })
    expect(result.current.attachments[0].rejection).toBe("too_large")
    expect(mockUpload).not.toHaveBeenCalled()
  })

  it("marks the file as errored when the upload throws", async () => {
    mockUpload.mockRejectedValueOnce(new Error("upload failed: 403"))
    const { result } = renderHook(() => useFeedbackAttachments())
    act(() => {
      result.current.addFiles([fileOf("shot.png", "image/png", 1024)])
    })
    await waitFor(() => {
      expect(result.current.attachments[0].status).toBe("error")
    })
    expect(result.current.attachments[0].rejection).toBe("upload_failed")
    expect(result.current.uploadedRefs).toEqual([])
  })

  it("removes an attachment and revokes its preview URL", async () => {
    mockUpload.mockResolvedValueOnce({
      kind: "image",
      object_key: "k",
      content_type: "image/png",
      size_bytes: 1,
    })
    const { result } = renderHook(() => useFeedbackAttachments())
    act(() => {
      result.current.addFiles([fileOf("shot.png", "image/png", 1)])
    })
    const id = result.current.attachments[0].id

    act(() => {
      result.current.remove(id)
    })
    expect(result.current.attachments).toHaveLength(0)
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:stub")
  })

  it("reset clears all attachments and revokes their preview URLs", () => {
    mockUpload.mockResolvedValue({
      kind: "image",
      object_key: "k",
      content_type: "image/png",
      size_bytes: 1,
    })
    const { result } = renderHook(() => useFeedbackAttachments())
    act(() => {
      result.current.addFiles([
        fileOf("a.png", "image/png", 1),
        fileOf("b.png", "image/png", 1),
      ])
    })
    act(() => {
      result.current.reset()
    })
    expect(result.current.attachments).toHaveLength(0)
    expect(URL.revokeObjectURL).toHaveBeenCalledTimes(2)
  })
})
