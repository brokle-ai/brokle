'use client'

import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { MoreVertical, RefreshCw, Trash } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { PageHeader } from '@/components/layout/page-header'
import { LoadingSpinner } from '@/components/guards/loading-spinner'
import {
  useDashboardQuery,
  useDeleteDashboardMutation,
  useUpdateDashboardMutation,
  useLockDashboardMutation,
  useUnlockDashboardMutation,
  useImportDashboardMutation,
} from '../hooks/use-dashboards-queries'
import { useDashboardQueries, useRefreshDashboardQueries } from '../hooks/use-widget-queries'
import { useDashboardVariables } from '../hooks/use-dashboard-variables'
import { useDashboardTimeRange } from '../hooks/use-dashboard-time-range'
import { useAutoSave } from '../hooks/use-auto-save'
import { exportDashboard } from '../api/dashboards-api'
import { DashboardGrid } from './dashboard-grid'
import { DashboardEditorToolbar } from './dashboard-editor-toolbar'
import { ImportDashboardDialog } from './import-dashboard-dialog'
import { VariablesBar } from './variables-bar'
import { WidgetEditDialog } from './widget-edit-dialog'
import { TimeRangePicker } from './time-range-picker'
import { useProjectOnly } from '@/features/projects/hooks/use-project-only'
import type { LayoutItem, WidgetType, Widget, DashboardImportRequest } from '../types'

interface DashboardDetailProps {
  projectSlug: string
  dashboardId: string
}

