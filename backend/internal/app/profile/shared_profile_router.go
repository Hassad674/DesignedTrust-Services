package profileapp

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"marketplace-backend/internal/domain/organization"
	"marketplace-backend/internal/port/repository"
)

// shared_profile_router.go grafts the agency shared-identity write
// redirection onto the legacy profile Service. Since migration 155 the
// agency photo / location / languages live on the organizations row
// (the org-shared model). To keep EVERY caller of the legacy
// PUT /api/v1/profile/location|languages endpoints (notably the mobile
// profile_tier1 feature, which has not been migrated to the
// /organization/* endpoints yet) consistent with the new read path,
// the service redirects those two writes to the org-shared writer when
// — and only when — the org is an agency. Non-agency orgs (enterprise)
// keep writing to the legacy profiles columns, byte-for-byte unchanged.
//
// The redirection is OPT-IN: it activates only when both collaborators
// are wired via WithAgencySharedRouter. Existing unit tests and any
// wiring that predates the agency refactor keep the pure-legacy
// behaviour, so this is a strictly additive change.

// OrgTypeReader is the narrow read contract the router needs to learn
// an org's type. The postgres OrganizationRepository satisfies it
// directly. Defined locally (not imported as the full
// repository.OrganizationRepository) so the profile service depends on
// the smallest possible surface — interface segregation. A nil reader
// disables the redirection (legacy path).
type OrgTypeReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error)
}

// WithAgencySharedRouter wires the optional agency shared-identity
// write redirection. When both the org-type reader and the org-shared
// writer are non-nil, UpdateLocation / UpdateLanguages route agency
// writes to the organizations row instead of the legacy profiles
// columns. Passing a nil for either argument leaves the legacy
// behaviour intact. Returns the same service for fluent wiring.
func (s *Service) WithAgencySharedRouter(orgs OrgTypeReader, writer repository.OrganizationSharedProfileWriter) *Service {
	if orgs != nil && writer != nil {
		s.orgTypeReader = orgs
		s.sharedWriter = writer
	}
	return s
}

// isAgencyOrg reports whether the redirection is active for orgID:
// the router must be wired AND the org must be of type agency. Any
// lookup error is treated as "not agency" so a transient org-read blip
// degrades to the legacy write path rather than failing the save —
// the caller still gets a successful, consistent write (just to the
// legacy columns, which the read path falls back to for that request).
func (s *Service) isAgencyOrg(ctx context.Context, orgID uuid.UUID) bool {
	if s.orgTypeReader == nil || s.sharedWriter == nil {
		return false
	}
	org, err := s.orgTypeReader.FindByID(ctx, orgID)
	if err != nil || org == nil {
		return false
	}
	return org.Type == organization.OrgTypeAgency
}

// writeAgencyLocation persists the location block onto the
// organizations row via the org-shared writer, then fires the same
// reindex + cache-invalidation hooks the legacy path uses so the
// agency persona stays fresh in Typesense and the Redis profile cache
// is busted. Mirrors the post-write side effects of the legacy
// profiles UpdateLocation exactly.
func (s *Service) writeAgencyLocation(ctx context.Context, orgID uuid.UUID, input repository.LocationInput) error {
	if err := s.sharedWriter.UpdateSharedLocation(ctx, orgID, repository.SharedProfileLocationInput{
		City:           input.City,
		CountryCode:    input.CountryCode,
		Latitude:       input.Latitude,
		Longitude:      input.Longitude,
		WorkMode:       input.WorkMode,
		TravelRadiusKm: input.TravelRadiusKm,
	}); err != nil {
		return fmt.Errorf("update location: persist (org-shared): %w", err)
	}
	s.publishReindex(ctx, orgID)
	s.invalidateCache(ctx, orgID)
	return nil
}

// writeAgencyLanguages persists the two language arrays onto the
// organizations row via the org-shared writer, then fires the reindex
// + cache-invalidation hooks. Mirrors the legacy UpdateLanguages
// side effects.
func (s *Service) writeAgencyLanguages(ctx context.Context, orgID uuid.UUID, professional, conversational []string) error {
	if err := s.sharedWriter.UpdateSharedLanguages(ctx, orgID, professional, conversational); err != nil {
		return fmt.Errorf("update languages (org-shared): %w", err)
	}
	s.publishReindex(ctx, orgID)
	s.invalidateCache(ctx, orgID)
	return nil
}
