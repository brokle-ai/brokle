// Public exports for projects feature

// Hooks
export { useProjectOnly } from './hooks/use-project-only'
export { useCreateProjectMutation, useUpdateProjectMutation } from './hooks/use-project-queries'
export {
  useAPIKeysQuery,
  useCreateAPIKeyMutation,
  useDeleteAPIKeyMutation,
  apiKeyQueryKeys
} from './hooks/use-api-key-queries'

// API
export {
  listAPIKeys,
  createAPIKey,
  deleteAPIKey,
  createKeyPreview,
  validateAPIKeyFormat
} from './api/api-keys-api'
export { updateProject, archiveProject, unarchiveProject } from './api/projects-api'

// Components
export { CreateProjectDialog } from './components/create-project-dialog'
export { Overview } from './components/overview'
export { RecentSales } from './components/recent-sales'
export { ProjectSelector } from './components/project-selector'
export { DashboardView } from './components/dashboard-view'
export { ProjectGeneralSection } from './components/project-general-section'
export { ProjectAPIKeysSection } from './components/project-api-keys-section'
export { ProjectIntegrationsSection } from './components/project-integrations-section'
export { ProjectSecuritySection } from './components/project-security-section'
export { ProjectDangerSection } from './components/project-danger-section'

// Store
export { useDashboardStore } from './stores/dashboard-store'

// Types
export type {
  APIKey,
  CreateAPIKeyRequest,
  APIKeyFilters,
  APIKeyListResponse,
  APIKeyResponse
} from './types/api-keys'
export type { UpdateProjectRequest, Project } from './api/projects-api'
