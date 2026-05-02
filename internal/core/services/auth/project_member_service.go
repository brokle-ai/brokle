package auth

import (
	"context"
	"slices"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	appErrors "brokle/pkg/errors"
)

// Discovery-gate permission names. Referenced from the
// CanDiscoverProject predicate (this file) and the user-bootstrap
// handler (transport/http/handlers/user). Named consts so a YAML
// rename in seeds/permissions.yaml without updating the consts is
// caught by TestDiscoveryGatePermissionsExist before reaching prod.
const (
	PermOrgProjectsList = "org_projects:list"
	PermProjectsRead    = "projects:read"
)

// ProjectMemberService manages project-level role grants. A project
// membership is OPTIONAL — users access projects through their org-level
// role by default; a project_members row ADDS scopes on top of the org
// role's project-tier projection. Effective permissions are ADDITIVE
// (UNION of org-projection + override grant) — the AWS IAM / Cedar /
// OpenFGA / GitHub / GitLab convention. See
// CheckUserPermissionsInScope and docs/adr/0001-rbac-additive-
// semantics.md.
type ProjectMemberService struct {
	projMemberRepo authDomain.ProjectMemberRepository
	orgMemberRepo  authDomain.OrganizationMemberRepository
	roleRepo       authDomain.RoleRepository
}

// NewProjectMemberService constructs the service.
func NewProjectMemberService(
	projMemberRepo authDomain.ProjectMemberRepository,
	orgMemberRepo authDomain.OrganizationMemberRepository,
	roleRepo authDomain.RoleRepository,
) *ProjectMemberService {
	return &ProjectMemberService{
		projMemberRepo: projMemberRepo,
		orgMemberRepo:  orgMemberRepo,
		roleRepo:       roleRepo,
	}
}

// AddMember grants an existing org member an additional project-tier
// role on a specific project. **Operator-facing entry point** —
// rejects assignments that would have no effect on the user's
// effective scopes (proposed role's project-tier scopes ⊆ org-role
// projection) so the UI can't silently display an ineffective
// "demotion." For system-initiated creator grants written as part of
// project creation, use AddCreatorGrant — which intentionally bypasses
// this guard because the row's purpose is creator-retention, not
// permission elevation.
//
// The user MUST already be a member of the project's organization —
// project membership is an additive grant on top of the org role, not
// a way to grant access to users outside the org.
//
// See round-25 reviewer rationale + ADR-0001 §"What gets worse" #1
// for the redundant-grant guard's purpose.
func (s *ProjectMemberService) AddMember(ctx context.Context, userID, projectID, orgID, roleID uuid.UUID) (*authDomain.ProjectMember, error) {
	if _, err := s.checkAddMemberPreconditions(ctx, userID, orgID, projectID, roleID); err != nil {
		return nil, err
	}
	if err := s.rejectIfRedundantGrant(ctx, userID, orgID, roleID); err != nil {
		return nil, err
	}
	return s.persistProjectMember(ctx, userID, projectID, roleID)
}

// AddCreatorGrant writes the per-resource admin grant for the creator
// of a newly-created project. **System-initiated entry point** —
// called only by ProjectService.CreateProject, never from operator-
// facing surface. Skips the redundant-grant guard that AddMember
// applies because:
//
//  1. There's no operator confusion to prevent: the system writes the
//     row deterministically as part of project creation, regardless
//     of the creator's org role.
//  2. Under the additive resolver (round 24), a "redundant" row only
//     ever adds zero perms (UNION with a subset is the same set); it
//     cannot reduce access. Industry pattern (GitHub / GitLab /
//     Linear) writes this row for every creator regardless of
//     org-tier role.
//  3. Creator-retention across org-tier role CHANGES: a user later
//     demoted from owner → viewer keeps the per-resource admin grant
//     on projects they created (the row outlives org-tier role
//     changes). Org-tier REMOVAL still cascades the grant via
//     OrganizationMemberRepository.DeleteAllInOrgForUser (round-21
//     invariant) — demotion preserves access; removal does not.
//     ADR-0001 §3.
//
// Cross-tenant scope validation, org-membership precondition, and
// the UPSERT-WHERE concurrency contract all still run via the shared
// helpers below.
func (s *ProjectMemberService) AddCreatorGrant(ctx context.Context, userID, projectID, orgID, roleID uuid.UUID) (*authDomain.ProjectMember, error) {
	if _, err := s.checkAddMemberPreconditions(ctx, userID, orgID, projectID, roleID); err != nil {
		return nil, err
	}
	return s.persistProjectMember(ctx, userID, projectID, roleID)
}

