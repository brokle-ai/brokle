/**
 * Dashboard Hooks
 *
 * This module exports all React hooks for the dashboards feature,
 * including query hooks and mutation hooks.
 */

export {
  dashboardQueryKeys,
  useDashboardsQuery,
  useDashboardQuery,
  useCreateDashboardMutation,
  useUpdateDashboardMutation,
  useDeleteDashboardMutation,
  useLockDashboardMutation,
  useUnlockDashboardMutation,
  useImportDashboardMutation,
} from './use-dashboards-queries'

export { useProjectDashboards } from './use-project-dashboards'

export {
  widgetQueryKeys,
  useDashboardQueries,
  useWidgetQuery,
  useRefreshDashboardQueries,
  useViewDefinitions,
} from './use-widget-queries'

export {
  templateQueryKeys,
  useTemplatesQuery,
  useTemplateQuery,
  useCreateFromTemplateMutation,
} from './use-templates'

export {
  variableQueryKeys,
  useDashboardVariables,
} from './use-dashboard-variables'

export { useDashboardTimeRange } from './use-dashboard-time-range'

export {
  useAutoSave,
  getAutoSaveStatusLabel,
  type AutoSaveStatus,
} from './use-auto-save'
