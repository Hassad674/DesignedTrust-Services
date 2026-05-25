import type { FeedbackAttachmentKind } from "../types"

// Client-side mirror of the backend attachment allowlist + size caps
// (backend/internal/domain/feedback/attachment_spec.go). Validating here
// gives the reporter an instant, localised error instead of a doomed
// presign round-trip. The backend remains the source of truth — these
// constants must stay in sync with it.

const MEGABYTE = 1024 * 1024

/** Per-kind size ceilings, matching MaxImageSizeBytes / MaxVideoSizeBytes. */
export const MAX_IMAGE_SIZE_BYTES = 10 * MEGABYTE
export const MAX_VIDEO_SIZE_BYTES = 50 * MEGABYTE

/** Accepted MIME types per kind (image/png|jpeg|webp, video/mp4|webm). */
export const IMAGE_MIME_TYPES = ["image/png", "image/jpeg", "image/webp"] as const
export const VIDEO_MIME_TYPES = ["video/mp4", "video/webm"] as const

/** `accept` attribute value for the file input (images + videos). */
export const ATTACHMENT_ACCEPT = [...IMAGE_MIME_TYPES, ...VIDEO_MIME_TYPES].join(",")

/** Normalises a MIME string the way the backend does (lowercase, no codecs param). */
function normaliseMime(contentType: string): string {
  const base = contentType.trim().toLowerCase()
  const semicolon = base.indexOf(";")
  return semicolon >= 0 ? base.slice(0, semicolon).trim() : base
}

/**
 * resolveAttachmentKind maps a file's MIME type to the feedback
 * attachment kind, or null when the type is outside the allowlist.
 */
export function resolveAttachmentKind(
  contentType: string,
): FeedbackAttachmentKind | null {
  const mime = normaliseMime(contentType)
  if ((IMAGE_MIME_TYPES as readonly string[]).includes(mime)) return "image"
  if ((VIDEO_MIME_TYPES as readonly string[]).includes(mime)) return "video"
  return null
}

/** The byte ceiling for a given kind. */
export function maxSizeForKind(kind: FeedbackAttachmentKind): number {
  return kind === "video" ? MAX_VIDEO_SIZE_BYTES : MAX_IMAGE_SIZE_BYTES
}

/** The reasons a file can be rejected before upload. */
export type AttachmentRejection = "unsupported_type" | "too_large"

/**
 * validateAttachment runs the type + size checks a file must pass before
 * it is uploaded. Returns the resolved kind on success, or a rejection
 * reason the caller maps to a localised message.
 */
export function validateAttachment(
  file: File,
):
  | { ok: true; kind: FeedbackAttachmentKind }
  | { ok: false; reason: AttachmentRejection } {
  const kind = resolveAttachmentKind(file.type)
  if (!kind) return { ok: false, reason: "unsupported_type" }
  if (file.size <= 0 || file.size > maxSizeForKind(kind)) {
    return { ok: false, reason: "too_large" }
  }
  return { ok: true, kind }
}
