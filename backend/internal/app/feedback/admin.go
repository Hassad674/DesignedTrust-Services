package feedback

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domain "marketplace-backend/internal/domain/feedback"
	"marketplace-backend/internal/port/repository"
)

// PresignInput is the validated payload for an attachment presign
// request. The endpoint is authenticated (media is logged-in only) so
// there is no anonymous path here.
type PresignInput struct {
	Kind        string
	ContentType string
	SizeBytes   int64
}

// PresignResult is the presign envelope returned to the client: a
// short-lived PUT URL, the server-minted object key (never derived from
// a client filename), and the resolved attachment kind.
type PresignResult struct {
	UploadURL string
	ObjectKey string
	Kind      string
}

// PresignAttachment validates the upload request against the
// content-type allowlist + per-kind size cap, mints a randomized,
// namespaced object key (reports/{uuid}/{uuid}.{ext}) — never the
// client filename — and asks the storage port for a short-lived
// presigned PUT URL. Returns a domain validation error on a disallowed
// type or oversized payload.
func (s *Service) PresignAttachment(ctx context.Context, in PresignInput) (*PresignResult, error) {
	kind, ext, err := domain.ValidatePresign(domain.AttachmentKind(in.Kind), in.ContentType, in.SizeBytes)
	if err != nil {
		return nil, err
	}

	objectKey := fmt.Sprintf("reports/%s/%s.%s", uuid.New().String(), uuid.New().String(), ext)
	uploadURL, err := s.storage.GetPresignedUploadURL(ctx, objectKey, in.ContentType, s.presignExpiry)
	if err != nil {
		return nil, fmt.Errorf("presign attachment: %w", err)
	}
	return &PresignResult{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		Kind:      string(kind),
	}, nil
}

// ListReports returns admin-facing report summaries filtered by the
// given criteria, cursor-paginated newest-first. The limit is clamped
// to a sane window.
func (s *Service) ListReports(ctx context.Context, filter repository.ReportFilter, cursor string, limit int) ([]*repository.ReportSummary, string, error) {
	limit = clampLimit(limit)
	return s.reports.ListReports(ctx, filter, cursor, limit)
}

// ReportDetail bundles a report with its attachments and notes for the
// admin detail view. Attachments carry a freshly presigned GET URL so
// the bucket stays private — see GetReport.
type ReportDetail struct {
	Report      *domain.Report
	Attachments []AttachmentView
	Notes       []*domain.Note
}

// AttachmentView is an attachment plus a short-lived presigned GET URL
// generated on demand for an admin. The object is NEVER served via a
// public URL — only admins, only through a one-shot signed link.
type AttachmentView struct {
	Attachment   *domain.Attachment
	PresignedURL string
}

// GetReport returns the full admin detail for a report: the report
// itself, its attachments (each with a presigned GET URL), and its
// notes (newest first). Returns domain.ErrNotFound when the report does
// not exist.
func (s *Service) GetReport(ctx context.Context, id uuid.UUID) (*ReportDetail, error) {
	report, err := s.reports.GetReport(ctx, id)
	if err != nil {
		return nil, err
	}

	rawAttachments, err := s.reports.ListAttachments(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get report: list attachments: %w", err)
	}
	attachments := make([]AttachmentView, 0, len(rawAttachments))
	for _, a := range rawAttachments {
		// Presign each object on demand. A failure to sign a single
		// object must not blank the whole detail view — emit the row
		// with an empty URL so the admin still sees the metadata.
		url, signErr := s.storage.GetPresignedDownloadURL(ctx, a.ObjectKey, s.presignExpiry)
		if signErr != nil {
			url = ""
		}
		attachments = append(attachments, AttachmentView{Attachment: a, PresignedURL: url})
	}

	notes, err := s.reports.ListNotes(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get report: list notes: %w", err)
	}

	return &ReportDetail{Report: report, Attachments: attachments, Notes: notes}, nil
}

// UpdateReport applies an admin triage update (status and/or severity)
// to a report and persists it. Transitioning into a terminal state
// stamps resolved_at / resolved_by with the admin's id. Emits a
// best-effort audit row. Returns domain.ErrNotFound when the report is
// gone, or a domain validation error for an invalid status / severity /
// empty update.
func (s *Service) UpdateReport(ctx context.Context, id uuid.UUID, adminID uuid.UUID, update domain.Update, ipAddress string) (*domain.Report, error) {
	report, err := s.reports.GetReport(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := update.Apply(report, adminID, time.Now()); err != nil {
		return nil, err
	}
	if err := s.reports.UpdateReport(ctx, report); err != nil {
		return nil, fmt.Errorf("update report: persist: %w", err)
	}

	s.recordAudit(ctx, AuditEvent{
		AdminUserID: adminID,
		Action:      AdminActionUpdate,
		ReportID:    id,
		IPAddress:   ipAddress,
		Metadata: map[string]any{
			"status":   string(report.Status),
			"severity": string(report.Severity),
		},
	})
	return report, nil
}

// AddNote creates an internal admin note on a report. Verifies the
// report exists first so a note never dangles. Emits a best-effort
// audit row. Returns domain.ErrNotFound when the report is gone or a
// domain validation error for an empty / oversized body.
func (s *Service) AddNote(ctx context.Context, reportID uuid.UUID, adminID uuid.UUID, body string, ipAddress string) (*domain.Note, error) {
	// Confirm the parent report exists before constructing the note so
	// the caller gets a clean 404 instead of an FK violation.
	if _, err := s.reports.GetReport(ctx, reportID); err != nil {
		return nil, err
	}
	note, err := domain.NewNote(domain.NewNoteInput{
		ReportID:    reportID,
		AdminUserID: adminID,
		Body:        body,
	})
	if err != nil {
		return nil, err
	}
	if err := s.reports.AddNote(ctx, note); err != nil {
		return nil, fmt.Errorf("add note: persist: %w", err)
	}

	s.recordAudit(ctx, AuditEvent{
		AdminUserID: adminID,
		Action:      AdminActionNote,
		ReportID:    reportID,
		IPAddress:   ipAddress,
	})
	return note, nil
}

// recordAudit records an admin mutation when an audit recorder is
// wired. Best-effort: a nil recorder is a no-op and the recorder's
// implementation must never fail the caller.
func (s *Service) recordAudit(ctx context.Context, event AuditEvent) {
	if s.audit == nil {
		return
	}
	s.audit.RecordAdminAction(ctx, event)
}

// clampLimit normalises a page size to the project's cursor-pagination
// bounds (default 20, max 100).
func clampLimit(limit int) int {
	if limit <= 0 || limit > 100 {
		return 20
	}
	return limit
}
