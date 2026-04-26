package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
)

// stubProjectMemberRepo is a minimal fake. Only the methods exercised by
// these tests return real values; everything else panics so a new caller
// shows up in test output instead of silently no-op'ing.
type stubProjectMemberRepo struct {
	effectivePerms []string
	effectiveErr   error
}

func (s *stubProjectMemberRepo) ListUserEffectivePermissionsInScope(ctx context.Context, userID, orgID, projectID uuid.UUID) ([]string, error) {
	return s.effectivePerms, s.effectiveErr
}

func (s *stubProjectMemberRepo) Create(context.Context, *authDomain.ProjectMember) error {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) GetByUserAndProject(context.Context, uuid.UUID, uuid.UUID) (*authDomain.ProjectMember, error) {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) UpdateRole(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) Delete(context.Context, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) ListByProject(context.Context, uuid.UUID) ([]*authDomain.ProjectMember, error) {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) ListByUser(context.Context, uuid.UUID) ([]*authDomain.ProjectMember, error) {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) IsMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) GetMemberCount(context.Context, uuid.UUID) (int, error) {
	panic("unimplemented")
}

// TestCheckUserPermissionsInScope_MapsResults exercises the result-shape
// contract: a permission appears in the result map iff the user holds it,
// even when the input list contains permissions the user does NOT hold
// (which must come back as false rather than missing keys).
func TestCheckUserPermissionsInScope_MapsResults(t *testing.T) {
	svc := NewProjectMemberService(
		&stubProjectMemberRepo{
			effectivePerms: []string{"traces:read", "scores:write"},
		},
		nil, // orgMemberRepo unused on this path
		nil, // roleRepo unused on this path
	)

	got, err := svc.CheckUserPermissionsInScope(
		context.Background(),
		uuid.New(), uuid.New(), uuid.New(),
		[]string{"traces:read", "scores:write", "projects:delete"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]bool{
		"traces:read":     true,
		"scores:write":    true,
		"projects:delete": false,
	}
	for perm, expected := range want {
		if got[perm] != expected {
			t.Errorf("permission %q: got %v, want %v", perm, got[perm], expected)
		}
	}
	if len(got) != len(want) {
		t.Errorf("expected %d entries, got %d (%v)", len(want), len(got), got)
	}
}

// TestCheckUserPermissionsInScope_EmptyRequestShortCircuits guards the
// no-op path: callers passing an empty permission slice should not
// trigger a DB lookup. We assert this by configuring the stub repo to
// return an error and checking it never gets called.
func TestCheckUserPermissionsInScope_EmptyRequestShortCircuits(t *testing.T) {
	svc := NewProjectMemberService(
		&stubProjectMemberRepo{effectiveErr: errors.New("repo must not be called")},
		nil, nil,
	)

	got, err := svc.CheckUserPermissionsInScope(
		context.Background(),
		uuid.New(), uuid.New(), uuid.New(),
		[]string{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

// TestCheckUserPermissionsInScope_RepoErrorPropagates ensures repo
// failures surface as 500 AppError (not silently dropped to false).
func TestCheckUserPermissionsInScope_RepoErrorPropagates(t *testing.T) {
	svc := NewProjectMemberService(
		&stubProjectMemberRepo{effectiveErr: errors.New("connection refused")},
		nil, nil,
	)

	_, err := svc.CheckUserPermissionsInScope(
		context.Background(),
		uuid.New(), uuid.New(), uuid.New(),
		[]string{"traces:read"},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
