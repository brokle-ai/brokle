'use client'

import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { Row } from '@tanstack/react-table'
import { MoreHorizontal, Eye, Edit, Trash2, Copy } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface RowAction {
  label: string
  icon?: React.ComponentType<{ className?: string }>
  action: string
  variant?: 'default' | 'destructive'
  disabled?: boolean
}

interface DataTableRowActionsProps<TData> {
  row: Row<TData>
  actions?: RowAction[]
  onRowAction?: (action: string, row: TData) => void
}

const defaultActions: RowAction[] = [
  {
    label: 'View',
    icon: Eye,
    action: 'view',
  },
  {
    label: 'Edit',
    icon: Edit,
    action: 'edit',
  },
  {
    label: 'Copy',
    icon: Copy,
    action: 'copy',
  },
  {
    label: 'Delete',
    icon: Trash2,
    action: 'delete',
    variant: 'destructive',
  },
]

export function DataTableRowActions<TData>({
  row,
  actions = defaultActions,
  onRowAction,
}: DataTableRowActionsProps<TData>) {
  const handleAction = (action: string) => {
    
    if (onRowAction) {
      onRowAction(action, row.original)
    }
  }

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          className='data-[state=open]:bg-muted flex h-8 w-8 p-0'
          onClick={(e) => e.stopPropagation()}
        >
          <MoreHorizontal className='h-4 w-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[160px]'>
        {actions.map((action, index) => {
          const isDestructive = action.variant === 'destructive'
          const Icon = action.icon
          
          return (
            <div key={action.action}>
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation() // Prevent row click when clicking menu items
                  handleAction(action.action)
                }}
                disabled={action.disabled}
                className={isDestructive ? 'text-destructive focus:text-destructive' : ''}
              >
                {Icon && <Icon className='mr-2 h-4 w-4' />}
                {action.label}
              </DropdownMenuItem>
              {index < actions.length - 1 && isDestructive && (
                <DropdownMenuSeparator />
              )}
            </div>
          )
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}