package response

import (
	"time"

	"github.com/google/uuid"

	domain "marketplace-backend/internal/domain/feedback"
)

// SubmitFeedbackResponse is the minimal acknowledgement returned to a
// reporter after a successful submission. Reporters (including admins
// submitting) never receive triage internals — just the id, type and
// status so a client can show a confirmation.
type SubmitFeedbackResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// SubmitFeedbackResponseFrom builds the submit acknowledgement.
func SubmitFeedbackResponseFrom(r *domain.Report) SubmitFeedbackResponse {
	return SubmitFeedbackResponse{
		ID:        r.ID.String(),
		Type:      string(r.Type),
		Status:    string(r.Status),
		CreatedAt: r.CreatedAt,
	}
}

// PresignFeedbackAttachmentResponse is the presign envelope: a
// short-lived PUT URL and the server-minted object key the client must
// echo back in the submit payload.
type PresignFeedbackAttachmentResponse struct {
	UploadURL string `json:"url"`
	ObjectKey string `json:"object_key"`
	Kind      string `json:"kind"`
}

// AdminFeedbackReportResponse is an admin list-row projection of a
// report plus its attachment + note counts.
type AdminFeedbackReportResponse struct {
	ID              string         `json:"id"`
	ReporterUserID  *string        `json:"reporter_user_id"`
	Type            string         `json:"type"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	Status          string         `json:"status"`
	Severity        string         `json:"severity"`
	PageURL         string         `json:"page_url"`
	ReporterEmail   string         `json:"reporter_email"`
	Context         map[string]any `json:"context"`
	AttachmentCount int            `json:"attachment_count"`
	NoteCount       int            `json:"note_count"`
	ResolvedAt      *time.Time     `json:"resolved_at"`
	ResolvedBy      *string        `json:"resolved_by"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// AdminFeedbackReportFrom builds an admin list-row response from a
// report and its aggregate counts.
func AdminFeedbackReportFrom(r *domain.Report, attachmentCount, noteCount int) AdminFeedbackReportResponse {
	return AdminFeedbackReportResponse{
		ID:              r.ID.String(),
		ReporterUserID:  uuidPtrString(r.ReporterID),
		Type:            string(r.Type),
		Title:           r.Title,
		Description:     r.Description,
		Status:          string(r.Status),
		Severity:        string(r.Severity),
		PageURL:         r.PageURL,
		ReporterEmail:   r.ReporterEmail,
		Context:         r.Context,
		AttachmentCount: attachmentCount,
		NoteCount:       noteCount,
		ResolvedAt:      r.ResolvedAt,
		ResolvedBy:      uuidPtrString(r.ResolvedBy),
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// FeedbackAttachmentResponse is an attachment plus its on-demand
// presigned GET URL (admin-only — the bucket stays private).
type FeedbackAttachmentResponse struct {
	ID           string    `json:"id"`
	Kind         string    `json:"kind"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	PresignedURL string    `json:"url"`
	CreatedAt    time.Time `json:"created_at"`
}

// FeedbackNoteResponse is an internal admin note.
type FeedbackNoteResponse struct {
	ID          string    `json:"id"`
	AdminUserID string    `json:"admin_user_id"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
}

// FeedbackNoteFrom builds a note response.
func FeedbackNoteFrom(n *domain.Note) FeedbackNoteResponse {
	return FeedbackNoteResponse{
		ID:          n.ID.String(),
		AdminUserID: n.AdminUserID.String(),
		Body:        n.Body,
		CreatedAt:   n.CreatedAt,
	}
}

// AdminFeedbackReportDetailResponse is the full admin detail view: the
// report row, its attachments (each with a presigned GET URL), and its
// notes (newest first).
type AdminFeedbackReportDetailResponse struct {
	AdminFeedbackReportResponse
	Attachments []FeedbackAttachmentResponse `json:"attachments"`
	Notes       []FeedbackNoteResponse       `json:"notes"`
}

// uuidPtrString renders a *uuid.UUID as a *string, preserving nil so an
// anonymous reporter / un-resolved report serialises the field as JSON
// null rather than an empty UUID.
func uuidPtrString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}
