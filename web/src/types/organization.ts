export interface Organization {
  id: string
  name: string
  plan: SubscriptionPlan
  billing_email?: string
  created_at: string
  updated_at: string
  members: OrganizationMember[]
  projects?: Project[]
  usage?: UsageStats
}

export interface OrganizationMember {
  id: string
  email: string
  role: OrganizationRole
  name: string
  avatar?: string
  joined_at: string
}

export interface Project {
  id: string
  name: string
  organizationId: string
  description?: string
  status: ProjectStatus
  environment: ProjectEnvironment
  metrics: ProjectMetrics
  created_at: string
  updated_at: string
  settings?: ProjectSettings
}

export interface ProjectMetrics {
  requests_today: number
  cost_today: number
  avg_latency: number
  error_rate: number
  total_requests?: number
  total_cost?: number
}

export interface ProjectSettings {
  default_provider: string
  enable_caching: boolean
  enable_analytics: boolean
  budget_limit?: number
  routing_preferences: RoutingPreferences
}

export interface RoutingPreferences {
  prioritize_latency: boolean
  prioritize_cost: boolean
  fallback_providers: string[]
}

export interface UsageStats {
  requests_this_month: number
  cost_this_month: number
  models_used: number
  last_request?: string
}

export type SubscriptionPlan = 'free' | 'pro' | 'business' | 'enterprise'

export type OrganizationRole = 'owner' | 'admin' | 'developer' | 'viewer'

export type ProjectStatus = 'active' | 'inactive' | 'archived'

export type ProjectEnvironment = 'development' | 'staging' | 'production'

// Context types for state management
export interface OrganizationContext {
  organizations: Organization[]
  currentOrganization: Organization | null
  currentProject: Project | null
  projects: Project[]
  isLoading: boolean
  error: string | null

  // Actions
  switchOrganization: (orgSlug: string) => Promise<void>
  switchProject: (projectSlug: string) => Promise<void>
  createOrganization: (data: CreateOrganizationData) => Promise<Organization>
  createProject: (data: CreateProjectData) => Promise<Project>

  // Utils
  // TODO: Remove deprecated access control functions - replaced with backend permissions
  // hasAccess: (orgSlug: string, projectSlug?: string) => boolean
  // getUserRole: (orgSlug: string) => OrganizationRole | null
  getProjectsByOrg: (orgSlug: string) => Project[]
}

export interface CreateOrganizationData {
  name: string
  description?: string
}

export interface CreateProjectData {
  name: string
  organizationId: string
  description?: string
  environment?: ProjectEnvironment
}

// Route parameter types for Next.js
export interface OrganizationParams {
  orgSlug: string
}

export interface ProjectParams {
  orgSlug: string
  projectSlug: string
}