package dispute

// Reviews flow regression tests — covers the three dispute-resolution
// paths that must emit a proposal_completed + evaluation_request
// system message pair AND a proposal_completed notification carrying
// the proposal-flow data payload (proposal_id + conversation_id +
// proposal_title) so the frontend can deep-link the conversation and
// auto-open the review modal.
//
// Paths under test:
//   1. RespondToCounter (Accept=true)  — amiable resolution
//   2. AdminResolve                    — mediation decision
//   3. Scheduler.autoResolve           — ghost path (no respondent reply)

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	disputedomain "marketplace-backend/internal/domain/dispute"
	"marketplace-backend/internal/domain/message"
	"marketplace-backend/internal/domain/proposal"
	"marketplace-backend/internal/port/service"
)

// fixtureCompletedProposal returns a proposal already linked to a
// dispute. Tests mutate the status to Disputed before the resolution
// path runs so the domain RestoreFromDispute call succeeds.
func fixtureCompletedProposal(id, convID, clientID, providerID uuid.UUID) *proposal.Proposal {
	now := time.Now()
	return &proposal.Proposal{
		ID:              id,
		ConversationID:  convID,
		ClientID:        clientID,
		ProviderID:      providerID,
		Title:           "Site web Acme",
		Status:          proposal.StatusDisputed,
		Amount:          100000,
		Version:         1,
		CreatedAt:       now,
		UpdatedAt:       now,
		ActiveDisputeID: nil,
	}
}

// findMessage returns the first captured system message whose type
// matches the target — or nil. Used to assert the close-out pair is
// emitted alongside the dispute-specific message.
func findMessage(in []service.SystemMessageInput, t message.MessageType) *service.SystemMessageInput {
	for i := range in {
		if in[i].Type == string(t) {
			return &in[i]
		}
	}
	return nil
}

// findNotification returns the first notification matching kind +
// recipient. Used to assert both parties got the proposal_completed
// prompt with the right payload shape.
func findNotification(sent []service.NotificationInput, kind string, userID uuid.UUID) *service.NotificationInput {
	for i := range sent {
		if sent[i].Type == kind && sent[i].UserID == userID {
			return &sent[i]
		}
	}
	return nil
}

func assertProposalNotifPayload(t *testing.T, raw json.RawMessage, proposalID, convID uuid.UUID, title string) {
	t.Helper()
	var got map[string]string
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("notification data is not a string map: %v", err)
	}
	assert.Equal(t, proposalID.String(), got["proposal_id"],
		"notification must carry proposal_id so the frontend can target the review modal")
	assert.Equal(t, convID.String(), got["conversation_id"],
		"notification must carry conversation_id for navigation")
	assert.Equal(t, title, got["proposal_title"],
		"notification must carry proposal_title for the toast/header")
}

// ---------------------------------------------------------------------------
// 1. AdminResolve emits the review CTA pair + proposal-shaped notification
// ---------------------------------------------------------------------------

