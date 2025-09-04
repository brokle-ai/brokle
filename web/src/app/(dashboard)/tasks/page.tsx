import { Suspense } from 'react'
import { Tasks } from '@/features/tasks'
import { Skeleton } from '@/components/ui/skeleton'

function TasksLoader() {
  return (
    <div className='space-y-4'>
      <div className='flex items-center justify-between'>
        <div className='space-y-2'>
          <Skeleton className='h-8 w-20' />
          <Skeleton className='h-4 w-96' />
        </div>
        <div className='flex gap-2'>
          <Skeleton className='h-10 w-20' />
          <Skeleton className='h-10 w-20' />
        </div>
      </div>
      <Skeleton className='h-96 w-full' />
    </div>
  )
}

export default function TasksPage() {
  return (
    <Suspense fallback={<TasksLoader />}>
      <Tasks />
    </Suspense>
  )
}