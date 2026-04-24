import { useNavigate } from '@tanstack/react-router'
import { Plus, FileText } from 'lucide-react'
import { useProjectOnly } from '@/features/projects'
import { useProtectedLabelsQuery, usePromptsQuery } from './api/queries'
import { usePromptsTableState } from './hooks/use-prompts-table-state'
import { PromptsTable } from './components/prompt-list/PromptList'
import { PageHeader } from '@/components/layout/page-header'
import { Button } from '@/components/ui/button'
import { DataTableEmptyState } from '@/components/data-table'
import { Skeleton } from '@/components/ui/skeleton'

interface PromptsProps {
  orgId: string
  projectId: string
}

export function Prompts({ orgId, projectId }: PromptsProps) {
  const navigate = useNavigate()
  const { hasProject } = useProjectOnly()
  const tableState = usePromptsTableState()
  const type = tableState.types.length > 0 ? tableState.types[0] : undefined

  const { data, isLoading, isFetching, error, refetch } = usePromptsQuery(
    projectId,
    {
      page: tableState.page,
      limit: tableState.pageSize,
      search: tableState.search || undefined,
      type,
    },
  )

  const { data: protectedLabels } = useProtectedLabelsQuery(projectId)

  const rows = data?.data ?? []
  const totalCount = data?.total ?? 0
  const hasActiveFilters = tableState.hasActiveFilters
  const isInitialLoad = isLoading && rows.length === 0
  const isEmptyProject = !isLoading && totalCount === 0 && !hasActiveFilters
  const errMsg =
    error instanceof Error ? error.message : error ? String(error) : null

  return (
    <>
      <PageHeader title="Prompts">
        <Button
          onClick={() =>
            void navigate({
              to: '/o/$orgId/p/$projectId/prompts/new',
              params: { orgId, projectId },
            })
          }
        >
          <Plus className="mr-2 h-4 w-4" />
          New Prompt
        </Button>
      </PageHeader>
      <div className="-mx-4 flex flex-1 flex-col overflow-auto px-4 py-1">
        {isInitialLoad && (
          <div className="space-y-4 py-4">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-[400px] w-full" />
          </div>
        )}

        {errMsg && !isInitialLoad && (
          <div className="flex flex-col items-center justify-center py-12 space-y-4">
            <div className="rounded-lg bg-destructive/10 p-6 text-center max-w-md">
              <h3 className="font-semibold text-destructive mb-2">
                Failed to load prompts
              </h3>
              <p className="text-sm text-muted-foreground mb-4">{errMsg}</p>
              <button
                onClick={() => void refetch()}
                className="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                Try Again
              </button>
            </div>
          </div>
        )}

        {!hasProject && !isInitialLoad && !errMsg && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <p className="text-muted-foreground">No project selected</p>
          </div>
        )}

        {!errMsg && hasProject && !isInitialLoad && isEmptyProject && (
          <DataTableEmptyState
            icon={<FileText className="h-full w-full" />}
            title="No prompts yet"
            description="Click 'New Prompt' above to create your first prompt."
          />
        )}

        {!errMsg && hasProject && !isInitialLoad && !isEmptyProject && (
          <PromptsTable
            data={rows}
            totalCount={totalCount}
            isFetching={isFetching}
            protectedLabels={protectedLabels || []}
            projectId={projectId}
          />
        )}
      </div>
    </>
  )
}

// Re-exports — types + hooks + components. `api/queries` is the
// canonical public surface for callers; the lower-level `api/prompts-api`
// is re-exported via `queries.ts`'s mutation fns and direct imports,
// so we don't blanket-export it here (would duplicate `createPrompt`).
export * from './types'
export * from './api/queries'

export { LabelBadge, LabelList } from './components/label-badge'
export { PromptTypeIcon } from './components/common/PromptTypeIcon'
export { PromptStatusBadge } from './components/common/PromptStatusBadge'

export {
  PromptEditor,
  TextEditor,
  ChatEditor,
} from './components/prompt-editor'
export { ChatMessageEditor } from './components/prompt-editor/ChatMessageEditor'
export { PromptTemplateInput } from './components/prompt-editor/PromptTemplateInput'
export {
  VariableBadge,
  VariableList,
} from './components/prompt-editor/VariableExtractor'

export { PromptsTable } from './components/prompt-list/PromptList'
export { PromptsDeleteDialog } from './components/prompt-list/prompts-delete-dialog'
export { createPromptsColumns } from './components/prompt-list/prompts-columns'

export { LabelSelector } from './components/label-management/LabelSelector'

export { DiffViewer } from './components/version-management/DiffViewer'
export { VersionDiffDialog } from './components/version-management/VersionDiffDialog'

export {
  PromptDetailLayout,
  VersionSidebar,
  VersionSidebarItem,
  PromptViewerPanel,
  JsonConfigViewer,
} from './components/prompt-detail'

export {
  PromptEditLayout,
  PromptEditPanel,
  SaveVersionDialog,
  JsonConfigEditor,
} from './components/prompt-edit'

export {
  usePromptDetailState,
  type UsePromptDetailStateReturn,
  type PromptDetailTab,
} from './hooks/use-prompt-detail-state'

export {
  usePromptEditState,
  type UsePromptEditStateReturn,
} from './hooks/use-prompt-edit-state'

export {
  usePromptsTableState,
  type UsePromptsTableStateReturn,
} from './hooks/use-prompts-table-state'

export { extractVariables } from './utils/variable-extraction'
