'use client'

import { useSearchParams, useRouter } from 'next/navigation'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import { useProjectOnly } from '@/features/projects'
import { ScoresTable } from './scores-table'
import { ScoreAnalyticsDashboard } from './analytics'
import { useScoresQuery } from '../hooks/use-scores'
import { useScoresTableState } from '../hooks/use-scores-table-state'

interface ScoresPageContentProps {
  projectSlug: string
}

export function ScoresPageContent({ projectSlug }: ScoresPageContentProps) {
  const searchParams = useSearchParams()
  const router = useRouter()
  const { currentProject, hasProject, isLoading: projectLoading } = useProjectOnly()

  // Use the centralized URL state hook
  const {
    page,
    pageSize,
    search,
    dataType,
    source,
    sortBy,
    sortOrder,
    setSearch,
    setDataType,
    setSource,
    setPagination,
    setSorting,
    resetAll,
    hasActiveFilters,
    toApiParams,
  } = useScoresTableState()

  const currentTab = searchParams.get('tab') || 'list'

  // Query scores with params from URL state
  const {
    data: scoresResponse,
    isLoading: scoresLoading,
    error: scoresError,
  } = useScoresQuery(currentProject?.id, toApiParams())

  const handleTabChange = (value: string) => {
    const newParams = new URLSearchParams(searchParams.toString())
    newParams.set('tab', value)
    if (value !== 'list') {
      // Clear table-specific params when switching away from list
      newParams.delete('page')
      newParams.delete('pageSize')
      newParams.delete('search')
      newParams.delete('dataType')
      newParams.delete('source')
      newParams.delete('sortBy')
      newParams.delete('sortOrder')
    }
    router.push(`?${newParams.toString()}`)
  }

  if (projectLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <Skeleton className="h-8 w-32 mb-2" />
            <Skeleton className="h-4 w-64" />
          </div>
        </div>
        <div className="space-y-4">
          <Skeleton className="h-10 w-[200px]" />
          <Skeleton className="h-[400px]" />
        </div>
      </div>
    )
  }

  if (!hasProject || !currentProject) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">No project selected</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Scores</h1>
          <p className="text-muted-foreground">
            View and analyze quality scores from evaluations
          </p>
        </div>
      </div>

      <Tabs value={currentTab} onValueChange={handleTabChange}>
        <TabsList>
          <TabsTrigger value="list">List</TabsTrigger>
          <TabsTrigger value="analytics">Analytics</TabsTrigger>
        </TabsList>

        <TabsContent value="list" className="mt-6">
          <ScoresTable
            data={scoresResponse?.data ?? []}
            pagination={
              scoresResponse?.pagination ?? {
                page: 1,
                limit: 50,
                total: 0,
                totalPages: 0,
                hasNext: false,
                hasPrev: false,
              }
            }
            projectSlug={projectSlug}
            loading={scoresLoading}
            error={scoresError?.message}
            // URL state
            search={search}
            dataType={dataType}
            source={source}
            sortBy={sortBy}
            sortOrder={sortOrder}
            // State setters
            onSearchChange={setSearch}
            onDataTypeChange={setDataType}
            onSourceChange={setSource}
            onPageChange={setPagination}
            onSortChange={setSorting}
            onReset={resetAll}
            hasActiveFilters={hasActiveFilters}
          />
        </TabsContent>

        <TabsContent value="analytics" className="mt-6">
          <ScoreAnalyticsDashboard projectId={currentProject.id} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
