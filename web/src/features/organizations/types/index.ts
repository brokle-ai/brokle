export interface Organization {
  id: string
  name: string
  slug?: string  // Computed from name + id using generateCompositeSlug()
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
  slug?: string  // Computed from name + id using generateCompositeSlug()
  organizationId: string
  description?: string
  status: ProjectStatus
  metrics: ProjectMetrics
  createdAt: string
  updatedAt: string
  settings?: ProjectSettings
}

export interface ProjectSummary {
  id: string
  name: string
  slug?: string // Legacy field
  compositeSlug: string
  description: string
  organizationId: string
  status: ProjectStatus
  metrics: ProjectMetrics
  settings?: ProjectSettings
  createdAt: string
  updatedAt: string
}

/**
 * Project metrics for observability dashboard
 * Note: These are observability metrics (traces observed), NOT gateway metrics (requests processed)
 */
export interface ProjectMetrics {
  // Observability metrics
  traces_collected: number
  traces_trend?: number
  observed_cost: number
  cost_trend?: number

  // Evaluation metrics
  active_rules: number
  running_experiments: number

  // Optional additional metrics
  scores_count?: number
  avg_score?: number

  // Legacy fields (deprecated - for backward compatibility during migration)
  /** @deprecated Use traces_collected instead */
  requests_today?: number
  /** @deprecated Use observed_cost instead */
  cost_today?: number
  /** @deprecated Not relevant for observability platform */
  avg_latency?: number
  /** @deprecated Not relevant for observability platform */
  error_rate?: number
  /** @deprecated Use traces_collected instead */
  total_requests?: number
  /** @deprecated Use observed_cost instead */
  total_cost?: number
}

/**
 * Organization-level aggregated stats
 * Used in the organization dashboard stats row
 */
export interface OrganizationStats {
  traces_collected: number
  traces_trend: number
  spans_analyzed: number
  scores_recorded: number
  observed_cost: number
  cost_trend: number
  projects_count: number
  members_count?: number
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

/**
 * Usage statistics for organization
 * Note: These track observed AI usage, not gateway requests
 */
export interface UsageStats {
  traces_this_month: number
  observed_cost_this_month: number
  models_observed: number
  last_trace?: string

  // Legacy fields (deprecated - for backward compatibility)
  /** @deprecated Use traces_this_month instead */
  requests_this_month?: number
  /** @deprecated Use observed_cost_this_month instead */
  cost_this_month?: number
  /** @deprecated Use models_observed instead */
  models_used?: number
  /** @deprecated Use last_trace instead */
  last_request?: string
}

export type SubscriptionPlan = 'free' | 'pro' | 'business' | 'enterprise'

export type OrganizationRole = 'owner' | 'admin' | 'developer' | 'viewer'

export type ProjectStatus = 'active' | 'archived'

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
  /** Reserved for future backend use (not currently persisted) */
  description?: string
}

export interface CreateProjectData {
  name: string
  organizationId: string
  /** Reserved for future use (optional, not required in UI) */
  description?: string
}

// Route parameter types for Next.js
// Note: Index signature required by Next.js 15 Params constraint
export interface OrganizationParams {
  orgSlug: string
  [key: string]: string | string[] | undefined
}

export interface ProjectParams {
  orgSlug: string
  projectSlug: string
  [key: string]: string | string[] | undefined
}