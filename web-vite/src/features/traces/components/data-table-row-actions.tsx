import { Eye, Trash, Database } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import type { TraceListItem } from '../api/types'

interface TracesRowActionsProps {
  trace: TraceListItem
  onViewDetail?: (trace: TraceListItem) => void
  onAddToDataset?: (trace: TraceListItem) => void
  onDelete?: (trace: TraceListItem) => void
}

/**
 * Per-row action menu. Callbacks are wired by the parent so the menu
 * stays framework-agnostic (no TanStack Router / TanStack Query usage
 * in here — the parent owns the `navigate` / mutation wiring).
 */
export function TracesRowActions({
  trace,
  onViewDetail,
  onAddToDataset,
  onDelete,
}: TracesRowActionsProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="data-[state=open]:bg-muted flex size-8 p-0"
          onClick={(e) => e.stopPropagation()}
        >
          <DotsHorizontalIcon className="size-4" />
          <span className="sr-only">Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="end"
        className="w-44"
        onClick={(e) => e.stopPropagation()}
      >
        {onViewDetail && (
          <DropdownMenuItem onClick={() => onViewDetail(trace)}>
            <Eye className="mr-2 size-4" />
            View Detail
          </DropdownMenuItem>
        )}
        {onAddToDataset && (
          <DropdownMenuItem onClick={() => onAddToDataset(trace)}>
            <Database className="mr-2 size-4" />
            Add to Dataset
          </DropdownMenuItem>
        )}
        {onDelete && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => onDelete(trace)}
              className="text-destructive"
            >
              <Trash className="mr-2 size-4" />
              Delete
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
