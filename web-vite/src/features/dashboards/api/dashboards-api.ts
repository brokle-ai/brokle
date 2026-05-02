// Dashboard CRUD API wrappers built on `rawFetch`. The web/ variant
// uses the gin-era BrokleAPIClient; web-vite's rawFetch already
// handles auth refresh + CSRF + typed error throwing. We keep the
// function surface identical so hooks don't change.

import { rawFetch } from '@/lib/api/client'
import type {
  Dashboard,
  DashboardListResponse,
  CreateDashboardRequest,
  UpdateDashboardRequest,
  DashboardFilter,
  DuplicateDashboardRequest,
  DashboardExport,
  DashboardImportRequest,
} from '../types'

function buildQuery(params: Record<string, string | number | boolean | undefined>): string {
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue
    search.set(key, String(value))
  }
  const qs = search.toString()
  return qs ? `?${qs}` : ''
}

async function requestJson<T>(input: string, init: RequestInit = {}): Promise<T> {
  const resp = await rawFetch(input, init)
  if (resp.status === 204) return undefined as T
  return (await resp.json()) as T
}

async function requestVoid(input: string, init: RequestInit = {}): Promise<void> {
  await rawFetch(input, init)
}

export const getDashboards = async (
  projectId: string,
  filter?: DashboardFilter,
): Promise<DashboardListResponse> => {
  const query = buildQuery({
    name: filter?.name,
    limit: filter?.limit,
    offset: filter?.offset,
  })
  return requestJson<DashboardListResponse>(
    `/api/v1/projects/${projectId}/dashboards${query}`,
    { method: 'GET' },
  )
}

export const getDashboardById = async (
  projectId: string,
  dashboardId: string,
): Promise<Dashboard> => {
  return requestJson<Dashboard>(
    `/api/v1/projects/${projectId}/dashboards/${encodeURIComponent(dashboardId)}`,
    { method: 'GET' },
  )
}

export const createDashboard = async (
  projectId: string,
  data: CreateDashboardRequest,
): Promise<Dashboard> => {
  return requestJson<Dashboard>(`/api/v1/projects/${projectId}/dashboards`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
}

export const updateDashboard = async (
  projectId: string,
  dashboardId: string,
  data: UpdateDashboardRequest,
): Promise<Dashboard> => {
  return requestJson<Dashboard>(
    `/api/v1/projects/${projectId}/dashboards/${encodeURIComponent(dashboardId)}`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
}

export const deleteDashboard = async (
  projectId: string,
  dashboardId: string,
): Promise<void> => {
  await requestVoid(
    `/api/v1/projects/${projectId}/dashboards/${encodeURIComponent(dashboardId)}`,
    { method: 'DELETE' },
  )
}

export const duplicateDashboard = async (
  projectId: string,
  dashboardId: string,
  data: DuplicateDashboardRequest,
): Promise<Dashboard> => {
  return requestJson<Dashboard>(
    `/api/v1/projects/${projectId}/dashboards/${encodeURIComponent(dashboardId)}/duplicate`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
}

export const lockDashboard = async (
  projectId: string,
  dashboardId: string,
): Promise<Dashboard> => {
  return requestJson<Dashboard>(
    `/api/v1/projects/${projectId}/dashboards/${encodeURIComponent(dashboardId)}/lock`,
    { method: 'POST' },
  )
}

export const unlockDashboard = async (
  projectId: string,
  dashboardId: string,
): Promise<Dashboard> => {
  return requestJson<Dashboard>(
    `/api/v1/projects/${projectId}/dashboards/${encodeURIComponent(dashboardId)}/unlock`,
    { method: 'POST' },
  )
}

export const exportDashboard = async (
  projectId: string,
  dashboardId: string,
): Promise<DashboardExport> => {
  return requestJson<DashboardExport>(
    `/api/v1/projects/${projectId}/dashboards/${encodeURIComponent(dashboardId)}/export`,
    { method: 'GET' },
  )
}

export const importDashboard = async (
  projectId: string,
  data: DashboardImportRequest,
): Promise<Dashboard> => {
  return requestJson<Dashboard>(
    `/api/v1/projects/${projectId}/dashboards/import`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
}
