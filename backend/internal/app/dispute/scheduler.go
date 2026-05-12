package dispute

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	disputedomain "marketplace-backend/internal/domain/dispute"
	"marketplace-backend/internal/domain/message"
	proposaldomain "marketplace-backend/internal/domain/proposal"
	"marketplace-backend/internal/port/repository"
	"marketplace-backend/internal/port/service"
	"marketplace-backend/internal/system"
)

// schedulerDisputes is the local composite the dispute scheduler
// needs: it lists pending disputes (Reader) and updates them when
// auto-resolution or escalation fires (Writer). No segregated child
// covers both — composing locally keeps the wide port out of the
// dependency graph.
type schedulerDisputes interface {
	repository.DisputeReader
	repository.DisputeWriter
}

// SchedulerDeps groups dependencies for the dispute scheduler.
//
// Proposals reuses the disputeProposals composite — the auto-resolve
// path reads the source proposal and updates its escrow / dispute
// flags when the respondent never replies.
type SchedulerDeps struct {
	Svc           *Service // canonical escalation routine lives here
	Disputes      schedulerDisputes
	Proposals     disputeProposals
	Messages      service.MessageSender
	Notifications service.NotificationSender
	Payments      service.PaymentProcessor
}

// Scheduler periodically checks for disputes that need auto-resolution
// or escalation. Runs as a background goroutine.
//
// Escalation is fully delegated to Service.escalate so that timed and
// manual (force-escalate) escalations produce strictly identical state.
// Auto-resolution (the "ghost" path when the respondent never replies)
// stays here because it has no manual counterpart.
type Scheduler struct {
	svc           *Service
	disputes      schedulerDisputes
	proposals     disputeProposals
	messages      service.MessageSender
	notifications service.NotificationSender
	payments      service.PaymentProcessor
}

func NewScheduler(deps SchedulerDeps) *Scheduler {
	return &Scheduler{
		svc:           deps.Svc,
		disputes:      deps.Disputes,
		proposals:     deps.Proposals,
		messages:      deps.Messages,
		notifications: deps.Notifications,
		payments:      deps.Payments,
	}
}

