import { presignFeedbackAttachment } from "./feedback-api"
import type { FeedbackAttachmentKind, FeedbackAttachmentRef } from "../types"

// Feedback media upload — the presigned-PUT mechanism shared with the
// messaging/video flows, scoped to the feedback contract.
//
// WHY A FEEDBACK-LOCAL HELPER — the video helper
// (shared/lib/upload/direct-video-upload.ts) speaks a different
// envelope: it returns `{ upload_url, file_key, public_url }` and runs a
// third "complete" round-trip that persists + moderates the asset. The
// feedback presign returns `{ url, object_key, kind }` and has NO
// complete step — the `object_key` is carried straight into the
// submit's `attachment_keys`. Coupling the two flows behind one generic
// would break both, so the feature owns its own presign→PUT here.
//
// The R2 origin is already whitelisted in the CSP `connect-src`
// (see src/shared/lib/csp.ts), so the cross-origin PUT is allowed.

/**
 * uploadFeedbackAttachment runs presign → PUT-bytes-to-R2 and returns
 * the attachment reference the submit body expects. Auth is required
 * (the presign endpoint 401s anonymous callers); the caller is expected
 * to gate the UI so this is only invoked for logged-in reporters.
 *
 * `fetch` does NOT throw on a non-2xx PUT, so we explicitly guard
 * `uploadRes.ok` to avoid returning a key for bytes that never landed.
 */
export async function uploadFeedbackAttachment(
  file: File,
  kind: FeedbackAttachmentKind,
): Promise<FeedbackAttachmentRef> {
  const contentType = file.type || "application/octet-stream"

  // 1. Ask the backend for a presigned PUT URL + server-minted key.
  const presigned = await presignFeedbackAttachment({
    kind,
    content_type: contentType,
    size_bytes: file.size,
    filename: file.name,
  })

  // 2. PUT the bytes DIRECTLY to R2 (absolute URL → bypasses the proxy).
  const uploadRes = await fetch(presigned.url, {
    method: "PUT",
    body: file,
    headers: { "Content-Type": contentType },
  })
  if (!uploadRes.ok) {
    throw new Error(`upload failed: ${uploadRes.status}`)
  }

  // 3. Hand back the reference the submit body carries verbatim.
  return {
    kind: presigned.kind,
    object_key: presigned.object_key,
    content_type: contentType,
    size_bytes: file.size,
  }
}
