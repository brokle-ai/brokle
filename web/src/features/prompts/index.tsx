'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { useProjectOnly } from '@/features/projects'
import { useProjectPrompts } from './hooks/use-project-prompts'
import { useProtectedLabelsQuery } from './hooks/use-prompts-queries'
import { PromptsTable } from './components/prompt-list/PromptList'
import type { PromptType } from './types'

interface PromptsProps {
  projectSlug: string
  orgSlug: string
}

export function Prompts({ projectSlug, orgSlug }: PromptsProps) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { currentProject, hasProject } = useProjectOnly()
  const { data, totalCount, page, pageSize, totalPages, isLoading, error, refetch } =
    useProjectPrompts()
  const { data: protectedLabels } = useProtectedLabelsQuery(currentProject?.id)

  const handlePageChange = (newPage: number) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set('page', String(newPage))
    router.push(`?${params.toString()}`)
  }

  const handleSearch = (query: string) => {
    const params = new URLSearchParams(searchParams.toString())
    if (query) {
      params.set('search', query)
    } else {
      params.delete('search')
    }
    params.set('page', '1')
    router.push(`?${params.toString()}`)
  }

  const handleTypeFilter = (type: PromptType | undefined) => {
    const params = new URLSearchParams(searchParams.toString())
    if (type) {
      params.set('type', type)
    } else {
      params.delete('type')
    }
    params.set('page', '1')
    router.push(`?${params.toString()}`)
  }

  return (
    <>
      <div className="mb-6 flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Prompts</h2>
          <p className="text-muted-foreground">
            Manage prompt templates with versioning and labels
          </p>
        </div>
      </div>
      <div className="-mx-4 flex-1 overflow-auto px-4 py-1">
        {error && !isLoading && (
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

        {!hasProject && !isLoading && !error && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <p className="text-muted-foreground">No project selected</p>
          </div>
        )}

        {!error && hasProject && (
          <PromptsTable
            data={data}
            totalCount={totalCount}
            page={page}
            pageSize={pageSize}
            totalPages={totalPages}
            isLoading={isLoading}
            protectedLabels={protectedLabels || []}
            projectSlug={projectSlug}
            orgSlug={orgSlug}
            onPageChange={handlePageChange}
            onSearch={handleSearch}
            onTypeFilter={handleTypeFilter}
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
