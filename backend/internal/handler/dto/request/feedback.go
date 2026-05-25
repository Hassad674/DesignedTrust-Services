package request

// SubmitFeedbackAttachment is a single media reference attached to a
// feedback submission by a logged-in reporter. The object_key is the
// server-minted key returned by the presign endpoint; the client never
// invents it.
type SubmitFeedbackAttachment struct {
	Kind        string `json:"kind"`
	ObjectKey   string `json:"object_key"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

// SubmitFeedbackRequest is the payload for POST /api/v1/feedback. Anyone
// (including anonymous visitors) may submit text; attachment_keys are
// honoured only for authenticated reporters and rejected otherwise.
//
// The `hp` field is a honeypot — a hidden form field a human never
// fills. A non-empty value marks the submission as a bot and the server
// silently drops it (returns 200 without persisting).
type SubmitFeedbackRequest struct {
	Type           string                     `json:"type"`
	Title          string                     `json:"title"`
	Description    string                     `json:"description"`
	PageURL        string                     `json:"page_url"`
	Context        *FeedbackContext           `json:"context"`
	ReporterEmail  string                     `json:"reporter_email"`
	AttachmentKeys []SubmitFeedbackAttachment `json:"attachment_keys"`
	HP             string                     `json:"hp"`
}

// FeedbackContext is the structured client context captured with a
// submission. Every field is optional; the whole object is optional.
// Kept as a typed struct (not a free map) so the OpenAPI contract is
// explicit and the handler rejects unknown context fields.
type FeedbackContext struct {
	Role       string `json:"role"`
	Locale     string `json:"locale"`
	Platform   string `json:"platform"`
	AppVersion string `json:"app_version"`
	Viewport   string `json:"viewport"`
	UserAgent  string `json:"user_agent"`
}

// PresignFeedbackAttachmentRequest is the payload for
// POST /api/v1/feedback/attachments/presign (auth required). The
// filename is accepted only for the human-facing UI — it never
// influences the server-minted storage key.
type PresignFeedbackAttachmentRequest struct {
	Kind        string `json:"kind"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	Filename    string `json:"filename"`
}

// UpdateFeedbackReportRequest is the payload for
// PATCH /api/v1/admin/feedback/{id}. Both fields are optional pointers
// so an admin can change status, severity, or both. A nil pointer means
// "leave unchanged"; an explicit empty-string severity clears the
// rating.
type UpdateFeedbackReportRequest struct {
	Status   *string `json:"status"`
	Severity *string `json:"severity"`
}

// AddFeedbackNoteRequest is the payload for
// POST /api/v1/admin/feedback/{id}/notes.
type AddFeedbackNoteRequest struct {
	Body string `json:"body"`
}
