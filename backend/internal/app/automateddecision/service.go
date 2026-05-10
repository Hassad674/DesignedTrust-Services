// Package automateddecision implements the FileAppeal use case for the
// RGPD art. 22 disclosure surface.
//
// The handler builds a FileAppealInput from the JSON body + the
// authenticated user, the service validates it via the domain
// constructor, persists it, and best-effort emails the configured
// admin contact so the human reviewer is notified within minutes.
// Email failures are logged but never abort the use case — the row
// is the source of truth.
package automateddecision

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"marketplace-backend/internal/domain/automateddecision"
	"marketplace-backend/internal/port/repository"
	"marketplace-backend/internal/port/service"
)

// ServiceDeps groups every collaborator the FileAppeal use case
// touches. Bundled into a struct so the constructor stays under the
// project's 4-parameter ceiling.
type ServiceDeps struct {
	Repo       repository.AutomatedDecisionAppealRepository
	Email      service.EmailService
	AdminEmail string // recipient for the human-reviewer notification
}

// Service exposes the FileAppeal use case. Constructed once at boot
// in cmd/api/wire_*.go and shared across handler invocations.
type Service struct {
	repo       repository.AutomatedDecisionAppealRepository
	email      service.EmailService
	adminEmail string
}

// NewService is the canonical constructor. The repo and email
// dependencies are port interfaces — tests inject fakes; production
// wires the postgres + Resend implementations.
func NewService(deps ServiceDeps) *Service {
	return &Service{
		repo:       deps.Repo,
		email:      deps.Email,
		adminEmail: strings.TrimSpace(deps.AdminEmail),
	}
}

// FileAppealInput is the handler-shaped input. The service hands the
// fields straight to the domain constructor so the boundary stays
// thin.
type FileAppealInput struct {
	UserID       uuid.UUID
	DecisionType string
	ReferenceID  string
	Reason       string
}

// FileAppeal validates the input, persists the appeal, then best-effort
// notifies the admin contact. Returns the persisted entity so the
// handler can echo the canonical id and timestamps.
func (s *Service) FileAppeal(
	ctx context.Context,
	in FileAppealInput,
) (*automateddecision.Appeal, error) {
	appeal, err := automateddecision.New(automateddecision.NewInput{
		UserID:       in.UserID,
		DecisionType: automateddecision.DecisionType(strings.TrimSpace(in.DecisionType)),
		ReferenceID:  in.ReferenceID,
		Reason:       in.Reason,
	})
	if err != nil {
		return nil, fmt.Errorf("automated_decision: build appeal: %w", err)
	}

	if err := s.repo.Create(ctx, appeal); err != nil {
		return nil, fmt.Errorf("automated_decision: persist appeal: %w", err)
	}

	s.notifyAdminBestEffort(ctx, appeal)
	return appeal, nil
}

// notifyAdminBestEffort sends the admin notification email. Wrapped
// so a transient SMTP / Resend outage never bubbles up to the user
// — the appeal row is already committed.
func (s *Service) notifyAdminBestEffort(
	ctx context.Context,
	appeal *automateddecision.Appeal,
) {
	if s.email == nil || s.adminEmail == "" {
		return
	}
	subject := fmt.Sprintf("[RGPD art. 22] Nouvelle demande de revue humaine — %s", appeal.DecisionType)
	body := buildAdminEmailHTML(appeal)
	if err := s.email.SendNotification(ctx, s.adminEmail, subject, body); err != nil {
		slog.Error("automated_decision: admin notification failed",
			"appeal_id", appeal.ID,
			"error", err.Error())
	}
}

// buildAdminEmailHTML assembles a minimal HTML body for the admin
// inbox. Every dynamic field is run through html.EscapeString so a
// hostile reason payload cannot inject markup into the admin's email
// client.
func buildAdminEmailHTML(appeal *automateddecision.Appeal) string {
	return fmt.Sprintf(
		`<p>Une nouvelle demande de revue humaine vient d'être déposée.</p>`+
			`<ul>`+
			`<li><strong>ID :</strong> %s</li>`+
			`<li><strong>Type de décision :</strong> %s</li>`+
			`<li><strong>Référence :</strong> %s</li>`+
			`<li><strong>Utilisateur :</strong> %s</li>`+
			`<li><strong>Reçu le :</strong> %s UTC</li>`+
			`</ul>`+
			`<p><strong>Motif :</strong></p>`+
			`<blockquote>%s</blockquote>`,
		html.EscapeString(appeal.ID.String()),
		html.EscapeString(string(appeal.DecisionType)),
		html.EscapeString(appeal.ReferenceID),
		html.EscapeString(appeal.UserID.String()),
		html.EscapeString(appeal.CreatedAt.Format("2006-01-02 15:04:05")),
		html.EscapeString(appeal.Reason),
	)
}
