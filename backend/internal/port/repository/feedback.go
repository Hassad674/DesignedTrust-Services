package repository

import (
	"context"

	"github.com/google/uuid"

	"marketplace-backend/internal/domain/feedback"
)

// ReportFilter narrows an admin report listing. Empty fields mean "no
// filter on this dimension". Bundled into a struct so the repository
// method stays under the project's 4-parameter rule as filters grow.
type ReportFilter struct {
	Type     string // "" = any; otherwise "bug" | "security"
	Status   string // "" = any; otherwise a ReportStatus value
	Severity string // "" = any; otherwise a Severity value
	Search   string // "" = no search; otherwise an ILIKE match on title
}

// ReportSummary is a list-row projection: the report plus the cheap
// aggregate counts the admin queue needs, computed in a single query to
// avoid an N+1 over attachments / notes.
type ReportSummary struct {
	Report          *feedback.Report
	AttachmentCount int
	NoteCount       int
}

// FeedbackRepository persists platform feedback reports, their media
// attachments, and internal admin notes.
//
// The interface is deliberately focused (ISP): the public submit path
// uses only CreateReport (+ AddAttachment for logged-in media); the
// admin surface uses the list / get / update / note methods. Ownership
// is by reporter_user_id authorship — there is intentionally NO
// org-scoped query here because platform feedback is not business
// state.
type FeedbackRepository interface {
	// CreateReport persists a new report and its attachments atomically.
	// attachments may be empty (text-only / anonymous submissions). The
	// implementation inserts the report row then every attachment row in
	// one transaction so a partial write never leaves orphaned media.
	CreateReport(ctx context.Context, report *feedback.Report, attachments []*feedback.Attachment) error

	// ListReports returns admin-facing report summaries filtered by the
	// given criteria, cursor-paginated newest-first. limit is clamped by
	// the implementation. Each summary carries attachment + note counts.
	ListReports(ctx context.Context, filter ReportFilter, cursor string, limit int) ([]*ReportSummary, string, error)

	// GetReport returns a single report by id, or feedback.ErrNotFound
	// when it does not exist.
	GetReport(ctx context.Context, id uuid.UUID) (*feedback.Report, error)

	// UpdateReport persists the triage fields (status, severity,
	// resolved_at, resolved_by, updated_at) of an already-mutated report.
	// Returns feedback.ErrNotFound when the row is gone.
	UpdateReport(ctx context.Context, report *feedback.Report) error

	// ListAttachments returns every attachment for a report, oldest
	// first. Used by the admin detail view to render presigned GET links.
	ListAttachments(ctx context.Context, reportID uuid.UUID) ([]*feedback.Attachment, error)

	// AddNote persists a new internal admin note.
	AddNote(ctx context.Context, note *feedback.Note) error

	// ListNotes returns every note for a report, newest first.
	ListNotes(ctx context.Context, reportID uuid.UUID) ([]*feedback.Note, error)
}
