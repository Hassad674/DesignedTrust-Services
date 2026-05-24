package profileapp

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/domain/organization"
	"marketplace-backend/internal/domain/profile"
	"marketplace-backend/internal/port/repository"
)

// --- shared mocks for the agency shared-identity router + overlay ---

type mockOrgTypeReader struct {
	findByIDFn func(ctx context.Context, id uuid.UUID) (*organization.Organization, error)
}

func (m *mockOrgTypeReader) FindByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func orgOfType(id uuid.UUID, t organization.OrgType) *organization.Organization {
	return &organization.Organization{ID: id, Type: t}
}

type mockSharedWriter struct {
	updateLocationFn  func(ctx context.Context, orgID uuid.UUID, in repository.SharedProfileLocationInput) error
	updateLanguagesFn func(ctx context.Context, orgID uuid.UUID, pro, conv []string) error
	updatePhotoFn     func(ctx context.Context, orgID uuid.UUID, photoURL string) error
	getSharedFn       func(ctx context.Context, orgID uuid.UUID) (*repository.OrganizationSharedProfile, error)

	locationCalls  int
	languagesCalls int
}

func (m *mockSharedWriter) UpdateSharedLocation(ctx context.Context, orgID uuid.UUID, in repository.SharedProfileLocationInput) error {
	m.locationCalls++
	if m.updateLocationFn != nil {
		return m.updateLocationFn(ctx, orgID, in)
	}
	return nil
}

func (m *mockSharedWriter) UpdateSharedLanguages(ctx context.Context, orgID uuid.UUID, pro, conv []string) error {
	m.languagesCalls++
	if m.updateLanguagesFn != nil {
		return m.updateLanguagesFn(ctx, orgID, pro, conv)
	}
	return nil
}

func (m *mockSharedWriter) UpdateSharedPhotoURL(ctx context.Context, orgID uuid.UUID, photoURL string) error {
	if m.updatePhotoFn != nil {
		return m.updatePhotoFn(ctx, orgID, photoURL)
	}
	return nil
}

func (m *mockSharedWriter) GetSharedProfile(ctx context.Context, orgID uuid.UUID) (*repository.OrganizationSharedProfile, error) {
	if m.getSharedFn != nil {
		return m.getSharedFn(ctx, orgID)
	}
	return &repository.OrganizationSharedProfile{}, nil
}

type stubProfileReader struct {
	fn func(ctx context.Context, orgID uuid.UUID) (*profile.Profile, error)
}

func (s *stubProfileReader) GetProfile(ctx context.Context, orgID uuid.UUID) (*profile.Profile, error) {
	return s.fn(ctx, orgID)
}

// =====================================================================
// Read overlay decorator
// =====================================================================

func TestAgencySharedProfileReader_OverlaysForAgency(t *testing.T) {
	orgID := uuid.New()
	lat, lng := 48.8566, 2.3522
	radius := 30

	inner := &stubProfileReader{fn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
		// The legacy profiles read returns its OWN (now-non-authoritative)
		// values — the overlay must replace every shared field.
		return &profile.Profile{
			OrganizationID:          orgID,
			Title:                   "Agence Soleil",
			About:                   "We craft brands.",
			PhotoURL:                "https://old/profiles-photo.jpg",
			City:                    "OldCity",
			CountryCode:             "ES",
			WorkMode:                []string{"on_site"},
			LanguagesProfessional:   []string{"es"},
			LanguagesConversational: []string{},
		}, nil
	}}
	orgs := &mockOrgTypeReader{findByIDFn: func(_ context.Context, id uuid.UUID) (*organization.Organization, error) {
		return orgOfType(id, organization.OrgTypeAgency), nil
	}}
	shared := &mockSharedWriter{getSharedFn: func(_ context.Context, _ uuid.UUID) (*repository.OrganizationSharedProfile, error) {
		return &repository.OrganizationSharedProfile{
			PhotoURL:                "https://new/org-photo.jpg",
			City:                    "Paris",
			CountryCode:             "FR",
			Latitude:                &lat,
			Longitude:               &lng,
			WorkMode:                []string{"remote", "hybrid"},
			TravelRadiusKm:          &radius,
			LanguagesProfessional:   []string{"fr", "en"},
			LanguagesConversational: []string{"de"},
		}, nil
	}}

	reader := NewAgencySharedProfileReader(inner, orgs, shared)
	got, err := reader.GetProfile(context.Background(), orgID)
	require.NoError(t, err)

	// Shared fields come from the org row.
	assert.Equal(t, "https://new/org-photo.jpg", got.PhotoURL)
	assert.Equal(t, "Paris", got.City)
	assert.Equal(t, "FR", got.CountryCode)
	require.NotNil(t, got.Latitude)
	assert.InDelta(t, 48.8566, *got.Latitude, 1e-9)
	require.NotNil(t, got.TravelRadiusKm)
	assert.Equal(t, 30, *got.TravelRadiusKm)
	assert.Equal(t, []string{"remote", "hybrid"}, got.WorkMode)
	assert.Equal(t, []string{"fr", "en"}, got.LanguagesProfessional)
	assert.Equal(t, []string{"de"}, got.LanguagesConversational)

	// Non-shared fields are preserved from the legacy read.
	assert.Equal(t, "Agence Soleil", got.Title)
	assert.Equal(t, "We craft brands.", got.About)
}

