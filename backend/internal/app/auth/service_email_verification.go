package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"marketplace-backend/internal/domain/audit"
	"marketplace-backend/internal/domain/session"
	"marketplace-backend/internal/domain/twofactor"
	"marketplace-backend/internal/domain/user"
)

// VerifyEmail completes the signup email-verification flow. The caller
// posts the 6-digit code they received at registration; on a correct
// code the service:
//
//  1. Verifies the code against the latest pending email_verification
//     challenge (purpose-scoped — a login_2fa code can never satisfy it).
//  2. Flips users.email_verified to true.
//  3. RE-ISSUES a fresh access+refresh pair carrying email_verified=true
//     in the claims, mirroring refresh rotation, so the client is not
//     stuck on the stale "false" claim baked into its current token.
//
// Returns the new tokens + the (now-verified) user. Verify-code failures
// surface the underlying twofactor sentinel so the handler maps them to
// the same user-facing codes the 2FA verify endpoint uses.
//
// Idempotency: if the account is already verified, the call is a no-op
// success that simply re-issues tokens — the user does not need a valid
// code to "re-verify" an already-verified email. This keeps a
// double-submit (slow network, retried tap) from returning a confusing
// error.
func (s *Service) VerifyEmail(ctx context.Context, userID uuid.UUID, code string, fp SessionFingerprint) (*AuthOutput, error) {
	if s.twoFactorGate == nil {
		return nil, fmt.Errorf("auth: email verification gate not configured")
	}
	if userID == uuid.Nil {
		return nil, user.ErrUnauthorized
	}

	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth: load user for verify-email: %w", err)
	}

	// Already verified → idempotent success, skip the code check entirely.
	if !u.EmailVerified {
		if vErr := s.twoFactorGate.VerifyChallengeWithPurpose(ctx, userID, code, twofactor.PurposeEmailVerification); vErr != nil {
			s.logAudit(ctx, audit.NewEntryInput{
				UserID:       &userID,
				Action:       audit.ActionLoginFailure,
				ResourceType: audit.ResourceTypeUser,
				ResourceID:   &userID,
				Metadata:     map[string]any{"reason": "email_verification_failed"},
			})
			return nil, vErr
		}
		if err := s.users.SetEmailVerified(ctx, userID, true); err != nil {
			return nil, fmt.Errorf("auth: set email_verified: %w", err)
		}
		u.EmailVerified = true
		s.logAudit(ctx, audit.NewEntryInput{
			UserID:       &userID,
			Action:       audit.ActionLoginSuccess,
			ResourceType: audit.ResourceTypeUser,
			ResourceID:   &userID,
			Metadata:     map[string]any{"email": u.Email, "email_verified": true},
		})
	}

	// Re-issue the token pair so the new claims carry email_verified=true.
	return s.reissueTokensForVerifiedEmail(ctx, u, fp)
}

// reissueTokensForVerifiedEmail mints a fresh access+refresh pair for a
// user whose email_verified flag was just flipped, and records the new
// session row. Mirrors the tail of Login so the post-verify state is
// indistinguishable from a fresh login (same org context resolution,
// same session bookkeeping).
func (s *Service) reissueTokensForVerifiedEmail(ctx context.Context, u *user.User, fp SessionFingerprint) (*AuthOutput, error) {
	orgCtx, err := s.resolveOrgContext(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	accessToken, err := s.tokens.GenerateAccessToken(buildAccessInput(u, orgCtx))
	if err != nil {
		return nil, fmt.Errorf("auth: generate access token after verify-email: %w", err)
	}
	refreshToken, err := s.tokens.GenerateRefreshToken(u.ID)
	if err != nil {
		return nil, fmt.Errorf("auth: generate refresh token after verify-email: %w", err)
	}
	s.recordSession(ctx, u.ID, refreshToken, "", session.LoginMethodPassword, fp)
	return buildAuthOutput(u, orgCtx, accessToken, refreshToken), nil
}

// ResendVerification issues a fresh email_verification OTP for the
// authenticated user. When the account is ALREADY verified the call is a
// 200 no-op (signalled by the returned alreadyVerified=true) so the
// endpoint is safe to hit redundantly. The rate limiter on the route is
// the abuse guard against inbox-bombing.
//
// Errors from the challenge-issue path (email outage) ARE surfaced so
// the client can show a "couldn't send, retry" message — unlike the
// register auto-send which is fire-and-forget because registration must
// not fail on the side-effect.
func (s *Service) ResendVerification(ctx context.Context, userID uuid.UUID, fp SessionFingerprint) (alreadyVerified bool, err error) {
	if s.twoFactorGate == nil {
		return false, fmt.Errorf("auth: email verification gate not configured")
	}
	if userID == uuid.Nil {
		return false, user.ErrUnauthorized
	}

	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("auth: load user for resend-verification: %w", err)
	}
	if u.EmailVerified {
		return true, nil
	}

	if _, err := s.twoFactorGate.RequestChallenge(ctx, TwoFactorChallengeRequest{
		UserID:        u.ID,
		EmailTo:       u.Email,
		ClientIP:      fp.IPAnonymized,
		UserAgentHash: fp.UserAgentHash,
		Purpose:       twofactor.PurposeEmailVerification,
	}); err != nil {
		slog.Warn("auth: resend email-verification challenge failed",
			"user_id", u.ID, "error", err)
		return false, fmt.Errorf("auth: resend verification: %w", err)
	}
	return false, nil
}
