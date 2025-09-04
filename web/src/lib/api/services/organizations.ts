import { BrokleAPIClient } from '../core/client'
import type { RequestOptions } from '../core/types'
import type { Organization, Project, OrganizationMember } from '@/types/organization'

/**
 * Organization API response types matching backend
 */
interface OrganizationAPIResponse {
  id: string
  name: string
  slug: string
  billing_email: string
  subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
  created_at: string
  updated_at: string
}

interface ProjectAPIResponse {
  id: string
  organization_id: string
  name: string
  slug: string
  description?: string
  environments: EnvironmentAPIResponse[]
  created_at: string
  updated_at: string
}

interface EnvironmentAPIResponse {
  id: string
  project_id: string
  name: string
  created_at: string
  updated_at: string
}

interface ProjectMetricsAPIResponse {
  requests_today: number
  cost_today: number
  avg_latency_ms: number
  error_rate: number
  total_requests: number
  total_cost: number
  last_request_at?: string
}

interface OrganizationMemberAPIResponse {
  user_id: string
  email: string
  first_name: string
  last_name: string
  role: 'owner' | 'admin' | 'developer' | 'viewer'
  joined_at: string
  avatar_url?: string
}

/**
 * Organization API Client
 * Handles all organization and project related API calls with context headers
 */
export class OrganizationAPIClient extends BrokleAPIClient {
  
  constructor() {
    super('/auth') // Organizations are managed through auth service
  }

  /**
   * Get user's organizations (no context headers needed - this gets the orgs)
   */
  async getUserOrganizations(): Promise<Organization[]> {
    const response = await this.get<OrganizationAPIResponse[]>(
      '/v1/organizations/user'
      // No headers needed - getting user's orgs
    )

    return response.map(this.mapOrganizationFromAPI)
  }

  /**
   * Resolve organization slug to organization data
   * Used for URL-based context resolution
   */
  async resolveOrganizationSlug(slug: string): Promise<Organization> {
    const response = await this.get<OrganizationAPIResponse[] | OrganizationAPIResponse>(
      `/v1/organizations/slug/${slug}`
      // No headers needed - slug resolution
    )

    console.debug('[OrganizationAPI] Slug resolution response:', { slug, response, isArray: Array.isArray(response) })

    let orgData: OrganizationAPIResponse

    // Handle both array and single object responses
    if (Array.isArray(response)) {
      if (response.length === 0) {
        console.warn('[OrganizationAPI] Empty array response for slug:', slug)
        throw new Error(`Organization with slug '${slug}' not found`)
      }
      orgData = response[0]
    } else if (response && typeof response === 'object') {
      // Single object response
      orgData = response
    } else {
      console.warn('[OrganizationAPI] Invalid response type for slug:', slug, response)
      throw new Error(`Organization with slug '${slug}' not found`)
    }

    if (!orgData || !orgData.id) {
      console.error('[OrganizationAPI] Invalid organization data:', orgData)
      throw new Error(`Invalid organization data for slug '${slug}'`)
    }

    return this.mapOrganizationFromAPI(orgData)
  }

  /**
   * Get single organization by ID (includes X-Org-ID header)
   */
  async getOrganization(organizationId: string): Promise<Organization> {
    const response = await this.get<OrganizationAPIResponse>(
      `/v1/organizations/${organizationId}`,
      {},
      { 
        includeOrgContext: true,
        customOrgId: organizationId // Override context with specific org ID
      }
    )

    return this.mapOrganizationFromAPI(response)
  }

  /**
   * Get organization members (requires X-Org-ID header)
   */
  async getOrganizationMembers(organizationId: string): Promise<OrganizationMember[]> {
    const response = await this.get<OrganizationMemberAPIResponse[]>(
      `/v1/organizations/${organizationId}/members`,
      {},
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )

    return response.map(this.mapOrganizationMemberFromAPI)
  }

