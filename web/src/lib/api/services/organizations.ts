// Organizations API - Latest endpoints for dashboard application
// Direct functions using optimal backend endpoints

import { BrokleAPIClient } from '../core/client'
import type { RequestOptions } from '../core/types'
import type { Organization, Project, OrganizationMember } from '@/types/organization'

// API response types matching backend
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

// Flexible base client - versions specified per endpoint
const client = new BrokleAPIClient('/api')

// Direct organization functions - latest & optimal endpoints
export const getUserOrganizations = async (): Promise<Organization[]> => {
    const response = await client.get<OrganizationAPIResponse[]>('/v2/organizations')
    return response.map(mapOrganizationFromAPI)
  }

export const resolveOrganizationSlug = async (slug: string): Promise<Organization> => {
    const response = await client.get<OrganizationAPIResponse[] | OrganizationAPIResponse>(
      `/organizations/slug/${slug}`
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
      orgData = response
    } else {
      console.warn('[OrganizationAPI] Invalid response type for slug:', slug, response)
      throw new Error(`Organization with slug '${slug}' not found`)
    }

    if (!orgData || !orgData.id) {
      console.error('[OrganizationAPI] Invalid organization data:', orgData)
      throw new Error(`Invalid organization data for slug '${slug}'`)
    }

    return mapOrganizationFromAPI(orgData)
  }

export const getOrganization = async (organizationId: string): Promise<Organization> => {
    const response = await client.get<OrganizationAPIResponse>(
      `/organizations/${organizationId}`,
      {},
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )

    return mapOrganizationFromAPI(response)
  }

export const getOrganizationMembers = async (organizationId: string): Promise<OrganizationMember[]> => {
    const response = await client.get<OrganizationMemberAPIResponse[]>(
      `/organizations/${organizationId}/members`,
      {},
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )

    return response.map(mapOrganizationMemberFromAPI)
  }

export const getOrganizationProjects = async (organizationId: string): Promise<Project[]> => {
    const response = await client.get<ProjectAPIResponse[]>(
      `/organizations/${organizationId}/projects`,
      {},
      { 
        // TODO: Re-enable when backend CORS is configured
        // includeOrgContext: true,
        // customOrgId: organizationId
      }
    )

    return response.map(mapProjectFromAPI)
  }

export const resolveProjectSlug = async (organizationId: string, slug: string): Promise<Project> => {
    const response = await client.get<ProjectAPIResponse>(
      `/organizations/${organizationId}/projects/slug/${slug}`,
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

    return mapProjectFromAPI(response)
  }

export const getProject = async (organizationId: string, projectId: string): Promise<Project> => {
    const response = await client.get<ProjectAPIResponse>(
      `/organizations/${organizationId}/projects/${projectId}`,
      {},
      { 
        includeOrgContext: true,
        includeProjectContext: true,
        customOrgId: organizationId,
        customProjectId: projectId
      }
    )

    return mapProjectFromAPI(response)
  }

export const getProjectMetrics = async (organizationId: string, projectId: string, environmentId?: string): Promise<ProjectMetricsAPIResponse> => {
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

    return await client.get<ProjectMetricsAPIResponse>(
      `/organizations/${organizationId}/projects/${projectId}/metrics`,
      {},
      options
    )
  }

export const createOrganization = async (data: {
    name: string
    slug?: string
    billing_email: string
    subscription_plan?: 'free' | 'pro' | 'business' | 'enterprise'
  }): Promise<Organization> => {
    const response = await client.post<OrganizationAPIResponse>('/v2/organizations', data)
    return mapOrganizationFromAPI(response)
  }

export const updateOrganization = async (organizationId: string, data: Partial<{
    name: string
    billing_email: string
    subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
  }>): Promise<Organization> => {
    const response = await client.patch<OrganizationAPIResponse>(
      `/organizations/${organizationId}`,
      data,
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )

    return mapOrganizationFromAPI(response)
  }

export const createProject = async (organizationId: string, data: {
    name: string
    slug?: string
    description?: string
  }): Promise<Project> => {
    const response = await client.post<ProjectAPIResponse>(
      `/organizations/${organizationId}/projects`,
      data,
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )

    return mapProjectFromAPI(response)
  }

export const updateProject = async (organizationId: string, projectId: string, data: Partial<{
    name: string
    description: string
  }>): Promise<Project> => {
    const response = await client.patch<ProjectAPIResponse>(
      `/organizations/${organizationId}/projects/${projectId}`,
      data,
      { 
        includeOrgContext: true,
        includeProjectContext: true,
        customOrgId: organizationId,
        customProjectId: projectId
      }
    )

    return mapProjectFromAPI(response)
  }

export const deleteProject = async (organizationId: string, projectId: string): Promise<void> => {
    await client.delete(
      `/organizations/${organizationId}/projects/${projectId}`,
      { 
        includeOrgContext: true,
        includeProjectContext: true,
        customOrgId: organizationId,
        customProjectId: projectId
      }
    )
  }

export const inviteUser = async (organizationId: string, email: string, role: 'admin' | 'developer' | 'viewer'): Promise<void> => {
    await client.post(
      `/organizations/${organizationId}/invitations`,
      { email, role },
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )
  }

export const removeUser = async (organizationId: string, userId: string): Promise<void> => {
    await client.delete(
      `/organizations/${organizationId}/members/${userId}`,
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )
  }

export const updateUserRole = async (organizationId: string, userId: string, role: 'admin' | 'developer' | 'viewer'): Promise<void> => {
    await client.patch(
      `/organizations/${organizationId}/members/${userId}`,
      { role },
      { 
        includeOrgContext: true,
        customOrgId: organizationId
      }
    )
  }

  // Private mapping functions
  const mapOrganizationFromAPI = (apiOrg: OrganizationAPIResponse): Organization => {
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

  const mapProjectFromAPI = (apiProject: ProjectAPIResponse): Project => {
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

  const mapOrganizationMemberFromAPI = (apiMember: OrganizationMemberAPIResponse): OrganizationMember => {
    return {
      id: apiMember.user_id,
      email: apiMember.email,
      role: apiMember.role,
      name: `${apiMember.first_name} ${apiMember.last_name}`.trim(),
      avatar: apiMember.avatar_url,
      joined_at: apiMember.joined_at,
    }
  }