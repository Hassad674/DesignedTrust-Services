package search

import "fmt"

// completion_gate.go centralises the profile-completion visibility gate
// applied to the PUBLIC search + listing surface. A profile must reach
// ProfileCompletionGateMin to appear; below that it is hidden by a
// Typesense filter_by clause appended at query time (no reindex — every
// document already carries profile_completion_score).
//
// Both query paths converge here so the gate is identical:
//   - the server-side proxy (PersonaScopedClient.composeFilter), and
//   - the frontend scoped key (handler.ScopedKey embeds the same base).

// ProfileCompletionGateMin is the minimum profile_completion_score a
// freelance / agency / referrer profile must have to be returned by
// public search. MUST stay equal to
// profilecompletion.SearchVisibilityThreshold so the number the user is
// shown ("≥50% to be visible") matches what actually gates them. The
// two constants live in different packages (no shared import) but a
// drift test could pin them; the value is a stable product constant.
const ProfileCompletionGateMin = 50

// ProfileCompletionGateClause is the Typesense filter_by fragment that
// hides incomplete profiles. Built once as a package var so callers
// cannot typo the field name or operator.
var ProfileCompletionGateClause = fmt.Sprintf("profile_completion_score:>=%d", ProfileCompletionGateMin)

// BasePersonaFilter returns the mandatory base filter for a persona:
//
//	persona:<persona> && is_published:true [&& profile_completion_score:>=50]
//
// The completion gate is appended UNLESS includeIncomplete is true
// (admin / internal callers). Public callers always pass false so the
// gate is on. Defense-in-depth: this is the single place both the
// server-side scoped client and the frontend scoped-key handler build
// their base clause, so the gate cannot be applied on one path and
// forgotten on the other.
func BasePersonaFilter(persona Persona, includeIncomplete bool) string {
	base := fmt.Sprintf("persona:%s && is_published:true", persona)
	if includeIncomplete {
		return base
	}
	return base + " && " + ProfileCompletionGateClause
}