  /**
   * Get organization projects (requires X-Org-ID header)
   * TEMPORARY: Headers disabled due to CORS configuration - backend needs X-Org-ID in Access-Control-Allow-Headers
   */
  async getOrganizationProjects(organizationId: string): Promise<Project[]> {
    const response = await this.get<ProjectAPIResponse[]>(
      `/v1/organizations/${organizationId}/projects`,
      {},
      { 
        // TODO: Re-enable when backend CORS is configured
        // includeOrgContext: true,
        // customOrgId: organizationId
      }
    )

    return response.map(this.mapProjectFromAPI)
  }

  /**
   * Resolve project slug to project data within an organization
   * Used for URL-based context resolution
   * TEMPORARY: Headers disabled due to CORS configuration - backend needs X-Org-ID in Access-Control-Allow-Headers
   */
  async resolveProjectSlug(organizationId: string, slug: string): Promise<Project> {
    const response = await this.get<ProjectAPIResponse>(
      `/v1/organizations/${organizationId}/projects/slug/${slug}`,
      {},
      { 
        // TODO: Re-enable when backend CORS is configured
        // includeOrgContext: true,
        // customOrgId: organizationId
      }
    )

    console.debug('[OrganizationAPI] Project slug resolution response:', { organizationId, slug, response })

    if (!response || !response.id) {
      console.error('[OrganizationAPI] Invalid project data:', response)
      throw new Error(`Invalid project data for slug '${slug}' in organization '${organizationId}'`)
    }

    return this.mapProjectFromAPI(response)
  }

  /**
   * Get single project by ID (requires X-Org-ID and X-Project-ID headers)
   */
  async getProject(organizationId: string, projectId: string): Promise<Project> {
    const response = await this.get<ProjectAPIResponse>(
      `/v1/organizations/${organizationId}/projects/${projectId}`,
      {},
      { 
        includeOrgContext: true,
        includeProjectContext: true,
        customOrgId: organizationId,
        customProjectId: projectId
      }
    )

    return this.mapProjectFromAPI(response)
  }

  /**
   * Get project metrics (requires X-Org-ID and X-Project-ID headers)
   */
  async getProjectMetrics(organizationId: string, projectId: string, environmentId?: string): Promise<ProjectMetricsAPIResponse> {
    const options: RequestOptions = {
      includeOrgContext: true,
      includeProjectContext: true,
      customOrgId: organizationId,
      customProjectId: projectId,
    }

    // Include environment context if provided
    if (environmentId) {
      options.includeEnvironmentContext = true
      options.customEnvironmentId = environmentId
    }

    return await this.get<ProjectMetricsAPIResponse>(
      `/v1/organizations/${organizationId}/projects/${projectId}/metrics`,
      {},
      options
    )
  }

  /**
   * Create new organization (no context headers needed)
   */
  async createOrganization(data: {
    name: string
    slug?: string
    billing_email: string
    subscription_plan?: 'free' | 'pro' | 'business' | 'enterprise'
  }): Promise<Organization> {
    const response = await this.post<OrganizationAPIResponse>(
      '/v1/organizations',
      data
      // No headers needed - creating new org
    )

    return this.mapOrganizationFromAPI(response)
  }

  /**
   * Update organization (requires X-Org-ID header)
   */
  async updateOrganization(organizationId: string, data: Partial<{
    name: string
    billing_email: string
    subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
  }>): Promise<Organization> {
    const response = await this.patch<OrganizationAPIResponse>(
      `/v1/organizations/${organizationId}`,
      data,
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )

    return this.mapOrganizationFromAPI(response)
  }

  /**
   * Create new project (requires X-Org-ID header)
   */
  async createProject(organizationId: string, data: {
    name: string
    slug?: string
    description?: string
  }): Promise<Project> {
    const response = await this.post<ProjectAPIResponse>(
      `/v1/organizations/${organizationId}/projects`,
      data,
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )

    return this.mapProjectFromAPI(response)
  }