func TestAgencySharedProfileReader_PassthroughForNonAgency(t *testing.T) {
	for _, orgType := range []organization.OrgType{
		organization.OrgTypeEnterprise,
		organization.OrgTypeProviderPersonal,
	} {
		t.Run(string(orgType), func(t *testing.T) {
			orgID := uuid.New()
			base := &profile.Profile{
				OrganizationID: orgID,
				PhotoURL:       "https://legacy/photo.jpg",
				City:           "Lyon",
				CountryCode:    "FR",
			}
			inner := &stubProfileReader{fn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
				return base, nil
			}}
			orgs := &mockOrgTypeReader{findByIDFn: func(_ context.Context, id uuid.UUID) (*organization.Organization, error) {
				return orgOfType(id, orgType), nil
			}}
			shared := &mockSharedWriter{getSharedFn: func(_ context.Context, _ uuid.UUID) (*repository.OrganizationSharedProfile, error) {
				t.Fatalf("GetSharedProfile must NOT be called for org type %s", orgType)
				return nil, nil
			}}

			reader := NewAgencySharedProfileReader(inner, orgs, shared)
			got, err := reader.GetProfile(context.Background(), orgID)
			require.NoError(t, err)
			// Byte-identical to the legacy read — no overlay.
			assert.Equal(t, "https://legacy/photo.jpg", got.PhotoURL)
			assert.Equal(t, "Lyon", got.City)
		})
	}
}

func TestAgencySharedProfileReader_InnerErrorPropagates(t *testing.T) {
	want := errors.New("db down")
	inner := &stubProfileReader{fn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
		return nil, want
	}}
	reader := NewAgencySharedProfileReader(inner, &mockOrgTypeReader{}, &mockSharedWriter{})
	_, err := reader.GetProfile(context.Background(), uuid.New())
	assert.ErrorIs(t, err, want)
}

func TestAgencySharedProfileReader_DegradesWhenOrgLookupFails(t *testing.T) {
	orgID := uuid.New()
	base := &profile.Profile{OrganizationID: orgID, PhotoURL: "https://legacy/photo.jpg", City: "Nice"}
	inner := &stubProfileReader{fn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
		return base, nil
	}}
	orgs := &mockOrgTypeReader{findByIDFn: func(_ context.Context, _ uuid.UUID) (*organization.Organization, error) {
		return nil, errors.New("org read blip")
	}}
	shared := &mockSharedWriter{getSharedFn: func(_ context.Context, _ uuid.UUID) (*repository.OrganizationSharedProfile, error) {
		t.Fatal("shared read must be skipped when org lookup fails")
		return nil, nil
	}}
	reader := NewAgencySharedProfileReader(inner, orgs, shared)
	got, err := reader.GetProfile(context.Background(), orgID)
	require.NoError(t, err)
	// Falls back to the legacy read rather than failing.
	assert.Equal(t, "https://legacy/photo.jpg", got.PhotoURL)
	assert.Equal(t, "Nice", got.City)
}

func TestAgencySharedProfileReader_DegradesWhenSharedReadFails(t *testing.T) {
	orgID := uuid.New()
	base := &profile.Profile{OrganizationID: orgID, City: "Nantes"}
	inner := &stubProfileReader{fn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
		return base, nil
	}}
	orgs := &mockOrgTypeReader{findByIDFn: func(_ context.Context, id uuid.UUID) (*organization.Organization, error) {
		return orgOfType(id, organization.OrgTypeAgency), nil
	}}
	shared := &mockSharedWriter{getSharedFn: func(_ context.Context, _ uuid.UUID) (*repository.OrganizationSharedProfile, error) {
		return nil, errors.New("shared read blip")
	}}
	reader := NewAgencySharedProfileReader(inner, orgs, shared)
	got, err := reader.GetProfile(context.Background(), orgID)
	require.NoError(t, err)
	assert.Equal(t, "Nantes", got.City)
}

