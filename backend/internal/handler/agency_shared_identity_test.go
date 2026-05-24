package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/domain/organization"
	"marketplace-backend/internal/domain/profile"
	"marketplace-backend/internal/handler/middleware"
	"marketplace-backend/internal/port/repository"
)

// Tests for the agency shared-identity backend wiring (migration 155):
//   - UploadPhoto redirects the photo write to the organizations row
//     for agency orgs, and keeps the legacy profiles write otherwise.
//   - OrganizationSharedProfileHandler busts the public agency profile
//     cache after every shared-profile mutation.

// --- shared test doubles -------------------------------------------------

type fakeOrgTypeReader struct {
	orgType organization.OrgType
	err     error
}

func (f *fakeOrgTypeReader) FindByID(_ context.Context, id uuid.UUID) (*organization.Organization, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &organization.Organization{ID: id, Type: f.orgType}, nil
}

type fakeSharedWriter struct {
	photoCalls     int
	lastPhoto      string
	locationCalls  int
	languagesCalls int
	updateErr      error
}

func (f *fakeSharedWriter) UpdateSharedLocation(_ context.Context, _ uuid.UUID, _ repository.SharedProfileLocationInput) error {
	f.locationCalls++
	return f.updateErr
}

func (f *fakeSharedWriter) UpdateSharedLanguages(_ context.Context, _ uuid.UUID, _, _ []string) error {
	f.languagesCalls++
	return f.updateErr
}

func (f *fakeSharedWriter) UpdateSharedPhotoURL(_ context.Context, _ uuid.UUID, photoURL string) error {
	f.photoCalls++
	f.lastPhoto = photoURL
	return f.updateErr
}

func (f *fakeSharedWriter) GetSharedProfile(_ context.Context, _ uuid.UUID) (*repository.OrganizationSharedProfile, error) {
	return &repository.OrganizationSharedProfile{}, nil
}

type spyCacheInvalidator struct {
	calls int
	last  uuid.UUID
}

func (s *spyCacheInvalidator) Invalidate(_ context.Context, orgID uuid.UUID) error {
	s.calls++
	s.last = orgID
	return nil
}

func withUserAndOrgCtx(req *http.Request, userID, orgID uuid.UUID) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, middleware.ContextKeyOrganizationID, orgID)
	return req.WithContext(ctx)
}

// =====================================================================
// UploadPhoto agency redirection
// =====================================================================

func TestUploadPhoto_AgencyRoutesToOrgShared(t *testing.T) {
	userID, orgID := uuid.New(), uuid.New()
	// Default mock Upload returns "https://storage.example.com/<key>".
	storage := &mockStorageService{}
	// The legacy profiles write MUST NOT happen for an agency.
	profiles := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, _ uuid.UUID) (*profile.Profile, error) {
			t.Fatal("legacy profiles read/write must not run for an agency photo upload")
			return nil, nil
		},
	}
	writer := &fakeSharedWriter{}

	h := NewUploadHandler(storage, profiles, nil).
		WithAgencyPhotoRouter(&fakeOrgTypeReader{orgType: organization.OrgTypeAgency}, writer)

	req := buildMultipartRequest(http.MethodPost, "/api/v1/upload/photo", "file", "p.jpg", "image/jpeg", validJPEG())
	req = withUserAndOrgCtx(req, userID, orgID)
	rec := httptest.NewRecorder()

	h.UploadPhoto(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.Equal(t, 1, writer.photoCalls, "agency photo must be stamped on the organizations row")
	assert.True(t, strings.HasPrefix(writer.lastPhoto, "https://storage.example.com/"))
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.NotEmpty(t, resp["url"])
}

func TestUploadPhoto_NonAgencyKeepsLegacyProfilesWrite(t *testing.T) {
	userID, orgID := uuid.New(), uuid.New()
	storage := &mockStorageService{}
	var legacyUpdated bool
	var stampedURL string
	profiles := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, oid uuid.UUID) (*profile.Profile, error) {
			return &profile.Profile{OrganizationID: oid}, nil
		},
		updateFn: func(_ context.Context, p *profile.Profile) error {
			legacyUpdated = true
			stampedURL = p.PhotoURL
			return nil
		},
	}
	writer := &fakeSharedWriter{}

	h := NewUploadHandler(storage, profiles, nil).
		WithAgencyPhotoRouter(&fakeOrgTypeReader{orgType: organization.OrgTypeEnterprise}, writer)

	req := buildMultipartRequest(http.MethodPost, "/api/v1/upload/photo", "file", "p.png", "image/png", validPNG())
	req = withUserAndOrgCtx(req, userID, orgID)
	rec := httptest.NewRecorder()

	h.UploadPhoto(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.True(t, legacyUpdated, "enterprise photo must keep stamping the legacy profiles row")
	assert.NotEmpty(t, stampedURL)
	assert.Equal(t, 0, writer.photoCalls, "org-shared writer must NOT run for enterprise")
}

