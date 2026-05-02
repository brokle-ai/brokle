/**
 * Backend API Response Types
 *
 * These types match the exact structure returned by Go backend handlers.
 * They serve as the contract between frontend and backend API responses.
 */

/**
 * User profile data structure
 * Matches: internal/core/domain/user/user.go ProfileData
 */
export interface UserProfileData {
  bio: string | null
  location: string | null
  website: string | null
  twitter_url: string | null
  linkedin_url: string | null
  github_url: string | null
  timezone: string
  language: string
  theme: string
}

/**
 * Backend project summary response
 * Matches: internal/transport/http/handlers/user/types.go projectSummary
 *
 * `role` and `scopes` are part of the Langfuse-style session-bootstrap
 * payload — the user's OVERRIDE-resolved project role and resolved
 * project-tier permission set, computed server-side once per /users/me
 * fetch. The frontend's useHasProjectAccess reads `scopes` directly,
 * eliminating per-render /scopes/check round trips.
 */
export interface BackendProjectSummary {
  id: string
  name: string
  composite_slug: string
  description: string
  organization_id: string
  status: string
  role: string
  scopes: string[]
  created_at: string
  updated_at: string
}

/**
 * Backend organization with projects response
 * Matches: internal/transport/http/handlers/user/types.go organizationWithProjects
 *
 * `scopes` is the user's resolved org-tier permission set for this
 * organization. See BackendProjectSummary for rationale.
 */
export interface BackendOrganizationWithProjects {
  id: string
  name: string
  composite_slug: string
  plan: string
  role: string
  scopes: string[]
  created_at: string
  updated_at: string
  projects: BackendProjectSummary[]
}

/**
 * Enhanced user profile response from /v1/users/me endpoint
 * Matches: internal/transport/http/handlers/user/user.go EnhancedUserProfileResponse
 */
export interface EnhancedUserProfileResponse {
  id: string
  email: string
  name: string
  first_name: string
  last_name: string
  avatar_url: string
  is_email_verified: boolean
  is_active: boolean
  created_at: string
  updated_at: string
  last_login_at: string | null
  default_organization_id: string | null | undefined
  profile: UserProfileData | null
  completeness: number
  organizations: BackendOrganizationWithProjects[]
}
