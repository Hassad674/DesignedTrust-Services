package rules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// new_account_cap_test.go covers §7.5 of docs/ranking-v1.md — the
// final-score cap enforced by applyNewAccountCap.
//
// The tests exercise the helper directly + via the BusinessRules.Apply
// integration so the wired path is also covered.

func TestApplyNewAccountCap_NoFlaggedCandidates_NoOp(t *testing.T) {
	candidates := []Candidate{
		{DocumentID: "a", Score: Score{Final: 90}},
		{DocumentID: "b", Score: Score{Final: 70}},
		{DocumentID: "c", Score: Score{Final: 50}},
	}
	mod := applyNewAccountCap(candidates)
	assert.Equal(t, 0, mod, "no flagged candidate must return 0 modifications")
	assert.Equal(t, 90.0, candidates[0].Score.Final)
	assert.Equal(t, 70.0, candidates[1].Score.Final)
	assert.Equal(t, 50.0, candidates[2].Score.Final)
}

func TestApplyNewAccountCap_CapsAboveMedian(t *testing.T) {
	// 5 mature candidates with Final scores 90, 80, 70, 60, 50.
	// Median = 70.
	// 1 capped candidate with Final = 95 — should be capped at 70.
	candidates := []Candidate{
		{DocumentID: "fresh", Score: Score{Final: 95, Adjusted: 0.95}, NewAccountCapped: true},
		{DocumentID: "m1", Score: Score{Final: 90}},
		{DocumentID: "m2", Score: Score{Final: 80}},
		{DocumentID: "m3", Score: Score{Final: 70}},
		{DocumentID: "m4", Score: Score{Final: 60}},
		{DocumentID: "m5", Score: Score{Final: 50}},
	}
	mod := applyNewAccountCap(candidates)
	assert.Equal(t, 1, mod)
	assert.InDelta(t, 70.0, candidates[0].Score.Final, 1e-9,
		"capped candidate Final must equal cohort median")
	// Adjusted scaled in lockstep: original 0.95 × (70/95) ≈ 0.7.
	assert.InDelta(t, 0.95*70.0/95.0, candidates[0].Score.Adjusted, 1e-9)
	// Mature candidates untouched.
	assert.Equal(t, 90.0, candidates[1].Score.Final)
}

func TestApplyNewAccountCap_BelowMedian_Untouched(t *testing.T) {
	// Capped candidate is already below median — no-op.
	candidates := []Candidate{
		{DocumentID: "fresh", Score: Score{Final: 30, Adjusted: 0.3}, NewAccountCapped: true},
		{DocumentID: "m1", Score: Score{Final: 90}},
		{DocumentID: "m2", Score: Score{Final: 70}},
		{DocumentID: "m3", Score: Score{Final: 50}},
	}
	mod := applyNewAccountCap(candidates)
	assert.Equal(t, 0, mod, "fresh below median must stay untouched")
	assert.Equal(t, 30.0, candidates[0].Score.Final)
	assert.Equal(t, 0.3, candidates[0].Score.Adjusted)
}

func TestApplyNewAccountCap_AllCapped_ReferenceMedianZero(t *testing.T) {
	// Edge case: every candidate is flagged. Median falls to 0 and
	// every candidate's Final is capped at min(Final, 0) = 0.
	candidates := []Candidate{
		{DocumentID: "f1", Score: Score{Final: 80, Adjusted: 0.8}, NewAccountCapped: true},
		{DocumentID: "f2", Score: Score{Final: 50, Adjusted: 0.5}, NewAccountCapped: true},
	}
	mod := applyNewAccountCap(candidates)
	assert.Equal(t, 2, mod)
	assert.Equal(t, 0.0, candidates[0].Score.Final)
	assert.Equal(t, 0.0, candidates[1].Score.Final)
}

