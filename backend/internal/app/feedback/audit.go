package feedback

import (
	"context"

	"github.com/google/uuid"
)

// AdminAction labels the admin mutation being audited. Kept as a small
// local enum so the feedback feature does not depend on the platform's
// global audit-action catalogue (which would couple the two and break
// the "deleting this feature compiles cleanly" invariant). The wiring
// layer maps these onto canonical audit actions.
type AdminAction string

const (
	// AdminActionUpdate is emitted when an admin changes a report's
	// status and/or severity.
	AdminActionUpdate AdminAction = "feedback.report_updated"
	// AdminActionNote is emitted when an admin adds an internal note.
	AdminActionNote AdminAction = "feedback.note_added"
)

// AuditEvent is the feature-local description of an admin mutation worth
// recording. The wiring layer's auditRecorder implementation translates
// it into the platform's append-only audit row.
type AuditEvent struct {
	AdminUserID uuid.UUID
	Action      AdminAction
	ReportID    uuid.UUID
	IPAddress   string
	Metadata    map[string]any
}

// auditRecorder is the narrow port the feedback service uses to record
// admin mutations. Optional at construction — when nil the service
// simply skips auditing. The production implementation (wired in
// cmd/api/main.go) bridges to the platform AuditRepository, keeping the
// audit-domain coupling out of this feature package.
//
// RecordAdminAction MUST be best-effort and MUST NOT return an error
// that fails the caller's main operation; the service logs and
// continues regardless.
type auditRecorder interface {
	RecordAdminAction(ctx context.Context, event AuditEvent)
}
