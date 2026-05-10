package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"marketplace-backend/internal/domain/automateddecision"
)

// AutomatedDecisionAppealRepository is the PostgreSQL adapter for the
// automated_decision_appeals table created in migration 144. The
// repository is currently insert-only; admin status mutations will
// land in a follow-up dispatch.
type AutomatedDecisionAppealRepository struct {
	db *sql.DB
}

// NewAutomatedDecisionAppealRepository wires the adapter to a sql.DB
// handle. The caller (cmd/api wire layer) owns the lifecycle of the
// underlying pool.
func NewAutomatedDecisionAppealRepository(db *sql.DB) *AutomatedDecisionAppealRepository {
	return &AutomatedDecisionAppealRepository{db: db}
}

// Create inserts a new automated_decision_appeals row. Uses the
// entity's ID and CreatedAt verbatim — the domain layer owns identity
// + timestamps so the persisted row matches what the service handed
// back to the handler.
func (r *AutomatedDecisionAppealRepository) Create(
	ctx context.Context,
	appeal *automateddecision.Appeal,
) error {
	if appeal == nil {
		return fmt.Errorf("automated_decision_appeals: nil appeal")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const stmt = `
		INSERT INTO automated_decision_appeals (
			id, user_id, decision_type, reference_id,
			reason, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if _, err := r.db.ExecContext(
		ctx,
		stmt,
		appeal.ID,
		appeal.UserID,
		string(appeal.DecisionType),
		appeal.ReferenceID,
		appeal.Reason,
		string(appeal.Status),
		appeal.CreatedAt,
		appeal.UpdatedAt,
	); err != nil {
		return fmt.Errorf("automated_decision_appeals: insert: %w", err)
	}
	return nil
}
