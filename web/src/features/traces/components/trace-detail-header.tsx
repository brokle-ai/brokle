'use client'

import * as React from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import {
  ChevronUp,
  ChevronDown,
  Maximize2,
  ExternalLink,
  X,
  ListTree,
  ArrowLeft,
  Database,
  Trash2,
  Star,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Trace, Span } from '../data/schema'
import { CopyIdsDropdown } from './copy-ids-dropdown'
import { TraceTagsEditor } from './trace-tags-editor'
import { ExportTraceButton } from './export-trace-button'
import { AddTraceToDatasetDialog } from '@/features/datasets/components/add-from-traces-dialog'
import { useDeleteTrace } from '../hooks/use-delete-trace'
import { useUpdateTraceBookmark } from '../hooks/use-update-trace-bookmark'
import { CommentsDrawer } from './comments-drawer'
import { AnnotationsDrawer } from './annotations-drawer'
import { AddToQueueButton } from '@/features/annotation-queues'

interface TraceDetailHeaderProps {
  trace: Trace
  spans: Span[]
  projectId: string
  selectedSpanId?: string
  onPrevious?: () => void
  onNext?: () => void
  onExpand: () => void
  onOpenInNewTab: () => void
  onClose?: () => void
  onBack?: () => void
  hasPrevious: boolean
  hasNext: boolean
  context: 'peek' | 'page'
  className?: string
}

/**
 * TraceDetailHeader - Header for trace detail view
 *
 * Displays trace identification and navigation controls.
 * Adapts based on context:
 * - peek: Shows prev/next navigation, expand button, close button
 * - page: Shows back button, open in new tab
 *
 * Features:
 * - Copy IDs dropdown (trace ID, span ID, URL, JSON)
 * - Add to Dataset button
 * - Delete with confirmation
 */
