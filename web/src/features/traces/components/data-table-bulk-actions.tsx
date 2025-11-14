'use client'

import { useState } from 'react'
import { Download, Trash2 } from 'lucide-react'
import { type Table } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { BulkActionsToolbar } from '@/components/bulk-actions-toolbar'
import { TracesMultiDeleteDialog } from './traces-multi-delete-dialog'
import type { Trace } from '../data/schema'
import { toast } from 'sonner'

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

type DataTableBulkActionsProps<TData> = {
  table: Table<TData>
}

export function DataTableBulkActions<TData>({
  table,
}: DataTableBulkActionsProps<TData>) {
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  // Get all selected rows across all pages
  const rowSelection = table.getState().rowSelection
  const selectedRows = table.getSelectedRowModel().flatRows
  const selectedTraces = selectedRows.map((row) => row.original as Trace)

  const handleBulkExport = () => {
    if (selectedTraces.length === 0) return

    toast.promise(sleep(2000), {
      loading: 'Exporting traces...',
      success: () => {
        table.resetRowSelection()
        return `Exported ${selectedTraces.length} trace${selectedTraces.length > 1 ? 's' : ''} to CSV.`
      },
      error: 'Error',
    })
  }

  return (
    <>
      <BulkActionsToolbar table={table} entityName='trace'>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant='outline'
              size='icon'
              onClick={handleBulkExport}
              className='size-8'
              aria-label='Export traces'
              title='Export traces'
            >
              <Download />
              <span className='sr-only'>Export traces</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Export traces</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant='destructive'
              size='icon'
              onClick={() => setShowDeleteConfirm(true)}
              className='size-8'
              aria-label='Delete selected traces'
              title='Delete selected traces'
            >
              <Trash2 />
              <span className='sr-only'>Delete selected traces</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Delete selected traces</p>
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
