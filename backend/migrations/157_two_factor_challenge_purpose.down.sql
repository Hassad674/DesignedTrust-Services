-- Reverse 157: drop the purpose-aware index, restore the original
-- purpose-blind partial index, then drop the constraint and column.
--
-- Order matters: recreate the old index BEFORE dropping the column so the
-- verify path is never left without a covering index mid-rollback.

CREATE INDEX IF NOT EXISTS idx_2fa_user_pending
    ON two_factor_challenges(user_id, used_at, expires_at)
    WHERE used_at IS NULL;

DROP INDEX IF EXISTS idx_2fa_user_purpose_pending;

ALTER TABLE two_factor_challenges
    DROP CONSTRAINT IF EXISTS two_factor_challenges_purpose_check;

ALTER TABLE two_factor_challenges
    DROP COLUMN IF EXISTS purpose;
