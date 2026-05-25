package main

import (
	"context"
	"database/sql"
	"log/slog"

	"marketplace-backend/internal/adapter/postgres"
	feedbackapp "marketplace-backend/internal/app/feedback"
	"marketplace-backend/internal/domain/audit"
	"marketplace-backend/internal/handler"
	"marketplace-backend/internal/port/repository"
	"marketplace-backend/internal/port/service"
)

// feedbackWiring carries the products of the platform feedback feature:
// both HTTP handlers (public submit + admin triage). The repository and
// service are private — no other feature reaches into them.
type feedbackWiring struct {
	Handler      *handler.FeedbackHandler
	AdminHandler *handler.AdminFeedbackHandler
}

// feedbackDeps captures the upstream collaborators the feedback feature
// needs: the SQL pool, the shared storage port (presigned R2 uploads /
// downloads), the audit repository (best-effort admin-mutation trail),
// and the GDPR anonymization salt (used to hash submitter IPs — the raw
// IP is never persisted).
type feedbackDeps struct {
	DB                *sql.DB
	Storage           service.StorageService
	Audit             repository.AuditRepository
	AnonymizationSalt string
}

// wireFeedback brings up the platform feedback feature: repository, app
// service (with the audit bridge + IP-hash salt), and both HTTP
// handlers. Fully isolated — deleting this wire + the feedback packages
// removes the feature with zero impact elsewhere.
func wireFeedback(deps feedbackDeps) feedbackWiring {
	feedbackRepo := postgres.NewFeedbackRepository(deps.DB)
	feedbackSvc := feedbackapp.NewService(feedbackapp.ServiceDeps{
		Reports:           feedbackRepo,
		Storage:           deps.Storage,
		Audit:             feedbackAuditBridge{audit: deps.Audit},
		AnonymizationSalt: deps.AnonymizationSalt,
	})
	return feedbackWiring{
		Handler:      handler.NewFeedbackHandler(feedbackSvc),
		AdminHandler: handler.NewAdminFeedbackHandler(feedbackSvc),
	}
}

// feedbackAuditBridge adapts the feedback feature's feature-local
// AuditEvent onto the platform's canonical append-only audit log. It
// keeps the audit-domain coupling out of the feedback app package (so
// the feature stays removable) while still recording admin mutations in
// the same trail as every other regulated action.
//
// Best-effort by contract: a build/insert failure is logged and
// swallowed — auditing must never fail the admin's operation.
type feedbackAuditBridge struct {
	audit repository.AuditRepository
}

// RecordAdminAction maps a feedback AuditEvent to a canonical audit
// Entry and logs it. The feature-local action label is translated to
// the matching audit.Action constant; an unknown label is dropped with
// a warning rather than recording a mislabeled row.
func (b feedbackAuditBridge) RecordAdminAction(ctx context.Context, event feedbackapp.AuditEvent) {
	if b.audit == nil {
		return
	}
	action, ok := feedbackAuditAction(event.Action)
	if !ok {
		slog.Warn("feedback audit: unknown action label", "action", event.Action)
		return
	}
	adminID := event.AdminUserID
	reportID := event.ReportID
	entry, err := audit.NewEntry(audit.NewEntryInput{
		UserID:       &adminID,
		Action:       action,
		ResourceType: audit.ResourceTypeFeedbackReport,
		ResourceID:   &reportID,
		Metadata:     event.Metadata,
		IPAddress:    event.IPAddress,
	})
	if err != nil {
		slog.Warn("feedback audit: build entry failed", "action", action, "error", err)
		return
	}
	if err := b.audit.Log(ctx, entry); err != nil {
		slog.Warn("feedback audit: insert failed", "action", action, "error", err)
	}
}

// feedbackAuditAction maps a feature-local admin action label to the
// canonical audit action constant.
func feedbackAuditAction(a feedbackapp.AdminAction) (audit.Action, bool) {
	switch a {
	case feedbackapp.AdminActionUpdate:
		return audit.ActionFeedbackReportUpdated, true
	case feedbackapp.AdminActionNote:
		return audit.ActionFeedbackNoteAdded, true
	default:
		return "", false
	}
}