// checkAddMemberPreconditions validates everything required before
// writing a project_members row, regardless of which entry point
// (operator AddMember vs system AddCreatorGrant) initiated the
// request:
//
//   - user is an active org member,
//   - role exists,
//   - role is assignable as a project membership (cross-tenant +
//     scope-type checks via validateProjectAssignableRole).
//
// Returns the resolved Role for callers that may need it (none today;
// kept in the signature for future composition).
func (s *ProjectMemberService) checkAddMemberPreconditions(ctx context.Context, userID, orgID, projectID, roleID uuid.UUID) (*authDomain.Role, error) {
	isOrgMember, err := s.orgMemberRepo.Exists(ctx, userID, orgID)
	if err != nil {
		return nil, appErrors.Internal("failed to check organization membership", err)
	}
	if !isOrgMember {
		return nil, appErrors.InvalidParam("user_id", "User is not a member of the project's organization",
			appErrors.WithDetails("user must join the organization before being assigned a project role"))
	}

	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if appErrors.IsNotFound(err) {
			return nil, appErrors.NotFound("role")
		}
		return nil, appErrors.Internal("failed to load role", err)
	}
	if err := s.validateProjectAssignableRole(ctx, role, projectID, orgID); err != nil {
		return nil, err
	}
	return role, nil
}

// persistProjectMember writes the project_members row via the
// UPSERT-WHERE shape and translates the rows-affected = 0 case into
// a Conflict pointing at UpdateMemberRole. Concurrency contract:
//
//   - Filtered IsMember pre-check rejects visible duplicates fast;
//     two admins who BOTH pass the pre-check race on the (user_id,
//     project_id) PK at INSERT time.
//   - Repository classifies unique-violation as AlreadyExists; the
//     rows-affected = 0 branch translates to Conflict.
//   - Hidden-orphan case takes the UPSERT UPDATE branch (rows = 1)
//     and is repaired atomically by the WHERE clause. See SQL query
//     comment + CLAUDE.md 2026-04-30 (project-rbac) for the rationale
//     against the prior UPSERT shape.
func (s *ProjectMemberService) persistProjectMember(ctx context.Context, userID, projectID, roleID uuid.UUID) (*authDomain.ProjectMember, error) {
	exists, err := s.projMemberRepo.IsMember(ctx, userID, projectID)
	if err != nil {
		return nil, appErrors.Internal("failed to check project membership", err)
	}
	if exists {
		return nil, appErrors.Conflict("project_member", "user already has a project-level role; use UpdateMemberRole instead")
	}

	member := authDomain.NewProjectMember(userID, projectID, roleID)
	rows, err := s.projMemberRepo.Create(ctx, member)
	if err != nil {
		return nil, appErrors.Internal("failed to create project membership", err)
	}
	if rows == 0 {
		return nil, appErrors.Conflict("project_member", "user already has a project-level role; use UpdateMemberRole instead")
	}
	return member, nil
}

