-- 155_backfill_agency_shared_profile_from_profiles.down.sql
--
-- Reverts migration 155 by restoring the agency org shared columns
-- (photo, location, languages) to the exact pre-backfill state captured
-- in the _mig155_agency_shared_backup snapshot table, then drops the
-- snapshot table.
--
-- This is a byte-for-byte restore: the up migration snapshotted the
-- original organizations.{photo_url, city, country_code, latitude,
-- longitude, work_mode, travel_radius_km, languages_professional,
-- languages_conversational} for every agency org BEFORE overwriting
-- them, so the down can put them back precisely.
--
-- Safe to run when the snapshot table does not exist (e.g. the up was
-- never applied): the UPDATE simply matches no rows and the DROP is
-- guarded by IF EXISTS.
--
-- NOTE: this only touches agency orgs (the snapshot only contains
-- agency rows) and only the nine shared columns — every other column,
-- and every non-agency org, is left exactly as-is.

BEGIN;

UPDATE organizations o
SET photo_url                = b.photo_url,
    city                     = b.city,
    country_code             = b.country_code,
    latitude                 = b.latitude,
    longitude                = b.longitude,
    work_mode                = b.work_mode,
    travel_radius_km         = b.travel_radius_km,
    languages_professional   = b.languages_professional,
    languages_conversational = b.languages_conversational,
    updated_at               = now()
FROM _mig155_agency_shared_backup b
WHERE o.id = b.organization_id;

DROP TABLE IF EXISTS _mig155_agency_shared_backup;

COMMIT;
