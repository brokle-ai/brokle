'use client'

import { useMemo } from 'react'
import { ClipboardList } from 'lucide-react'
import { QueuesTable } from './queues-table'
import { QueueDialogs } from './queue-dialogs'
import { CreateQueueDialog } from './create-queue-dialog'
import { AnnotationQueuesProvider, useAnnotationQueues } from '../context/annotation-queues-context'
import { useAnnotationQueuesQuery } from '../hooks/use-annotation-queues'
import { useQueuesTableState } from '../hooks/use-queues-table-state'
import { useProjectOnly } from '@/features/projects'
import type { QueueWithStats, AnnotationQueue } from '../types'

interface QueuesContentProps {
  projectSlug: string
}

export function QueuesContent({ projectSlug }: QueuesContentProps) {
  return (
    <AnnotationQueuesProvider projectSlug={projectSlug}>
      <QueuesContentInner projectSlug={projectSlug} />
    </AnnotationQueuesProvider>
  )
}

function QueuesContentInner({ projectSlug }: { projectSlug: string }) {
  const { currentProject } = useProjectOnly()
  const { setOpen, setCurrentRow } = useAnnotationQueues()
  const projectId = currentProject?.id
  const { data: queues, isLoading, error } = useAnnotationQueuesQuery(projectId)

  // URL state management
  const {
    search,
    status,
    sortBy,
    sortOrder,
    setSearch,
    setStatus,
    setSorting,
    resetAll,
    hasActiveFilters,
  } = useQueuesTableState()

  // Filter and sort data locally (server-side pagination can be added later)
  const filteredAndSortedQueues = useMemo(() => {
    if (!queues) return []

    let result = [...queues]

    // Apply search filter
    if (search) {
      const lowerSearch = search.toLowerCase()
      result = result.filter(
        (q) =>
          q.queue.name.toLowerCase().includes(lowerSearch) ||
          q.queue.description?.toLowerCase().includes(lowerSearch)
      )
    }

    // Apply status filter
    if (status) {
      result = result.filter((q) => q.queue.status === status)
    }

    // Apply sorting
    result.sort((a, b) => {
      let comparison = 0
      switch (sortBy) {
        case 'name':
          comparison = a.queue.name.localeCompare(b.queue.name)
          break
        case 'status':
          comparison = a.queue.status.localeCompare(b.queue.status)
          break
        case 'created_at':
          comparison = new Date(a.queue.created_at).getTime() - new Date(b.queue.created_at).getTime()
          break
        case 'updated_at':
          comparison = new Date(a.queue.updated_at).getTime() - new Date(b.queue.updated_at).getTime()
          break
        default:
          comparison = 0
      }
      return sortOrder === 'desc' ? -comparison : comparison
    })

    return result
  }, [queues, search, status, sortBy, sortOrder])

  // Action handlers
  const handleEdit = (queue: AnnotationQueue) => {
    setCurrentRow(queue)
    setOpen('edit')
  }

  const handleAddItems = (queue: AnnotationQueue) => {
    setCurrentRow(queue)
    setOpen('add-items')
  }

  const handleDelete = (queue: AnnotationQueue) => {
    setCurrentRow(queue)
    setOpen('delete')
  }

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

      {/* Content */}
      {isLoading ? (
        <QueuesTable
          data={[]}
          projectSlug={projectSlug}
          loading={true}
          search={search}
          status={status}
          sortBy={sortBy}
          sortOrder={sortOrder}
          onSearchChange={setSearch}
          onStatusChange={setStatus}
          onSortChange={setSorting}
          onReset={resetAll}
          hasActiveFilters={hasActiveFilters}
        />
      ) : error ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="text-muted-foreground mb-2">Failed to load annotation queues</p>
          <p className="text-sm text-destructive">{(error as Error).message}</p>
        </div>
      ) : filteredAndSortedQueues.length === 0 && !hasActiveFilters ? (
        <EmptyState />
      ) : (
        <QueuesTable
          data={filteredAndSortedQueues}
          projectSlug={projectSlug}
          search={search}
          status={status}
          sortBy={sortBy}
          sortOrder={sortOrder}
          onSearchChange={setSearch}
          onStatusChange={setStatus}
          onSortChange={setSorting}
          onReset={resetAll}
          hasActiveFilters={hasActiveFilters}
          onEdit={handleEdit}
          onAddItems={handleAddItems}
          onDelete={handleDelete}
        />
      )}

      {/* Dialogs */}
      <QueueDialogs />
    </div>
  )
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <div className="rounded-full bg-muted p-4 mb-4">
        <ClipboardList className="h-8 w-8 text-muted-foreground" />
      </div>
      <h3 className="font-semibold mb-1">No annotation queues yet</h3>
      <p className="text-sm text-muted-foreground mb-4">
        Create a queue to start collecting human feedback on your AI outputs.
      </p>
    </div>
  )
}