// UpdateMemberRole changes the project-level role of an existing project
// member. Returns 404 if no project_members row exists for (user, project).
//
// Caller MUST pass the project's orgID — already pinned by
// RequireProjectAccess into httpctx.MustGetOrganizationID, so handlers
// read it directly and thread it through. Same strict scope validation
// as AddMember closes the cross-tenant role injection on the update
// path: an admin in org B cannot rotate an org-A project member's role
// to a custom org-B role. Same redundant-grant check as AddMember
// rejects updates that would no-op against the user's org-role
// projection (see rejectIfRedundantGrant).
func (s *ProjectMemberService) UpdateMemberRole(ctx context.Context, userID, projectID, orgID, roleID uuid.UUID) error {
	exists, err := s.projMemberRepo.IsMember(ctx, userID, projectID)
	if err != nil {
		return appErrors.Internal("failed to check project membership", err)
	}
	if !exists {
		return appErrors.NotFound("project member")
	}

	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if appErrors.IsNotFound(err) {
			return appErrors.NotFound("role")
		}
		return appErrors.Internal("failed to load role", err)
	}
	if err := s.validateProjectAssignableRole(ctx, role, projectID, orgID); err != nil {
		return err
	}
	if err := s.rejectIfRedundantGrant(ctx, userID, orgID, role.ID); err != nil {
		return err
	}

	if err := s.projMemberRepo.UpdateRole(ctx, userID, projectID, roleID); err != nil {
		return appErrors.Internal("failed to update project member role", err)
	}
	return nil
}

// rejectIfRedundantGrant returns 409 when the proposed role's project-
// tier scopes are a subset of the user's org-role projection. Under the
// additive resolver, a per-resource grant that's a subset of the org
// projection contributes zero new permissions — assignment would be a
// silent no-op (the most dangerous case being a "demotion" attempt
// where an admin tries to assign project viewer to themselves and the
// UI/API displays success while the user's effective access is
// unchanged).
//
// The five concrete cases this rejects (one error message for all):
//   - Proposed role equals user's org role.
//   - Proposed role's project-tier scopes ⊊ org-role projection
//     (the "demote" footgun: e.g. org admin → project viewer).
//   - Proposed role grants no project-tier perms (vacuously a subset).
//   - Proposed role's project-tier scopes equal user's org-role
//     projection but role IDs differ (e.g. duplicate templates).
//   - Built-in template assignment to a user who already has equal
//     or broader access via inheritance.
//
// Cases that proceed:
//   - Proposed role grants AT LEAST ONE project-tier scope the user
//     doesn't already inherit (genuine elevation).
//   - Mixed scope sets where proposed adds something new even if it
//     also overlaps.
func (s *ProjectMemberService) rejectIfRedundantGrant(ctx context.Context, userID, orgID, proposedRoleID uuid.UUID) error {
	orgMembership, err := s.orgMemberRepo.GetByUserAndOrganization(ctx, userID, orgID)
	if err != nil {
		// User isn't an org member — caller already validated this
		// upstream in AddMember; for UpdateMemberRole the membership
		// row exists by construction (project_members.IsMember was
		// true). Either way, treat repo errors as Internal.
		if appErrors.IsNotFound(err) {
			return appErrors.Internal("user organization membership not found during redundant-grant check", err)
		}
		return appErrors.Internal("failed to load organization membership for redundant-grant check", err)
	}

	orgRolePerms, err := s.roleRepo.GetRolePermissions(ctx, orgMembership.RoleID)
	if err != nil {
		return appErrors.Internal("failed to load org role permissions for redundant-grant check", err)
	}
	proposedRolePerms, err := s.roleRepo.GetRolePermissions(ctx, proposedRoleID)
	if err != nil {
		return appErrors.Internal("failed to load proposed role permissions for redundant-grant check", err)
	}

	orgProj := projectTierScopeSet(orgRolePerms)
	proposed := projectTierScopeSet(proposedRolePerms)
	for name := range proposed {
		if _, ok := orgProj[name]; !ok {
			return nil // proposed grants something new
		}
	}
	return appErrors.Conflict("project_member",
		"user already has equivalent or broader project access via their organization role; this assignment would have no effect")
}

// projectTierScopeSet projects a permission slice down to a set of
// `resource:action` names whose ScopeLevel is project. Used by the
// redundant-grant check to compare proposed-role and org-role
// project-tier projections set-wise.
func projectTierScopeSet(perms []*authDomain.Permission) map[string]struct{} {
	out := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		if p == nil {
			continue
		}
		if p.ScopeLevel != authDomain.ScopeLevelProject {
			continue
		}
		out[p.Name] = struct{}{}
	}
	return out
}

