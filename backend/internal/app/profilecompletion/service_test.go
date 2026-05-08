package profilecompletion_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/app/profilecompletion"
	"marketplace-backend/internal/domain/invoicing"
	"marketplace-backend/internal/domain/organization"
	"marketplace-backend/internal/domain/profile"
	"marketplace-backend/internal/domain/user"
)

// completionFixture wires a Service with a deterministic set of
// readers — every reader is configurable per test so we can exercise
// "all empty", "all filled", and the partial scenarios without
// duplicating boilerplate.
type completionFixture struct {
	users *userReaderStub
	orgs  *orgReaderStub
	deps  profilecompletion.Deps
}

func newFixture() *completionFixture {
	return &completionFixture{
		users: &userReaderStub{},
		orgs:  &orgReaderStub{},
	}
}

func (f *completionFixture) build(t *testing.T) *profilecompletion.Service {
	t.Helper()
	d := f.deps
	d.Users = f.users
	d.Organizations = f.orgs
	svc, err := profilecompletion.NewService(d)
	require.NoError(t, err)
	return svc
}

// ---------------------------------------------------------------
// Stubs (reader doubles)
// ---------------------------------------------------------------

type userReaderStub struct {
	user *user.User
	err  error
}

func (s *userReaderStub) GetByID(_ context.Context, _ uuid.UUID) (*user.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

type orgReaderStub struct {
	org *organization.Organization
	err error
}

func (s *orgReaderStub) FindByID(_ context.Context, _ uuid.UUID) (*organization.Organization, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.org, nil
}

// fakeSharedReader builds the SharedProfile inline so each test case
// can declare the exact shape it cares about.
type fakeSharedReader struct {
	ph    string
	c     string
	cc    string
	langs []string
	err   error
}

func (s *fakeSharedReader) GetSharedProfile(_ context.Context, _ uuid.UUID) (*profilecompletion.SharedProfile, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &profilecompletion.SharedProfile{
		PhotoURL:              s.ph,
		City:                  s.c,
		CountryCode:           s.cc,
		LanguagesProfessional: s.langs,
	}, nil
}

// fakeFreelanceReader builds a FreelanceProfileSnapshot.
type fakeFreelanceReader struct {
	profileID    uuid.UUID
	title        string
	about        string
	video        string
	expertises   []string
	availability profile.AvailabilityStatus
	err          error
}

func (f *fakeFreelanceReader) GetByOrgID(_ context.Context, _ uuid.UUID) (*profilecompletion.FreelanceProfileSnapshot, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &profilecompletion.FreelanceProfileSnapshot{
		ProfileID:          f.profileID,
		Title:              f.title,
		About:              f.about,
		VideoURL:           f.video,
		ExpertiseDomains:   f.expertises,
		AvailabilityStatus: f.availability,
	}, nil
}

// fakeLegacyReader returns a *profile.Profile pointer.
type fakeLegacyReader struct {
	p   *profile.Profile
	err error
}

func (f *fakeLegacyReader) GetByOrgID(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.p, nil
}

// skillsCounterStub returns a fixed count.
type skillsCounterStub struct {
	n   int
	err error
}

func (s *skillsCounterStub) CountByOrg(_ context.Context, _ uuid.UUID) (int, error) {
	return s.n, s.err
}

// socialLinksCounterStub returns per-persona counts.
type socialLinksCounterStub struct {
	freelance int
	referrer  int
	agency    int
	err       error
}

func (s *socialLinksCounterStub) CountByOrgPersona(_ context.Context, _ uuid.UUID, p profile.SocialLinkPersona) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	switch p {
	case profile.PersonaFreelance:
		return s.freelance, nil
	case profile.PersonaReferrer:
		return s.referrer, nil
	case profile.PersonaAgency:
		return s.agency, nil
	}
	return 0, nil
}

type portfolioCounterStub struct {
	n int
}

func (s *portfolioCounterStub) CountByOrganization(_ context.Context, _ uuid.UUID) (int, error) {
	return s.n, nil
}

