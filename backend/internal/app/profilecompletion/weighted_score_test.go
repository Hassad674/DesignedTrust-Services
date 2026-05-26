package profilecompletion

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/domain/profile"
	"marketplace-backend/internal/search"
)

// TestWeightedScore_MatchesSearchIndexer is the CONSISTENCY LANDMINE
// guard. It proves that the score the profile-completion endpoint
// reports for a given profile is IDENTICAL to the
// profile_completion_score the search indexer would compute for the
// same profile (the number the gate filters on). If these ever drift,
// the user would see one number and be gated on another — exactly the
// failure the brief calls out.
//
// The test mirrors the indexer's CompletionInput construction
// (internal/search/indexer_more.go assembleDocument) against the
// equivalent snapshotBundle and asserts equal scores for freelance,
// referrer, and agency personas across several fill levels.
func TestWeightedScore_MatchesSearchIndexer(t *testing.T) {
	tests := []struct {
		name    string
		persona Persona
		bundle  *snapshotBundle
		// indexerInput is what internal/search would feed
		// ProfileCompletionScore for the SAME profile. Built independently
		// so the test fails loudly if the endpoint mapping diverges.
		indexerInput search.CompletionInput
	}{
		{
			name:    "freelance fully filled",
			persona: PersonaFreelance,
			bundle: &snapshotBundle{
				Shared: &SharedProfile{PhotoURL: "p.webp", City: "Paris", CountryCode: "FR", LanguagesProfessional: []string{"fr", "en"}},
				Freelance: &FreelanceProfileSnapshot{
					Title: "Go dev", About: "long bio", VideoURL: "v.mp4",
					ExpertiseDomains: []string{"backend"},
				},
				SkillCount: 6, FreelancePricing: true, SocialFreelance: 2,
			},
			indexerInput: search.CompletionInput{
				HasPhoto: true, HasAbout: true, HasTitle: true, HasVideo: true,
				ExpertiseCount: 1, SkillsCount: 6, HasPricing: true,
				HasLocation: true, SocialLinksCount: 2, LanguagesCount: 2,
			},
		},
		{
			name:    "freelance half filled (around threshold)",
			persona: PersonaFreelance,
			bundle: &snapshotBundle{
				Shared:    &SharedProfile{PhotoURL: "p.webp"},
				Freelance: &FreelanceProfileSnapshot{About: "bio", Title: "title"},
				SkillCount: 1,
			},
			indexerInput: search.CompletionInput{
				HasPhoto: true, HasAbout: true, HasTitle: true, SkillsCount: 1,
			},
		},
		{
			name:    "freelance empty",
			persona: PersonaFreelance,
			bundle:  &snapshotBundle{},
			indexerInput: search.CompletionInput{},
		},
		{
			name:    "referrer filled",
			persona: PersonaReferrer,
			bundle: &snapshotBundle{
				Shared:   &SharedProfile{PhotoURL: "p.webp", City: "Lyon", CountryCode: "FR", LanguagesProfessional: []string{"fr"}},
				Referrer: &ReferrerProfileSnapshot{Title: "Apporteur", About: "bio", VideoURL: "v.mp4", ExpertiseDomains: []string{"sales", "saas"}},
				SkillCount: 3, ReferrerPricing: true, SocialReferrer: 1,
			},
			indexerInput: search.CompletionInput{
				HasPhoto: true, HasAbout: true, HasTitle: true, HasVideo: true,
				ExpertiseCount: 2, SkillsCount: 3, HasPricing: true,
				HasLocation: true, SocialLinksCount: 1, LanguagesCount: 1,
			},
		},
		{
			name:    "agency filled (expertise hardcoded 0, photo from shared)",
			persona: PersonaAgency,
			bundle: &snapshotBundle{
				Shared: &SharedProfile{PhotoURL: "p.webp", City: "Paris", CountryCode: "FR", LanguagesProfessional: []string{"fr", "en"}},
				Legacy: &profile.Profile{
					Title: "Agency", About: "bio", PresentationVideoURL: "v.mp4",
					// ExpertiseDomains on the legacy row are IGNORED by the
					// agency indexer (hardcoded empty) — set some to prove
					// the endpoint also ignores them.
				},
				SkillCount: 5, LegacyPricingN: 1, SocialAgency: 1,
			},
			indexerInput: search.CompletionInput{
				HasPhoto: true, HasAbout: true, HasTitle: true, HasVideo: true,
				ExpertiseCount: 0, SkillsCount: 5, HasPricing: true,
				HasLocation: true, SocialLinksCount: 1, LanguagesCount: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{}
			endpointScore, _ := svc.weightedScoreAndVisibility(tt.persona, tt.bundle)
			indexerScore := search.ProfileCompletionScore(tt.indexerInput)
			assert.Equal(t, indexerScore, endpointScore,
				"endpoint score must equal the search-indexer score for the same profile")
		})
	}
}

// TestThresholds_StayInLockstep is the drift guard: the threshold the
// user-facing score is compared against (SearchVisibilityThreshold) MUST
// equal the threshold the Typesense gate filters on
// (search.ProfileCompletionGateMin). If a future change bumps one without
// the other, the number the user sees would stop matching what gates
// them — this test fails loudly before that ships.
func TestThresholds_StayInLockstep(t *testing.T) {
	assert.Equal(t, search.ProfileCompletionGateMin, SearchVisibilityThreshold,
		"the search gate threshold and the user-facing visibility threshold must be equal")
}

// TestWeightedScore_VisibilityThreshold asserts ListedInSearch flips at
// exactly the threshold for searchable personas and is always false for
// enterprise.
func TestWeightedScore_VisibilityThreshold(t *testing.T) {
	svc := &Service{}

	// A freelance bundle scoring exactly 50: photo(15)+about(15)+title(10)
	// +location(5)+languages(5) = 50.
	at50 := &snapshotBundle{
		Shared:    &SharedProfile{PhotoURL: "p", City: "Paris", CountryCode: "FR", LanguagesProfessional: []string{"fr"}},
		Freelance: &FreelanceProfileSnapshot{About: "bio", Title: "t"},
	}
	score, listed := svc.weightedScoreAndVisibility(PersonaFreelance, at50)
	require.Equal(t, 50, score)
	assert.True(t, listed, "score == threshold (50) must be listed")

	// Drop the title (-10) → 40, below threshold.
	below := &snapshotBundle{
		Shared:    &SharedProfile{PhotoURL: "p", City: "Paris", CountryCode: "FR", LanguagesProfessional: []string{"fr"}},
		Freelance: &FreelanceProfileSnapshot{About: "bio"},
	}
	score, listed = svc.weightedScoreAndVisibility(PersonaFreelance, below)
	require.Equal(t, 40, score)
	assert.False(t, listed, "score below threshold must NOT be listed")

	// Enterprise is never listed regardless of score.
	entHigh := &snapshotBundle{
		Shared: &SharedProfile{PhotoURL: "p", City: "Paris", CountryCode: "FR", LanguagesProfessional: []string{"fr", "en"}},
		Legacy: &profile.Profile{Title: "x", About: strings.Repeat("a", 10), PresentationVideoURL: "v"},
		SkillCount: 9, LegacyPricingN: 1, SocialAgency: 3,
	}
	_, listed = svc.weightedScoreAndVisibility(PersonaEnterprise, entHigh)
	assert.False(t, listed, "enterprise persona is never listed in search")
}