func TestAgencySharedProfileReader_NilCollaboratorsPassthrough(t *testing.T) {
	orgID := uuid.New()
	base := &profile.Profile{OrganizationID: orgID, City: "Brest"}
	inner := &stubProfileReader{fn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
		return base, nil
	}}
	reader := NewAgencySharedProfileReader(inner, nil, nil)
	got, err := reader.GetProfile(context.Background(), orgID)
	require.NoError(t, err)
	assert.Equal(t, "Brest", got.City)
}

func TestAgencySharedProfileReader_OverlayNilArraysBecomeEmpty(t *testing.T) {
	orgID := uuid.New()
	inner := &stubProfileReader{fn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
		return &profile.Profile{OrganizationID: orgID}, nil
	}}
	orgs := &mockOrgTypeReader{findByIDFn: func(_ context.Context, id uuid.UUID) (*organization.Organization, error) {
		return orgOfType(id, organization.OrgTypeAgency), nil
	}}
	shared := &mockSharedWriter{getSharedFn: func(_ context.Context, _ uuid.UUID) (*repository.OrganizationSharedProfile, error) {
		// nil arrays from the org row must surface as non-nil empties.
		return &repository.OrganizationSharedProfile{
			WorkMode:                nil,
			LanguagesProfessional:   nil,
			LanguagesConversational: nil,
		}, nil
	}}
	reader := NewAgencySharedProfileReader(inner, orgs, shared)
	got, err := reader.GetProfile(context.Background(), orgID)
	require.NoError(t, err)
	assert.NotNil(t, got.WorkMode)
	assert.NotNil(t, got.LanguagesProfessional)
	assert.NotNil(t, got.LanguagesConversational)
	assert.Len(t, got.WorkMode, 0)
}

// =====================================================================
// Write redirection (UpdateLocation / UpdateLanguages)
// =====================================================================

func newRoutedService(repo *mockProfileRepo, orgs OrgTypeReader, writer repository.OrganizationSharedProfileWriter) *Service {
	if repo == nil {
		repo = &mockProfileRepo{}
	}
	return NewService(repo).WithAgencySharedRouter(orgs, writer)
}

func TestService_UpdateLocation_RoutesAgencyToOrgShared(t *testing.T) {
	orgID := uuid.New()
	var legacyCalled bool
	repo := &mockProfileRepo{
		updateLocationFn: func(_ context.Context, _ uuid.UUID, _ repository.LocationInput) error {
			legacyCalled = true
			return nil
		},
	}
	orgs := &mockOrgTypeReader{findByIDFn: func(_ context.Context, id uuid.UUID) (*organization.Organization, error) {
		return orgOfType(id, organization.OrgTypeAgency), nil
	}}
	var gotInput repository.SharedProfileLocationInput
	writer := &mockSharedWriter{updateLocationFn: func(_ context.Context, _ uuid.UUID, in repository.SharedProfileLocationInput) error {
		gotInput = in
		return nil
	}}

	svc := newRoutedService(repo, orgs, writer)
	err := svc.UpdateLocation(context.Background(), orgID, UpdateLocationInput{
		City:        "Paris",
		CountryCode: "fr", // lower — service must upper-normalise
		WorkMode:    []string{"remote"},
	})
	require.NoError(t, err)
	assert.False(t, legacyCalled, "legacy profiles.UpdateLocation must NOT run for an agency")
	assert.Equal(t, 1, writer.locationCalls, "org-shared writer must run exactly once")
	assert.Equal(t, "Paris", gotInput.City)
	assert.Equal(t, "FR", gotInput.CountryCode, "country code normalised to upper")
	assert.Equal(t, []string{"remote"}, gotInput.WorkMode)
}

