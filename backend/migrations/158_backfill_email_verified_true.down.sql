-- Reverse 158 is intentionally a NO-OP.
--
-- The up migration is a data backfill, not a schema change. Flipping every
-- user back to email_verified=false on rollback would re-introduce the exact
-- production lockout the up migration exists to prevent, and would also clobber
-- any genuine verifications that happened after the backfill ran. A data
-- backfill that grandfathers existing users is not meaningfully reversible —
-- the safe, correct "undo" is to leave the data as-is and instead disable the
-- gate at the application layer (remove the RequireEmailVerified middleware
-- wiring) if the feature must be rolled back.
--
-- This file exists so `migrate down` past version 158 succeeds cleanly
-- (golang-migrate requires a .down.sql for every .up.sql).

SELECT 1;