func TestUploadPhoto_NoRouterKeepsLegacy(t *testing.T) {
	userID, orgID := uuid.New(), uuid.New()
	storage := &mockStorageService{}
	var legacyUpdated bool
	profiles := &mockProfileRepo{
		getByOrgIDFn: func(_ context.Context, oid uuid.UUID) (*profile.Profile, error) {
			return &profile.Profile{OrganizationID: oid}, nil
		},
		updateFn: func(_ context.Context, _ *profile.Profile) error { legacyUpdated = true; return nil },
	}
	// Router NOT wired → legacy path regardless of org type.
	h := NewUploadHandler(storage, profiles, nil)
	req := buildMultipartRequest(http.MethodPost, "/api/v1/upload/photo", "file", "p.jpg", "image/jpeg", validJPEG())
	req = withUserAndOrgCtx(req, userID, orgID)
	rec := httptest.NewRecorder()
	h.UploadPhoto(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.True(t, legacyUpdated)
}

// =====================================================================
// OrganizationSharedProfileHandler — cache invalidation
// =====================================================================

func TestOrganizationSharedHandler_UpdatePhoto_InvalidatesCache(t *testing.T) {
	orgID := uuid.New()
	writer := &fakeSharedWriter{}
	cache := &spyCacheInvalidator{}
	h := NewOrganizationSharedProfileHandler(writer).WithAgencyProfileCacheInvalidator(cache)

	body := strings.NewReader(`{"photo_url":"https://x/y.jpg"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/organization/photo", body)
	req = withUserAndOrgCtx(req, uuid.New(), orgID)
	rec := httptest.NewRecorder()

	h.UpdatePhoto(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.Equal(t, 1, cache.calls, "shared photo write must bust the agency profile cache")
	assert.Equal(t, orgID, cache.last)
}

func TestOrganizationSharedHandler_UpdateLanguages_InvalidatesCache(t *testing.T) {
	orgID := uuid.New()
	writer := &fakeSharedWriter{}
	cache := &spyCacheInvalidator{}
	h := NewOrganizationSharedProfileHandler(writer).WithAgencyProfileCacheInvalidator(cache)

	body := strings.NewReader(`{"professional":["fr"],"conversational":[]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/organization/languages", body)
	req = withUserAndOrgCtx(req, uuid.New(), orgID)
	rec := httptest.NewRecorder()

	h.UpdateLanguages(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.Equal(t, 1, cache.calls)
	assert.Equal(t, 1, writer.languagesCalls)
}

func TestOrganizationSharedHandler_UpdateLocation_InvalidatesCache(t *testing.T) {
	orgID := uuid.New()
	writer := &fakeSharedWriter{}
	cache := &spyCacheInvalidator{}
	h := NewOrganizationSharedProfileHandler(writer).WithAgencyProfileCacheInvalidator(cache)

	body := strings.NewReader(`{"city":"Paris","country_code":"FR","work_mode":["remote"]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/organization/location", body)
	req = withUserAndOrgCtx(req, uuid.New(), orgID)
	rec := httptest.NewRecorder()

	h.UpdateLocation(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.Equal(t, 1, cache.calls)
	assert.Equal(t, 1, writer.locationCalls)
}

func TestOrganizationSharedHandler_NilCacheInvalidatorIsNoop(t *testing.T) {
	// No invalidator wired — the write must still succeed (the entry
	// just ages out on its TTL). Guards the nil-safe path.
	orgID := uuid.New()
	writer := &fakeSharedWriter{}
	h := NewOrganizationSharedProfileHandler(writer)

	body := strings.NewReader(`{"photo_url":""}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/organization/photo", body)
	req = withUserAndOrgCtx(req, uuid.New(), orgID)
	rec := httptest.NewRecorder()

	h.UpdatePhoto(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.Equal(t, 1, writer.photoCalls)
}
