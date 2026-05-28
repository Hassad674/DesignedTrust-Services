package twofactor

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurpose_IsValid(t *testing.T) {
	tests := []struct {
		name string
		p    Purpose
		want bool
	}{
		{"login_2fa", PurposeLogin2FA, true},
		{"email_verification", PurposeEmailVerification, true},
		{"empty", Purpose(""), false},
		{"unknown", Purpose("password_reset"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.p.IsValid())
		})
	}
}

// TestNew_PurposeHandling covers the constructor's purpose contract:
// empty defaults to login_2fa (DB-default parity), a known purpose is
// preserved verbatim, and an unknown purpose is rejected before it can
// be persisted.
func TestNew_PurposeHandling(t *testing.T) {
	tests := []struct {
		name        string
		purpose     Purpose
		wantErr     error
		wantPurpose Purpose
	}{
		{name: "empty defaults to login_2fa", purpose: "", wantPurpose: PurposeLogin2FA},
		{name: "login_2fa preserved", purpose: PurposeLogin2FA, wantPurpose: PurposeLogin2FA},
		{name: "email_verification preserved", purpose: PurposeEmailVerification, wantPurpose: PurposeEmailVerification},
		{name: "unknown rejected", purpose: Purpose("bogus"), wantErr: ErrInvalidPurpose},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(NewChallengeInput{
				UserID:   uuid.New(),
				CodeHash: "hashed",
				Purpose:  tt.purpose,
			})
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, c)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPurpose, c.Purpose)
		})
	}
}
