package rules

import "sort"

// new_account_cap.go implements §7.5 of docs/ranking-v1.md — the
// final-score cap on new accounts.
//
// The anti-gaming pipeline §7.5 detects accounts younger than the
// configured age threshold (default 7 days) and flags them via
// Candidate.NewAccountCapped. The rule stated by the spec is :
//
//	"a profile younger than 7 days can at best rank at the median —
//	 not in the top"
//
// Scope of this file
//
// We operate on the cohort actually being ranked (the slice handed to
// BusinessRules.Apply) — that is the persona-scoped candidate list
// the query path retrieved from Typesense. The "persona median" the
// spec mentions is therefore the median Final score of non-capped
// candidates in the same cohort.
//
// Why non-capped only ?
//
// If we included capped accounts in the median, two new accounts
// gaming each other could drag the median down and bypass the rule.
// By computing the median over candidates whose Final the rule will
// NOT touch, we guarantee a stable reference point.
//
// Tie-breaking
//
// When several candidates are capped, the cap is applied per-candidate
// — they keep their relative order via sort.SliceStable later in the
// pipeline. Within a tie at exactly the median, the original order
// from scoreCandidates is preserved (stable sort all the way down).
//
// Edge cases
//
//   - Zero candidates → no-op.
//   - All candidates capped → median = 0 (no eligible reference); the
//     rule then caps everyone at the existing minimum which preserves
//     order without amplifying any candidate.
//   - Capped candidate already below the median → Final stays unchanged
//     (min(Final, median) == Final).

// applyNewAccountCap caps Score.Final for every candidate whose
// NewAccountCapped flag is set, at the median Final score of the
// non-capped cohort. Mutates Candidate.Score.Final in place.
//
// The function also caps Score.Adjusted in lockstep with Final so
// downstream tools (LTR feature logging, breakdown UI) see consistent
// values. Score.Base remains untouched — it represents the pre-penalty
// composite the scorer produced before any rule fired, and tampering
// with it would muddy LTR training data later.
//
// Returns the number of candidates the cap modified, useful for
// telemetry / log correlation.
func applyNewAccountCap(candidates []Candidate) int {
	if len(candidates) == 0 {
		return 0
	}

	// Fast path: if no candidate is flagged, the rule is a no-op and
	// we save the median allocation.
	hasCapped := false
	for i := range candidates {
		if candidates[i].NewAccountCapped {
			hasCapped = true
			break
		}
	}
	if !hasCapped {
		return 0
	}

	median := medianFinalNonCapped(candidates)

	// Cap each flagged candidate's Final at min(Final, median). The
	// adjusted score is brought down proportionally so a non-zero base
	// score with a zero adjusted is never produced (which would be a
	// nonsensical state for the breakdown UI).
	modified := 0
	for i := range candidates {
		if !candidates[i].NewAccountCapped {
			continue
		}
		current := candidates[i].Score.Final
		if current <= median {
			continue
		}
		candidates[i].Score.Final = median
		// Adjusted is in [0, 1]; Final = adjusted × 100 in normal
		// conditions. Track the same ratio so a future LTR scorer
		// reading Adjusted (rather than Final) sees the cap too.
		if current > 0 {
			ratio := median / current
			candidates[i].Score.Adjusted *= ratio
		} else {
			candidates[i].Score.Adjusted = 0
		}
		modified++
	}
	return modified
}

// medianFinalNonCapped returns the median Final score of candidates
// whose NewAccountCapped flag is FALSE. Returns 0 when every candidate
// is flagged — see file header for the rationale.
func medianFinalNonCapped(candidates []Candidate) float64 {
	scores := make([]float64, 0, len(candidates))
	for i := range candidates {
		if candidates[i].NewAccountCapped {
			continue
		}
		scores = append(scores, candidates[i].Score.Final)
	}
	if len(scores) == 0 {
		return 0
	}
	sort.Float64s(scores)
	n := len(scores)
	if n%2 == 1 {
		return scores[n/2]
	}
	return (scores[n/2-1] + scores[n/2]) / 2
}
