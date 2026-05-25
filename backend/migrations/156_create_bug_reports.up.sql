-- Platform feedback feature: bug reports + security/vulnerability
-- disclosures. PLATFORM feedback, NOT business state — ownership is by
-- reporter_user_id authorship (the documented user_id-for-authorship
-- exception), never organization_id. Anonymous submissions are allowed
-- (reporter_user_id NULL). Within-feature FKs (attachments/notes ->
-- bug_reports) are allowed; the only cross-feature FK is to users(id)
-- for authorship, with ON DELETE SET NULL so deleting a user never
-- cascades away the feedback record itself.

CREATE TABLE IF NOT EXISTS bug_reports (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    type             TEXT NOT NULL CHECK (type IN ('bug', 'security')),
    title            TEXT NOT NULL,
    description      TEXT NOT NULL,
    status           TEXT NOT NULL DEFAULT 'new'
                         CHECK (status IN ('new', 'triaged', 'in_progress', 'resolved', 'rejected')),
    severity         TEXT CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    page_url         TEXT,
    context          JSONB NOT NULL DEFAULT '{}',
    reporter_email   TEXT,
    ip_hash          TEXT,
    resolved_at      TIMESTAMPTZ,
    resolved_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bug_reports_status ON bug_reports(status);
CREATE INDEX IF NOT EXISTS idx_bug_reports_type ON bug_reports(type);
CREATE INDEX IF NOT EXISTS idx_bug_reports_created_at ON bug_reports(created_at DESC);
-- Index the authorship FK so a per-user lookup / ON DELETE SET NULL scan
-- does not seq-scan (PostgreSQL does not auto-index FKs).
CREATE INDEX IF NOT EXISTS idx_bug_reports_reporter ON bug_reports(reporter_user_id)
    WHERE reporter_user_id IS NOT NULL;

CREATE TRIGGER set_bug_reports_updated_at
    BEFORE UPDATE ON bug_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Media attachments (image | video) — logged-in reporters only. The
-- bug_report_id FK is WITHIN this feature so cascading is allowed:
-- deleting a report removes its attachments and notes.
CREATE TABLE IF NOT EXISTS bug_report_attachments (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bug_report_id UUID NOT NULL REFERENCES bug_reports(id) ON DELETE CASCADE,
    kind          TEXT NOT NULL CHECK (kind IN ('image', 'video')),
    object_key    TEXT NOT NULL,
    content_type  TEXT,
    size_bytes    BIGINT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bug_report_attachments_report
    ON bug_report_attachments(bug_report_id);

CREATE TRIGGER set_bug_report_attachments_updated_at
    BEFORE UPDATE ON bug_report_attachments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Internal admin triage notes. Authored by an admin user; never shown
-- to reporters.
CREATE TABLE IF NOT EXISTS bug_report_notes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bug_report_id UUID NOT NULL REFERENCES bug_reports(id) ON DELETE CASCADE,
    admin_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body          TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bug_report_notes_report
    ON bug_report_notes(bug_report_id);

CREATE TRIGGER set_bug_report_notes_updated_at
    BEFORE UPDATE ON bug_report_notes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
