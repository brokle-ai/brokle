// Re-export AuthState from store
export type { AuthState } from '../stores/auth-store'

// Import organization types for local use (enables usage in interfaces below)
import type {
  UsageStats,
  ProjectMetrics,
  ProjectStatus,
  ProjectSettings,
  RoutingPreferences,
  ProjectSummary
} from '@/features/organizations/types'

// Re-export organization types for external consumers
export type {
  UsageStats,
  ProjectMetrics,
  ProjectStatus,
  ProjectSettings,
  RoutingPreferences,
  ProjectSummary
}

export interface User {
  id: string
  email: string
  name?: string
  firstName?: string
  lastName?: string
  avatar?: string
  role: UserRole  // TODO: Backend compatibility only
  organizationId: string
  defaultOrganizationId?: string
  projects: string[]
  createdAt: string
  updatedAt: string
  lastLoginAt?: string
  isEmailVerified: boolean
  onboardingCompletedAt?: string | null
  organizations?: OrganizationWithProjects[]  // NEW: Organizations with nested projects

  // TODO: Add when implementing backend-integrated permission system
  // permissions?: Permission[]  // User's calculated permissions for current context
  // organizationMemberships?: Array<{
  //   organizationId: string
  //   role: OrganizationRole  // Backend compatibility
  //   permissions: Permission[]  // Calculated permissions for this org
  // }>
}

export interface Organization {
  id: string
  name: string
  plan: SubscriptionPlan
  members: OrganizationMember[]
  apiKeys: ApiKey[]
  usage: UsageStats
  createdAt: string
  updatedAt: string
}

export interface OrganizationMember {
  userId: string
  user: User
  role: OrganizationRole
  joinedAt: string
}

export interface Project {
  id: string
  name: string
  organizationId: string
  environment: ProjectEnvironment
  apiKeys: ApiKey[]
  settings: ProjectSettings
  createdAt: string
  updatedAt: string
}

export interface ApiKey {
  id: string
  name: string
  key: string
  permissions: Permission[]
  lastUsed?: string
  createdAt: string
  expiresAt?: string
}

// TODO: These role types are kept for backend compatibility
// Frontend should use Permission-based access control instead of role checking
export type UserRole = 'user' | 'admin' | 'super_admin'
export type OrganizationRole = 'owner' | 'admin' | 'developer' | 'viewer'

export type SubscriptionPlan = 'free' | 'pro' | 'business' | 'enterprise'

export type ProjectEnvironment = 'development' | 'staging' | 'production'

export type Permission = 
  | 'auth:read' 
  | 'auth:write' 
  | 'analytics:read' 
  | 'analytics:write'
  | 'models:read' 
  | 'models:write'
  | 'costs:read' 
  | 'costs:write'
  | 'settings:read' 
  | 'settings:write'
  | 'members:read'
  | 'members:write'
  | 'members:manage'
  | 'billing:read'
  | 'billing:write'
  | 'projects:read'
  | 'projects:write'
  | 'projects:create'
  | 'projects:delete'

// TODO: Utility types for future permission-based system
// export type PermissionCategory = 'auth' | 'analytics' | 'models' | 'costs' | 'settings' | 'members' | 'billing' | 'projects'
// export type PermissionAction = 'read' | 'write' | 'create' | 'delete' | 'manage'
// 
// export interface PermissionCheck {
//   required: Permission | Permission[]
//   requireAll?: boolean
// }
//
// export interface UserPermissions {
//   organizationId: string
//   permissions: Permission[]
//   lastUpdated: string
// }

// UsageStats removed - now imported from organizations (snake_case)

export interface AuthTokens {
  accessToken: string
  refreshToken: string
  expiresIn: number
  tokenType: 'Bearer'
}

export interface LoginCredentials {
  email: string
  password: string
  rememberMe?: boolean
}

export interface SignUpCredentials {
  email: string
  password: string
  firstName: string
  lastName: string
  role: string
  organizationName?: string
  referralSource?: string
  invitationToken?: string
}

export interface InvitationDetails {
  organizationName: string
  organizationId: string
  inviterName: string
  role: string
  email: string
  expiresAt: string
  isExpired: boolean
}

export interface AuthResponse {
  user: User
  organization: Organization
  expiresAt: number  // Unix timestamp in milliseconds (when token expires)
  expiresIn: number  // Duration in milliseconds (time until expiry)
}

export interface RefreshTokenRequest {
  refresh_token: string
}

// Raw API response types (snake_case)
export interface LoginResponse {
  access_token: string
  refresh_token: string
  user_id: string
  default_organization_id?: string
  expires_in: number
}

export interface UserResponse {
  id: string
  email: string
  first_name: string
  last_name: string
  is_email_verified: boolean
  onboarding_completed_at: string | null
  onboarding_step: number
  timezone: string
  language: string
  is_active: boolean
  login_count: number
  default_organization_id?: string
  created_at: string
  organizations?: OrganizationWithProjects[]
}

// ProjectSettings and RoutingPreferences removed - now imported from organizations (snake_case)

// Workspace context types for unified provider
export interface OrganizationWithProjects {
  id: string
  name: string
  slug?: string // Legacy support
  compositeSlug: string
  plan: SubscriptionPlan
  role: OrganizationRole
  billing_email?: string
  createdAt: string
  updatedAt: string
  projects: ProjectSummary[]
  members: OrganizationMember[]
  usage: UsageStats
}

// ProjectSummary removed - now imported from organizations (moved to canonical location)