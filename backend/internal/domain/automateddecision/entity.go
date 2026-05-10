// Package automateddecision implements RGPD art. 22 disclosure: the
// data subject's right to obtain human review of decisions taken
// solely on the basis of automated processing.
//
// The marketplace runs three such decisions:
//   - AI moderation (text via OpenAI omni-moderation, media via AWS
//     Rekognition) — auto-rejects content when confidence > threshold;
//   - search ranking (custom 5-stage pipeline) — determines visibility;
//   - Stripe payment risk scoring — can block payment intents.
//
// This package owns the Appeal entity persisted by the repository
// adapter and exposed by POST /api/v1/me/automated-decision-appeals.
package automateddecision

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DecisionType identifies which automated surface the appeal targets.
// Mirrors the DB CHECK constraint in migration 144.
type DecisionType string

const (
	DecisionMod     DecisionType = "moderation"
	DecisionRanking DecisionType = "ranking"
	DecisionPayment DecisionType = "payment"
)

// IsValid reports whether the value matches one of the three pinned
// surfaces. Anything else is rejected at the domain boundary so a
// typo never reaches the database.
func (d DecisionType) IsValid() bool {
	switch d {
	case DecisionMod, DecisionRanking, DecisionPayment:
		return true
	default:
		return false
	}
}

// Status tracks the lifecycle of the appeal. The DB CHECK constraint
// in migration 144 enforces the same enum.
type Status string

const (
	StatusPending     Status = "pending"
	StatusReviewing   Status = "reviewing"
	StatusUpheld      Status = "upheld"
	StatusOverturned  Status = "overturned"
)

// Appeal is one row of automated_decision_appeals. The entity is
// constructed by New, persisted by the repository adapter, and
// surfaced read-only to the user that filed it.
type Appeal struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	DecisionType DecisionType
	ReferenceID  string
	Reason       string
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// MaxReasonLength caps the reason payload at 5_000 bytes. Anything
// longer signals a copy-paste of the entire content rather than a
// short justification — and protects the table from bloat.
const MaxReasonLength = 5000

// Sentinel validation errors. Per backend/CLAUDE.md the domain layer
// never wraps these — the service translates them into HTTP status
// codes via errors.Is.
var (
	ErrInvalidDecisionType  = errors.New("automated_decision: invalid decision_type")
	ErrReferenceIDRequired  = errors.New("automated_decision: reference_id must be set")
	ErrReasonRequired       = errors.New("automated_decision: reason must be set")
	ErrReasonTooLong        = errors.New("automated_decision: reason exceeds maximum length")
	ErrUserIDRequired       = errors.New("automated_decision: user_id must be set")
)

// NewInput groups the constructor parameters so the service layer can
// hand a single struct to New without juggling positional arguments.
type NewInput struct {
	UserID       uuid.UUID
	DecisionType DecisionType
	ReferenceID  string
	Reason       string
}

// New validates the input and returns an Appeal stamped with a fresh
// UUID, default Pending status, and the current UTC wall-clock time
// for both timestamps. The repository MUST NOT overwrite ID or
// CreatedAt — the domain owns identity + timestamping.
func New(in NewInput) (*Appeal, error) {
	if in.UserID == uuid.Nil {
		return nil, ErrUserIDRequired
	}
	if !in.DecisionType.IsValid() {
		return nil, ErrInvalidDecisionType
	}
	referenceID := strings.TrimSpace(in.ReferenceID)
	if referenceID == "" {
		return nil, ErrReferenceIDRequired
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		return nil, ErrReasonRequired
	}
	if len(reason) > MaxReasonLength {
		return nil, ErrReasonTooLong
	}
	now := time.Now().UTC()
	return &Appeal{
		ID:           uuid.New(),
		UserID:       in.UserID,
		DecisionType: in.DecisionType,
		ReferenceID:  referenceID,
		Reason:       reason,
		Status:       StatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