  /**
   * Update project (requires X-Org-ID and X-Project-ID headers)
   */
  async updateProject(organizationId: string, projectId: string, data: Partial<{
    name: string
    description: string
  }>): Promise<Project> {
    const response = await this.patch<ProjectAPIResponse>(
      `/v1/organizations/${organizationId}/projects/${projectId}`,
      data,
      { 
        includeOrgContext: true,
        includeProjectContext: true,
        customOrgId: organizationId,
        customProjectId: projectId
      }
    )

    return this.mapProjectFromAPI(response)
  }

  /**
   * Delete project (requires X-Org-ID and X-Project-ID headers)
   */
  async deleteProject(organizationId: string, projectId: string): Promise<void> {
    await this.delete(
      `/v1/organizations/${organizationId}/projects/${projectId}`,
      { 
        includeOrgContext: true,
        includeProjectContext: true,
        customOrgId: organizationId,
        customProjectId: projectId
      }
    )
  }

  /**
   * Invite user to organization (requires X-Org-ID header)
   */
  async inviteUser(organizationId: string, email: string, role: 'admin' | 'developer' | 'viewer'): Promise<void> {
    await this.post(
      `/v1/organizations/${organizationId}/invitations`,
      { email, role },
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )
  }

  /**
   * Remove user from organization (requires X-Org-ID header)
   */
  async removeUser(organizationId: string, userId: string): Promise<void> {
    await this.delete(
      `/v1/organizations/${organizationId}/members/${userId}`,
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )
  }

  /**
   * Update user role in organization (requires X-Org-ID header)
   */
  async updateUserRole(organizationId: string, userId: string, role: 'admin' | 'developer' | 'viewer'): Promise<void> {
    await this.patch(
      `/v1/organizations/${organizationId}/members/${userId}`,
      { role },
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )
  }

  // Private mapping functions

  private mapOrganizationFromAPI(apiOrg: OrganizationAPIResponse): Organization {
    if (!apiOrg) {
      throw new Error('Organization API response is null or undefined')
    }
    
    if (!apiOrg.id) {
      console.error('[OrganizationAPI] Missing required fields in API response:', apiOrg)
      throw new Error('Organization API response missing required id field')
    }

    return {
      id: apiOrg.id,
      name: apiOrg.name || '',
      slug: apiOrg.slug || '',
      plan: apiOrg.subscription_plan || 'free',
      billing_email: apiOrg.billing_email || '',
      created_at: apiOrg.created_at || '',
      updated_at: apiOrg.updated_at || '',
      members: [], // Will be populated separately if needed
      usage: {
        requests_this_month: 0, // Will be populated from metrics API
        cost_this_month: 0,
        models_used: 0,
      },
    }
  }

  private mapProjectFromAPI(apiProject: ProjectAPIResponse): Project {
    if (!apiProject) {
      throw new Error('Project API response is null or undefined')
    }
    
    if (!apiProject.id) {
      console.error('[OrganizationAPI] Missing required fields in project API response:', apiProject)
      throw new Error('Project API response missing required id field')
    }

    return {
      id: apiProject.id,
      name: apiProject.name || '',
      slug: apiProject.slug || '',
      organizationId: apiProject.organization_id || '',
      description: apiProject.description || '',
      status: 'active', // Default status, will be determined by backend
      environment: 'development', // Will be determined by selected environment
      metrics: {
        requests_today: 0, // Will be populated from metrics API
        cost_today: 0,
        avg_latency: 0,
        error_rate: 0,
        total_requests: 0,
        total_cost: 0,
      },
      created_at: apiProject.created_at,
      updated_at: apiProject.updated_at,
      settings: {
        default_provider: 'openai', // Default settings, will be from backend
        enable_caching: true,
        enable_analytics: true,
        routing_preferences: {
          prioritize_latency: true,
          prioritize_cost: false,
          fallback_providers: ['anthropic'],
        },
      },
    }
  }

  private mapOrganizationMemberFromAPI(apiMember: OrganizationMemberAPIResponse): OrganizationMember {
    return {
      id: apiMember.user_id,
      email: apiMember.email,
      role: apiMember.role,
      name: `${apiMember.first_name} ${apiMember.last_name}`.trim(),
      avatar: apiMember.avatar_url,
      joined_at: apiMember.joined_at,
    }
  }
}