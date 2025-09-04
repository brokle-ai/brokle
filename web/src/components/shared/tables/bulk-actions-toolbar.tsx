'use client'

import { type Table } from '@tanstack/react-table'
import { 
  Download, 
  Mail, 
  MoreHorizontal, 
  Trash2, 
  UserCheck, 
  UserX,
  X 
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Badge } from '@/components/ui/badge'

interface BulkActionsToolbarProps<TData> {
  table: Table<TData>
  onDelete?: () => void
  onExport?: () => void
  onBulkAction?: (action: string) => void
}

export function BulkActionsToolbar<TData>({
  table,
  onDelete,
  onExport,
  onBulkAction,
}: BulkActionsToolbarProps<TData>) {
  const selectedCount = table.getFilteredSelectedRowModel().rows.length

  if (selectedCount === 0) return null

  return (
    <div className='flex items-center justify-between rounded-lg border bg-muted/50 px-4 py-2'>
      <div className='flex items-center gap-2'>
        <Badge variant='secondary' className='font-medium'>
          {selectedCount} selected
        </Badge>
        <span className='text-sm text-muted-foreground'>
          {selectedCount} of {table.getFilteredRowModel().rows.length} row(s) selected
        </span>
      </div>

      <div className='flex items-center gap-2'>
        <Button
          variant='outline'
          size='sm'
          onClick={() => table.resetRowSelection()}
        >
          <X className='h-4 w-4' />
          Clear
        </Button>

        {onExport && (
          <Button
            variant='outline'
            size='sm'
            onClick={onExport}
          >
            <Download className='h-4 w-4' />
            Export
          </Button>
        )}

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant='outline' size='sm'>
              Actions
              <MoreHorizontal className='ml-2 h-4 w-4' />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end' className='w-[200px]'>
            <DropdownMenuItem
              onClick={() => onBulkAction?.('activate')}
            >
              <UserCheck className='mr-2 h-4 w-4' />
              Activate users
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => onBulkAction?.('deactivate')}
            >
              <UserX className='mr-2 h-4 w-4' />
              Deactivate users
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => onBulkAction?.('email')}
            >
              <Mail className='mr-2 h-4 w-4' />
              Send email
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            {onDelete && (
              <DropdownMenuItem
                onClick={onDelete}
                className='text-destructive'
              >
                <Trash2 className='mr-2 h-4 w-4' />
                Delete users
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}