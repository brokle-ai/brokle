'use client'

import { useCallback, useMemo } from 'react'
import { Scale } from 'lucide-react'
import { RulesProvider, useRules } from '../context/rules-context'
import { RulesTable } from './rules-table'
import { RulesDialogs } from './rules-dialogs'
import { CreateRuleDialog } from './create-rule-dialog'
import { useEvaluationRulesQuery, useActivateEvaluationRuleMutation, useDeactivateEvaluationRuleMutation, useDeleteEvaluationRuleMutation } from '../hooks/use-evaluation-rules'
import { useRulesTableState } from '../hooks/use-rules-table-state'
import { useProjectOnly } from '@/features/projects'
import { PageHeader } from '@/components/layout/page-header'
import { DataTableEmptyState } from '@/components/data-table'
import { LoadingSpinner } from '@/components/guards/loading-spinner'
import type { EvaluationRule, RuleStatus } from '../types'
import type { Pagination } from '@/lib/api/core/types'

interface RulesProps {
  projectSlug: string
}

function RulesContent() {
  const { projectId } = useRules()
  const { currentProject, hasProject, isLoading: isProjectLoading } = useProjectOnly()

  // URL state management
  const tableState = useRulesTableState()
  const apiParams = tableState.toApiParams()

  // Fetch rules with pagination and filtering
  const {
    data: rulesResponse,
    isLoading: isQueryLoading,
    isFetching,
    error,
    refetch,
  } = useEvaluationRulesQuery(projectId, apiParams)

  // Mutations for table actions
  const activateMutation = useActivateEvaluationRuleMutation(projectId ?? '')
  const deactivateMutation = useDeactivateEvaluationRuleMutation(projectId ?? '')
  const deleteMutation = useDeleteEvaluationRuleMutation(projectId ?? '')

  const isLoading = isProjectLoading || isQueryLoading

  // Only show spinner on true initial load (no data at all)
  const isInitialLoad = isLoading && !rulesResponse

  // Extract data and pagination from response (memoized for stable reference)
  const rules = useMemo(() => rulesResponse?.rules ?? [], [rulesResponse?.rules])
  const totalCount = rulesResponse?.total ?? 0

  // Build pagination object
  const pagination: Pagination = {
    page: rulesResponse?.page ?? tableState.page,
    limit: rulesResponse?.limit ?? tableState.pageSize,
    total: totalCount,
    totalPages: Math.ceil(totalCount / (rulesResponse?.limit ?? tableState.pageSize)),
    hasNext: (rulesResponse?.page ?? tableState.page) < Math.ceil(totalCount / (rulesResponse?.limit ?? tableState.pageSize)),
    hasPrev: (rulesResponse?.page ?? tableState.page) > 1,
  }

  // Determine if project is truly empty (no data ever, not just filtered to zero)
  const isEmptyProject = !isLoading && totalCount === 0 && !tableState.hasActiveFilters

  // Handle status toggle (inline toggle from Opik pattern)
  const handleStatusToggle = useCallback(
    (ruleId: string, newStatus: RuleStatus) => {
      const rule = rules.find((r) => r.id === ruleId)
      if (!rule) return

      if (newStatus === 'active') {
        activateMutation.mutate({ ruleId, ruleName: rule.name })
      } else {
        deactivateMutation.mutate({ ruleId, ruleName: rule.name })
      }
    },
    [rules, activateMutation, deactivateMutation]
  )

  // Handle edit action - opens the edit dialog
  const handleEdit = useCallback((rule: EvaluationRule) => {
    // TODO: Open edit dialog via context
    // For now, navigate to detail page
    window.location.href = `/projects/${currentProject?.slug}/evaluation-rules/${rule.id}`
  }, [currentProject?.slug])

  // Handle duplicate action
  const handleDuplicate = useCallback((rule: EvaluationRule) => {
    // TODO: Open create dialog with pre-filled data
    console.log('Duplicate rule:', rule.id)
  }, [])

  // Handle view logs action
  const handleViewLogs = useCallback((rule: EvaluationRule) => {
    // Navigate to rule detail page with executions tab
    window.location.href = `/projects/${currentProject?.slug}/evaluation-rules/${rule.id}?tab=executions`
  }, [currentProject?.slug])

  // Handle delete action
  const handleDelete = useCallback(
    (rule: EvaluationRule) => {
      if (confirm(`Are you sure you want to delete "${rule.name}"? This action cannot be undone.`)) {
        deleteMutation.mutate({ ruleId: rule.id, ruleName: rule.name })
      }
    },
    [deleteMutation]
  )

  return (
    <>
      <PageHeader title="Evaluation Rules">
        {projectId && <CreateRuleDialog projectId={projectId} orgId={currentProject?.organizationId} />}
      </PageHeader>
      <div className="-mx-4 flex flex-1 flex-col overflow-auto px-4 py-1">
        {isInitialLoad && (
          <div className="flex flex-1 items-center justify-center py-16">
            <LoadingSpinner message="Loading evaluation rules..." />
          </div>
        )}

        {error && !isInitialLoad && (
          <div className="flex flex-col items-center justify-center py-12 space-y-4">
            <div className="rounded-lg bg-destructive/10 p-6 text-center max-w-md">
              <h3 className="font-semibold text-destructive mb-2">
                Failed to load evaluation rules
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
            title="No evaluation rules yet"
            description="Create evaluation rules to automatically score incoming spans using LLM, built-in scorers, or regex patterns."
          />
        )}

        {!error && hasProject && !isInitialLoad && !isEmptyProject && (
          <RulesTable
            data={rules}
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
      <RulesDialogs />
    </>
  )
}

export function Rules({ projectSlug }: RulesProps) {
  return (
    <RulesProvider projectSlug={projectSlug}>
      <RulesContent />
    </RulesProvider>
  )
}
