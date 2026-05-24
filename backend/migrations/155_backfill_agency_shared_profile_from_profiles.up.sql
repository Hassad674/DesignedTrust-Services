-- 155_backfill_agency_shared_profile_from_profiles.up.sql
--
-- Migrates the AGENCY "shared identity" fields — photo, location,
-- languages — from the legacy `profiles` row onto the `organizations`
-- row (the org-shared model introduced by migration 096).
--
-- WHY
-- ===
-- Migration 096 added the shared columns to `organizations` and did a
-- ONE-TIME backfill from `profiles`. For provider_personal orgs the
-- write path then moved to the org row (PUT /api/v1/organization/...),
-- so `organizations` stayed authoritative for freelance/referrer.
--
-- For AGENCY orgs, however, the legacy write path
-- (PUT /api/v1/profile/location|languages, photo via the legacy upload
-- handler) kept writing to `profiles` and NEVER to `organizations`.
-- Result: every agency edit since 096 landed in `profiles` while the
-- `organizations` shared columns went stale or stayed empty. The two
-- diverged silently (096 backfilled once, never maintained).
--
-- The agency read/write refactor (this round) points the agency
-- profile read AND write at the org-shared model so the agency reuses
-- the SAME components as freelance for these 3 sections. Without this
-- backfill, the moment the read flips to `organizations` every agency
-- would appear to LOSE its photo / city / languages — a data-loss
-- regression. This migration carries the authoritative `profiles`
-- values forward onto `organizations` so the flip is lossless.
--
-- AUTHORITATIVE SOURCE
-- ====================
-- `profiles` is the source of truth for agency shared fields: it is
-- the ONLY table the agency write path ever touched. Any value in the
-- `organizations` shared columns for an agency is either (a) empty or
-- (b) a stale leftover from 096's one-time backfill. We therefore copy
-- `profiles -> organizations` per field, but ONLY when the `profiles`
-- value is meaningful (non-empty / non-default) AND differs from the
-- org value. That rule guarantees we:
--   * never clobber a real org value with an empty profiles value, and
--   * do correct a stale org value toward the authoritative profiles
--     value (the one the agency actually edited).
--
-- SCOPE: agency orgs only. provider_personal (freelance/referrer) and
-- enterprise rows are left untouched — their org-shared columns are
-- already authoritative (freelance/referrer) or unused for these three
-- groups (enterprise only shares photo, which its own flow maintains).
--
-- IDEMPOTENT: re-running is a no-op once the org values already match
-- the authoritative profiles values (the IS DISTINCT FROM guards make
-- every UPDATE a self-stabilising fixpoint).
--
-- REVERSIBLE: the up snapshots the pre-backfill org shared columns into
-- a dedicated table so the down can restore them exactly. The snapshot
-- table is created IF NOT EXISTS and the snapshot insert is guarded by
-- ON CONFLICT DO NOTHING so a re-run never overwrites the original
-- snapshot with already-backfilled values.

BEGIN;

-- 1. Snapshot the CURRENT (pre-backfill) shared columns for every
--    agency org so the down migration can restore them byte-for-byte.
--    The snapshot is taken only on first apply (ON CONFLICT DO NOTHING),
--    so a re-run after a partial apply keeps the genuine originals.
CREATE TABLE IF NOT EXISTS _mig155_agency_shared_backup (
    organization_id          UUID PRIMARY KEY,
    photo_url                TEXT,
    city                     TEXT,
    country_code             TEXT,
    latitude                 DOUBLE PRECISION,
    longitude                DOUBLE PRECISION,
    work_mode                TEXT[],
    travel_radius_km         INTEGER,
    languages_professional   TEXT[],
    languages_conversational TEXT[]
);

INSERT INTO _mig155_agency_shared_backup (
    organization_id, photo_url, city, country_code, latitude, longitude,
    work_mode, travel_radius_km, languages_professional, languages_conversational
)
SELECT o.id, o.photo_url, o.city, o.country_code, o.latitude, o.longitude,
       o.work_mode, o.travel_radius_km, o.languages_professional, o.languages_conversational
FROM organizations o
WHERE o.type = 'agency'
ON CONFLICT (organization_id) DO NOTHING;

-- 2. Backfill the agency org shared columns from the authoritative
--    profiles row. Each field copies only when the profiles value is
--    meaningful AND differs from the org value (COALESCE keeps the
--    existing org value otherwise). The location block (city, country,
--    lat, lng, work_mode, travel_radius) is copied as a coherent unit:
--    when profiles carries a non-empty city/country we trust the whole
--    profiles location block, so the coordinates stay consistent with
--    the canonical municipality the agency actually picked.
UPDATE organizations o
SET
    photo_url = CASE
        WHEN p.photo_url <> '' AND p.photo_url IS DISTINCT FROM o.photo_url
            THEN p.photo_url
        ELSE o.photo_url
    END,
    city = CASE
        WHEN p.city <> '' AND p.city IS DISTINCT FROM o.city
            THEN p.city
        ELSE o.city
    END,
    country_code = CASE
        WHEN p.country_code <> '' AND p.country_code IS DISTINCT FROM o.country_code
            THEN p.country_code
        ELSE o.country_code
    END,
    -- Coordinates follow the city: when profiles has a non-empty city
    -- that differs from the org's, adopt the profiles coordinates too
    -- (including NULLing them if profiles never geocoded). Otherwise
    -- leave the org coordinates untouched.
    latitude = CASE
        WHEN p.city <> '' AND p.city IS DISTINCT FROM o.city
            THEN p.latitude
        ELSE o.latitude
    END,
    longitude = CASE
        WHEN p.city <> '' AND p.city IS DISTINCT FROM o.city
            THEN p.longitude
        ELSE o.longitude
    END,
    work_mode = CASE
        WHEN p.work_mode <> '{}' AND p.work_mode IS DISTINCT FROM o.work_mode
            THEN p.work_mode
        ELSE o.work_mode
    END,
    travel_radius_km = CASE
        WHEN p.work_mode <> '{}' AND p.work_mode IS DISTINCT FROM o.work_mode
            THEN p.travel_radius_km
        ELSE o.travel_radius_km
    END,
    languages_professional = CASE
        WHEN p.languages_professional <> '{}'
             AND p.languages_professional IS DISTINCT FROM o.languages_professional
            THEN p.languages_professional
        ELSE o.languages_professional
    END,
    languages_conversational = CASE
        WHEN p.languages_conversational <> '{}'
             AND p.languages_conversational IS DISTINCT FROM o.languages_conversational
            THEN p.languages_conversational
        ELSE o.languages_conversational
    END,
    updated_at = now()
FROM profiles p
WHERE p.organization_id = o.id
  AND o.type = 'agency'
  AND (
        (p.photo_url <> '' AND p.photo_url IS DISTINCT FROM o.photo_url)
     OR (p.city <> '' AND p.city IS DISTINCT FROM o.city)
     OR (p.country_code <> '' AND p.country_code IS DISTINCT FROM o.country_code)
     OR (p.work_mode <> '{}' AND p.work_mode IS DISTINCT FROM o.work_mode)
     OR (p.languages_professional <> '{}'
            AND p.languages_professional IS DISTINCT FROM o.languages_professional)
     OR (p.languages_conversational <> '{}'
            AND p.languages_conversational IS DISTINCT FROM o.languages_conversational)
  );

COMMIT;
