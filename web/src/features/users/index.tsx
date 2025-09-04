'use client'

import { useUsersColumns } from './components/users-columns'
import { UsersDialogs } from './components/users-dialogs'
import { UsersPrimaryButtons } from './components/users-primary-buttons'
import { UsersTable } from './components/users-table'
import UsersProvider from './context/users-context'
import { useUsersData } from '@/hooks/use-users-data'
import { UsersErrorBoundary } from '@/components/error-boundary/users-error-boundary'
import { Button } from '@/components/ui/button'
import { AlertCircle, HelpCircle, RefreshCw } from 'lucide-react'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'

function UsersLoading() {
  return (
    <div className="flex items-center justify-center min-h-[400px]">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      <span className="ml-3 text-muted-foreground">Loading users...</span>
    </div>
  )
}

function UsersError({ error, onRetry }: { error: string; onRetry: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] space-y-4">
      <div className="flex items-center space-x-2 text-destructive">
        <AlertCircle className="h-5 w-5" />
        <h3 className="font-semibold">Failed to load users</h3>
      </div>
      
      <div className="text-center space-y-2">
        <p className="text-sm text-muted-foreground">
          Unable to fetch user data from the server.
        </p>
        <details className="text-xs text-muted-foreground">
          <summary className="cursor-pointer hover:text-foreground">
            Error details
          </summary>
          <pre className="mt-2 p-2 bg-muted rounded text-left max-w-md overflow-auto">
            {error}
          </pre>
        </details>
      </div>

      <Button onClick={onRetry} variant="outline" className="space-x-2">
        <RefreshCw className="h-4 w-4" />
        <span>Try Again</span>
      </Button>
    </div>
  )
}

function UsersContent() {
  const { data, pagination, loading, error, refetch } = useUsersData()
  const columns = useUsersColumns()

  // Show error state only for critical errors, not loading states
  if (error && !loading) {
    return <UsersError error={error} onRetry={refetch} />
  }

  return (
    <div className="flex-1">
      <UsersTable 
        data={data} 
        columns={columns} 
        pagination={pagination}
        loading={loading}
      />
    </div>
  )
}

export default function Users() {
  return (
    <UsersErrorBoundary>
      <UsersProvider>
        <div className="mb-4 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <h2 className="text-2xl font-bold tracking-tight">Users</h2>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="sm" className="h-6 w-6 p-0">
                    <HelpCircle className="h-4 w-4 text-muted-foreground" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Manage user accounts, roles, and permissions</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          <UsersPrimaryButtons />
        </div>
        <UsersContent />

        <UsersDialogs />
      </UsersProvider>
    </UsersErrorBoundary>
  )
}