type billingReaderStub struct {
	bp  *invoicing.BillingProfile
	err error
}

func (s *billingReaderStub) FindByOrganization(_ context.Context, _ uuid.UUID) (*invoicing.BillingProfile, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.bp, nil
}

type pricingExistsStub struct {
	exists bool
}

func (s *pricingExistsStub) ExistsByProfileID(_ context.Context, _ uuid.UUID) (bool, error) {
	return s.exists, nil
}

type legacyPricingCounterStub struct {
	n int
}

func (s *legacyPricingCounterStub) CountByOrgID(_ context.Context, _ uuid.UUID) (int, error) {
	return s.n, nil
}

// ---------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------

func providerUser() *user.User {
	uid := uuid.New()
	return &user.User{
		ID:    uid,
		Email: "p@example.com",
		Role:  user.RoleProvider,
	}
}

func agencyUser() *user.User {
	uid := uuid.New()
	return &user.User{
		ID:    uid,
		Email: "a@example.com",
		Role:  user.RoleAgency,
	}
}

func enterpriseUser() *user.User {
	uid := uuid.New()
	return &user.User{
		ID:    uid,
		Email: "e@example.com",
		Role:  user.RoleEnterprise,
	}
}

func providerOrg() *organization.Organization {
	return &organization.Organization{
		ID:          uuid.New(),
		OwnerUserID: uuid.New(),
		Type:        organization.OrgTypeProviderPersonal,
		Name:        "Solo",
	}
}

func agencyOrg() *organization.Organization {
	return &organization.Organization{
		ID:          uuid.New(),
		OwnerUserID: uuid.New(),
		Type:        organization.OrgTypeAgency,
		Name:        "Agency",
	}
}

func enterpriseOrg() *organization.Organization {
	return &organization.Organization{
		ID:          uuid.New(),
		OwnerUserID: uuid.New(),
		Type:        organization.OrgTypeEnterprise,
		Name:        "Enterprise",
	}
}

// ---------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------

func TestNewService_RequiresUsersReader(t *testing.T) {
	_, err := profilecompletion.NewService(profilecompletion.Deps{
		Organizations: &orgReaderStub{},
	})
	assert.Error(t, err)
}

func TestNewService_RequiresOrganizationsReader(t *testing.T) {
	_, err := profilecompletion.NewService(profilecompletion.Deps{
		Users: &userReaderStub{},
	})
	assert.Error(t, err)
}

// ---------------------------------------------------------------
// Compute — provider/freelance role
// ---------------------------------------------------------------

func TestCompute_Provider_AllEmpty_ZeroPercentExceptDefaults(t *testing.T) {
	f := newFixture()
	f.users.user = providerUser()
	f.orgs.org = providerOrg()

	// Every optional reader is nil → every "external" section
	// collapses to false. The freelance reader returns nil so
	// availability/title/about all evaluate as empty.

	svc := f.build(t)
	r, err := svc.Compute(context.Background(), f.users.user.ID, f.orgs.org.ID)
	require.NoError(t, err)

	assert.Equal(t, "provider", r.Role)
	assert.Equal(t, "freelance", r.Persona)
	assert.Equal(t, 13, r.TotalSections, "freelance persona has 13 sections")
	assert.Equal(t, 0, r.FilledSections)
	assert.Equal(t, 0, r.Percent)

	// Sanity-check the section keys are stable across releases.
	assertHasSectionKey(t, r, "photo")
	assertHasSectionKey(t, r, "title")
	assertHasSectionKey(t, r, "billing_profile")
	assertHasSectionKey(t, r, "kyc")
}

