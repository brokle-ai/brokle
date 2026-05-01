// Service-level tests for OrganizationSettingsService — focused on the
// repo-error contract:
//
//   - GetByKey returning NotFound is the "no existing row" signal for
//     CreateSetting; service must NOT translate it as a 500.
//   - GetByKey returning Internal must propagate as Internal; the
//     previous blanket NotFound("setting") translation in
//     UpdateSetting / DeleteSetting hid 500s as 404s.
//
// These tests pin the round-29 reviewer regression fix.
package organization

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	orgDomain "brokle/internal/core/domain/organization"
	appErrors "brokle/pkg/errors"
)

// ---- fake repositories ---------------------------------------------------

// fakeMemberRepo answers IsMember with a fixed bool. Other methods panic
// so an unintended caller surfaces in test output.
type fakeMemberRepo struct {
	isMember    bool
	isMemberErr error
}

func (f *fakeMemberRepo) IsMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return f.isMember, f.isMemberErr
}
func (f *fakeMemberRepo) Create(context.Context, *orgDomain.Member) error {
	panic("unimplemented")
}
func (f *fakeMemberRepo) GetByID(context.Context, uuid.UUID) (*orgDomain.Member, error) {
	panic("unimplemented")
}
func (f *fakeMemberRepo) GetByUserAndOrg(context.Context, uuid.UUID, uuid.UUID) (*orgDomain.Member, error) {
	panic("unimplemented")
}
func (f *fakeMemberRepo) GetByUserAndOrganization(context.Context, uuid.UUID, uuid.UUID) (*orgDomain.Member, error) {
	panic("unimplemented")
}
func (f *fakeMemberRepo) Update(context.Context, *orgDomain.Member) error {
	panic("unimplemented")
}
func (f *fakeMemberRepo) Delete(context.Context, uuid.UUID) error {
	panic("unimplemented")
}
func (f *fakeMemberRepo) DeleteByUserAndOrg(context.Context, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (f *fakeMemberRepo) GetMembersByOrganizationID(context.Context, uuid.UUID) ([]*orgDomain.Member, error) {
	panic("unimplemented")
}
func (f *fakeMemberRepo) GetByOrganizationID(context.Context, uuid.UUID) ([]*orgDomain.Member, error) {
	panic("unimplemented")
}
func (f *fakeMemberRepo) GetMembersByUserID(context.Context, uuid.UUID) ([]*orgDomain.Member, error) {
	panic("unimplemented")
}
func (f *fakeMemberRepo) UpdateMemberRole(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (f *fakeMemberRepo) GetMemberRole(context.Context, uuid.UUID, uuid.UUID) (uuid.UUID, error) {
	panic("unimplemented")
}
func (f *fakeMemberRepo) CountByOrganizationAndRole(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	panic("unimplemented")
}
func (f *fakeMemberRepo) GetMemberCount(context.Context, uuid.UUID) (int, error) {
	panic("unimplemented")
}

// fakeSettingsRepo records calls and returns configurable responses.
type fakeSettingsRepo struct {
	getByKeyResp *orgDomain.OrganizationSettings
	getByKeyErr  error
	createCalled int
	createErr    error
	updateCalled int
	updateErr    error
}

func (f *fakeSettingsRepo) GetByKey(context.Context, uuid.UUID, string) (*orgDomain.OrganizationSettings, error) {
	return f.getByKeyResp, f.getByKeyErr
}
func (f *fakeSettingsRepo) Create(context.Context, *orgDomain.OrganizationSettings) error {
	f.createCalled++
	return f.createErr
}
func (f *fakeSettingsRepo) Update(context.Context, *orgDomain.OrganizationSettings) error {
	f.updateCalled++
	return f.updateErr
}
func (f *fakeSettingsRepo) GetByID(context.Context, uuid.UUID) (*orgDomain.OrganizationSettings, error) {
	panic("unimplemented")
}
func (f *fakeSettingsRepo) Delete(context.Context, uuid.UUID) error {
	panic("unimplemented")
}
func (f *fakeSettingsRepo) GetAllByOrganizationID(context.Context, uuid.UUID) ([]*orgDomain.OrganizationSettings, error) {
	panic("unimplemented")
}
func (f *fakeSettingsRepo) GetSettingsMap(context.Context, uuid.UUID) (map[string]any, error) {
	panic("unimplemented")
}
func (f *fakeSettingsRepo) DeleteByKey(context.Context, uuid.UUID, string) error {
	panic("unimplemented")
}
func (f *fakeSettingsRepo) UpsertSetting(context.Context, uuid.UUID, string, any) (*orgDomain.OrganizationSettings, error) {
	panic("unimplemented")
}
func (f *fakeSettingsRepo) CreateMultiple(context.Context, []*orgDomain.OrganizationSettings) error {
	panic("unimplemented")
}
func (f *fakeSettingsRepo) GetByKeys(context.Context, uuid.UUID, []string) ([]*orgDomain.OrganizationSettings, error) {
	panic("unimplemented")
}
func (f *fakeSettingsRepo) DeleteMultiple(context.Context, uuid.UUID, []string) error {
	panic("unimplemented")
}

// ---- helpers -------------------------------------------------------------

func newServiceWithStubs(settings *fakeSettingsRepo) *OrganizationSettingsService {
	return NewOrganizationSettingsService(settings, &fakeMemberRepo{isMember: true})
}

// ---- tests ---------------------------------------------------------------

// TestCreateSetting_BrandNewKey_Inserts pins the round-29 reviewer
// regression fix: GetByKey returning NotFound("organization_settings")
// is the "no existing row" signal — the service must proceed to Create,
// not 500 on a stringly-typed error compare.
func TestCreateSetting_BrandNewKey_Inserts(t *testing.T) {
	repo := &fakeSettingsRepo{
		getByKeyErr: appErrors.NotFound("organization_settings"),
	}
	svc := newServiceWithStubs(repo)

	got, err := svc.CreateSetting(context.Background(), uuid.New(), uuid.New(),
		&orgDomain.CreateOrganizationSettingRequest{Key: "feature.x", Value: true})
	if err != nil {
		t.Fatalf("expected nil error on brand-new key, got %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil setting on insert")
	}
	if repo.createCalled != 1 {
		t.Errorf("expected exactly one Create call, got %d", repo.createCalled)
	}
}

// TestCreateSetting_DuplicateKey_Returns409 — when GetByKey returns a
// non-nil setting (key already in use), service must reject with
// AlreadyExists, not write a duplicate.
func TestCreateSetting_DuplicateKey_Returns409(t *testing.T) {
	existing := &orgDomain.OrganizationSettings{ID: uuid.New(), Key: "feature.x"}
	repo := &fakeSettingsRepo{getByKeyResp: existing}
	svc := newServiceWithStubs(repo)

	_, err := svc.CreateSetting(context.Background(), uuid.New(), uuid.New(),
		&orgDomain.CreateOrganizationSettingRequest{Key: "feature.x", Value: true})
	if err == nil {
		t.Fatal("expected AlreadyExists, got nil")
	}
	if !appErrors.IsAlreadyExists(err) {
		t.Errorf("expected AlreadyExists, got %v", err)
	}
	if repo.createCalled != 0 {
		t.Errorf("expected NO Create call on duplicate, got %d", repo.createCalled)
	}
}

// TestCreateSetting_RepoInternalPropagates — GetByKey returning a
// repo-Internal error must surface as Internal at the service boundary.
// Previously the stringly-typed compare wrapped EVERY GetByKey error
// (including legitimate Internal infra failures) into Internal — that
// part still holds — but with the fix, a NotFound (the most common
// error) flows through cleanly while Internal is still surfaced.
func TestCreateSetting_RepoInternalPropagates(t *testing.T) {
	infraErr := appErrors.Internal("db connection lost", errors.New("conn refused"))
	repo := &fakeSettingsRepo{getByKeyErr: infraErr}
	svc := newServiceWithStubs(repo)

	_, err := svc.CreateSetting(context.Background(), uuid.New(), uuid.New(),
		&orgDomain.CreateOrganizationSettingRequest{Key: "feature.x", Value: true})
	if err == nil {
		t.Fatal("expected Internal, got nil")
	}
	if !appErrors.IsReason(err, appErrors.ReasonInternal) {
		t.Errorf("expected Internal reason, got %v", err)
	}
	if repo.createCalled != 0 {
		t.Errorf("expected NO Create call on infra failure, got %d", repo.createCalled)
	}
}

// TestUpdateSetting_RepoNotFoundPropagates — repo NotFound flows
// through as the typed *Error, not as a generic NotFound("setting")
// wrap that hid the repo's resource name.
func TestUpdateSetting_RepoNotFoundPropagates(t *testing.T) {
	repo := &fakeSettingsRepo{
		getByKeyErr: appErrors.NotFound("organization_settings"),
	}
	svc := newServiceWithStubs(repo)

	_, err := svc.UpdateSetting(context.Background(), uuid.New(), "feature.x", uuid.New(),
		&orgDomain.UpdateOrganizationSettingRequest{Value: false})
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	if !appErrors.IsNotFound(err) {
		t.Errorf("expected NotFound, got %v", err)
	}
	if repo.updateCalled != 0 {
		t.Errorf("expected NO Update call when GetByKey returns NotFound, got %d", repo.updateCalled)
	}
}

// TestUpdateSetting_RepoInternalPropagates — pin the bug fix: the
// previous blanket `if err != nil { return NotFound("setting") }`
// hid 500s as 404s. After the fix, repo-Internal must propagate as
// Internal.
func TestUpdateSetting_RepoInternalPropagates(t *testing.T) {
	infraErr := appErrors.Internal("db down", errors.New("timeout"))
	repo := &fakeSettingsRepo{getByKeyErr: infraErr}
	svc := newServiceWithStubs(repo)

	_, err := svc.UpdateSetting(context.Background(), uuid.New(), "feature.x", uuid.New(),
		&orgDomain.UpdateOrganizationSettingRequest{Value: false})
	if err == nil {
		t.Fatal("expected Internal, got nil")
	}
	if !appErrors.IsReason(err, appErrors.ReasonInternal) {
		t.Errorf("expected Internal reason (no longer masked as NotFound), got %v", err)
	}
	if repo.updateCalled != 0 {
		t.Errorf("expected NO Update call on infra failure, got %d", repo.updateCalled)
	}
}
