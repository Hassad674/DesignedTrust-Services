package crypto

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/port/service"
)

// TestJWTService_EmailVerified_RoundTrip asserts the `ev` claim survives
// a generate→validate round-trip for both true and false.
func TestJWTService_EmailVerified_RoundTrip(t *testing.T) {
	svc := newTestJWTService()
	tests := []struct {
		name string
		ev   bool
	}{
		{"verified", true},
		{"unverified", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := svc.GenerateAccessToken(service.AccessTokenInput{
				UserID:        uuid.New(),
				Role:          "provider",
				EmailVerified: tt.ev,
			})
			require.NoError(t, err)

			claims, err := svc.ValidateAccessToken(token)
			require.NoError(t, err)
			assert.Equal(t, tt.ev, claims.EmailVerified)
		})
	}
}

// TestJWTService_EmailVerified_LegacyTokenDefaultsVerified asserts a
// token forged WITHOUT the `ev` claim (a legacy access token minted
// before signup-OTP shipped) decodes to verified=true so the gate does
// not lock out in-flight sessions during the deploy.
func TestJWTService_EmailVerified_LegacyTokenDefaultsVerified(t *testing.T) {
	svc := newTestJWTService()
	// Forge an access token with no `ev` field at all.
	token := signWithClaims(t, jwtClaims{
		"user_id": uuid.New().String(),
		"role":    "agency",
		"type":    "access",
		"exp":     jwt.NewNumericDate(time.Now().Add(time.Hour)).Unix(),
		"iat":     jwt.NewNumericDate(time.Now()).Unix(),
	})

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.True(t, claims.EmailVerified,
		"a legacy token with no ev claim must decode to verified=true")
}

// TestJWTService_EmailVerified_ExplicitFalseHonored asserts that a token
// carrying ev=false is decoded as unverified (the gate must be able to
// reject a genuinely-unverified fresh account).
func TestJWTService_EmailVerified_ExplicitFalseHonored(t *testing.T) {
	svc := newTestJWTService()
	token := signWithClaims(t, jwtClaims{
		"user_id": uuid.New().String(),
		"role":    "provider",
		"type":    "access",
		"ev":      false,
		"exp":     jwt.NewNumericDate(time.Now().Add(time.Hour)).Unix(),
		"iat":     jwt.NewNumericDate(time.Now()).Unix(),
	})

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.False(t, claims.EmailVerified)
}
