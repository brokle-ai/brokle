import { BrokleAPIClient } from '@/lib/api/core/client'
import type { PaginatedResponse } from '@/lib/api/core/types'
import type {
  Dashboard,
  CreateDashboardRequest,
  UpdateDashboardRequest,
  DashboardFilter,
  DuplicateDashboardRequest,
  DashboardExport,
  DashboardImportRequest,
} from '../types'

const client = new BrokleAPIClient('/api')

export const getDashboards = async (
  projectId: string,
  filter?: DashboardFilter,
): Promise<PaginatedResponse<Dashboard>> => {
  // Backend mounts the list at /api/v1/projects/{projectId}/dashboards
  // and returns the canonical {data, pagination} envelope. Mirrors
  // getOrganizationMembers (members-api.ts).
  const queryParams: Record<string, string | number | boolean> = {}

  if (filter?.name) queryParams.name = filter.name
  if (filter?.limit) queryParams.limit = filter.limit
  if (filter?.offset !== undefined) queryParams.offset = filter.offset

  return client.getPaginated<Dashboard>(
    `/v1/projects/${projectId}/dashboards`,
    queryParams,
  )
}

export const getDashboardById = async (
  projectId: string,
  dashboardId: string
): Promise<Dashboard> => {
  return client.get<Dashboard>(
    `/v1/projects/${projectId}/dashboards/${dashboardId}`
  )
}

export const createDashboard = async (
  projectId: string,
  data: CreateDashboardRequest
): Promise<Dashboard> => {
  return client.post<Dashboard>(`/v1/projects/${projectId}/dashboards`, data)
}

export const updateDashboard = async (
  projectId: string,
  dashboardId: string,
  data: UpdateDashboardRequest
): Promise<Dashboard> => {
  return client.put<Dashboard>(
    `/v1/projects/${projectId}/dashboards/${dashboardId}`,
    data
  )
}

export const deleteDashboard = async (
  projectId: string,
  dashboardId: string
): Promise<void> => {
  await client.delete(`/v1/projects/${projectId}/dashboards/${dashboardId}`)
}

export const duplicateDashboard = async (
  projectId: string,
  dashboardId: string,
  data: DuplicateDashboardRequest
): Promise<Dashboard> => {
  return client.post<Dashboard>(
    `/v1/projects/${projectId}/dashboards/${dashboardId}/duplicate`,
    data
  )
}

export const lockDashboard = async (
  projectId: string,
  dashboardId: string
): Promise<Dashboard> => {
  return client.post<Dashboard>(
    `/v1/projects/${projectId}/dashboards/${dashboardId}/lock`
  )
}

export const unlockDashboard = async (
  projectId: string,
  dashboardId: string
): Promise<Dashboard> => {
  return client.post<Dashboard>(
    `/v1/projects/${projectId}/dashboards/${dashboardId}/unlock`
  )
}

export const exportDashboard = async (
  projectId: string,
  dashboardId: string
): Promise<DashboardExport> => {
  return client.get<DashboardExport>(
    `/v1/projects/${projectId}/dashboards/${dashboardId}/export`
  )
}

export const importDashboard = async (
  projectId: string,
  data: DashboardImportRequest
): Promise<Dashboard> => {
  return client.post<Dashboard>(
    `/v1/projects/${projectId}/dashboards/import`,
    data
  )
}