func TestCompute_Provider_AllFilled_HundredPercent(t *testing.T) {
	f := newFixture()
	f.users.user = providerUser()
	org := providerOrg()
	stripeAcct := "acct_test"
	org.StripeAccountID = &stripeAcct
	f.orgs.org = org

	f.deps.Shared = &fakeSharedReader{ph: "photo.jpg", c: "Paris", cc: "FR", langs: []string{"fr"}}
	f.deps.FreelanceProfile = &fakeFreelanceReader{
		profileID:   uuid.New(),
		title:       "Senior dev",
		about:       "10 years",
		video:       "https://video",
		expertises:  []string{"backend"},
		availability: profile.AvailabilityNow,
	}
	f.deps.Skills = &skillsCounterStub{n: 5}
	f.deps.SocialLinks = &socialLinksCounterStub{freelance: 1}
	f.deps.BillingProfile = &billingReaderStub{bp: completeBillingProfile()}
	f.deps.FreelancePricing = &pricingExistsStub{exists: true}

	svc := f.build(t)
	r, err := svc.Compute(context.Background(), f.users.user.ID, f.orgs.org.ID)
	require.NoError(t, err)

	assert.Equal(t, 13, r.TotalSections)
	assert.Equal(t, 13, r.FilledSections)
	assert.Equal(t, 100, r.Percent)
}

func TestCompute_Provider_PartialFilled_RoundedFraction(t *testing.T) {
	f := newFixture()
	f.users.user = providerUser()
	f.orgs.org = providerOrg()

	// Fill 4 sections out of 13 → 4*100/13 = 30 (integer division).
	f.deps.Shared = &fakeSharedReader{ph: "photo.jpg"}
	f.deps.FreelanceProfile = &fakeFreelanceReader{
		profileID: uuid.New(),
		title:     "Hi",
		about:     "Yo",
	}
	f.deps.Skills = &skillsCounterStub{n: 1}
	f.deps.SocialLinks = &socialLinksCounterStub{}
	f.deps.BillingProfile = &billingReaderStub{}

	svc := f.build(t)
	r, err := svc.Compute(context.Background(), f.users.user.ID, f.orgs.org.ID)
	require.NoError(t, err)

	assert.Equal(t, 4, r.FilledSections)
	assert.Equal(t, 30, r.Percent)
}

// ---------------------------------------------------------------
// Compute — agency role
// ---------------------------------------------------------------

func TestCompute_Agency_EmptyAndFilled(t *testing.T) {
	f := newFixture()
	f.users.user = agencyUser()
	f.orgs.org = agencyOrg()

	svc := f.build(t)
	empty, err := svc.Compute(context.Background(), f.users.user.ID, f.orgs.org.ID)
	require.NoError(t, err)
	assert.Equal(t, "agency", empty.Persona)
	assert.Equal(t, 12, empty.TotalSections)
	assert.Equal(t, 0, empty.FilledSections)

	// Now fill every reader.
	stripeAcct := "acct_x"
	org := agencyOrg()
	org.StripeAccountID = &stripeAcct
	f.orgs.org = org

	f.deps.Shared = &fakeSharedReader{ph: "p.jpg", c: "Lyon", cc: "FR", langs: []string{"fr"}}
	f.deps.LegacyProfile = &fakeLegacyReader{p: &profile.Profile{
		Title:                 "Agency",
		About:                 "We do",
		AvailabilityStatus:    profile.AvailabilityNow,
		LanguagesProfessional: []string{"fr"},
	}}
	f.deps.Skills = &skillsCounterStub{n: 2}
	f.deps.SocialLinks = &socialLinksCounterStub{agency: 2}
	f.deps.Portfolio = &portfolioCounterStub{n: 3}
	f.deps.BillingProfile = &billingReaderStub{bp: completeBillingProfile()}
	f.deps.LegacyPricing = &legacyPricingCounterStub{n: 1}

	svc = f.build(t)
	full, err := svc.Compute(context.Background(), f.users.user.ID, org.ID)
	require.NoError(t, err)
	assert.Equal(t, 12, full.FilledSections)
	assert.Equal(t, 100, full.Percent)
}

// ---------------------------------------------------------------
// Compute — enterprise role
// ---------------------------------------------------------------

