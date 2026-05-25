package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"marketplace-backend/internal/domain/feedback"
	"marketplace-backend/internal/port/repository"
	"marketplace-backend/pkg/cursor"
)

// FeedbackRepository is the PostgreSQL adapter for the platform
// feedback feature (migration 156: bug_reports, bug_report_attachments,
// bug_report_notes). It implements repository.FeedbackRepository.
//
// Ownership is by reporter_user_id authorship — there is no org-scoped
// query here because platform feedback is not business state. The table
// is intentionally NOT RLS-protected: anonymous submissions have no
// tenant, and the admin read surface needs to see every report.
type FeedbackRepository struct {
	db *sql.DB
}

// NewFeedbackRepository creates a PostgreSQL-backed feedback repository.
func NewFeedbackRepository(db *sql.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

// CreateReport inserts the report and all its attachments in a single
// transaction so a partial write never leaves orphaned media. The
// attachments slice may be empty (text-only / anonymous submissions).
func (r *FeedbackRepository) CreateReport(ctx context.Context, report *feedback.Report, attachments []*feedback.Attachment) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	contextJSON, err := json.Marshal(report.Context)
	if err != nil {
		return fmt.Errorf("feedback create: marshal context: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("feedback create: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, queryInsertBugReport,
		report.ID,
		nullableUUID(report.ReporterID),
		string(report.Type),
		report.Title,
		report.Description,
		string(report.Status),
		nullableSeverity(report.Severity),
		nullableText(report.PageURL),
		contextJSON,
		nullableText(report.ReporterEmail),
		nullableText(report.IPHash),
		report.ResolvedAt,
		nullableUUID(report.ResolvedBy),
		report.CreatedAt,
		report.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("feedback create: insert report: %w", err)
	}

	for _, a := range attachments {
		_, err = tx.ExecContext(ctx, queryInsertBugReportAttachment,
			a.ID, a.ReportID, string(a.Kind), a.ObjectKey,
			nullableText(a.ContentType), a.SizeBytes, a.CreatedAt, a.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("feedback create: insert attachment: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("feedback create: commit: %w", err)
	}
	return nil
}

// GetReport returns a single report by id, or feedback.ErrNotFound.
func (r *FeedbackRepository) GetReport(ctx context.Context, id uuid.UUID) (*feedback.Report, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	report, err := scanFeedbackReport(r.db.QueryRowContext(ctx, queryGetBugReportByID, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, feedback.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("feedback get: %w", err)
	}
	return report, nil
}

// UpdateReport persists the triage fields of an already-mutated report.
func (r *FeedbackRepository) UpdateReport(ctx context.Context, report *feedback.Report) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	result, err := r.db.ExecContext(ctx, queryUpdateBugReport,
		report.ID,
		string(report.Status),
		nullableSeverity(report.Severity),
		report.ResolvedAt,
		nullableUUID(report.ResolvedBy),
		report.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("feedback update: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return feedback.ErrNotFound
	}
	return nil
}

// ListReports returns admin report summaries filtered + cursor-paginated
// newest-first. Empty filter fields disable the corresponding predicate.
func (r *FeedbackRepository) ListReports(ctx context.Context, filter repository.ReportFilter, cursorStr string, limit int) ([]*repository.ReportSummary, string, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var cursorCreatedAt *time.Time
	var cursorID uuid.UUID
	if cursorStr != "" {
		c, err := cursor.Decode(cursorStr)
		if err != nil {
			return nil, "", fmt.Errorf("feedback list: decode cursor: %w", err)
		}
		cursorCreatedAt = &c.CreatedAt
		cursorID = c.ID
	}

	rows, err := r.db.QueryContext(ctx, queryListBugReports,
		filter.Type, filter.Status, filter.Severity, filter.Search,
		cursorCreatedAt, cursorID, limit+1,
	)
	if err != nil {
		return nil, "", fmt.Errorf("feedback list: query: %w", err)
	}
	defer rows.Close()

	return collectReportSummaries(rows, limit)
}

// ListAttachments returns every attachment for a report, oldest first.
func (r *FeedbackRepository) ListAttachments(ctx context.Context, reportID uuid.UUID) ([]*feedback.Attachment, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, queryListBugReportAttachments, reportID)
	if err != nil {
		return nil, fmt.Errorf("feedback list attachments: %w", err)
	}
	defer rows.Close()

	var out []*feedback.Attachment
	for rows.Next() {
		a, scanErr := scanAttachment(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("feedback scan attachment: %w", scanErr)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("feedback attachments iteration: %w", err)
	}
	return out, nil
}

// AddNote persists a new internal admin note.
func (r *FeedbackRepository) AddNote(ctx context.Context, note *feedback.Note) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, queryInsertBugReportNote,
		note.ID, note.ReportID, note.AdminUserID, note.Body, note.CreatedAt, note.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("feedback add note: %w", err)
	}
	return nil
}

// ListNotes returns every note for a report, newest first.
func (r *FeedbackRepository) ListNotes(ctx context.Context, reportID uuid.UUID) ([]*feedback.Note, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, queryListBugReportNotes, reportID)
	if err != nil {
		return nil, fmt.Errorf("feedback list notes: %w", err)
	}
	defer rows.Close()

	var out []*feedback.Note
	for rows.Next() {
		n, scanErr := scanNote(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("feedback scan note: %w", scanErr)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("feedback notes iteration: %w", err)
	}
	return out, nil
}
