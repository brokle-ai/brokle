export interface User {
  id: string
  email: string
  name?: string
  firstName?: string
  lastName?: string
  avatar?: string
  role: UserRole
  organizationId: string
  defaultOrganizationId?: string
  projects: string[]
  createdAt: string
  updatedAt: string
  lastLoginAt?: string
  isEmailVerified: boolean
  onboardingCompleted?: boolean
}

export interface Organization {
  id: string
  name: string
  slug: string
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
  slug: string
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

export interface UsageStats {
  requestsThisMonth: number
  costsThisMonth: number
  modelsUsed: number
  lastRequest?: string
}

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
  organizationName?: string
}

export interface AuthResponse {
  user: User
  organization: Organization
  accessToken: string
  refreshToken: string
  expiresIn: number
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
  onboarding_completed: boolean
  onboarding_step: number
  timezone: string
  language: string
  is_active: boolean
  login_count: number
  default_organization_id?: string
  created_at: string
}

interface ProjectSettings {
  defaultProvider: string
  enableCaching: boolean
  enableAnalytics: boolean
  budgetLimit?: number
  routingPreferences: RoutingPreferences
}

interface RoutingPreferences {
  prioritizeLatency: boolean
  prioritizeCost: boolean
  fallbackProviders: string[]
}