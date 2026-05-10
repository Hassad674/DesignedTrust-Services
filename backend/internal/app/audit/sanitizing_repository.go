// Package audit holds the app-layer wrappers around the audit log
// repository.
//
// There is no AuditService — audit writes are emitted directly from
// each business service (auth, admin, moderation, referral, …) so a
// centralized service would only forward calls. Instead, the app
// layer ships ONE concern: sanitization of personally identifiable
// information (PII) carried in `metadata` before it touches the DB.
//
// SanitizingRepository wraps a `repository.AuditRepository`,
// transforms `entry.Metadata` via `domain/audit.SanitizeMetadata` on
// every Log call, and delegates ListByResource / ListByUser
// unchanged. Wired in cmd/api/wire_infra.go so every business
// service that writes to the audit log goes through the redacting
// path with zero per-call-site changes.
package audit

import (
	"context"

	"github.com/google/uuid"

	"marketplace-backend/internal/domain/audit"
	"marketplace-backend/internal/port/repository"
)

// SanitizingRepository decorates a `repository.AuditRepository` with
// PII redaction. It satisfies the same interface, so it is a
// drop-in substitute at the wiring layer.
type SanitizingRepository struct {
	inner repository.AuditRepository
}

// NewSanitizingRepository wraps `inner` so every Log call has its
// metadata redacted by `audit.SanitizeMetadata` before persistence.
// List methods are forwarded unchanged — sanitization is a
// pre-storage transform, not a query-time one, so existing rows
// (post-backfill migration) and new rows return identically.
func NewSanitizingRepository(inner repository.AuditRepository) *SanitizingRepository {
	return &SanitizingRepository{inner: inner}
}

// Log redacts `entry.Metadata` in place against the sensitive-key
// allow-list and forwards to the wrapped repository. The transform
// is intentionally applied to a copy returned by SanitizeMetadata —
// the Entry is mutated by reassigning the field, but the original
// map provided by the caller is left untouched (see
// `audit.SanitizeMetadata` test contract). The caller usually
// constructs the Entry inline so this matters for one case only:
// tests that retain a reference to the input map.
func (r *SanitizingRepository) Log(ctx context.Context, entry *audit.Entry) error {
	if entry != nil {
		entry.Metadata = audit.SanitizeMetadata(entry.Metadata)
	}
	return r.inner.Log(ctx, entry)
}

// ListByResource forwards to the inner repository unchanged. Stored
// rows are already sanitized (new rows by Log, legacy rows by the
// 144_audit_logs_sanitize_pii.up.sql backfill).
func (r *SanitizingRepository) ListByResource(
	ctx context.Context,
	resourceType audit.ResourceType,
	resourceID uuid.UUID,
	cursor string,
	limit int,
) ([]*audit.Entry, string, error) {
	return r.inner.ListByResource(ctx, resourceType, resourceID, cursor, limit)
}

// ListByUser forwards to the inner repository unchanged.
func (r *SanitizingRepository) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	cursor string,
	limit int,
) ([]*audit.Entry, string, error) {
	return r.inner.ListByUser(ctx, userID, cursor, limit)
}