export function DashboardDetail({ projectSlug, dashboardId }: DashboardDetailProps) {
  const router = useRouter()
  const { currentProject } = useProjectOnly()
  const projectId = currentProject?.id ?? ''

  // Edit mode state
  const [isEditMode, setIsEditMode] = useState(false)
  const [pendingLayout, setPendingLayout] = useState<LayoutItem[] | null>(null)
  const [pendingWidgets, setPendingWidgets] = useState<Widget[] | null>(null)
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false)

  // Widget edit dialog state
  const [editingWidget, setEditingWidget] = useState<Widget | null>(null)
  const [isWidgetDialogOpen, setIsWidgetDialogOpen] = useState(false)
  const [pendingWidgetCreation, setPendingWidgetCreation] = useState<{
    type: WidgetType
    defaultSize: { w: number; h: number }
  } | null>(null)

  // Queries
  const { data: dashboard, isLoading, error } = useDashboardQuery(projectId, dashboardId)

  // Time range with URL state persistence
  const {
    timeRange,
    setTimeRange,
    resetTimeRange,
    isCustomized: isTimeRangeCustomized,
  } = useDashboardTimeRange(dashboard?.config?.time_range)

  // Dashboard variables
  const {
    values: variableValues,
    setValue: setVariableValue,
    resetToDefaults: resetVariables,
    hasActiveVariables,
    getVariableOptions,
    isLoadingOptions: isLoadingVariableOptions,
  } = useDashboardVariables(projectId, dashboard?.config?.variables)

  // Mutations
  const deleteMutation = useDeleteDashboardMutation(projectId)
  const updateMutation = useUpdateDashboardMutation(projectId, dashboardId)
  const lockMutation = useLockDashboardMutation(projectId)
  const unlockMutation = useUnlockDashboardMutation(projectId)
  const importMutation = useImportDashboardMutation(projectId)

  // Execute widget queries with URL-persisted time range, variables, and optional auto-refresh
  const {
    data: queryResults,
    isLoading: isQueryLoading,
    isFetching: isQueryFetching,
  } = useDashboardQueries(projectId, dashboardId, {
    enabled: !!dashboard,
    timeRange,
    variableValues,
    refetchInterval: dashboard?.config?.refresh_rate
      ? dashboard.config.refresh_rate * 1000
      : 0,
  })

  const refreshMutation = useRefreshDashboardQueries(projectId, dashboardId)

  // Derived state
  const widgets = pendingWidgets ?? dashboard?.config?.widgets ?? []
  const layout = pendingLayout ?? dashboard?.layout ?? []
  const hasChanges = pendingLayout !== null || pendingWidgets !== null
  const isLocked = dashboard?.is_locked ?? false

  // Auto-save (only active in edit mode)
  const {
    status: autoSaveStatus,
    error: autoSaveError,
    scheduleSave,
    cancelPendingSave,
  } = useAutoSave(projectId, dashboardId, dashboard, {
    enabled: isEditMode && !isLocked,
    debounceMs: 1500, // 1.5 second debounce
    onSaveSuccess: () => {
      // Clear pending state after successful auto-save
      setPendingLayout(null)
      setPendingWidgets(null)
    },
  })

  // Handlers
  const handleRefresh = useCallback(() => {
    refreshMutation.mutate({ time_range: timeRange })
  }, [refreshMutation, timeRange])

  const handleDelete = useCallback(async () => {
    if (!dashboard) return
    await deleteMutation.mutateAsync({
      dashboardId: dashboard.id,
      dashboardName: dashboard.name,
    })
    router.push(`/projects/${projectSlug}/dashboards`)
  }, [dashboard, deleteMutation, router, projectSlug])

  const handleEditModeToggle = useCallback(() => {
    if (isEditMode) {
      // Exiting edit mode - discard changes
      setPendingLayout(null)
      setPendingWidgets(null)
    }
    setIsEditMode(!isEditMode)
  }, [isEditMode])

  const handleLayoutChange = useCallback(
    (newLayout: LayoutItem[]) => {
      setPendingLayout(newLayout)
      // Schedule auto-save with the new layout
      if (dashboard) {
        scheduleSave({
          config: {
            ...dashboard.config,
            widgets: pendingWidgets ?? dashboard.config?.widgets ?? [],
          },
          layout: newLayout,
        })
      }
    },
    [dashboard, pendingWidgets, scheduleSave]
  )

  const handleSave = useCallback(async () => {
    if (!dashboard || !hasChanges) return

    await updateMutation.mutateAsync({
      config: {
        ...dashboard.config,
        widgets: pendingWidgets ?? dashboard.config?.widgets ?? [],
      },
      layout: pendingLayout ?? dashboard.layout ?? [],
    })

    setPendingLayout(null)
    setPendingWidgets(null)
    setIsEditMode(false)
  }, [dashboard, hasChanges, updateMutation, pendingWidgets, pendingLayout])

  const handleCancel = useCallback(() => {
    cancelPendingSave()
    setPendingLayout(null)
    setPendingWidgets(null)
    setIsEditMode(false)
  }, [cancelPendingSave])

  const handleLock = useCallback(async () => {
    if (!dashboard) return
    await lockMutation.mutateAsync(dashboard.id)
  }, [dashboard, lockMutation])

  const handleUnlock = useCallback(async () => {
    if (!dashboard) return
    await unlockMutation.mutateAsync(dashboard.id)
  }, [dashboard, unlockMutation])

  const handleExport = useCallback(async () => {
    if (!dashboard) return

    try {
      const exportData = await exportDashboard(projectId, dashboard.id)

      // Create and download JSON file
      const blob = new Blob([JSON.stringify(exportData, null, 2)], {
        type: 'application/json',
      })
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `${dashboard.name.replace(/[^a-z0-9]/gi, '_').toLowerCase()}_dashboard.json`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
    } catch {
      // Error is handled by the API client
    }
  }, [dashboard, projectId])

  const handleImport = useCallback(() => {
    setIsImportDialogOpen(true)
  }, [])

  const handleImportSubmit = useCallback(
    async (data: DashboardImportRequest) => {
      await importMutation.mutateAsync(data)
      setIsImportDialogOpen(false)
    },
    [importMutation]
  )

  // Open dialog to create a new widget (triggered from palette)
  const handleAddWidget = useCallback(
    (type: WidgetType, defaultSize: { w: number; h: number }) => {
      setPendingWidgetCreation({ type, defaultSize })
      setEditingWidget(null)
      setIsWidgetDialogOpen(true)
    },
    []
  )

  // Save widget from dialog (create new or update existing)
  const handleWidgetSave = useCallback(
    (widgetData: Omit<Widget, 'id'> | Widget) => {
      if (!dashboard) return

      const currentLayout = pendingLayout ?? dashboard.layout ?? []
      const currentWidgets = pendingWidgets ?? dashboard.config?.widgets ?? []

      let updatedWidgets: Widget[]
      let updatedLayout = currentLayout

      if ('id' in widgetData && widgetData.id) {
        // Editing existing widget - update in place
        updatedWidgets = currentWidgets.map((w) =>
          w.id === widgetData.id ? (widgetData as Widget) : w
        )
        setPendingWidgets(updatedWidgets)
      } else {
        // Creating new widget - generate ID and add
        const newWidget: Widget = {
          id: `widget_${Date.now()}`,
          ...(widgetData as Omit<Widget, 'id'>),
        }

        // Find the next available y position for layout
        const maxY = currentLayout.reduce(
          (max, item) => Math.max(max, item.y + item.h),
          0
        )

        const newLayoutItem: LayoutItem = {
          widget_id: newWidget.id,
          x: 0,
          y: maxY,
          w: pendingWidgetCreation?.defaultSize.w ?? 4,
          h: pendingWidgetCreation?.defaultSize.h ?? 3,
        }

        updatedWidgets = [...currentWidgets, newWidget]
        updatedLayout = [...currentLayout, newLayoutItem]
        setPendingWidgets(updatedWidgets)
        setPendingLayout(updatedLayout)
      }

      // Schedule auto-save with updated widgets
      scheduleSave({
        config: {
          ...dashboard.config,
          widgets: updatedWidgets,
        },
        layout: updatedLayout,
      })

      // Close dialog and reset state
      setIsWidgetDialogOpen(false)
      setEditingWidget(null)
      setPendingWidgetCreation(null)
    },
    [dashboard, pendingLayout, pendingWidgets, pendingWidgetCreation, scheduleSave]
  )

  // Open dialog to edit an existing widget
  const handleEditWidget = useCallback((widget: Widget) => {
    setEditingWidget(widget)
    setPendingWidgetCreation(null)
    setIsWidgetDialogOpen(true)
  }, [])

  // Delete a widget from the dashboard
  const handleDeleteWidget = useCallback(
    (widgetId: string) => {
      if (!dashboard) return

      const currentLayout = pendingLayout ?? dashboard.layout ?? []
      const currentWidgets = pendingWidgets ?? dashboard.config?.widgets ?? []

      const updatedWidgets = currentWidgets.filter((w) => w.id !== widgetId)
      const updatedLayout = currentLayout.filter((l) => l.widget_id !== widgetId)

      setPendingWidgets(updatedWidgets)
      setPendingLayout(updatedLayout)

      // Schedule auto-save with updated widgets
      scheduleSave({
        config: {
          ...dashboard.config,
          widgets: updatedWidgets,
        },
        layout: updatedLayout,
      })
    },
    [dashboard, pendingLayout, pendingWidgets, scheduleSave]
  )

  if (isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center py-16">
        <LoadingSpinner message="Loading dashboard..." />
      </div>
    )
  }

  if (error || !dashboard) {
    return (
      <div className="flex flex-col items-center justify-center py-12 space-y-4">
        <div className="rounded-lg bg-destructive/10 p-6 text-center max-w-md">
          <h3 className="font-semibold text-destructive mb-2">Failed to load dashboard</h3>
          <p className="text-sm text-muted-foreground mb-4">
            {error instanceof Error ? error.message : 'Dashboard not found'}
          </p>
          <Button asChild>
            <Link href={`/projects/${projectSlug}/dashboards`}>Go Back</Link>
          </Button>
        </div>
      </div>
    )
  }

  // Transform query results for DashboardGrid
  const gridQueryResults: Record<string, { data: unknown; error?: string }> | undefined =
    queryResults?.results
      ? Object.fromEntries(
          Object.entries(queryResults.results).map(([widgetId, result]) => [
            widgetId,
            { data: result.data, error: result.error },
          ])
        )
      : undefined

  return (
    <>
      <PageHeader
        title={dashboard.name}
        backHref={`/projects/${projectSlug}/dashboards`}
        description={dashboard.description}
      >
        <TimeRangePicker value={timeRange} onChange={setTimeRange} />
        <Button
          variant="outline"
          size="icon"
          onClick={handleRefresh}
          disabled={refreshMutation.isPending || isQueryFetching}
        >
          <RefreshCw
            className={`h-4 w-4 ${isQueryFetching ? 'animate-spin' : ''}`}
          />
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="icon">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              onClick={handleDelete}
              className="text-destructive focus:text-destructive"
            >
              <Trash className="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </PageHeader>

      <div className="flex-1 overflow-auto p-4 space-y-4">
        <DashboardEditorToolbar
          isEditMode={isEditMode}
          isLocked={isLocked}
          hasChanges={hasChanges}
          isLockPending={lockMutation.isPending}
          isUnlockPending={unlockMutation.isPending}
          isSavePending={updateMutation.isPending}
          autoSaveStatus={autoSaveStatus}
          autoSaveError={autoSaveError}
          onEditModeToggle={handleEditModeToggle}
          onSave={handleSave}
          onCancel={handleCancel}
          onLock={handleLock}
          onUnlock={handleUnlock}
          onExport={handleExport}
          onImport={handleImport}
          onAddWidget={handleAddWidget}
        />

        {dashboard?.config?.variables && dashboard.config.variables.length > 0 && (
          <VariablesBar
            variables={dashboard.config.variables}
            values={variableValues}
            onValueChange={setVariableValue}
            onReset={resetVariables}
            hasActiveVariables={hasActiveVariables}
            getVariableOptions={getVariableOptions}
            isLoadingOptions={isLoadingVariableOptions}
          />
        )}

        <DashboardGrid
          widgets={widgets}
          layout={layout}
          queryResults={gridQueryResults}
          isLoading={isQueryLoading && !queryResults}
          isEditMode={isEditMode}
          isLocked={isLocked}
          projectSlug={projectSlug}
          timeRange={timeRange}
          onLayoutChange={handleLayoutChange}
          onEditWidget={handleEditWidget}
          onDeleteWidget={handleDeleteWidget}
        />
      </div>

      <ImportDashboardDialog
        open={isImportDialogOpen}
        onOpenChange={setIsImportDialogOpen}
        onImport={handleImportSubmit}
        isPending={importMutation.isPending}
      />

      <WidgetEditDialog
        open={isWidgetDialogOpen}
        onOpenChange={(open) => {
          setIsWidgetDialogOpen(open)
          if (!open) {
            setEditingWidget(null)
            setPendingWidgetCreation(null)
          }
        }}
        widget={editingWidget ?? undefined}
        defaultType={pendingWidgetCreation?.type}
        onSave={handleWidgetSave}
      />
    </>
  )
}
