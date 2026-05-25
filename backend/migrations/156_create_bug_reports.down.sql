-- Reverse of 156_create_bug_reports.up.sql. Drop child tables first
-- (they FK into bug_reports), each with its trigger, then the parent.
-- IF EXISTS everywhere so the rollback is safe to retry on a partially
-- applied state.

DROP TRIGGER IF EXISTS set_bug_report_notes_updated_at ON bug_report_notes;
DROP TABLE IF EXISTS bug_report_notes;

DROP TRIGGER IF EXISTS set_bug_report_attachments_updated_at ON bug_report_attachments;
DROP TABLE IF EXISTS bug_report_attachments;

DROP TRIGGER IF EXISTS set_bug_reports_updated_at ON bug_reports;
DROP TABLE IF EXISTS bug_reports;
