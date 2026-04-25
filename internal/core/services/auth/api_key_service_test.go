package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	orgDomain "brokle/internal/core/domain/organization"
	appErrors "brokle/pkg/errors"
)

// stubAPIKeyRepo is a minimal fake that only exercises the error paths
// under test. Unused methods panic so a new caller added later is visible
// in the test output rather than silently returning zero values.
type stubAPIKeyRepo struct {
	getByIDErr    error
	createErr     error
	getByHashErr  error
}

func (s *stubAPIKeyRepo) Create(ctx context.Context, _ *authDomain.APIKey) error {
	return s.createErr
}
func (s *stubAPIKeyRepo) GetByID(ctx context.Context, _ uuid.UUID) (*authDomain.APIKey, error) {
	return nil, s.getByIDErr
}
func (s *stubAPIKeyRepo) GetByKeyHash(ctx context.Context, _ string) (*authDomain.APIKey, error) {
	return nil, s.getByHashErr
}
func (s *stubAPIKeyRepo) Update(context.Context, *authDomain.APIKey) error { panic("unimplemented") }
func (s *stubAPIKeyRepo) Delete(context.Context, uuid.UUID) error          { panic("unimplemented") }
func (s *stubAPIKeyRepo) GetByUserID(context.Context, uuid.UUID) ([]*authDomain.APIKey, error) {
	panic("unimplemented")
}
func (s *stubAPIKeyRepo) GetByOrganizationID(context.Context, uuid.UUID) ([]*authDomain.APIKey, error) {
	panic("unimplemented")
}
func (s *stubAPIKeyRepo) GetByProjectID(context.Context, uuid.UUID) ([]*authDomain.APIKey, error) {
	panic("unimplemented")
}
func (s *stubAPIKeyRepo) UpdateLastUsed(context.Context, uuid.UUID) error  { panic("unimplemented") }
func (s *stubAPIKeyRepo) CleanupExpiredAPIKeys(context.Context) error      { panic("unimplemented") }
func (s *stubAPIKeyRepo) GetByFilters(context.Context, *authDomain.APIKeyFilters) ([]*authDomain.APIKey, error) {
	panic("unimplemented")
}
func (s *stubAPIKeyRepo) CountByFilters(context.Context, *authDomain.APIKeyFilters) (int64, error) {
	panic("unimplemented")
}
func (s *stubAPIKeyRepo) GetAPIKeyCount(context.Context, uuid.UUID) (int, error) {
	panic("unimplemented")
}
func (s *stubAPIKeyRepo) GetActiveAPIKeyCount(context.Context, uuid.UUID) (int, error) {
	panic("unimplemented")
}

// noopLogger keeps test output quiet without letting nil-deref panics hide bugs.
func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newServiceUnderTest(repo *stubAPIKeyRepo) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: repo,
		logger:     noopLogger(),
	}
}

