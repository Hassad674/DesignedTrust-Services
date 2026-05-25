import type { components } from "@/shared/types/api"

// Feature-local aliases over the OpenAPI-generated contract. Keeping a
// thin indirection here means the components/api never reach into the
// deep `components["schemas"][...]` path and the feature owns a single
// place to adjust if the contract is regenerated.

/** Request body for POST /api/v1/feedback (anonymous allowed). */
export type SubmitFeedbackRequest =
  components["schemas"]["SubmitFeedbackRequest"]

/** Response data for a successful POST /api/v1/feedback. */
export type SubmitFeedbackResponse =
  components["schemas"]["SubmitFeedbackResponse"]

/** A single server-minted attachment reference carried in the submit body. */
export type FeedbackAttachmentRef =
  SubmitFeedbackRequest["attachment_keys"][number]

/** Structured client context auto-captured with a submission. */
export type FeedbackContext = NonNullable<SubmitFeedbackRequest["context"]>

/** Request body for POST /api/v1/feedback/attachments/presign (auth only). */
export type PresignFeedbackAttachmentRequest =
  components["schemas"]["PresignFeedbackAttachmentRequest"]

/** Response data for a successful presign call. */
export type PresignFeedbackAttachmentResponse =
  components["schemas"]["PresignFeedbackAttachmentResponse"]

/** The two report categories the UI exposes. */
export type FeedbackType = "bug" | "security"

/** The two media kinds the attachment zone accepts. */
export type FeedbackAttachmentKind = "image" | "video"
