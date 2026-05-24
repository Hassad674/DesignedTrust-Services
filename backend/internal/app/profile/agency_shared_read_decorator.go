package profileapp

import (
	"context"

	"github.com/google/uuid"

	"marketplace-backend/internal/domain/organization"
	"marketplace-backend/internal/domain/profile"
	"marketplace-backend/internal/port/repository"
)

// agency_shared_read_decorator.go overlays the agency shared-identity
// fields (photo, location, languages) from the organizations row onto
// the legacy profiles read.
//
// Since migration 155 the agency photo / location / languages are
// authoritative on the organizations row (the org-shared model). The
// legacy ProfileRepository still reads those columns from the profiles
// row to keep its own write→read round-trip self-consistent, so this
// decorator does the agency overlay one layer up: it fetches the base
// profile, and — only for AGENCY orgs — replaces the three shared
// field groups with the values read from the organizations row.
//
// Placement (wiring): the decorator sits BETWEEN the Redis public
// profile cache and the underlying service —
//
//	cache(GetProfile) -> AgencySharedProfileReader(GetProfile) -> service(GetProfile)
//
// so the MERGED result is what gets cached, and the org-shared write
// path (which busts the same profile:agency:{org} cache key) keeps the
// cached merge fresh. Both the owner self-read (GET /api/v1/profile)
// and the public read (GET /api/v1/profiles/{orgId}) flow through this
// reader, so the editor and the public page see the same values the
// org-shared editor writes.
//
// NON-agency orgs (enterprise) pass through untouched — their profile
// is returned exactly as the inner reader produced it.

// SharedProfileReader is the narrow read the decorator needs to fetch
// the org-shared block. The postgres OrganizationRepository satisfies
// it via GetSharedProfile (it implements OrganizationSharedProfileWriter,
// which embeds this method). Reusing the existing org-shared read keeps
// the SQL in one place — no duplicate query.
type SharedProfileReader interface {
	GetSharedProfile(ctx context.Context, orgID uuid.UUID) (*repository.OrganizationSharedProfile, error)
}

// AgencySharedProfileReader decorates a PublicProfileReader with the
// agency shared-identity overlay. Implements the same GetProfile
// contract so callers (the cache, the handler) never learn the overlay
// is in play.
type AgencySharedProfileReader struct {
	inner  ProfileReader
	orgs   OrgTypeReader
	shared SharedProfileReader
}

// ProfileReader is the single-method read contract the decorator wraps.
// Matches both the profile Service and the Redis cache decorator.
type ProfileReader interface {
	GetProfile(ctx context.Context, orgID uuid.UUID) (*profile.Profile, error)
}

// NewAgencySharedProfileReader wraps inner with the agency overlay. When
// either collaborator is nil the decorator degrades to a pure
// pass-through (it just delegates to inner) so a misconfigured wiring
// can never drop the base read.
func NewAgencySharedProfileReader(inner ProfileReader, orgs OrgTypeReader, shared SharedProfileReader) *AgencySharedProfileReader {
	return &AgencySharedProfileReader{inner: inner, orgs: orgs, shared: shared}
}

// GetProfile fetches the base profile from the inner reader, then — for
// AGENCY orgs — overlays the three shared field groups from the
// organizations row. Any failure to resolve the org type or the shared
// block is non-fatal: the base profile is returned unchanged so a
// transient org-read blip degrades to the legacy profiles values rather
// than failing the whole read.
func (d *AgencySharedProfileReader) GetProfile(ctx context.Context, orgID uuid.UUID) (*profile.Profile, error) {
	p, err := d.inner.GetProfile(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if d.orgs == nil || d.shared == nil {
		return p, nil
	}
	org, err := d.orgs.FindByID(ctx, orgID)
	if err != nil || org == nil || org.Type != organization.OrgTypeAgency {
		return p, nil
	}
	shared, err := d.shared.GetSharedProfile(ctx, orgID)
	if err != nil || shared == nil {
		return p, nil
	}
	overlaySharedProfile(p, shared)
	return p, nil
}

// overlaySharedProfile copies the three shared field groups (photo,
// location, languages) from the org-shared block onto the profile. The
// slices are copied through nilToEmptyShared so the DTO marshals empty
// arrays as [] rather than null, matching the legacy read's guarantee.
func overlaySharedProfile(p *profile.Profile, shared *repository.OrganizationSharedProfile) {
	p.PhotoURL = shared.PhotoURL
	p.City = shared.City
	p.CountryCode = shared.CountryCode
	p.Latitude = shared.Latitude
	p.Longitude = shared.Longitude
	p.WorkMode = nilToEmptyShared(shared.WorkMode)
	p.TravelRadiusKm = shared.TravelRadiusKm
	p.LanguagesProfessional = nilToEmptyShared(shared.LanguagesProfessional)
	p.LanguagesConversational = nilToEmptyShared(shared.LanguagesConversational)
}

// nilToEmptyShared guarantees a non-nil slice so the response DTO emits
// [] instead of null for an empty array — mirrors the adapter's
// nilToEmpty helper without crossing the package boundary.
func nilToEmptyShared(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
