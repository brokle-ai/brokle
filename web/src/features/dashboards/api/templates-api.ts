/**
 * Dashboard Templates API Client
 *
 * API functions for dashboard template operations using BrokleAPIClient.
 */

import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  Dashboard,
  DashboardTemplate,
  CreateFromTemplateRequest,
} from '../types'

const client = new BrokleAPIClient('/api')

/**
 * List all available dashboard templates
 */
export const getTemplates = async (): Promise<DashboardTemplate[]> => {
  return client.get<DashboardTemplate[]>('/v1/dashboard-templates')
}

/**
 * Get a single template by ID
 */
export const getTemplateById = async (
  templateId: string
): Promise<DashboardTemplate> => {
  return client.get<DashboardTemplate>(`/v1/dashboard-templates/${templateId}`)
}

/**
 * Create a new dashboard from a template
 */
export const createFromTemplate = async (
  projectId: string,
  data: CreateFromTemplateRequest
): Promise<Dashboard> => {
  return client.post<Dashboard>(
    `/v1/projects/${projectId}/dashboards/from-template`,
    data
  )
}
