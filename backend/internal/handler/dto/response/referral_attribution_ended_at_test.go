package response

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/domain/referral"
)

// TestAttributionResponse_EndedAt ensures the WALLET-UNIFY Run D fix
// projects the domain `Attribution.EndedAt` onto the DTO so the per-
// attribution "Intro terminée" badge persists across reloads on
// web + mobile. The mutation response already returns ended_at; the
// list endpoint must too — that is the bug this fix closes.
func TestAttributionResponse_EndedAt(t *testing.T) {
	referrerID := uuid.New()
	clientID := uuid.New()
	endedAt := time.Date(2026, 5, 11, 14, 30, 0, 0, time.UTC)

	newRow := func(ended *time.Time) attributionWithStats {
		return attributionWithStats{
			Attribution: &referral.Attribution{
				ID:              uuid.New(),
				ReferralID:      uuid.New(),
				ProposalID:      uuid.New(),
				ProviderID:      uuid.New(),
				ClientID:        clientID,
				RatePctSnapshot: 5,
				AttributedAt:    time.Now().UTC(),
				EndedAt:         ended,
			},
			ProposalTitle:        "Mission test",
			ProposalStatus:       "in_progress",
			TotalCommissionCents: 1000,
			MilestonesPaid:       1,
			MilestonesTotal:      3,
		}
	}

	cases := []struct {
		name           string
		ended          *time.Time
		viewerID       uuid.UUID
		expectField    bool
		expectIsClient bool
	}{
		{
			name:        "active attribution → ended_at omitted",
			ended:       nil,
			viewerID:    referrerID,
			expectField: false,
		},
		{
			name:        "ended attribution → ended_at present (RFC3339)",
			ended:       &endedAt,
			viewerID:    referrerID,
			expectField: true,
		},
		{
			name:           "client viewer + ended → ended_at still present, commissions stripped",
			ended:          &endedAt,
			viewerID:       clientID,
			expectField:    true,
			expectIsClient: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := NewAttributionListFromStats(
				[]attributionWithStats{newRow(tc.ended)},
				tc.viewerID,
				clientID,
			)
			require.Len(t, out, 1)

			// Round-trip through JSON to validate the json tag +
			// omitempty + RFC3339 formatting in one shot — same path
			// the real handler takes.
			raw, err := json.Marshal(out[0])
			require.NoError(t, err)

			var decoded map[string]any
			require.NoError(t, json.Unmarshal(raw, &decoded))

			if tc.expectField {
				gotEndedAt, ok := decoded["ended_at"].(string)
				require.True(t, ok, "ended_at must be a string in JSON: %s", string(raw))
				assert.Equal(t, "2026-05-11T14:30:00Z", gotEndedAt,
					"ended_at must be UTC RFC3339-formatted")
			} else {
				_, present := decoded["ended_at"]
				assert.False(t, present, "ended_at must be omitted when nil")
			}

			// Modèle A — client viewers see no commission amounts even
			// when the attribution is ended.
			if tc.expectIsClient {
				_, hasTotal := decoded["total_commission_cents"]
				assert.False(t, hasTotal,
					"client viewer must NOT see total_commission_cents")
				_, hasRate := decoded["rate_pct_snapshot"]
				assert.False(t, hasRate,
					"client viewer must NOT see rate_pct_snapshot")
			} else {
				_, hasTotal := decoded["total_commission_cents"]
				assert.True(t, hasTotal,
					"non-client viewer must see total_commission_cents")
			}
		})
	}
}
