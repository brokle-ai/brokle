import { useMemo, useState } from 'react'
import { Download, Trash2 } from 'lucide-react'
import { type Table } from '@tanstack/react-table'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { BulkActionsToolbar } from '@/components/bulk-actions-toolbar'
import { BulkAddToQueueButton } from '@/features/annotation-queues/components/bulk-add-to-queue-button'
import { TracesMultiDeleteDialog } from './traces-multi-delete-dialog'
import type { TraceListItem } from '../api/types'

type TracesBulkActionsProps = {
  table: Table<TraceListItem>
  projectId: string
}

/**
 * Bulk-actions strip shown when rows are selected. Export + delete are
 * disabled until the backend exposes bulk endpoints — the button shapes
 * exist so UI parity is preserved and enabling them is a one-line flip.
 */
export function TracesBulkActions({
  table,
  projectId,
}: TracesBulkActionsProps) {
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  // Trace IDs for the currently-selected rows. Memoized so the
  // BulkAddToQueueButton's mutation closure doesn't get a fresh array
  // identity on unrelated re-renders.
  const selectedTraceIds = useMemo(
    () =>
      table
        .getSelectedRowModel()
        .rows.map((r) => r.original.trace_id)
        .filter((id): id is string => Boolean(id)),
    [table],
  )

  const handleBulkExport = () => {
    toast.error('Export functionality is not yet available', {
      description:
        'This feature requires backend implementation and will be available in a future update.',
    })
  }

  const handleBulkDelete = () => {
    toast.error('Delete functionality is not yet available', {
      description:
        'This feature requires backend implementation and will be available in a future update.',
    })
  }

  return (
    <>
      <BulkActionsToolbar table={table} entityName="trace">
        <BulkAddToQueueButton
          projectId={projectId}
          objectIds={selectedTraceIds}
          objectType="trace"
        />
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              onClick={handleBulkExport}
              disabled
              className="size-8 cursor-not-allowed opacity-50"
              aria-label="Export traces (not available)"
              title="Export traces (not available)"
            >
              <Download />
              <span className="sr-only">Export traces (not available)</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p className="font-semibold">Export not available</p>
            <p className="text-xs text-muted-foreground mt-1">
              Backend endpoint not yet implemented
            </p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="destructive"
              size="icon"
              onClick={handleBulkDelete}
              disabled
              className="size-8 cursor-not-allowed opacity-50"
              aria-label="Delete selected traces (not available)"
              title="Delete selected traces (not available)"
            >
              <Trash2 />
              <span className="sr-only">
                Delete selected traces (not available)
              </span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p className="font-semibold">Delete not available</p>
            <p className="text-xs text-muted-foreground mt-1">
              Backend endpoint not yet implemented
            </p>
          </TooltipContent>
        </Tooltip>
      </BulkActionsToolbar>

      <TracesMultiDeleteDialog
        open={showDeleteConfirm}
        onOpenChange={setShowDeleteConfirm}
        table={table}
      />
    </>
  )
}
