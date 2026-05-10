package repository

import (
	"context"

	"marketplace-backend/internal/domain/automateddecision"
)

// AutomatedDecisionAppealRepository persists Appeal rows. Create-only
// from the user-facing surface — admin handlers (out of scope for B.5)
// will own the future status mutations.
type AutomatedDecisionAppealRepository interface {
	Create(ctx context.Context, appeal *automateddecision.Appeal) error
}
