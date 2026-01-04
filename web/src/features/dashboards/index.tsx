/**
 * Dashboards Feature
 *
 * This module exports all public components, hooks, and types
 * for the dashboards feature.
 */

// Types
export type {
  WidgetType,
  WidgetViewType,
  FilterOperator,
  QueryFilter,
  TimeRange,
  WidgetQuery,
  Widget,
  LayoutItem,
  Variable,
  DashboardConfig,
  Dashboard,
  DashboardListItem,
  CreateDashboardRequest,
  UpdateDashboardRequest,
  DashboardListResponse,
  DashboardFilter,
  GetDashboardsParams,
  // Query execution types
  QueryExecutionParams,
  QueryMetadata,
  WidgetQueryResult,
  DashboardQueryResults,
  // View definition types
  MeasureDefinition,
  DimensionDefinition,
  ViewDefinition,
  ViewDefinitionsResponse,
  // Template types
  TemplateCategory,
  DashboardTemplate,
  CreateFromTemplateRequest,
  DuplicateDashboardRequest,
  // Export/Import types
  DashboardExport,
  DashboardImportRequest,
  // Grid layout types
  ReactGridLayoutItem,
} from './types'

// Hooks
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
  useProjectDashboards,
  // Widget query hooks
  widgetQueryKeys,
  useDashboardQueries,
  useWidgetQuery,
  useRefreshDashboardQueries,
  useViewDefinitions,
  // Template hooks
  templateQueryKeys,
  useTemplatesQuery,
  useTemplateQuery,
  useCreateFromTemplateMutation,
} from './hooks'

// Context
export { DashboardsProvider, useDashboards } from './context/dashboards-context'

// Components
export { Dashboards } from './components/dashboards-content'
export { DashboardList } from './components/dashboard-list'
export { DashboardCard } from './components/dashboard-card'
export { DashboardForm } from './components/dashboard-form'
export { CreateDashboardDialog } from './components/create-dashboard-dialog'
export { DashboardsDialogs } from './components/dashboards-dialogs'
export { DashboardDetail } from './components/dashboard-detail'
export { TemplateSelector } from './components/template-selector'
export { TimeRangePicker, getTimeRangeDates } from './components/time-range-picker'
export { AutoRefreshControl, useAutoRefresh } from './components/auto-refresh-control'

// API functions (for direct use when needed)
export {
  getDashboards,
  getDashboardById,
  createDashboard,
  updateDashboard,
  deleteDashboard,
  duplicateDashboard,
  lockDashboard,
  unlockDashboard,
  exportDashboard,
  importDashboard,
} from './api/dashboards-api'

// Template API functions
export {
  getTemplates,
  getTemplateById,
  createFromTemplate,
} from './api/templates-api'

// Widget query API functions
export {
  executeDashboardQueries,
  executeWidgetQuery,
  getViewDefinitions,
} from './api/widget-queries-api'

// Widget components
export {
  WidgetRenderer,
  StatWidget,
  TimeSeriesWidget,
  TableWidget,
  BarWidget,
  PieWidget,
  HeatmapWidget,
  HistogramWidget,
  TraceListWidget,
  TextWidget,
  type StatData,
  type TimeSeriesData,
  type TableData,
  type ColumnDefinition,
  type BarData,
  type PieData,
  type HeatmapData,
  type HistogramData,
  type HistogramStats,
  type TraceListData,
  type TraceListItem,
  type TextData,
  // Skeletons
  StatSkeleton,
  TimeSeriesSkeleton,
  BarSkeleton,
  PieSkeleton,
  TableSkeleton,
  HeatmapSkeleton,
  HistogramSkeleton,
  TraceListSkeleton,
  TextSkeleton,
  GenericWidgetSkeleton,
  getWidgetSkeleton,
  WidgetSkeletonRenderer,
} from './components/widgets'

// Query builder components
export {
  QueryBuilder,
  ViewSelector,
  MeasureSelector,
  DimensionSelector,
  QueryPreview,
} from './components/query-builder'

// Widget editing
export { WidgetForm } from './components/widget-form'
export { WidgetEditDialog, AddWidgetButton } from './components/widget-edit-dialog'

// Error handling
export { WidgetErrorBoundary, WidgetErrorFallback } from './components/widget-error-boundary'

// Grid layout
export { DashboardGrid } from './components/dashboard-grid'

// Widget palette
export { WidgetPalette, WIDGET_TYPES, type WidgetTypeDefinition } from './components/widget-palette'

// Dashboard editor
export { DashboardEditorToolbar } from './components/dashboard-editor-toolbar'
export { ImportDashboardDialog } from './components/import-dashboard-dialog'
