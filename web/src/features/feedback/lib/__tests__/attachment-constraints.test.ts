import { describe, it, expect } from "vitest"
import {
  resolveAttachmentKind,
  maxSizeForKind,
  validateAttachment,
  ATTACHMENT_ACCEPT,
  MAX_IMAGE_SIZE_BYTES,
  MAX_VIDEO_SIZE_BYTES,
} from "../attachment-constraints"

function fileOf(type: string, size: number): File {
  const f = new File([new Uint8Array(1)], "f", { type })
  Object.defineProperty(f, "size", { value: size })
  return f
}

describe("resolveAttachmentKind", () => {
  it.each([
    ["image/png", "image"],
    ["image/jpeg", "image"],
    ["image/webp", "image"],
    ["video/mp4", "video"],
    ["video/webm", "video"],
  ])("maps %s to kind %s", (mime, expected) => {
    expect(resolveAttachmentKind(mime)).toBe(expected)
  })

  it("normalises casing and codecs params", () => {
    expect(resolveAttachmentKind("VIDEO/MP4; codecs=avc1")).toBe("video")
  })

  it("returns null for an unsupported type", () => {
    expect(resolveAttachmentKind("application/pdf")).toBeNull()
    expect(resolveAttachmentKind("image/gif")).toBeNull()
  })
})

describe("maxSizeForKind", () => {
  it("uses 10 MB for images and 50 MB for videos", () => {
    expect(maxSizeForKind("image")).toBe(MAX_IMAGE_SIZE_BYTES)
    expect(maxSizeForKind("video")).toBe(MAX_VIDEO_SIZE_BYTES)
    expect(MAX_IMAGE_SIZE_BYTES).toBe(10 * 1024 * 1024)
    expect(MAX_VIDEO_SIZE_BYTES).toBe(50 * 1024 * 1024)
  })
})

describe("validateAttachment", () => {
  it("accepts an in-bounds image", () => {
    expect(validateAttachment(fileOf("image/png", 1024))).toEqual({
      ok: true,
      kind: "image",
    })
  })

  it("rejects an unsupported type", () => {
    expect(validateAttachment(fileOf("application/zip", 10))).toEqual({
      ok: false,
      reason: "unsupported_type",
    })
  })

  it("rejects an image over the 10 MB cap", () => {
    expect(
      validateAttachment(fileOf("image/png", MAX_IMAGE_SIZE_BYTES + 1)),
    ).toEqual({ ok: false, reason: "too_large" })
  })

  it("accepts a video up to the 50 MB cap", () => {
    expect(
      validateAttachment(fileOf("video/mp4", MAX_VIDEO_SIZE_BYTES)),
    ).toEqual({ ok: true, kind: "video" })
  })

  it("rejects a zero-byte file", () => {
    expect(validateAttachment(fileOf("image/png", 0))).toEqual({
      ok: false,
      reason: "too_large",
    })
  })
})

describe("ATTACHMENT_ACCEPT", () => {
  it("lists all five accepted MIME types", () => {
    expect(ATTACHMENT_ACCEPT).toBe(
      "image/png,image/jpeg,image/webp,video/mp4,video/webm",
    )
  })
})