export function TraceDetailHeader({
  trace,
  spans,
  projectId,
  selectedSpanId,
  onPrevious,
  onNext,
  onExpand,
  onOpenInNewTab,
  onClose,
  onBack,
  hasPrevious,
  hasNext,
  context,
  className,
}: TraceDetailHeaderProps) {
  const router = useRouter()
  const isPeek = context === 'peek'

  // Dataset dialog state
  const [datasetDialogOpen, setDatasetDialogOpen] = React.useState(false)

  // Delete mutation
  const deleteMutation = useDeleteTrace(projectId)

  // Bookmark mutation
  const bookmarkMutation = useUpdateTraceBookmark(projectId)

  const handleToggleBookmark = () => {
    bookmarkMutation.mutate({
      traceId: trace.trace_id,
      bookmarked: !trace.bookmarked,
    })
  }

  const handleDelete = async () => {
    await deleteMutation.mutateAsync(trace.trace_id)
    // Navigate back to traces list after deletion
    if (isPeek && onClose) {
      onClose()
    } else if (onBack) {
      onBack()
    } else {
      router.push(`/projects/${projectId}/traces`)
    }
  }

  return (
    <div className={cn('border-b bg-background', className)}>
      {/* Minimal header: ID + Copy + Navigation */}
      <div className='flex items-center justify-between px-4 py-3'>
        <div className='flex items-center gap-2'>
          {/* Back button for page context */}
          {!isPeek && onBack && (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant='ghost'
                    size='icon'
                    className='h-8 w-8 mr-1'
                    onClick={onBack}
                  >
                    <ArrowLeft className='h-4 w-4' />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Back to traces</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )}

          {/* Trace prefix */}
          <div className='flex items-center gap-1.5'>
            <ListTree className='h-4 w-4' />
            <span className='text-sm font-medium'>Trace</span>
          </div>
          {/* Trace ID with copy dropdown */}
          <span className='text-sm font-medium font-mono'>
            {trace.trace_id}
          </span>
          <CopyIdsDropdown
            trace={trace}
            projectId={projectId}
            selectedSpanId={selectedSpanId}
          />
        </div>

        {/* Actions and Navigation Controls */}
        <div className='flex items-center gap-1'>
          {/* Add to Dataset */}
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-8 w-8'
                  onClick={() => setDatasetDialogOpen(true)}
                >
                  <Database className='h-4 w-4' />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Add to Dataset</TooltipContent>
            </Tooltip>
          </TooltipProvider>

          {/* Add to Annotation Queue */}
          <AddToQueueButton
            projectId={projectId}
            objectId={trace.trace_id}
            objectType="trace"
          />

          {/* Bookmark toggle */}
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-8 w-8'
                  onClick={handleToggleBookmark}
                  disabled={bookmarkMutation.isPending}
                >
                  <Star
                    className={cn(
                      'h-4 w-4',
                      trace.bookmarked && 'fill-yellow-400 text-yellow-400'
                    )}
                  />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                {trace.bookmarked ? 'Remove bookmark' : 'Bookmark trace'}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>

          {/* Comments drawer */}
          <CommentsDrawer projectId={projectId} traceId={trace.trace_id} />

          {/* Annotations drawer */}
          <AnnotationsDrawer projectId={projectId} traceId={trace.trace_id} />

          {/* Export dropdown */}
          <ExportTraceButton trace={trace} spans={spans} />

          {/* Delete with confirmation */}
          <AlertDialog>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant='ghost'
                      size='icon'
                      className='h-8 w-8 text-muted-foreground hover:text-destructive'
                      disabled={deleteMutation.isPending}
                    >
                      <Trash2 className='h-4 w-4' />
                    </Button>
                  </AlertDialogTrigger>
                </TooltipTrigger>
                <TooltipContent>Delete Trace</TooltipContent>
              </Tooltip>
            </TooltipProvider>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Delete Trace</AlertDialogTitle>
                <AlertDialogDescription>
                  This will permanently delete this trace and all its spans.
                  This action cannot be undone.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={handleDelete}
                  className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
                  disabled={deleteMutation.isPending}
                >
                  {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>

          <div className='w-px h-6 bg-border mx-1' />

          {/* Prev/Next - only in peek mode */}
          {isPeek && (
            <>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant='ghost'
                      size='icon'
                      className='h-8 w-8'
                      onClick={onPrevious}
                      disabled={!hasPrevious}
                    >
                      <ChevronUp className='h-4 w-4' />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Previous trace (←)</TooltipContent>
                </Tooltip>
              </TooltipProvider>

              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant='ghost'
                      size='icon'
                      className='h-8 w-8'
                      onClick={onNext}
                      disabled={!hasNext}
                    >
                      <ChevronDown className='h-4 w-4' />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Next trace (→)</TooltipContent>
                </Tooltip>
              </TooltipProvider>

              <div className='w-px h-6 bg-border mx-1' />
            </>
          )}

          {/* Expand to full page - only in peek mode */}
          {isPeek && (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant='ghost'
                    size='icon'
                    className='h-8 w-8'
                    onClick={onExpand}
                  >
                    <Maximize2 className='h-4 w-4' />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Open full page</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )}

          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-8 w-8'
                  onClick={onOpenInNewTab}
                >
                  <ExternalLink className='h-4 w-4' />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Open in new tab</TooltipContent>
            </Tooltip>
          </TooltipProvider>

          {/* Close - only in peek mode */}
          {isPeek && onClose && (
            <>
              <div className='w-px h-6 bg-border mx-1' />

              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant='ghost'
                      size='icon'
                      className='h-8 w-8'
                      onClick={onClose}
                    >
                      <X className='h-4 w-4' />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Close (Esc)</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </>
          )}
        </div>
      </div>

      {/* Tags row */}
      <div className='px-4 pb-3'>
        <TraceTagsEditor
          projectId={projectId}
          traceId={trace.trace_id}
          tags={trace.tags || []}
        />
      </div>

      {/* Add to Dataset Dialog */}
      <AddTraceToDatasetDialog
        projectId={projectId}
        traceId={trace.trace_id}
        traceName={trace.name}
        open={datasetDialogOpen}
        onOpenChange={setDatasetDialogOpen}
      />
    </div>
  )
}
