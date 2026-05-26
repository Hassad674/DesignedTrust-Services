package profilecompletion

import (
	"strings"

	"marketplace-backend/internal/search"
)

// weighted_score.go reconciles the profile-completion endpoint with the
// search visibility gate. Both must report the SAME number for the same
// profile so the "≥50% to appear in search" message the user reads is
// the number that actually gates them.
//
// Single source of truth: search.ProfileCompletionScore — the exact
// function the Typesense indexer calls to compute the
// profile_completion_score stored on each document (see
// internal/search/indexer_more.go assembleDocument). By feeding it the
// same field signals here, the endpoint's `score` equals the indexed
// score the `profile_completion_score:>=50` filter gates on.
//
// The CompletionInput field → signal mapping below mirrors the indexer's
// mapping 1:1:
//
//	HasPhoto         = photo URL present
//	HasAbout         = about text present
//	HasTitle         = title present
//	HasVideo         = video URL present
//	ExpertiseCount   = number of expertise domains
//	SkillsCount      = number of skills
//	HasPricing       = a pricing row exists
//	HasLocation      = city AND country present
//	SocialLinksCount = number of social links
//	LanguagesCount   = number of professional languages

// weightedScoreAndVisibility returns the 0-100 weighted score and the
// search-visibility boolean for the persona. The boolean is true only
// when the persona is searchable (freelance/agency/referrer) AND the
// score clears the threshold. Enterprise profiles are never listed, so
// they always return listed=false regardless of score.
func (s *Service) weightedScoreAndVisibility(persona Persona, bundle *snapshotBundle) (int, bool) {
	score := search.ProfileCompletionScore(completionInputForPersona(persona, bundle))
	listed := personaIsSearchable(persona) && score >= SearchVisibilityThreshold
	return score, listed
}

// personaIsSearchable reports whether a persona is indexed into the
// public search collection. Only freelance / agency / referrer appear
// in search; enterprise (client) profiles never do.
func personaIsSearchable(p Persona) bool {
	switch p {
	case PersonaFreelance, PersonaAgency, PersonaReferrer:
		return true
	}
	return false
}

// completionInputForPersona assembles the search.CompletionInput from
// the snapshot for the given persona. The freelance and referrer
// personas read their own profile rows; the agency persona reads the
// legacy profile row. Enterprise has no searchable offering so it maps
// onto the legacy row too (its score is informational — listed stays
// false).
func completionInputForPersona(persona Persona, bundle *snapshotBundle) search.CompletionInput {
	switch persona {
	case PersonaFreelance:
		return freelanceCompletionInput(bundle)
	case PersonaReferrer:
		return referrerCompletionInput(bundle)
	default: // agency + enterprise read the legacy profile row
		return agencyCompletionInput(bundle)
	}
}

func freelanceCompletionInput(b *snapshotBundle) search.CompletionInput {
	fp := b.Freelance
	hasFP := fp != nil
	in := search.CompletionInput{
		HasPhoto:         hasPhoto(b.Shared),
		HasLocation:      hasLocation(b.Shared),
		LanguagesCount:   languagesCount(b.Shared),
		SkillsCount:      b.SkillCount,
		HasPricing:       b.FreelancePricing,
		SocialLinksCount: b.SocialFreelance,
	}
	if hasFP {
		in.HasAbout = strings.TrimSpace(fp.About) != ""
		in.HasTitle = strings.TrimSpace(fp.Title) != ""
		in.HasVideo = strings.TrimSpace(fp.VideoURL) != ""
		in.ExpertiseCount = len(fp.ExpertiseDomains)
	}
	return in
}

func referrerCompletionInput(b *snapshotBundle) search.CompletionInput {
	rp := b.Referrer
	has := rp != nil
	in := search.CompletionInput{
		HasPhoto:         hasPhoto(b.Shared),
		HasLocation:      hasLocation(b.Shared),
		LanguagesCount:   languagesCount(b.Shared),
		SkillsCount:      b.SkillCount,
		HasPricing:       b.ReferrerPricing,
		SocialLinksCount: b.SocialReferrer,
	}
	if has {
		in.HasAbout = strings.TrimSpace(rp.About) != ""
		in.HasTitle = strings.TrimSpace(rp.Title) != ""
		in.HasVideo = strings.TrimSpace(rp.VideoURL) != ""
		in.ExpertiseCount = len(rp.ExpertiseDomains)
	}
	return in
}

func agencyCompletionInput(b *snapshotBundle) search.CompletionInput {
	legacy := b.Legacy
	hasLegacy := legacy != nil
	// CONSISTENCY: the agency search indexer reads photo / city / country
	// / languages from the organizations (shared) block — NOT the legacy
	// profiles row (see search_document_repository.loadAgencySignals,
	// which selects o.photo_url / o.city / … ). It also HARDCODES
	// expertise_domains to an empty array for agencies and reads the
	// title/about/presentation_video_url from the legacy row. We mirror
	// that EXACT mapping here so the score equals the indexed value the
	// gate filters on — even though the agency checklist (sections) uses
	// the more lenient OR-with-legacy semantics for its UX percent.
	in := search.CompletionInput{
		HasPhoto:         hasPhoto(b.Shared),
		HasLocation:      hasLocation(b.Shared),
		LanguagesCount:   languagesCount(b.Shared),
		SkillsCount:      b.SkillCount,
		HasPricing:       b.LegacyPricingN > 0,
		SocialLinksCount: b.SocialAgency,
		ExpertiseCount:   0, // indexer hardcodes ARRAY[]::text[] for agencies
	}
	if hasLegacy {
		in.HasAbout = strings.TrimSpace(legacy.About) != ""
		in.HasTitle = strings.TrimSpace(legacy.Title) != ""
		in.HasVideo = strings.TrimSpace(legacy.PresentationVideoURL) != ""
	}
	return in
}

// languagesCount returns the number of professional languages declared
// on the shared block (0 when the block is nil).
func languagesCount(s *SharedProfile) int {
	if s == nil {
		return 0
	}
	return len(s.LanguagesProfessional)
}
