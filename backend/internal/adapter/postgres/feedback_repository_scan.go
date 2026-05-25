package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"marketplace-backend/internal/domain/feedback"
	"marketplace-backend/internal/port/repository"
	"marketplace-backend/pkg/cursor"
)

// feedbackRowScanner is satisfied by both *sql.Row and *sql.Rows.
type feedbackRowScanner interface {
	Scan(dest ...any) error
}

// scanReport materialises a bug_reports row into a domain Report. The
// optional columns (reporter_user_id, severity, page_url, reporter_email,
// ip_hash, resolved_at, resolved_by) are decoded from their nullable SQL
// types; context JSONB is decoded into a non-nil map.
func scanFeedbackReport(s feedbackRowScanner) (*feedback.Report, error) {
	var (
		report      feedback.Report
		reporterID  uuid.NullUUID
		severity    sql.NullString
		pageURL     sql.NullString
		contextJSON []byte
		email       sql.NullString
		ipHash      sql.NullString
		resolvedAt  sql.NullTime
		resolvedBy  uuid.NullUUID
		typeStr     string
		statusStr   string
	)
	err := s.Scan(
		&report.ID,
		&reporterID,
		&typeStr,
		&report.Title,
		&report.Description,
		&statusStr,
		&severity,
		&pageURL,
		&contextJSON,
		&email,
		&ipHash,
		&resolvedAt,
		&resolvedBy,
		&report.CreatedAt,
		&report.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	report.Type = feedback.ReportType(typeStr)
	report.Status = feedback.ReportStatus(statusStr)
	if reporterID.Valid {
		id := reporterID.UUID
		report.ReporterID = &id
	}
	if severity.Valid {
		report.Severity = feedback.Severity(severity.String)
	}
	if pageURL.Valid {
		report.PageURL = pageURL.String
	}
	if email.Valid {
		report.ReporterEmail = email.String
	}
	if ipHash.Valid {
		report.IPHash = ipHash.String
	}
	if resolvedAt.Valid {
		t := resolvedAt.Time
		report.ResolvedAt = &t
	}
	if resolvedBy.Valid {
		id := resolvedBy.UUID
		report.ResolvedBy = &id
	}
	report.Context = decodeContext(contextJSON)
	return &report, nil
}

// scanReportSummary materialises a listing row (report columns + the two
// aggregate counts) into a repository.ReportSummary.
func scanReportSummary(rows *sql.Rows) (*repository.ReportSummary, error) {
	var (
		report      feedback.Report
		reporterID  uuid.NullUUID
		severity    sql.NullString
		pageURL     sql.NullString
		contextJSON []byte
		email       sql.NullString
		ipHash      sql.NullString
		resolvedAt  sql.NullTime
		resolvedBy  uuid.NullUUID
		typeStr     string
		statusStr   string
		attachments int
		notes       int
	)
	err := rows.Scan(
		&report.ID,
		&reporterID,
		&typeStr,
		&report.Title,
		&report.Description,
		&statusStr,
		&severity,
		&pageURL,
		&contextJSON,
		&email,
		&ipHash,
		&resolvedAt,
		&resolvedBy,
		&report.CreatedAt,
		&report.UpdatedAt,
		&attachments,
		&notes,
	)
	if err != nil {
		return nil, err
	}

	report.Type = feedback.ReportType(typeStr)
	report.Status = feedback.ReportStatus(statusStr)
	if reporterID.Valid {
		id := reporterID.UUID
		report.ReporterID = &id
	}
	if severity.Valid {
		report.Severity = feedback.Severity(severity.String)
	}
	if pageURL.Valid {
		report.PageURL = pageURL.String
	}
	if email.Valid {
		report.ReporterEmail = email.String
	}
	if ipHash.Valid {
		report.IPHash = ipHash.String
	}
	if resolvedAt.Valid {
		t := resolvedAt.Time
		report.ResolvedAt = &t
	}
	if resolvedBy.Valid {
		id := resolvedBy.UUID
		report.ResolvedBy = &id
	}
	report.Context = decodeContext(contextJSON)

	return &repository.ReportSummary{
		Report:          &report,
		AttachmentCount: attachments,
		NoteCount:       notes,
	}, nil
}

// collectReportSummaries walks the result set, peels the extra row used
// to detect "has more", and encodes the next cursor from the last item
// that stays in the returned slice.
func collectReportSummaries(rows *sql.Rows, limit int) ([]*repository.ReportSummary, string, error) {
	var summaries []*repository.ReportSummary
	for rows.Next() {
		summary, err := scanReportSummary(rows)
		if err != nil {
			return nil, "", fmt.Errorf("feedback scan summary: %w", err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("feedback summaries iteration: %w", err)
	}

	var nextCursor string
	if len(summaries) > limit {
		last := summaries[limit-1].Report
		nextCursor = cursor.Encode(last.CreatedAt, last.ID)
		summaries = summaries[:limit]
	}
	return summaries, nextCursor, nil
}

// scanAttachment materialises a bug_report_attachments row.
func scanAttachment(rows *sql.Rows) (*feedback.Attachment, error) {
	var (
		a           feedback.Attachment
		kindStr     string
		contentType sql.NullString
	)
	err := rows.Scan(
		&a.ID,
		&a.ReportID,
		&kindStr,
		&a.ObjectKey,
		&contentType,
		&a.SizeBytes,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	a.Kind = feedback.AttachmentKind(kindStr)
	if contentType.Valid {
		a.ContentType = contentType.String
	}
	return &a, nil
}

// scanNote materialises a bug_report_notes row.
func scanNote(rows *sql.Rows) (*feedback.Note, error) {
	var n feedback.Note
	err := rows.Scan(
		&n.ID,
		&n.ReportID,
		&n.AdminUserID,
		&n.Body,
		&n.CreatedAt,
		&n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// decodeContext turns a JSONB context column into a non-nil map. Empty
// or corrupt bytes yield an empty map so callers never dereference nil.
func decodeContext(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	if out == nil {
		return map[string]any{}
	}
	return out
}

// nullableText returns a NULL-able string: empty becomes SQL NULL so the
// column stays clean (no empty-string sentinels).
func nullableText(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// nullableUUID returns the UUID string or SQL NULL for a nil pointer.
func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}

// nullableSeverity returns the severity string or SQL NULL for the empty
// (un-triaged) severity. The bug_reports.severity column is nullable and
// CHECK-constrained, so an empty severity must be stored as NULL — not
// "" which would violate the check.
func nullableSeverity(s feedback.Severity) any {
	if s == "" {
		return nil
	}
	return string(s)
}
