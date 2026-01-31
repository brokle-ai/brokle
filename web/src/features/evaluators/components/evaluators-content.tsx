'use client'

import { useCallback, useMemo } from 'react'
import { Scale } from 'lucide-react'
import { EvaluatorsProvider, useEvaluators } from '../context/evaluators-context'
import { EvaluatorsTable } from './evaluators-table'
import { EvaluatorsDialogs } from './evaluators-dialogs'
import { CreateEvaluatorDialog } from './create-evaluator-dialog'
import { useEvaluatorsQuery, useActivateEvaluatorMutation, useDeactivateEvaluatorMutation, useDeleteEvaluatorMutation } from '../hooks/use-evaluators'
import { useEvaluatorsTableState } from '../hooks/use-evaluators-table-state'
import { useProjectOnly } from '@/features/projects'
import { PageHeader } from '@/components/layout/page-header'
import { DataTableEmptyState } from '@/components/data-table'
import { LoadingSpinner } from '@/components/guards/loading-spinner'
import type { Evaluator, EvaluatorStatus } from '../types'
import type { Pagination } from '@/lib/api/core/types'

interface EvaluatorsProps {
  projectSlug: string
}

function EvaluatorsContent() {
  const { projectId } = useEvaluators()
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()

  // URL state management
  const tableState = useEvaluatorsTableState()
  const apiParams = tableState.toApiParams()

  // Fetch evaluators with pagination and filtering
  const {
    data: evaluatorsResponse,
    isLoading: isQueryLoading,
    isFetching,
    error,
    refetch,
  } = useEvaluatorsQuery(projectId, apiParams)

  // Mutations for table actions
  const activateMutation = useActivateEvaluatorMutation(projectId ?? '')
  const deactivateMutation = useDeactivateEvaluatorMutation(projectId ?? '')
  const deleteMutation = useDeleteEvaluatorMutation(projectId ?? '')

  const isLoading = isProjectLoading || isQueryLoading

  // Only show spinner on true initial load (no data at all)
  const isInitialLoad = isLoading && !evaluatorsResponse

  // Extract data from response using generic PaginatedResponse fields
  const evaluators = useMemo(
    () => evaluatorsResponse?.data ?? [],
    [evaluatorsResponse?.data]
  )

  // Use pagination directly from response
  const pagination: Pagination = evaluatorsResponse?.pagination ?? {
    page: tableState.page,
    limit: tableState.pageSize,
    total: 0,
    totalPages: 0,
    hasNext: false,
    hasPrev: false,
  }
  const totalCount = pagination.total

  // Determine if project is truly empty (no data ever, not just filtered to zero)
  const isEmptyProject = !isLoading && totalCount === 0 && !tableState.hasActiveFilters

  // Handle status toggle (inline toggle from Opik pattern)
  const handleStatusToggle = useCallback(
    (evaluatorId: string, newStatus: EvaluatorStatus) => {
      const evaluator = evaluators.find((e) => e.id === evaluatorId)
      if (!evaluator) return

      if (newStatus === 'active') {
        activateMutation.mutate({ evaluatorId, evaluatorName: evaluator.name })
      } else {
        deactivateMutation.mutate({ evaluatorId, evaluatorName: evaluator.name })
      }
    },
    [evaluators, activateMutation, deactivateMutation]
  )

  // Handle edit action - opens the edit dialog
  const handleEdit = useCallback((evaluator: Evaluator) => {
    // TODO: Open edit dialog via context
    // For now, navigate to detail page
    window.location.href = `/projects/${currentProject?.slug}/evaluators/${evaluator.id}`
  }, [currentProject?.slug])

  // Handle duplicate action
  const handleDuplicate = useCallback((evaluator: Evaluator) => {
    // TODO: Open create dialog with pre-filled data
    console.log('Duplicate evaluator:', evaluator.id)
  }, [])

  // Handle view logs action
  const handleViewLogs = useCallback((evaluator: Evaluator) => {
    // Navigate to evaluator detail page with executions tab
    window.location.href = `/projects/${currentProject?.slug}/evaluators/${evaluator.id}?tab=executions`
  }, [currentProject?.slug])

  // Handle delete action
  const handleDelete = useCallback(
    (evaluator: Evaluator) => {
      if (confirm(`Are you sure you want to delete "${evaluator.name}"? This action cannot be undone.`)) {
        deleteMutation.mutate({ evaluatorId: evaluator.id, evaluatorName: evaluator.name })
      }
    },
    [deleteMutation]
  )

  return (
    <>
      <PageHeader title="Evaluators">
        {projectId && <CreateEvaluatorDialog projectId={projectId} orgId={currentProject?.organizationId} />}
      </PageHeader>
      <div className="-mx-4 flex flex-1 flex-col overflow-auto px-4 py-1">
        {isInitialLoad && (
          <div className="flex flex-1 items-center justify-center py-16">
            <LoadingSpinner message="Loading evaluators..." />
          </div>
        )}

        {error && !isInitialLoad && (
          <div className="flex flex-col items-center justify-center py-12 space-y-4">
            <div className="rounded-lg bg-destructive/10 p-6 text-center max-w-md">
              <h3 className="font-semibold text-destructive mb-2">
                Failed to load evaluators
              </h3>
              <p className="text-sm text-muted-foreground mb-4">{error.message}</p>
              <button
                onClick={() => refetch()}
                className="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                Try Again
              </button>
            </div>
          </div>
        )}

        {!hasProject && !isInitialLoad && !error && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <p className="text-muted-foreground">No project selected</p>
          </div>
        )}

        {!error && hasProject && !isInitialLoad && isEmptyProject && (
          <DataTableEmptyState
            icon={<Scale className="h-full w-full" />}
            title="No evaluators yet"
            description="Create evaluators to automatically score incoming spans using LLM, built-in scorers, or regex patterns."
          />
        )}

        {!error && hasProject && !isInitialLoad && !isEmptyProject && (
          <EvaluatorsTable
            data={evaluators}
            pagination={pagination}
            projectSlug={currentProject?.slug ?? ''}
            loading={isFetching}
            search={tableState.search}
            scorerType={tableState.scorerType}
            status={tableState.status}
            sortBy={tableState.sortBy}
            sortOrder={tableState.sortOrder}
            onSearchChange={tableState.setSearch}
            onScorerTypeChange={tableState.setScorerType}
            onStatusChange={tableState.setStatus}
            onPageChange={tableState.setPagination}
            onSortChange={tableState.setSorting}
            onReset={tableState.resetAll}
            hasActiveFilters={tableState.hasActiveFilters}
            onStatusToggle={handleStatusToggle}
            onEdit={handleEdit}
            onDuplicate={handleDuplicate}
            onViewLogs={handleViewLogs}
            onDelete={handleDelete}
          />
        )}
      </div>
      <EvaluatorsDialogs />
    </>
  )
}

export function Evaluators({ projectSlug }: EvaluatorsProps) {
  return (
    <EvaluatorsProvider projectSlug={projectSlug}>
      <EvaluatorsContent />
    </EvaluatorsProvider>
  )
}
