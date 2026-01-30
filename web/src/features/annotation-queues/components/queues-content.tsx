'use client'

import { ClipboardList, Loader2 } from 'lucide-react'
import { useSearchParams } from 'next/navigation'
import { QueueList } from './queue-list'
import { QueueDialogs } from './queue-dialogs'
import { CreateQueueDialog } from './create-queue-dialog'
import { AnnotationQueuesProvider } from '../context/annotation-queues-context'
import { useAnnotationQueuesQuery } from '../hooks/use-annotation-queues'
import { useProjectOnly } from '@/features/projects'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useCardListNavigation } from '@/hooks/use-card-list-navigation'
import { CardListToolbar, CardListPagination } from '@/components/card-list'

interface QueuesContentProps {
  projectSlug: string
}

export function QueuesContent({ projectSlug }: QueuesContentProps) {
  return (
    <AnnotationQueuesProvider projectSlug={projectSlug}>
      <QueuesContentInner />
    </AnnotationQueuesProvider>
  )
}

function QueuesContentInner() {
  const searchParams = useSearchParams()
  const { currentProject } = useProjectOnly()
  const projectId = currentProject?.id
  const { filter, page, pageSize } = useTableSearchParams(searchParams)
  const { handleSearch, handleReset, handlePageChange, handlePageSizeChange } = useCardListNavigation({ searchParams })

  const { data: queuesResponse, isLoading, isFetching, error } = useAnnotationQueuesQuery(projectId, {
    page,
    limit: pageSize,
    search: filter || undefined,
  })
  const queues = queuesResponse?.data ?? []

  const totalCount = queuesResponse?.pagination?.total ?? 0
  const hasActiveFilters = !!filter

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Annotation Queues</h2>
          <p className="text-muted-foreground">
            Manage human-in-the-loop evaluation workflows for quality assessment.
          </p>
        </div>
        {projectId && <CreateQueueDialog projectId={projectId} />}
      </div>

      {/* Search */}
      <CardListToolbar
        searchPlaceholder="Filter queues..."
        searchValue={filter}
        onSearchChange={handleSearch}
        isPending={isFetching}
        onReset={handleReset}
        isFiltered={hasActiveFilters}
      />

      {/* Content */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : error ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="text-muted-foreground mb-2">Failed to load annotation queues</p>
          <p className="text-sm text-destructive">{(error as Error).message}</p>
        </div>
      ) : queues.length === 0 ? (
        <EmptyState hasSearch={hasActiveFilters} />
      ) : (
        <>
          <QueueList data={queues} />
          <CardListPagination
            page={page}
            pageSize={pageSize}
            totalCount={totalCount}
            onPageChange={handlePageChange}
            onPageSizeChange={handlePageSizeChange}
            isPending={isFetching}
          />
        </>
      )}

      {/* Dialogs */}
      <QueueDialogs />
    </div>
  )
}

function EmptyState({ hasSearch }: { hasSearch: boolean }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <div className="rounded-full bg-muted p-4 mb-4">
        <ClipboardList className="h-8 w-8 text-muted-foreground" />
      </div>
      {hasSearch ? (
        <>
          <h3 className="font-semibold mb-1">No queues found</h3>
          <p className="text-sm text-muted-foreground">
            Try adjusting your search term.
          </p>
        </>
      ) : (
        <>
          <h3 className="font-semibold mb-1">No annotation queues yet</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Create a queue to start collecting human feedback on your AI outputs.
          </p>
        </>
      )}
    </div>
  )
}
