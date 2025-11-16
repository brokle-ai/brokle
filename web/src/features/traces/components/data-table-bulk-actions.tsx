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
    // Feature not implemented yet
    toast.error('Export functionality is not yet available', {
      description: 'This feature requires backend implementation and will be available in a future update.',
    })
  }

  const handleBulkDelete = () => {
    // Feature not implemented yet
    toast.error('Delete functionality is not yet available', {
      description: 'This feature requires backend implementation and will be available in a future update.',
    })
  }

  return (
    <>
      <BulkActionsToolbar table={table} entityName='trace'>
        {/* Export Action - Disabled (backend not implemented) */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant='outline'
              size='icon'
              onClick={handleBulkExport}
              disabled
              className='size-8 cursor-not-allowed opacity-50'
              aria-label='Export traces (not available)'
              title='Export traces (not available)'
            >
              <Download />
              <span className='sr-only'>Export traces (not available)</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p className='font-semibold'>Export not available</p>
            <p className='text-xs text-muted-foreground mt-1'>
              Backend endpoint not yet implemented
            </p>
          </TooltipContent>
        </Tooltip>

        {/* Delete Action - Disabled (backend not implemented) */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant='destructive'
              size='icon'
              onClick={handleBulkDelete}
              disabled
              className='size-8 cursor-not-allowed opacity-50'
              aria-label='Delete selected traces (not available)'
              title='Delete selected traces (not available)'
            >
              <Trash2 />
              <span className='sr-only'>Delete selected traces (not available)</span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p className='font-semibold'>Delete not available</p>
            <p className='text-xs text-muted-foreground mt-1'>
              Backend endpoint not yet implemented
            </p>
          </TooltipContent>
        </Tooltip>
      </BulkActionsToolbar>

      {/* Delete dialog - keep for future use but won't open */}
      <TracesMultiDeleteDialog
        open={showDeleteConfirm}
        onOpenChange={setShowDeleteConfirm}
        table={table}
      />
    </>
  )
}