// Run blocks until ctx is cancelled. Ticks every interval + runs immediately.
//
// P8 — defensive system-actor wrap on the goroutine root context. The
// caller in cmd/api/wire_dispute.go already wraps once, but
// duplicating the tag here costs nothing (context.WithValue is
// O(1)) and removes the dependency on every caller getting the wrap
// right. Downstream repository calls (DisputeWriter.Update,
// ProposalReader.GetByID via system-actor branch) are safe under
// NOSUPERUSER NOBYPASSRLS only when this tag is in the chain.
func (s *Scheduler) Run(ctx context.Context, interval time.Duration) {
	ctx = system.WithSystemActor(ctx)
	s.tick(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	disputes, err := s.disputes.ListPendingForScheduler(ctx)
	if err != nil {
		slog.Error("dispute scheduler: list pending", "error", err)
		return
	}
	if len(disputes) == 0 {
		return
	}

	slog.Debug("dispute scheduler: processing", "count", len(disputes))

	for _, d := range disputes {
		if d.Status == disputedomain.StatusOpen && d.RespondentFirstReplyAt == nil {
			s.autoResolve(ctx, d)
		} else {
			s.escalate(ctx, d)
		}
	}
}

// autoResolve handles the ghost scenario: respondent never replied within 7 days.
// Funds go to the initiator.
//
// On top of the dispute_auto_resolved system message, this path emits the same
// post-completion close-out as Service.restoreProposalAndDistribute (the
// amiable + admin paths): a proposal_completed + evaluation_request
// system message pair AND a proposal_completed notification to both
// parties carrying the proposal-flow payload (proposal_id +
// conversation_id + proposal_title) so the frontend can deep-link the
// conversation and auto-open the review modal. Without these the
// auto-resolved mission would silently drop the 14-day review CTA.
func (s *Scheduler) autoResolve(ctx context.Context, d *disputedomain.Dispute) {
	if err := d.AutoResolveForInitiator(); err != nil {
		slog.Error("dispute scheduler: auto-resolve", "dispute_id", d.ID, "error", err)
		return
	}
	if err := s.disputes.Update(ctx, d); err != nil {
		slog.Error("dispute scheduler: update after auto-resolve", "dispute_id", d.ID, "error", err)
		return
	}

	p := s.restoreAndDistribute(ctx, d)

	s.broadcastSystemMessage(ctx, d.ConversationID,
		message.MessageTypeDisputeAutoResolved, buildAutoResolvedMetadata(d))
	s.notifyBoth(ctx, d, "dispute_auto_resolved",
		"Litige resolu automatiquement",
		"Le litige a ete resolu automatiquement faute de reponse dans les 7 jours.")

	// Close-out: the mission is now `completed` — emit the same
	// proposal_completed + evaluation_request system messages and
	// the proposal_completed notification that the normal completion
	// path emits, so both parties get the 14-day review CTA in the
	// conversation AND a clickable notification that deep-links into
	// the review modal.
	if p != nil {
		s.emitCompletionMessages(ctx, p)
	}

	slog.Info("dispute scheduler: auto-resolved (ghost)",
		"dispute_id", d.ID, "initiator_id", d.InitiatorID)
}

// escalate delegates to Service.escalate so the scheduler and the manual
// force-escalate endpoint share the same code path. The scheduler keeps
// only the "what to log when escalation is triggered by the timer" concern.
func (s *Scheduler) escalate(ctx context.Context, d *disputedomain.Dispute) {
	if err := s.svc.escalate(ctx, d); err != nil {
		slog.Error("dispute scheduler: escalate", "dispute_id", d.ID, "error", err)
		return
	}
	slog.Info("dispute scheduler: escalated to admin",
		"dispute_id", d.ID, "has_ai_summary", d.AISummary != nil)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// restoreAndDistribute moves the proposal to `completed`, settles
// escrow, and returns the loaded proposal so the caller can pipe it
// into the close-out emission helpers without a second DB round-trip.
// Returns nil if loading or restoring the proposal failed — the
// caller is expected to skip subsequent message emission in that case.
func (s *Scheduler) restoreAndDistribute(ctx context.Context, d *disputedomain.Dispute) *proposaldomain.Proposal {
	p, err := s.proposals.GetByID(ctx, d.ProposalID)
	if err != nil {
		slog.Error("dispute scheduler: get proposal for restore", "error", err)
		return nil
	}
	if err := p.RestoreFromDispute(proposaldomain.StatusCompleted); err != nil {
		slog.Error("dispute scheduler: restore proposal", "error", err)
		return nil
	}
	_ = s.proposals.Update(ctx, p)

	if s.payments != nil {
		if d.ResolutionAmountProvider != nil && *d.ResolutionAmountProvider > 0 {
			if err := s.payments.TransferPartialToProvider(ctx, d.ProposalID, *d.ResolutionAmountProvider); err != nil {
				slog.Error("dispute scheduler: transfer to provider",
					"proposal_id", d.ProposalID, "error", err)
			}
		}
		if d.ResolutionAmountClient != nil && *d.ResolutionAmountClient > 0 {
			if err := s.payments.RefundToClient(ctx, d.ProposalID, *d.ResolutionAmountClient); err != nil {
				slog.Error("dispute scheduler: refund to client",
					"proposal_id", d.ProposalID, "error", err)
			}
		}
	}
	return p
}

// emitCompletionMessages fires the post-completion close-out:
// proposal_completed + evaluation_request system messages in the
// conversation + proposal_completed notification to both parties
// with the proposal-flow data payload. Mirrors what
// Service.restoreProposalAndDistribute does for the amiable + admin
// resolution paths so the auto-resolve ghost path produces the same
// "leave a review" CTA experience.
//
// Failures are best-effort — fund distribution has already happened
// in restoreAndDistribute, and a missed message is recoverable on
// the next view of the conversation.
func (s *Scheduler) emitCompletionMessages(ctx context.Context, p *proposaldomain.Proposal) {
	completedMeta := buildProposalCompletedMetadata(p)
	s.broadcastSystemMessage(ctx, p.ConversationID,
		message.MessageType("proposal_completed"), completedMeta)
	s.broadcastSystemMessage(ctx, p.ConversationID,
		message.MessageType("evaluation_request"), completedMeta)

	data := buildProposalCompletedNotificationData(p)
	for _, uid := range []uuid.UUID{p.ClientID, p.ProviderID} {
		if err := s.notifications.Send(ctx, service.NotificationInput{
			UserID: uid,
			Type:   "proposal_completed",
			Title:  "Mission terminée",
			Body:   "La mission est marquée comme terminée après résolution du litige. Laissez un avis avant la fin de la fenêtre de 14 jours.",
			Data:   data,
		}); err != nil {
			slog.Warn("dispute scheduler: send completion notification", "user_id", uid, "error", err)
		}
	}
}

func (s *Scheduler) broadcastSystemMessage(ctx context.Context, convID uuid.UUID, msgType message.MessageType, metadata json.RawMessage) {
	if err := s.messages.SendSystemMessage(ctx, service.SystemMessageInput{
		ConversationID: convID,
		SenderID:       uuid.Nil,
		Content:        "",
		Type:           string(msgType),
		Metadata:       metadata,
	}); err != nil {
		slog.Warn("dispute scheduler: send system message", "type", msgType, "error", err)
	}
}

func (s *Scheduler) notifyBoth(ctx context.Context, d *disputedomain.Dispute, notifType, title, body string) {
	data, _ := json.Marshal(map[string]string{"dispute_id": d.ID.String()})
	for _, uid := range []uuid.UUID{d.InitiatorID, d.RespondentID} {
		if err := s.notifications.Send(ctx, service.NotificationInput{
			UserID: uid,
			Type:   notifType,
			Title:  title,
			Body:   body,
			Data:   data,
		}); err != nil {
			slog.Warn("dispute scheduler: notify", "user_id", uid, "error", err)
		}
	}
}
