import { apiClient } from "@/shared/lib/api-client"
import type { Post } from "@/shared/lib/api-paths"
import type {
  PresignFeedbackAttachmentRequest,
  PresignFeedbackAttachmentResponse,
  SubmitFeedbackRequest,
  SubmitFeedbackResponse,
} from "../types"

const SUBMIT_PATH = "/api/v1/feedback"
const PRESIGN_PATH = "/api/v1/feedback/attachments/presign"

/**
 * submitFeedback posts a bug or security report. PUBLIC — an anonymous
 * visitor may submit text-only; only a logged-in reporter can carry
 * `attachment_keys`. The backend recognises the session via the
 * httpOnly cookie that `apiClient` forwards (`credentials: include`),
 * so the caller never passes an identity. The response is the bare
 * report summary (id/type/status/created_at) — these endpoints encode
 * their payload directly, not inside a `{ data }` envelope.
 */
export function submitFeedback(
  body: SubmitFeedbackRequest,
): Promise<SubmitFeedbackResponse> {
  return apiClient<Post<typeof SUBMIT_PATH>>(SUBMIT_PATH, {
    method: "POST",
    body,
  })
}

/**
 * presignFeedbackAttachment asks the backend for a short-lived PUT URL
 * plus the server-minted object key for a single media file. AUTH
 * REQUIRED — an anonymous caller gets a 401 (surfaced as an ApiError).
 * The returned `object_key` is later echoed back in the submit body's
 * `attachment_keys`; the client never invents it.
 */
export function presignFeedbackAttachment(
  body: PresignFeedbackAttachmentRequest,
): Promise<PresignFeedbackAttachmentResponse> {
  return apiClient<Post<typeof PRESIGN_PATH>>(PRESIGN_PATH, {
    method: "POST",
    body,
  })
}