// validateProjectAssignableRole rejects roles that cannot legitimately be
// attached to a project_members row for the given (project, org). Three
// rejection classes:
//
//   - Role is system-scoped (e.g., a future platform-admin template) —
//     project_members must not carry platform-wide grants.
//   - Role is org-scoped to a DIFFERENT org (cross-tenant role injection).
//     Built-in templates have ScopeID == nil and are accepted regardless of
//     org (they are seed-managed templates available to every org).
//   - Role is project-scoped to a DIFFERENT project.
//
// Universal pattern: GitHub teams (org-bound roles), Snowflake (rejects
// cross-DB role grants), Microsoft Entra (directoryScopeId enforced).
//
// Historical note: a fourth rejection class — "role lacks `projects:read`
// floor" — protected against a self-lockout vector under the round-11
// OVERRIDE resolver, where the override REPLACED the org role's
// projection. The round-24 additive flip eliminates the lockout vector
// structurally (org-projection always unions in, regardless of override
// content), so the floor-scope assignment guard is no longer needed and
// has been dropped. ADR-0001 §"Decision" has the full rationale.
func (s *ProjectMemberService) validateProjectAssignableRole(ctx context.Context, role *authDomain.Role, projectID, orgID uuid.UUID) error {
	switch role.ScopeType {
	case authDomain.ScopeOrganization:
		if role.ScopeID != nil && *role.ScopeID != orgID {
			return appErrors.InvalidParam("role_id", "Role does not belong to this organization",
				appErrors.WithDetails("role_id is scoped to a different organization"),
			)
		}
	case authDomain.ScopeProject:
		if role.ScopeID == nil || *role.ScopeID != projectID {
			return appErrors.InvalidParam("role_id", "Role does not belong to this project",
				appErrors.WithDetails("role_id is scoped to a different project"),
			)
		}
	default: // ScopeSystem or unknown
		return appErrors.InvalidParam("role_id", "Invalid role scope",
			appErrors.WithDetails("role_id must be an organization-template or project role"),
		)
	}
	return nil
}


// RemoveMember deletes the project-level role grant; the user keeps
// whatever access their org-level role grants them via projection.
func (s *ProjectMemberService) RemoveMember(ctx context.Context, userID, projectID uuid.UUID) error {
	exists, err := s.projMemberRepo.IsMember(ctx, userID, projectID)
	if err != nil {
		return appErrors.Internal("failed to check project membership", err)
	}
	if !exists {
		return appErrors.NotFound("project member")
	}

	if err := s.projMemberRepo.Delete(ctx, userID, projectID); err != nil {
		return appErrors.Internal("failed to remove project membership", err)
	}
	return nil
}

// GetMember returns a single project membership row, or NotFound.
func (s *ProjectMemberService) GetMember(ctx context.Context, userID, projectID uuid.UUID) (*authDomain.ProjectMember, error) {
	m, err := s.projMemberRepo.GetByUserAndProject(ctx, userID, projectID)
	if err != nil {
		if appErrors.IsNotFound(err) {
			return nil, appErrors.NotFound("project member")
		}
		return nil, appErrors.Internal("failed to get project member", err)
	}
	return m, nil
}

// ListProjectMembers returns every project_members row for a project (the
// users with explicit role overrides). Org-level-only members do NOT appear
// here — they're surfaced via OrganizationMemberService.
func (s *ProjectMemberService) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*authDomain.ProjectMember, error) {
	members, err := s.projMemberRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, appErrors.Internal("failed to list project members", err)
	}
	return members, nil
}

// IsMember checks for a project_members row. Note: this does NOT mean
// "can the user access the project" — that's RequireProjectAccess
// (which checks org membership too).
func (s *ProjectMemberService) IsMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	return s.projMemberRepo.IsMember(ctx, userID, projectID)
}

