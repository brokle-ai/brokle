'use client'

import { Eye, Trash } from 'lucide-react'
import { type Row } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { useRouter, useParams } from 'next/navigation'
import type { Trace } from '../data/schema'
import { toast } from 'sonner'

type DataTableRowActionsProps<TData> = {
  row: Row<TData>
}

export function DataTableRowActions<TData>({
  row,
}: DataTableRowActionsProps<TData>) {
  const router = useRouter()
  const params = useParams()
  const projectSlug = params?.projectSlug as string
  const trace = row.original as Trace

  const handleViewDetail = () => {
    router.push(`/projects/${projectSlug}/traces/${trace.trace_id}`)
  }

  const handleDelete = () => {
    // TODO: Replace with actual API call
    toast.promise(
      new Promise((resolve) => setTimeout(resolve, 1000)),
      {
        loading: 'Deleting trace...',
        success: 'Trace deleted successfully',
        error: 'Failed to delete trace',
      }
    )
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          className='data-[state=open]:bg-muted flex size-8 p-0'
        >
          <DotsHorizontalIcon className='size-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-40'>
        <DropdownMenuItem onClick={handleViewDetail}>
          <Eye className='mr-2 size-4' />
          View Detail
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleDelete} className='text-destructive'>
          <Trash className='mr-2 size-4' />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
