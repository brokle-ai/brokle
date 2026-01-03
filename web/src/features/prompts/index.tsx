'use client'

import { useRouter } from 'next/navigation'
import { Plus, FileText } from 'lucide-react'
import { useProjectOnly } from '@/features/projects'
import { useProjectPrompts } from './hooks/use-project-prompts'
import { useProtectedLabelsQuery } from './hooks/use-prompts-queries'
import { PromptsTable } from './components/prompt-list/PromptList'
import { PageHeader } from '@/components/layout/page-header'
import { Button } from '@/components/ui/button'
import { DataTableEmptyState } from '@/components/data-table'
import { LoadingSpinner } from '@/components/guards/loading-spinner'

interface PromptsProps {
  projectSlug: string
  orgSlug: string
}

export function Prompts({ projectSlug, orgSlug }: PromptsProps) {
  const router = useRouter()
  const { currentProject, hasProject } = useProjectOnly()
  const { data, totalCount, isLoading, isFetching, error, refetch, tableState } = useProjectPrompts()
  const { data: protectedLabels } = useProtectedLabelsQuery(currentProject?.id)

  // Check if there are active filters (from nuqs state)
  const hasActiveFilters = tableState.hasActiveFilters

  // Only show spinner on true initial load (no data at all)
  const isInitialLoad = isLoading && data.length === 0

  // Determine if project is truly empty (no data ever, not just filtered to zero)
  const isEmptyProject = !isLoading && totalCount === 0 && !hasActiveFilters

  return (
    <>
      <PageHeader title="Prompts">
        <Button onClick={() => router.push(`/projects/${projectSlug}/prompts/new`)}>
          <Plus className="mr-2 h-4 w-4" />
          New Prompt
        </Button>
      </PageHeader>
      <div className="-mx-4 flex flex-1 flex-col overflow-auto px-4 py-1">
        {/* Initial loading (first load, no cache) */}
        {isInitialLoad && (
          <div className="flex flex-1 items-center justify-center py-16">
            <LoadingSpinner message="Loading prompts..." />
          </div>
        )}

        {/* Error state */}
        {error && !isInitialLoad && (
          <div className="flex flex-col items-center justify-center py-12 space-y-4">
            <div className="rounded-lg bg-destructive/10 p-6 text-center max-w-md">
              <h3 className="font-semibold text-destructive mb-2">Failed to load prompts</h3>
              <p className="text-sm text-muted-foreground mb-4">{error}</p>
              <button
                onClick={() => refetch()}
                className="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                Try Again
              </button>
            </div>
          </div>
        )}

        {/* No project selected */}
        {!hasProject && !isInitialLoad && !error && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <p className="text-muted-foreground">No project selected</p>
          </div>
        )}

        {/* Empty project (never had data) */}
        {!error && hasProject && !isInitialLoad && isEmptyProject && (
          <DataTableEmptyState
            icon={<FileText className="h-full w-full" />}
            title="No prompts yet"
            description="Click 'New Prompt' above to create your first prompt."
          />
        )}

        {/* Table (has data OR has active filters) */}
        {!error && hasProject && !isInitialLoad && !isEmptyProject && (
          <PromptsTable
            data={data}
            totalCount={totalCount}
            isFetching={isFetching}
            protectedLabels={protectedLabels || []}
            projectSlug={projectSlug}
            orgSlug={orgSlug}
          />
        )}
      </div>
    </>
  )
}

// Types
export * from './types'

// API
export * from './api/prompts-api'

// Hooks
export * from './hooks/use-prompts-queries'
export * from './hooks/use-project-prompts'

// Components - Organized by feature area per design doc

// Common components
export { LabelBadge, LabelList } from './components/label-badge'
export { PromptTypeIcon } from './components/common/PromptTypeIcon'
export { PromptStatusBadge } from './components/common/PromptStatusBadge'

// Prompt Editor components (nested directory)
export { PromptEditor, TextEditor, ChatEditor } from './components/prompt-editor'
export { ChatMessageEditor } from './components/prompt-editor/ChatMessageEditor'
export { PromptTemplateInput } from './components/prompt-editor/PromptTemplateInput'
export { PromptEditorToolbar } from './components/prompt-editor/PromptEditorToolbar'
export { VariableBadge, VariableList } from './components/prompt-editor/VariableExtractor'

// Prompt List components (nested directory)
export { PromptsTable } from './components/prompt-list/PromptList'
export { PromptCard } from './components/prompt-list/PromptCard'
export { PromptFilters } from './components/prompt-list/PromptFilters'
export { PromptsDeleteDialog } from './components/prompt-list/prompts-delete-dialog'
export { createPromptsColumns } from './components/prompt-list/prompts-columns'

// Label Management components (nested directory)
export { LabelSelector } from './components/label-management/LabelSelector'
export { ProtectedLabelsConfig } from './components/label-management/ProtectedLabelsConfig'

// Version Management components (nested directory)
export { VersionHistory } from './components/version-management/VersionHistory'
export { VersionDiff } from './components/version-management/VersionDiff'
export { VersionCompare } from './components/version-management/VersionCompare'
export { VersionDetails } from './components/version-management/VersionDetails'

// Utilities
export { extractVariables } from './utils/variable-extraction'
