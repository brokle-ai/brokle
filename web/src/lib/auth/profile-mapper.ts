// Pure mapping from backend EnhancedUserProfileResponse (snake_case)
// to the frontend domain shape (User + OrganizationWithProjects[]).
//
// Lifted out of context/workspace-context.tsx so both the server-side
// DAL (lib/auth/dal.ts) and the client-side WorkspaceProvider
// useQuery refetch path can reuse it without duplicating the
// snake_case ↔ camelCase translation. No runtime dependencies, no
// React, no axios — pure data transform.

import type {
  User,
  OrganizationWithProjects,
  SubscriptionPlan,
  OrganizationRole,
  ProjectStatus,
  OrganizationMember,
} from '@/features/authentication'
import type {
  EnhancedUserProfileResponse,
  BackendOrganizationWithProjects,
  BackendProjectSummary,
} from '@/types/api-responses'

export interface MappedProfile {
  user: User
  organizations: OrganizationWithProjects[]
}

export function mapEnhancedUserProfile(
  response: EnhancedUserProfileResponse,
): MappedProfile {
  const user: User = {
    id: response.id,
    email: response.email,
    firstName: response.first_name,
    lastName: response.last_name,
    name: `${response.first_name} ${response.last_name}`.trim(),
    role: 'user',
    organizationId: '',
    defaultOrganizationId: response.default_organization_id ?? undefined,
    projects: [],
    createdAt: response.created_at,
    updatedAt: response.updated_at,
    isEmailVerified: response.is_email_verified,
    organizations: [], // Populated below.
  }

  const organizations: OrganizationWithProjects[] = (response.organizations || []).map(
    (org: BackendOrganizationWithProjects) => ({
      id: org.id,
      name: org.name,
      compositeSlug: org.composite_slug,
      plan: org.plan as SubscriptionPlan,
      role: org.role as OrganizationRole,
      // Backend always emits a non-null array; coalesce defensively in
      // case an older endpoint shape leaks through during the cutover.
      scopes: org.scopes ?? [],
      createdAt: org.created_at,
      updatedAt: org.updated_at,
      projects: (org.projects || []).map((proj: BackendProjectSummary) => ({
        id: proj.id,
        name: proj.name,
        compositeSlug: proj.composite_slug,
        description: proj.description || '',
        organizationId: proj.organization_id,
        status: proj.status as ProjectStatus,
        role: proj.role,
        scopes: proj.scopes ?? [],
        createdAt: proj.created_at,
        updatedAt: proj.updated_at,
        metrics: {
          traces_collected: 0,
          observed_cost: 0,
          active_rules: 0,
          running_experiments: 0,
        },
      })),
      members: [] as OrganizationMember[], // Populated lazily by member endpoints.
      usage: {
        traces_this_month: 0,
        observed_cost_this_month: 0,
        models_observed: 0,
      },
    }),
  )

  user.organizations = organizations
  return { user, organizations }
}