func TestAdminResolve_EmitsReviewCTAAndProposalNotification(t *testing.T) {
	svc, dr, pr, ms, ns, _ := newTestService()
	adminID := uuid.New()
	clientID := uuid.New()
	providerID := uuid.New()
	proposalID := uuid.New()
	convID := uuid.New()
	disputeID := uuid.New()

	dr.getByIDFn = func(_ context.Context, _ uuid.UUID) (*disputedomain.Dispute, error) {
		return &disputedomain.Dispute{
			ID: disputeID, ProposalID: proposalID,
			ConversationID: convID,
			InitiatorID:    clientID, RespondentID: providerID,
			ClientID: clientID, ProviderID: providerID,
			Status: disputedomain.StatusEscalated, ProposalAmount: 100000,
			Version: 1,
		}, nil
	}
	pr.getByIDFn = func(_ context.Context, _ uuid.UUID) (*proposal.Proposal, error) {
		p := fixtureCompletedProposal(proposalID, convID, clientID, providerID)
		p.ActiveDisputeID = &disputeID
		return p, nil
	}

	err := svc.AdminResolve(actorCtx(), AdminResolveInput{
		DisputeID:      disputeID,
		AdminID:        adminID,
		AmountClient:   40000,
		AmountProvider: 60000,
		Note:           "Split decision.",
	})
	assert.NoError(t, err)

	// Three system messages on the same conversation:
	// dispute_resolved + proposal_completed + evaluation_request.
	assert.NotNil(t, findMessage(ms.inputs, message.MessageTypeDisputeResolved),
		"dispute_resolved must be emitted")
	assert.NotNil(t, findMessage(ms.inputs, message.MessageType("proposal_completed")),
		"proposal_completed system message must be emitted so the chat shows 'Mission terminée'")
	assert.NotNil(t, findMessage(ms.inputs, message.MessageType("evaluation_request")),
		"evaluation_request must be emitted so both parties see the 'leave a review' CTA")

	// Proposal_completed notification fires to BOTH parties with the
	// proposal-flow payload shape (not just dispute_id).
	clientNotif := findNotification(ns.sent, "proposal_completed", clientID)
	providerNotif := findNotification(ns.sent, "proposal_completed", providerID)
	assert.NotNil(t, clientNotif, "client must get proposal_completed notification")
	assert.NotNil(t, providerNotif, "provider must get proposal_completed notification")
	assertProposalNotifPayload(t, clientNotif.Data, proposalID, convID, "Site web Acme")
	assertProposalNotifPayload(t, providerNotif.Data, proposalID, convID, "Site web Acme")
}

// ---------------------------------------------------------------------------
// 2. RespondToCounter (amiable accept) emits the same close-out
// ---------------------------------------------------------------------------

func TestRespondToCounter_AmiableAccept_EmitsReviewCTAAndProposalNotification(t *testing.T) {
	svc, dr, pr, ms, ns, _ := newTestService()
	clientID := uuid.New()
	providerID := uuid.New()
	proposalID := uuid.New()
	convID := uuid.New()
	disputeID := uuid.New()
	cpID := uuid.New()

	dr.getByIDFn = func(_ context.Context, _ uuid.UUID) (*disputedomain.Dispute, error) {
		return &disputedomain.Dispute{
			ID: disputeID, ProposalID: proposalID,
			ConversationID: convID,
			InitiatorID:    clientID, RespondentID: providerID,
			ClientID: clientID, ProviderID: providerID,
			Status: disputedomain.StatusNegotiation, ProposalAmount: 100000,
			Version: 1,
		}, nil
	}
	cp := &disputedomain.CounterProposal{
		ID:             cpID,
		DisputeID:      disputeID,
		ProposerID:     clientID,
		AmountClient:   30000,
		AmountProvider: 70000,
		Status:         disputedomain.CPStatusPending,
	}
	dr.getCPByIDFn = func(_ context.Context, _ uuid.UUID) (*disputedomain.CounterProposal, error) {
		return cp, nil
	}
	pr.getByIDFn = func(_ context.Context, _ uuid.UUID) (*proposal.Proposal, error) {
		p := fixtureCompletedProposal(proposalID, convID, clientID, providerID)
		p.ActiveDisputeID = &disputeID
		return p, nil
	}

	err := svc.RespondToCounter(actorCtx(), RespondToCounterInput{
		DisputeID:         disputeID,
		CounterProposalID: cpID,
		UserID:            providerID,
		Accept:            true,
	})
	assert.NoError(t, err)

	// All three messages must be present
	assert.NotNil(t, findMessage(ms.inputs, message.MessageTypeDisputeResolved))
	assert.NotNil(t, findMessage(ms.inputs, message.MessageType("proposal_completed")))
	assert.NotNil(t, findMessage(ms.inputs, message.MessageType("evaluation_request")))

	// Notifications carry the navigable proposal-flow payload
	clientNotif := findNotification(ns.sent, "proposal_completed", clientID)
	providerNotif := findNotification(ns.sent, "proposal_completed", providerID)
	assert.NotNil(t, clientNotif)
	assert.NotNil(t, providerNotif)
	assertProposalNotifPayload(t, clientNotif.Data, proposalID, convID, "Site web Acme")
	assertProposalNotifPayload(t, providerNotif.Data, proposalID, convID, "Site web Acme")
}

