package automateddecision_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"marketplace-backend/internal/domain/automateddecision"
)

func TestDecisionType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		dt   automateddecision.DecisionType
		want bool
	}{
		{name: "moderation", dt: automateddecision.DecisionMod, want: true},
		{name: "ranking", dt: automateddecision.DecisionRanking, want: true},
		{name: "payment", dt: automateddecision.DecisionPayment, want: true},
		{name: "empty", dt: "", want: false},
		{name: "unknown", dt: "search", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.dt.IsValid())
		})
	}
}

func TestNew_Validation(t *testing.T) {
	uid := uuid.New()
	tests := []struct {
		name    string
		in      automateddecision.NewInput
		wantErr error
	}{
		{
			name: "valid moderation appeal",
			in: automateddecision.NewInput{
				UserID:       uid,
				DecisionType: automateddecision.DecisionMod,
				ReferenceID:  "moderation-result-id",
				Reason:       "I disagree with the rejection — content is on-topic and respectful.",
			},
		},
		{
			name: "valid ranking appeal trims whitespace",
			in: automateddecision.NewInput{
				UserID:       uid,
				DecisionType: automateddecision.DecisionRanking,
				ReferenceID:  "  trace-abc  ",
				Reason:       "  Profile is invisible despite high rating.  ",
			},
		},
		{
			name: "missing user_id",
			in: automateddecision.NewInput{
				DecisionType: automateddecision.DecisionMod,
				ReferenceID:  "ref",
				Reason:       "reason",
			},
			wantErr: automateddecision.ErrUserIDRequired,
		},
		{
			name: "invalid decision type",
			in: automateddecision.NewInput{
				UserID:       uid,
				DecisionType: "search",
				ReferenceID:  "ref",
				Reason:       "reason",
			},
			wantErr: automateddecision.ErrInvalidDecisionType,
		},
		{
			name: "empty reference id",
			in: automateddecision.NewInput{
				UserID:       uid,
				DecisionType: automateddecision.DecisionPayment,
				ReferenceID:  "   ",
				Reason:       "reason",
			},
			wantErr: automateddecision.ErrReferenceIDRequired,
		},
		{
			name: "empty reason",
			in: automateddecision.NewInput{
				UserID:       uid,
				DecisionType: automateddecision.DecisionPayment,
				ReferenceID:  "pi_123",
				Reason:       "",
			},
			wantErr: automateddecision.ErrReasonRequired,
		},
		{
			name: "reason too long",
			in: automateddecision.NewInput{
				UserID:       uid,
				DecisionType: automateddecision.DecisionMod,
				ReferenceID:  "ref",
				Reason:       strings.Repeat("a", automateddecision.MaxReasonLength+1),
			},
			wantErr: automateddecision.ErrReasonTooLong,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appeal, err := automateddecision.New(tt.in)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, appeal)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, appeal)
			assert.NotEqual(t, uuid.Nil, appeal.ID)
			assert.Equal(t, tt.in.UserID, appeal.UserID)
			assert.Equal(t, tt.in.DecisionType, appeal.DecisionType)
			assert.Equal(t, automateddecision.StatusPending, appeal.Status)
			assert.False(t, appeal.CreatedAt.IsZero())
			assert.False(t, appeal.UpdatedAt.IsZero())
			assert.Equal(t, strings.TrimSpace(tt.in.ReferenceID), appeal.ReferenceID)
			assert.Equal(t, strings.TrimSpace(tt.in.Reason), appeal.Reason)
		})
	}
}
