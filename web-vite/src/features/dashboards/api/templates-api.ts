/**
 * Dashboard Templates API
 */

import { rawFetch } from '@/lib/api/client'
import type {
  Dashboard,
  DashboardTemplate,
  CreateFromTemplateRequest,
} from '../types'

async function requestJson<T>(
  input: string,
  init: RequestInit = {},
): Promise<T> {
  const resp = await rawFetch(input, init)
  return (await resp.json()) as T
}

export const getTemplates = async (): Promise<DashboardTemplate[]> => {
  return requestJson<DashboardTemplate[]>('/api/v1/dashboard-templates', {
    method: 'GET',
  })
}

export const getTemplateById = async (
  templateId: string,
): Promise<DashboardTemplate> => {
  return requestJson<DashboardTemplate>(
    `/api/v1/dashboard-templates/${encodeURIComponent(templateId)}`,
    { method: 'GET' },
  )
}

export const createFromTemplate = async (
  projectId: string,
  data: CreateFromTemplateRequest,
): Promise<Dashboard> => {
  return requestJson<Dashboard>(
    `/api/v1/projects/${projectId}/dashboards/from-template`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
}