// CheckUserPermissionsInScope resolves the user's effective permissions
// for (orgID, optional projectID) using ADDITIVE semantics (round 24):
//
//   - Active-org-membership precondition. The user must hold an active
//     `organization_members` row in `orgID`. A stale `project_members`
//     row alone grants ZERO permissions — closes the privilege-retention
//     bug after org removal.
//   - Org-scoped permissions: resolved against the org role only
//     (project membership is irrelevant for org-tier verbs).
//   - Project-scoped permissions: UNION of (org role's project-tier
//     projection) ∪ (project_members.role's project-tier scopes when
//     present). Per-resource grants only ADD, never reduce — matches
//     AWS IAM / Cedar / OpenFGA / GitHub / GitLab.
//   - Pass `uuid.Nil` as projectID to resolve org-only (no project layer).
//
// Returns map[permission]granted for each requested permission. See
// ListUserEffectivePermissionsInScope SQL for the resolver and ADR-0001
// for the migration rationale.
func (s *ProjectMemberService) CheckUserPermissionsInScope(
	ctx context.Context,
	userID, orgID, projectID uuid.UUID,
	permissions []string,
) (map[string]bool, error) {
	if len(permissions) == 0 {
		return map[string]bool{}, nil
	}
	granted, err := s.projMemberRepo.ListUserEffectivePermissionsInScope(ctx, userID, orgID, projectID)
	if err != nil {
		return nil, appErrors.Internal("failed to resolve effective permissions", err)
	}
	grantedSet := make(map[string]struct{}, len(granted))
	for _, p := range granted {
		grantedSet[p] = struct{}{}
	}
	result := make(map[string]bool, len(permissions))
	for _, p := range permissions {
		_, ok := grantedSet[p]
		result[p] = ok
	}
	return result, nil
}

// EffectiveOrgScopes carries a user's resolved scopes for one organization
// plus the resolved project-tier scopes for every project in that org.
// Returned by ListUserEffectiveScopes for the dashboard bootstrap path
// (GET /api/v1/users/me) — replaces N×M per-render /scopes/check round
// trips with one server-side aggregation. RoleName on each project is
// the most-specific role name: project_members.role if present, else
// empty string (caller displays the inherited org RoleName). Effective
// scopes are additively resolved server-side; this struct just carries
// the result.
type EffectiveOrgScopes struct {
	OrganizationID uuid.UUID
	Scopes         []string
	Projects       map[uuid.UUID]EffectiveProjectScopes
}

// CanDiscoverProject reports whether the user can SEE project P in
// listings ("discovery" — distinct from CONTENT access). Two ways to
// pass:
//
//   - Org-tier authority: user holds `org_projects:list` on this org →
//     every project in the org is discoverable (GitHub org-admin
//     shape). The org-tier role wins for discovery; project-tier
//     overrides still restrict CONTENT.
//   - Per-project floor: user holds `projects:read` on P → that
//     specific project is discoverable (GitHub repo-collaborator
//     shape). Auto-injected by RoleService.autoInjectFloorScope when
//     any project-tier permission is granted, so this is also the
//     canonical "user has any access to this project" predicate.
//
// Mirrors the GitHub / GitLab / Linear / Vercel / PostHog filter-at-
// discovery pattern (5 of 7 surveyed peers). Outline / Langfuse use
// include-and-flag, which doesn't fit Brokle's per-tier-verb-minting
// model where discovery and content are separate permissions.
func (o *EffectiveOrgScopes) CanDiscoverProject(projectID uuid.UUID) bool {
	if o == nil {
		return false
	}
	if slices.Contains(o.Scopes, PermOrgProjectsList) {
		return true
	}
	proj, ok := o.Projects[projectID]
	if !ok {
		return false
	}
	return slices.Contains(proj.Scopes, PermProjectsRead)
}

// EffectiveProjectScopes carries a user's additively-resolved project-
// tier scopes plus the most-specific role name for one project.
type EffectiveProjectScopes struct {
	ProjectID uuid.UUID
	// RoleName is empty when the user has no project_members row for
	// this project — callers display the inherited org role in that
	// case. Non-empty when a project_members grant row exists. Note:
	// under the round-24 additive resolver this is only a display hint;
	// effective scopes are the UNION of org-projection + the
	// per-resource grant.
	RoleName string
	Scopes   []string
}