// ---------------------------------------------------------------------------
// 3. Scheduler.autoResolve emits the close-out pair + notification
// ---------------------------------------------------------------------------

func TestSchedulerAutoResolve_EmitsReviewCTAAndProposalNotification(t *testing.T) {
	clientID := uuid.New()
	providerID := uuid.New()
	proposalID := uuid.New()
	convID := uuid.New()
	disputeID := uuid.New()

	// Construct a dispute that hits the ghost path: respondent never
	// replied (RespondentFirstReplyAt is nil) and the AutoResolve
	// domain method can fire.
	d := &disputedomain.Dispute{
		ID: disputeID, ProposalID: proposalID,
		ConversationID: convID,
		InitiatorID:    clientID, RespondentID: providerID,
		ClientID: clientID, ProviderID: providerID,
		Status:         disputedomain.StatusOpen,
		ProposalAmount: 100000, RequestedAmount: 100000,
		CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-10 * 24 * time.Hour),
		Version:   1,
	}

	repo := &mockDisputeRepo{
		listPendingFn: func(_ context.Context) ([]*disputedomain.Dispute, error) {
			return []*disputedomain.Dispute{d}, nil
		},
	}
	pr := &mockProposalRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*proposal.Proposal, error) {
			p := fixtureCompletedProposal(proposalID, convID, clientID, providerID)
			p.ActiveDisputeID = &disputeID
			return p, nil
		},
	}
	ms := &mockMessageSender{}
	ns := &mockNotificationSender{}
	pp := &mockPaymentProcessor{}

	sch := NewScheduler(SchedulerDeps{
		Disputes:      repo,
		Proposals:     pr,
		Messages:      ms,
		Notifications: ns,
		Payments:      pp,
	})

	// Pre-cancel so Run exits after the immediate tick.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sch.Run(ctx, 1*time.Hour)

	// The auto-resolve path must emit all three system messages.
	assert.NotNil(t, findMessage(ms.inputs, message.MessageTypeDisputeAutoResolved),
		"dispute_auto_resolved must be emitted on ghost path")
	assert.NotNil(t, findMessage(ms.inputs, message.MessageType("proposal_completed")),
		"REGRESSION GUARD: auto-resolved missions MUST emit proposal_completed system message")
	assert.NotNil(t, findMessage(ms.inputs, message.MessageType("evaluation_request")),
		"REGRESSION GUARD: auto-resolved missions MUST emit evaluation_request so 'leave a review' shows up")

	// Notifications: dispute_auto_resolved (legacy) + proposal_completed
	// (the new review CTA) — one of each per party.
	clientCompleted := findNotification(ns.sent, "proposal_completed", clientID)
	providerCompleted := findNotification(ns.sent, "proposal_completed", providerID)
	assert.NotNil(t, clientCompleted,
		"REGRESSION GUARD: client must receive proposal_completed notification on auto-resolve")
	assert.NotNil(t, providerCompleted,
		"REGRESSION GUARD: provider must receive proposal_completed notification on auto-resolve")
	assertProposalNotifPayload(t, clientCompleted.Data, proposalID, convID, "Site web Acme")
	assertProposalNotifPayload(t, providerCompleted.Data, proposalID, convID, "Site web Acme")
}

// ---------------------------------------------------------------------------
// 4. Helper payload builder is stable
// ---------------------------------------------------------------------------

func TestBuildProposalCompletedNotificationData_ShapeIsStable(t *testing.T) {
	proposalID := uuid.New()
	convID := uuid.New()
	p := fixtureCompletedProposal(proposalID, convID, uuid.New(), uuid.New())

	data := buildProposalCompletedNotificationData(p)

	var got map[string]string
	assert.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, proposalID.String(), got["proposal_id"])
	assert.Equal(t, convID.String(), got["conversation_id"])
	assert.Equal(t, "Site web Acme", got["proposal_title"])
	assert.Len(t, got, 3, "payload must stay minimal: proposal_id, conversation_id, proposal_title")
}