// TestGetAPIKey_WrapsErrAPIKeyNotFound_To404AppError locks in the fix for P2:
// when the repository wraps a missing row with ErrAPIKeyNotFound, the service
// must translate it into an AppError of TypeNotFound (HTTP 404).
//
// Before the fix the service checked ErrAPIKeyNotFound but the repo emitted
// ErrNotFound, so the not-found branch was dead and callers saw 500.
func TestGetAPIKey_WrapsErrAPIKeyNotFound_To404AppError(t *testing.T) {
	repo := &stubAPIKeyRepo{
		// Mirror the exact wrap shape used by api_key_repository.go:GetByID.
		getByIDErr: fmt.Errorf("get api_key by ID xxx: %w", authDomain.ErrAPIKeyNotFound),
	}
	svc := newServiceUnderTest(repo)

	_, err := svc.GetAPIKey(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	appErr, ok := appErrors.IsAppError(err)
	if !ok {
		t.Fatalf("expected *AppError, got %T: %v", err, err)
	}
	if appErr.Type != appErrors.TypeNotFound {
		t.Errorf("expected Type=%q, got %q", appErrors.TypeNotFound, appErr.Type)
	}
}

// TestGetAPIKey_WrapsOtherError_ToInternalError ensures non-sentinel repo
// errors still take the internal-error branch.
func TestGetAPIKey_WrapsOtherError_ToInternalError(t *testing.T) {
	repo := &stubAPIKeyRepo{
		getByIDErr: errors.New("connection refused"),
	}
	svc := newServiceUnderTest(repo)

	_, err := svc.GetAPIKey(context.Background(), uuid.New())
	appErr, ok := appErrors.IsAppError(err)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Type != appErrors.TypeAPIError {
		t.Errorf("expected Type=%q, got %q", appErrors.TypeAPIError, appErr.Type)
	}
}

// TestGetAPIKeyContext_NotFound_Returns404 covers the second buggy site from P2.
func TestGetAPIKeyContext_NotFound_Returns404(t *testing.T) {
	repo := &stubAPIKeyRepo{
		getByIDErr: fmt.Errorf("get api_key by ID xxx: %w", authDomain.ErrAPIKeyNotFound),
	}
	svc := newServiceUnderTest(repo)

	_, err := svc.GetAPIKeyContext(context.Background(), uuid.New())
	appErr, ok := appErrors.IsAppError(err)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Type != appErrors.TypeNotFound {
		t.Errorf("expected Type=%q, got %q", appErrors.TypeNotFound, appErr.Type)
	}
}

// TestCanAPIKeyAccessResource_NotFound_Returns404 covers the third buggy site.
func TestCanAPIKeyAccessResource_NotFound_Returns404(t *testing.T) {
	repo := &stubAPIKeyRepo{
		getByIDErr: fmt.Errorf("get api_key by ID xxx: %w", authDomain.ErrAPIKeyNotFound),
	}
	svc := newServiceUnderTest(repo)

	_, err := svc.CanAPIKeyAccessResource(context.Background(), uuid.New(), "some:resource")
	appErr, ok := appErrors.IsAppError(err)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Type != appErrors.TypeNotFound {
		t.Errorf("expected Type=%q, got %q", appErrors.TypeNotFound, appErr.Type)
	}
}

// TestValidateAPIKey_NotFound_Returns401 locks in the security-shaped behaviour
// after the sentinel swap: a valid-shape key that doesn't exist in the DB must
// still return Unauthorized (not Forbidden, not NotFound — we never leak
// existence). The sentinel the repo wraps with changed in the fix; this test
// guards against a regression in the status code the handler surfaces.
func TestValidateAPIKey_NotFound_Returns401(t *testing.T) {
	// Real 40-char format: bk_ + 40 chars. Any valid-format key the fake
	// repo reports missing must route to Unauthorized.
	validFormatKey := "bk_" + strings.Repeat("a", 40)

	repo := &stubAPIKeyRepo{
		getByHashErr: fmt.Errorf("get api_key by hash: %w", authDomain.ErrAPIKeyNotFound),
	}
	svc := newServiceUnderTest(repo)

	_, err := svc.ValidateAPIKey(context.Background(), validFormatKey)
	appErr, ok := appErrors.IsAppError(err)
	if !ok {
		t.Fatalf("expected *AppError, got %T: %v", err, err)
	}
	if appErr.Type != appErrors.TypeAuthentication {
		t.Errorf("expected Type=%q (401), got %q", appErrors.TypeAuthentication, appErr.Type)
	}
}

// TestCreateAPIKey_HashCollision_ReturnsAccurateConflict locks in the P3 fix:
// when the repository reports ErrAPIKeyAlreadyExists (a key_hash UNIQUE
// violation), the service must return a Conflict whose message points callers
// at the real cause (regenerate the key), not at the non-existent name
// constraint.
func TestCreateAPIKey_HashCollision_ReturnsAccurateConflict(t *testing.T) {
	repo := &stubAPIKeyRepo{
		createErr: fmt.Errorf("create api_key: %w", authDomain.ErrAPIKeyAlreadyExists),
	}
	svc := newServiceUnderTest(repo)

	_, err := svc.CreateAPIKey(
		context.Background(),
		uuid.New(),
		&authDomain.CreateAPIKeyRequest{
			Name:      "whatever",
			ProjectID: uuid.New(),
		},
	)

	appErr, ok := appErrors.IsAppError(err)
	if !ok {
		t.Fatalf("expected *AppError, got %T: %v", err, err)
	}
	if appErr.Type != appErrors.TypeConflict {
		t.Errorf("expected Type=%q, got %q", appErrors.TypeConflict, appErr.Type)
	}
	// The pre-fix message used "name" wording. Guard against regression.
	if strings.Contains(strings.ToLower(appErr.Message), "name") {
		t.Errorf("message should not blame the name constraint (no name UNIQUE exists): %q", appErr.Message)
	}
	// Positive check: message should invite a retry, which is the actual fix path.
	if !strings.Contains(strings.ToLower(appErr.Message), "retry") {
		t.Errorf("message should direct callers to retry; got %q", appErr.Message)
	}
}

// Silence unused-import vet complaint if a future edit removes all uses.
var _ = orgDomain.AcceptInvitationResult{}
