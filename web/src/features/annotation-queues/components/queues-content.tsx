'use client'

import { ClipboardList, Search, Loader2 } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { useState, useMemo } from 'react'
import { QueueList } from './queue-list'
import { QueueDialogs } from './queue-dialogs'
import { CreateQueueDialog } from './create-queue-dialog'
import { AnnotationQueuesProvider } from '../context/annotation-queues-context'
import { useAnnotationQueuesQuery } from '../hooks/use-annotation-queues'
import { useProjectOnly } from '@/features/projects'

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
  const { currentProject } = useProjectOnly()
  const projectId = currentProject?.id
  const { data: queues, isLoading, error } = useAnnotationQueuesQuery(projectId)
  const [searchTerm, setSearchTerm] = useState('')

  const filteredQueues = useMemo(() => {
    if (!queues) return []
    if (!searchTerm) return queues
    const lowerSearch = searchTerm.toLowerCase()
    return queues.filter(
      (q) =>
        q.queue.name.toLowerCase().includes(lowerSearch) ||
        q.queue.description?.toLowerCase().includes(lowerSearch)
    )
  }, [queues, searchTerm])

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
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search queues..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
      </div>

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
      ) : filteredQueues.length === 0 ? (
        <EmptyState hasSearch={!!searchTerm} />
      ) : (
        <QueueList data={filteredQueues} />
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
