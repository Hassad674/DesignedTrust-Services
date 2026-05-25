package feedback

import (
	"time"

	"github.com/google/uuid"
)

// Update is a partial admin mutation of a report's triage fields. Both
// fields are optional pointers so an admin can change status, severity,
// or both in a single PATCH. A nil pointer means "leave unchanged".
type Update struct {
	Status   *ReportStatus
	Severity *Severity
}

// HasChanges reports whether the update carries at least one field.
func (u Update) HasChanges() bool {
	return u.Status != nil || u.Severity != nil
}

// Apply validates and applies the update to the report in place,
// stamping resolved_at / resolved_by when the status transitions into a
// terminal state and clearing them when it leaves one. The adminID is
// recorded as the resolver. Returns a sentinel domain error on an
// invalid status / severity or when no field was provided.
//
// Severity may be cleared by passing a pointer to the empty Severity
// ("") — this un-sets a previously assigned rating. Any non-empty
// severity must be a known value.
func (u Update) Apply(r *Report, adminID uuid.UUID, now time.Time) error {
	if !u.HasChanges() {
		return ErrNoUpdateFields
	}

	if u.Severity != nil {
		sev := *u.Severity
		if sev != "" && !sev.IsValid() {
			return ErrInvalidSeverity
		}
		r.Severity = sev
	}

	if u.Status != nil {
		next := *u.Status
		if !next.IsValid() {
			return ErrInvalidStatus
		}
		r.Status = next
		if next.IsTerminal() {
			resolvedAt := now.UTC()
			resolver := adminID
			r.ResolvedAt = &resolvedAt
			r.ResolvedBy = &resolver
		} else {
			// Leaving a terminal state clears the resolution stamp so the
			// row never claims to be both "in_progress" and "resolved at".
			r.ResolvedAt = nil
			r.ResolvedBy = nil
		}
	}

	r.UpdatedAt = now.UTC()
	return nil
}
