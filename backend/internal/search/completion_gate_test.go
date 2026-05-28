package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasePersonaFilter_GatedByDefault(t *testing.T) {
	tests := []struct {
		name              string
		persona           Persona
		includeIncomplete bool
		want              string
	}{
		{
			name:    "freelance public is gated",
			persona: PersonaFreelance,
			want:    "persona:freelance && is_published:true && profile_completion_score:>=50",
		},
		{
			name:    "agency public is gated",
			persona: PersonaAgency,
			want:    "persona:agency && is_published:true && profile_completion_score:>=50",
		},
		{
			name:    "referrer public is gated",
			persona: PersonaReferrer,
			want:    "persona:referrer && is_published:true && profile_completion_score:>=50",
		},
		{
			name:              "include_incomplete drops the gate",
			persona:           PersonaFreelance,
			includeIncomplete: true,
			want:              "persona:freelance && is_published:true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, BasePersonaFilter(tt.persona, tt.includeIncomplete))
		})
	}
}

// TestProfileCompletionGate_ThresholdConstant pins the gate threshold so
// a change is a deliberate, reviewed diff. It MUST equal the
// profilecompletion.SearchVisibilityThreshold (50) the user-facing
// number is computed against.
func TestProfileCompletionGate_ThresholdConstant(t *testing.T) {
	assert.Equal(t, 50, ProfileCompletionGateMin)
	assert.Equal(t, "profile_completion_score:>=50", ProfileCompletionGateClause)
}
