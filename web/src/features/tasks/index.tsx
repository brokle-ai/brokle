'use client'

import { TasksDialogs } from './components/tasks-dialogs'
import { TasksPrimaryButtons } from './components/tasks-primary-buttons'
import { TasksProvider } from './context/tasks-context'
import { TasksTable } from './components/tasks-table'
import { useProjectTasks } from './hooks/use-project-tasks'
import { PageHeader } from '@/components/layout/page-header'

interface TasksProps {
  projectSlug?: string
}

function TasksContent() {
  const { data, totalCount, isLoading } = useProjectTasks()

  return (
    <>
      <PageHeader title="Tasks">
        <TasksPrimaryButtons />
      </PageHeader>
      <div className='-mx-4 flex-1 overflow-auto px-4 py-1 lg:flex-row lg:space-y-0 lg:space-x-12'>
        {isLoading ? (
          <div className='flex items-center justify-center py-8 text-muted-foreground'>
            Loading tasks...
          </div>
        ) : (
          <TasksTable data={data} totalCount={totalCount} />
        )}
      </div>

      <TasksDialogs />
    </>
  )
}

export function Tasks({ projectSlug }: TasksProps) {
  return (
    <TasksProvider projectSlug={projectSlug}>
      <TasksContent />
    </TasksProvider>
  )
}