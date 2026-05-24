/**
 * video-upload-endpoints.test.ts
 *
 * Asserts every rewired video-upload function delegates to the shared
 * uploadVideoDirect helper with the CORRECT presign + complete endpoint
 * pair. uploadVideoDirect is mocked so these tests pin the wiring (which
 * endpoints each surface targets) without exercising fetch.
 *
 * This is the regression guard for the prod 413 fix: if any surface
 * reverts to the multipart path (or points at the wrong endpoint), one
 * of these assertions fails.
 */
import { describe, it, expect, vi, beforeEach } from "vitest"

const mockUploadVideoDirect = vi.fn()
vi.mock("@/shared/lib/upload/direct-video-upload", () => ({
  uploadVideoDirect: (...a: unknown[]) => mockUploadVideoDirect(...a),
}))

// Photo (multipart) path stays untouched — stub fetch so the photo
// helper in upload-api doesn't hit the network if a test touches it.
import { uploadFreelanceVideo } from "@/features/freelance-profile/api/freelance-video-api"
import { uploadReferrerVideo as uploadReferrerPersonaVideo } from "@/features/referrer-profile/api/referrer-video-api"
import { uploadVideo, uploadReferrerVideo } from "@/shared/lib/upload-api"
import { uploadPortfolioVideo } from "@/features/provider/api/portfolio-api"
import { uploadReviewVideo } from "@/shared/lib/review/review-api"

function makeFile(): File {
  return new File([new Uint8Array(8)], "v.mp4", { type: "video/mp4" })
}

beforeEach(() => {
  vi.clearAllMocks()
  mockUploadVideoDirect.mockResolvedValue({ url: "https://r2/x", video_url: "https://r2/x" })
})

describe("video upload endpoint wiring", () => {
  it("freelance persona video → /freelance-profile/video/{presign,complete}", async () => {
    await uploadFreelanceVideo(makeFile())
    expect(mockUploadVideoDirect.mock.calls[0][0]).toEqual({
      presign: "/api/v1/freelance-profile/video/presign",
      complete: "/api/v1/freelance-profile/video/complete",
    })
  })

  it("referrer persona video → /referrer-profile/video/{presign,complete}", async () => {
    await uploadReferrerPersonaVideo(makeFile())
    expect(mockUploadVideoDirect.mock.calls[0][0]).toEqual({
      presign: "/api/v1/referrer-profile/video/presign",
      complete: "/api/v1/referrer-profile/video/complete",
    })
  })

  it("legacy agency intro video → /upload/video/{presign,complete}", async () => {
    await uploadVideo(makeFile())
    expect(mockUploadVideoDirect.mock.calls[0][0]).toEqual({
      presign: "/api/v1/upload/video/presign",
      complete: "/api/v1/upload/video/complete",
    })
  })

  it("legacy agency referrer video → /upload/referrer-video/{presign,complete}", async () => {
    await uploadReferrerVideo(makeFile())
    expect(mockUploadVideoDirect.mock.calls[0][0]).toEqual({
      presign: "/api/v1/upload/referrer-video/presign",
      complete: "/api/v1/upload/referrer-video/complete",
    })
  })

  it("portfolio video → /upload/portfolio-video/{presign,complete}", async () => {
    await uploadPortfolioVideo(makeFile())
    expect(mockUploadVideoDirect.mock.calls[0][0]).toEqual({
      presign: "/api/v1/upload/portfolio-video/presign",
      complete: "/api/v1/upload/portfolio-video/complete",
    })
  })

  it("review video → /upload/review-video/{presign,complete} and returns url", async () => {
    mockUploadVideoDirect.mockResolvedValueOnce({ url: "https://r2/review.mp4" })
    const url = await uploadReviewVideo(makeFile())
    expect(mockUploadVideoDirect.mock.calls[0][0]).toEqual({
      presign: "/api/v1/upload/review-video/presign",
      complete: "/api/v1/upload/review-video/complete",
    })
    expect(url).toBe("https://r2/review.mp4")
  })
})