func TestCompute_Enterprise_EmptyAndFilled(t *testing.T) {
	f := newFixture()
	f.users.user = enterpriseUser()
	f.orgs.org = enterpriseOrg()

	svc := f.build(t)
	empty, err := svc.Compute(context.Background(), f.users.user.ID, f.orgs.org.ID)
	require.NoError(t, err)
	assert.Equal(t, "enterprise", empty.Persona)
	assert.Equal(t, 4, empty.TotalSections)
	assert.Equal(t, 0, empty.FilledSections)

	stripeAcct := "acct_e"
	org := enterpriseOrg()
	org.StripeAccountID = &stripeAcct
	f.orgs.org = org

	f.deps.Shared = &fakeSharedReader{ph: "p.jpg"}
	f.deps.LegacyProfile = &fakeLegacyReader{p: &profile.Profile{ClientDescription: "We buy"}}
	f.deps.BillingProfile = &billingReaderStub{bp: completeBillingProfile()}

	svc = f.build(t)
	full, err := svc.Compute(context.Background(), f.users.user.ID, org.ID)
	require.NoError(t, err)
	assert.Equal(t, 4, full.FilledSections)
	assert.Equal(t, 100, full.Percent)
}

// ---------------------------------------------------------------
// Error propagation
// ---------------------------------------------------------------

func TestCompute_UserReaderError_Surfaces(t *testing.T) {
	f := newFixture()
	f.users.err = errors.New("boom")

	svc := f.build(t)
	_, err := svc.Compute(context.Background(), uuid.New(), uuid.New())
	assert.ErrorContains(t, err, "load user")
}

func TestCompute_OrgReaderError_Surfaces(t *testing.T) {
	f := newFixture()
	f.users.user = providerUser()
	f.orgs.err = errors.New("kaput")

	svc := f.build(t)
	_, err := svc.Compute(context.Background(), uuid.New(), uuid.New())
	assert.ErrorContains(t, err, "load org")
}

func TestCompute_BillingReaderNotFound_TreatedAsEmpty(t *testing.T) {
	f := newFixture()
	f.users.user = providerUser()
	f.orgs.org = providerOrg()
	f.deps.BillingProfile = &billingReaderStub{err: profilecompletion.ErrNotFound}

	svc := f.build(t)
	r, err := svc.Compute(context.Background(), f.users.user.ID, f.orgs.org.ID)
	require.NoError(t, err)
	// Billing section must be empty (not filled).
	for _, s := range r.Sections {
		if s.Key == "billing_profile" {
			assert.False(t, s.Filled)
			return
		}
	}
	t.Fatal("billing_profile section missing")
}

// ---------------------------------------------------------------
// Section payload
// ---------------------------------------------------------------

func TestCompute_SectionsCarryLabelKeyAndCompletionPath(t *testing.T) {
	f := newFixture()
	f.users.user = providerUser()
	f.orgs.org = providerOrg()
	svc := f.build(t)

	r, err := svc.Compute(context.Background(), f.users.user.ID, f.orgs.org.ID)
	require.NoError(t, err)
	for _, s := range r.Sections {
		assert.NotEmpty(t, s.LabelKey, "every section must carry a label key")
		assert.Contains(t, s.LabelKey, "profile.completion.section.",
			"label key must be the dotted i18n path")
		assert.NotEmpty(t, s.CompletionPath, "every section must carry a completion path")
	}
}

// ---------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------

func assertHasSectionKey(t *testing.T, r *profilecompletion.Report, key string) {
	t.Helper()
	for _, s := range r.Sections {
		if string(s.Key) == key {
			return
		}
	}
	t.Errorf("missing section key %q", key)
}

func completeBillingProfile() *invoicing.BillingProfile {
	return &invoicing.BillingProfile{
		ProfileType:  invoicing.ProfileBusiness,
		LegalName:    "ACME SAS",
		AddressLine1: "1 rue test",
		PostalCode:   "75001",
		City:         "Paris",
		Country:      "US", // outside-EU branch — minimal fields suffice
		TaxID:        "",
	}
}
