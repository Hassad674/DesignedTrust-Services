package twofactor

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/domain/twofactor"
)

// TestService_RequestChallenge_DefaultsToLogin2FA asserts that an
// unspecified purpose keeps the historical login-2FA semantics so the
// existing 2FA flow stays byte-identical. The persisted row's purpose
// must be login_2fa and the email subject the unchanged 2FA copy.
func TestService_RequestChallenge_DefaultsToLogin2FA(t *testing.T) {
	repo := &mockChallengeRepo{}
	email := &mockEmail{}
	svc := NewService(ServiceDeps{Challenges: repo, Hasher: &mockHasher{}, Email: email})

	c, err := svc.RequestChallenge(context.Background(), RequestChallengeInput{
		UserID:  uuid.New(),
		EmailTo: "user@example.com",
		// Purpose left empty on purpose.
	})
	require.NoError(t, err)
	assert.Equal(t, twofactor.PurposeLogin2FA, c.Purpose)
	assert.Equal(t, twofactor.PurposeLogin2FA, repo.lastCreatePurpose)
	require.Len(t, email.subjects, 1)
	assert.Equal(t, "Code de vérification — Marketplace Service", email.subjects[0])
}

// TestService_RequestChallenge_EmailVerificationPurpose asserts that an
// email_verification request persists the right purpose and sends the
// signup-confirmation copy (distinct subject from the 2FA email).
func TestService_RequestChallenge_EmailVerificationPurpose(t *testing.T) {
	repo := &mockChallengeRepo{}
	email := &mockEmail{}
	svc := NewService(ServiceDeps{Challenges: repo, Hasher: &mockHasher{}, Email: email})

	c, err := svc.RequestChallenge(context.Background(), RequestChallengeInput{
		UserID:  uuid.New(),
		EmailTo: "newuser@example.com",
		Purpose: twofactor.PurposeEmailVerification,
	})
	require.NoError(t, err)
	assert.Equal(t, twofactor.PurposeEmailVerification, c.Purpose)
	assert.Equal(t, twofactor.PurposeEmailVerification, repo.lastCreatePurpose)
	require.Len(t, email.subjects, 1)
	assert.Equal(t, "Confirme ton adresse email — Marketplace Service", email.subjects[0])
	assert.Regexp(t, `<strong[^>]*>\d{6}</strong>`, email.sentBody[0])
}

// TestService_VerifyChallenge_ScopesByPurpose asserts the verify path
// forwards the requested purpose to the repository so a code minted for
// one flow can never resolve a challenge of the other flow. We assert
// the purpose the service handed the repo for both the default and the
// explicit email_verification case.
func TestService_VerifyChallenge_ScopesByPurpose(t *testing.T) {
	tests := []struct {
		name        string
		inPurpose   twofactor.Purpose
		wantForward twofactor.Purpose
	}{
		{name: "empty defaults to login_2fa", inPurpose: "", wantForward: twofactor.PurposeLogin2FA},
		{name: "login_2fa explicit", inPurpose: twofactor.PurposeLogin2FA, wantForward: twofactor.PurposeLogin2FA},
		{name: "email_verification", inPurpose: twofactor.PurposeEmailVerification, wantForward: twofactor.PurposeEmailVerification},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid := uuid.New()
			plain := "654321"
			repo := &mockChallengeRepo{}
			repo.findFn = func(ctx context.Context, _ uuid.UUID) (*twofactor.Challenge, error) {
				return twofactor.New(twofactor.NewChallengeInput{
					UserID:   uid,
					CodeHash: "h:" + plain,
					Purpose:  tt.wantForward,
				})
			}
			svc := NewService(ServiceDeps{Challenges: repo, Hasher: &mockHasher{}, Email: &mockEmail{}})

			_, err := svc.VerifyChallenge(context.Background(), VerifyChallengeInput{
				UserID:  uid,
				Code:    plain,
				Purpose: tt.inPurpose,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.wantForward, repo.lastFindPurpose,
				"verify must scope the lookup to the requested purpose")
		})
	}
}
