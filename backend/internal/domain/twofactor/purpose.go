package twofactor

// Purpose discriminates what a challenge proves. The email-OTP infra is
// reused by two distinct flows that must never cross-validate: a code minted
// to verify email ownership at signup must NOT satisfy a 2FA login gate, and
// vice-versa. The purpose is persisted on the row (migration 157) and every
// Create / find query is scoped by it.
//
// Stored as a TEXT column with a CHECK constraint at the DB level; the domain
// keeps the same string values so the adapter is a 1:1 mapping.
type Purpose string

const (
	// PurposeLogin2FA is the existing email-2FA login gate — the only
	// purpose that existed before signup verification. It is the DB default
	// so any row missing an explicit purpose keeps the historical meaning.
	PurposeLogin2FA Purpose = "login_2fa"

	// PurposeEmailVerification is the signup email-ownership proof. Issued
	// automatically on Register and re-issuable via /auth/resend-verification.
	PurposeEmailVerification Purpose = "email_verification"
)

// IsValid reports whether p is one of the known purposes. Used by the
// constructor to reject a typo'd or empty purpose before it can be persisted
// (the DB CHECK constraint is the second line of defence).
func (p Purpose) IsValid() bool {
	switch p {
	case PurposeLogin2FA, PurposeEmailVerification:
		return true
	}
	return false
}

// String returns the wire/DB representation.
func (p Purpose) String() string { return string(p) }
