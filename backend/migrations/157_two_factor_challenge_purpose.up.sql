-- Email OTP at signup — purpose-scope the 2FA challenge ledger.
--
-- Before this migration every two_factor_challenges row was implicitly a
-- login-2FA challenge. Signup email-verification reuses the exact same
-- infra (6-digit bcrypt-hashed code, 10-min TTL, 5 attempts), so we add a
-- `purpose` discriminator to keep the two flows strictly isolated: a code
-- minted for email verification must NEVER satisfy a 2FA login challenge
-- and vice-versa.
--
-- Values (enforced at the application + a CHECK constraint here):
--   'login_2fa'          — the existing email-2FA login gate (default).
--   'email_verification' — the signup email-ownership proof.
--
-- All pre-existing rows are backfilled to 'login_2fa' because that is the
-- only purpose challenges could have had before this migration. The DEFAULT
-- also lands on 'login_2fa' so any code path that forgets to set the column
-- (defence-in-depth) keeps the historical semantics rather than silently
-- behaving like an email-verification challenge.
--
-- The dominant verify query is "find the latest pending challenge for this
-- user AND this purpose". The existing partial index idx_2fa_user_pending
-- covers (user_id, used_at, expires_at) WHERE used_at IS NULL; we replace it
-- with a purpose-aware partial index so the per-purpose lookup stays a single
-- index probe instead of scanning every pending row for the user.
--
-- Down migration drops the purpose-aware index, restores the original
-- index, and drops the column + constraint.

ALTER TABLE two_factor_challenges
    ADD COLUMN IF NOT EXISTS purpose TEXT NOT NULL DEFAULT 'login_2fa';

-- Backfill is redundant given the DEFAULT (every existing row already
-- materialised 'login_2fa' when the column was added) but kept explicit so
-- the intent is auditable and the migration is safe to re-run on a partially
-- applied state.
UPDATE two_factor_challenges SET purpose = 'login_2fa' WHERE purpose IS NULL;

-- Constrain the column to the known set. Guarded so a retry on a
-- partially-applied state does not error on the duplicate constraint.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'two_factor_challenges_purpose_check'
    ) THEN
        ALTER TABLE two_factor_challenges
            ADD CONSTRAINT two_factor_challenges_purpose_check
            CHECK (purpose IN ('login_2fa', 'email_verification'));
    END IF;
END$$;

-- Purpose-aware partial index: supports "find latest pending challenge for
-- this user AND purpose" without scanning the other purpose's pending rows.
-- Replaces the purpose-blind idx_2fa_user_pending. CONCURRENTLY is omitted
-- because golang-migrate wraps each migration in a transaction; the table is
-- tiny (short-lived OTP rows) so the brief lock is negligible.
CREATE INDEX IF NOT EXISTS idx_2fa_user_purpose_pending
    ON two_factor_challenges(user_id, purpose, used_at, expires_at)
    WHERE used_at IS NULL;

DROP INDEX IF EXISTS idx_2fa_user_pending;