func TestService_UpdateLocation_NonAgencyKeepsLegacy(t *testing.T) {
	orgID := uuid.New()
	var legacyCalled bool
	repo := &mockProfileRepo{
		updateLocationFn: func(_ context.Context, _ uuid.UUID, _ repository.LocationInput) error {
			legacyCalled = true
			return nil
		},
	}
	orgs := &mockOrgTypeReader{findByIDFn: func(_ context.Context, id uuid.UUID) (*organization.Organization, error) {
		return orgOfType(id, organization.OrgTypeEnterprise), nil
	}}
	writer := &mockSharedWriter{}

	svc := newRoutedService(repo, orgs, writer)
	require.NoError(t, svc.UpdateLocation(context.Background(), orgID, UpdateLocationInput{
		City: "Lyon", CountryCode: "FR",
	}))
	assert.True(t, legacyCalled, "enterprise must keep writing the legacy profiles columns")
	assert.Equal(t, 0, writer.locationCalls, "org-shared writer must NOT run for enterprise")
}

func TestService_UpdateLocation_NoRouterKeepsLegacy(t *testing.T) {
	orgID := uuid.New()
	var legacyCalled bool
	repo := &mockProfileRepo{
		updateLocationFn: func(_ context.Context, _ uuid.UUID, _ repository.LocationInput) error {
			legacyCalled = true
			return nil
		},
	}
	// Router NOT wired → pure legacy behaviour regardless of org type.
	svc := NewService(repo)
	require.NoError(t, svc.UpdateLocation(context.Background(), orgID, UpdateLocationInput{
		City: "Lyon", CountryCode: "FR",
	}))
	assert.True(t, legacyCalled)
}

func TestService_UpdateLanguages_RoutesAgencyToOrgShared(t *testing.T) {
	orgID := uuid.New()
	var legacyCalled bool
	repo := &mockProfileRepo{
		updateLanguagesFn: func(_ context.Context, _ uuid.UUID, _, _ []string) error {
			legacyCalled = true
			return nil
		},
	}
	orgs := &mockOrgTypeReader{findByIDFn: func(_ context.Context, id uuid.UUID) (*organization.Organization, error) {
		return orgOfType(id, organization.OrgTypeAgency), nil
	}}
	var gotPro, gotConv []string
	writer := &mockSharedWriter{updateLanguagesFn: func(_ context.Context, _ uuid.UUID, pro, conv []string) error {
		gotPro, gotConv = pro, conv
		return nil
	}}

	svc := newRoutedService(repo, orgs, writer)
	// Clients send canonical lowercase ISO-639-1 codes; the normalizer
	// filters/dedups (it does not transcase) — an unknown/dup is dropped.
	require.NoError(t, svc.UpdateLanguages(context.Background(), orgID, []string{"fr", "en", "fr"}, []string{"de"}))
	assert.False(t, legacyCalled)
	assert.Equal(t, 1, writer.languagesCalls)
	assert.Equal(t, []string{"fr", "en"}, gotPro, "normalizer dedups but preserves order")
	assert.Equal(t, []string{"de"}, gotConv)
}

func TestService_UpdateLanguages_NonAgencyKeepsLegacy(t *testing.T) {
	orgID := uuid.New()
	var legacyCalled bool
	repo := &mockProfileRepo{
		updateLanguagesFn: func(_ context.Context, _ uuid.UUID, _, _ []string) error {
			legacyCalled = true
			return nil
		},
	}
	orgs := &mockOrgTypeReader{findByIDFn: func(_ context.Context, id uuid.UUID) (*organization.Organization, error) {
		return orgOfType(id, organization.OrgTypeProviderPersonal), nil
	}}
	writer := &mockSharedWriter{}
	svc := newRoutedService(repo, orgs, writer)
	require.NoError(t, svc.UpdateLanguages(context.Background(), orgID, []string{"fr"}, nil))
	assert.True(t, legacyCalled)
	assert.Equal(t, 0, writer.languagesCalls)
}

func TestService_UpdateLocation_OrgLookupErrorDegradesToLegacy(t *testing.T) {
	orgID := uuid.New()
	var legacyCalled bool
	repo := &mockProfileRepo{
		updateLocationFn: func(_ context.Context, _ uuid.UUID, _ repository.LocationInput) error {
			legacyCalled = true
			return nil
		},
	}
	orgs := &mockOrgTypeReader{findByIDFn: func(_ context.Context, _ uuid.UUID) (*organization.Organization, error) {
		return nil, errors.New("org read blip")
	}}
	writer := &mockSharedWriter{}
	svc := newRoutedService(repo, orgs, writer)
	require.NoError(t, svc.UpdateLocation(context.Background(), orgID, UpdateLocationInput{City: "X", CountryCode: "FR"}))
	assert.True(t, legacyCalled, "org lookup failure must degrade to the legacy write, not fail the save")
	assert.Equal(t, 0, writer.locationCalls)
}