// ListUserEffectiveScopes resolves the user's full org+project scope tree
// in one server-side pass — Langfuse session-bootstrap pattern (see
// competitors/langfuse/web/src/server/auth.ts:851-895). The dashboard
// calls this once on mount via GET /api/v1/users/me and stores the
// result client-side; UI permission checks are then synchronous
// in-memory lookups (zero per-check round trip).
//
// `topology` is the user's known (org → []project) shape, supplied by
// the caller (typically OrganizationService.GetUserOrganizationsWithProjects).
// We delegate topology discovery to the org domain rather than
// re-deriving it here so this method stays a pure scope-resolver and
// honours the bounded-context split.
//
// For each org we fan out the canonical resolver
// (ListUserEffectivePermissionsInScope) once with projectID=uuid.Nil
// (org-tier) plus once per project. For typical bootstrap sizes
// (1-3 orgs × 5-20 projects each ≈ 50 calls at ~3 ms p99) this finishes
// well under the 200 ms p99 dashboard-mount budget. If profiling later
// shows it as hot, the SQL CTE can be folded into one CROSS-JOINed
// query — same semantics, fewer round trips.
//
// Returns a map keyed by orgID for O(1) consumer lookups during DTO
// assembly.
func (s *ProjectMemberService) ListUserEffectiveScopes(
	ctx context.Context,
	userID uuid.UUID,
	topology map[uuid.UUID][]uuid.UUID,
) (map[uuid.UUID]*EffectiveOrgScopes, error) {
	if len(topology) == 0 {
		return map[uuid.UUID]*EffectiveOrgScopes{}, nil
	}

	// Resolve per-project grant role names ONCE for the whole user via
	// the existing ListByUser path. Avoids N per-project lookups.
	grants, err := s.projMemberRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, appErrors.Internal("failed to list project memberships", err)
	}
	grantedRoleByProject := make(map[uuid.UUID]uuid.UUID, len(grants))
	for _, m := range grants {
		grantedRoleByProject[m.ProjectID] = m.RoleID
	}
	roleNameCache := make(map[uuid.UUID]string, len(grants))
	resolveRoleName := func(roleID uuid.UUID) (string, error) {
		if name, ok := roleNameCache[roleID]; ok {
			return name, nil
		}
		role, err := s.roleRepo.GetByID(ctx, roleID)
		if err != nil {
			return "", err
		}
		roleNameCache[roleID] = role.Name
		return role.Name, nil
	}

	out := make(map[uuid.UUID]*EffectiveOrgScopes, len(topology))
	for orgID, projectIDs := range topology {
		orgScopes, err := s.projMemberRepo.ListUserEffectivePermissionsInScope(ctx, userID, orgID, uuid.Nil)
		if err != nil {
			return nil, appErrors.Internal("failed to resolve org-tier permissions", err)
		}
		entry := &EffectiveOrgScopes{
			OrganizationID: orgID,
			Scopes:         orgScopes,
			Projects:       make(map[uuid.UUID]EffectiveProjectScopes, len(projectIDs)),
		}
		for _, projectID := range projectIDs {
			projScopes, err := s.projMemberRepo.ListUserEffectivePermissionsInScope(ctx, userID, orgID, projectID)
			if err != nil {
				return nil, appErrors.Internal("failed to resolve project-tier permissions", err)
			}
			projectEntry := EffectiveProjectScopes{
				ProjectID: projectID,
				Scopes:    projScopes,
			}
			if roleID, hasGrant := grantedRoleByProject[projectID]; hasGrant {
				name, err := resolveRoleName(roleID)
				if err != nil {
					return nil, appErrors.Internal("failed to load granted role", err)
				}
				projectEntry.RoleName = name
			}
			entry.Projects[projectID] = projectEntry
		}
		out[orgID] = entry
	}
	return out, nil
}
