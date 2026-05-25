package postgres

// SQL statements for the feedback feature (bug_reports,
// bug_report_attachments, bug_report_notes). Kept in their own file so
// feedback_repository.go stays focused on orchestration and well under
// the repository size limit. Every statement is fully parameterised —
// no string concatenation, ever.

const queryInsertBugReport = `
	INSERT INTO bug_reports (
		id, reporter_user_id, type, title, description, status, severity,
		page_url, context, reporter_email, ip_hash, resolved_at, resolved_by,
		created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
	)`

const queryInsertBugReportAttachment = `
	INSERT INTO bug_report_attachments (
		id, bug_report_id, kind, object_key, content_type, size_bytes,
		created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

const queryGetBugReportByID = `
	SELECT id, reporter_user_id, type, title, description, status, severity,
	       page_url, context, reporter_email, ip_hash, resolved_at, resolved_by,
	       created_at, updated_at
	FROM bug_reports
	WHERE id = $1`

// queryUpdateBugReport persists the triage fields only. Immutable
// columns (reporter, type, title, description, page_url, context,
// ip_hash, created_at) are never touched by an admin update.
const queryUpdateBugReport = `
	UPDATE bug_reports
	SET status = $2, severity = $3, resolved_at = $4, resolved_by = $5, updated_at = $6
	WHERE id = $1`

const queryListBugReportAttachments = `
	SELECT id, bug_report_id, kind, object_key, content_type, size_bytes,
	       created_at, updated_at
	FROM bug_report_attachments
	WHERE bug_report_id = $1
	ORDER BY created_at ASC, id ASC`

const queryInsertBugReportNote = `
	INSERT INTO bug_report_notes (
		id, bug_report_id, admin_user_id, body, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6)`

const queryListBugReportNotes = `
	SELECT id, bug_report_id, admin_user_id, body, created_at, updated_at
	FROM bug_report_notes
	WHERE bug_report_id = $1
	ORDER BY created_at DESC, id DESC`

// queryListBugReports is the admin queue listing. The aggregate counts
// are computed with correlated sub-selects so the whole page is one
// round-trip (no N+1 over attachments / notes). Every filter is an
// "is-NULL-or-equals" predicate so a single prepared statement serves
// every combination of {type, status, severity, search} — empty /
// sentinel parameters disable the corresponding filter.
//
// The cursor predicate uses ($6 IS NULL OR (created_at,id) < ($6,$7))
// so the first page and subsequent pages share one statement.
const queryListBugReports = `
	SELECT
		r.id, r.reporter_user_id, r.type, r.title, r.description, r.status,
		r.severity, r.page_url, r.context, r.reporter_email, r.ip_hash,
		r.resolved_at, r.resolved_by, r.created_at, r.updated_at,
		(SELECT count(*) FROM bug_report_attachments a WHERE a.bug_report_id = r.id) AS attachment_count,
		(SELECT count(*) FROM bug_report_notes n WHERE n.bug_report_id = r.id) AS note_count
	FROM bug_reports r
	WHERE ($1 = '' OR r.type = $1)
	  AND ($2 = '' OR r.status = $2)
	  AND ($3 = '' OR r.severity = $3)
	  AND ($4 = '' OR r.title ILIKE '%' || $4 || '%')
	  AND ($5::timestamptz IS NULL OR (r.created_at, r.id) < ($5, $6))
	ORDER BY r.created_at DESC, r.id DESC
	LIMIT $7`