func TestApplyNewAccountCap_MultipleFlagged_SameMedianApplied(t *testing.T) {
	// Two flagged candidates above median — both capped at the
	// non-capped median.
	candidates := []Candidate{
		{DocumentID: "f1", Score: Score{Final: 95, Adjusted: 0.95}, NewAccountCapped: true},
		{DocumentID: "f2", Score: Score{Final: 88, Adjusted: 0.88}, NewAccountCapped: true},
		{DocumentID: "m1", Score: Score{Final: 90}},
		{DocumentID: "m2", Score: Score{Final: 70}},
		{DocumentID: "m3", Score: Score{Final: 50}},
	}
	mod := applyNewAccountCap(candidates)
	assert.Equal(t, 2, mod)
	// Non-capped median = median(90, 70, 50) = 70.
	assert.InDelta(t, 70.0, candidates[0].Score.Final, 1e-9)
	assert.InDelta(t, 70.0, candidates[1].Score.Final, 1e-9)
}

func TestApplyNewAccountCap_EvenCount_MeanOfTwoMiddle(t *testing.T) {
	// 4 non-capped candidates → median = (top1+top2)/2 of mid-pair.
	candidates := []Candidate{
		{DocumentID: "fresh", Score: Score{Final: 95, Adjusted: 0.95}, NewAccountCapped: true},
		{DocumentID: "m1", Score: Score{Final: 90}},
		{DocumentID: "m2", Score: Score{Final: 80}},
		{DocumentID: "m3", Score: Score{Final: 60}},
		{DocumentID: "m4", Score: Score{Final: 50}},
	}
	mod := applyNewAccountCap(candidates)
	assert.Equal(t, 1, mod)
	assert.InDelta(t, 70.0, candidates[0].Score.Final, 1e-9, "median = (60+80)/2")
}

func TestApplyNewAccountCap_EmptyInput(t *testing.T) {
	mod := applyNewAccountCap(nil)
	assert.Equal(t, 0, mod)
}

// TestBusinessRules_Apply_NewAccountCapIntegration proves the rule
// fires inside BusinessRules.Apply, before tier sort + randomise.
func TestBusinessRules_Apply_NewAccountCapIntegration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RandSeed = 42 // deterministic noise
	br := NewBusinessRules(cfg)

	candidates := []Candidate{
		// Fresh attacker at the very top before the rule.
		{
			DocumentID:         "fresh",
			Persona:            PersonaFreelance,
			Score:              Score{Final: 99, Adjusted: 0.99},
			AvailabilityStatus: "available_now",
			IsVerified:         true,
			NewAccountCapped:   true,
			AccountAgeDays:     3,
		},
		{
			DocumentID:         "mature1",
			Persona:            PersonaFreelance,
			Score:              Score{Final: 80, Adjusted: 0.80},
			AvailabilityStatus: "available_now",
			IsVerified:         true,
			AccountAgeDays:     400,
		},
		{
			DocumentID:         "mature2",
			Persona:            PersonaFreelance,
			Score:              Score{Final: 60, Adjusted: 0.60},
			AvailabilityStatus: "available_now",
			IsVerified:         true,
			AccountAgeDays:     400,
		},
		{
			DocumentID:         "mature3",
			Persona:            PersonaFreelance,
			Score:              Score{Final: 40, Adjusted: 0.40},
			AvailabilityStatus: "available_now",
			IsVerified:         true,
			AccountAgeDays:     400,
		},
	}
	out := br.Apply(context.Background(), candidates, PersonaFreelance)
	require.Equal(t, 4, len(out))

	// Locate the fresh candidate post-apply and verify its Final is
	// no longer 99 — the cap (median = 60) capped it down.
	var freshFinal float64
	for _, c := range out {
		if c.DocumentID == "fresh" {
			freshFinal = c.Score.Final
			break
		}
	}
	// Allow a small noise tolerance: at score 60-ish + tail multiplier
	// the σ is ~0.5 so we generously bound the post-noise window.
	assert.LessOrEqual(t, freshFinal, 65.0,
		"fresh candidate Final after cap+noise must be ≤ ~median+noise (got %.2f)", freshFinal)
}
