'use client'

import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { type Task } from '../data/schema'

type TasksMutateDrawerProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow?: Task
}

export function TasksMutateDrawer({ 
  open, 
  onOpenChange, 
  currentRow 
}: TasksMutateDrawerProps) {
  const isEdit = !!currentRow
  
  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className='sm:max-w-[540px]'>
        <SheetHeader>
          <SheetTitle>
            {isEdit ? 'Edit Task' : 'Create Task'}
          </SheetTitle>
          <SheetDescription>
            {isEdit ? 'Update the task details.' : 'Create a new task.'}
          </SheetDescription>
        </SheetHeader>
        <div className='py-4'>
          <p className='text-muted-foreground'>
            This feature is coming soon.
          </p>
        </div>
        <SheetFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button disabled>
            {isEdit ? 'Update' : 'Create'}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}