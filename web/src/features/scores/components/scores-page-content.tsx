'use client'

import { useSearchParams, useRouter } from 'next/navigation'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import { useProjectOnly } from '@/features/projects'
import { ScoresTable } from './scores-table'
import { ScoreAnalyticsDashboard } from './analytics'
import { useScoresQuery } from '../hooks/use-scores'

interface ScoresPageContentProps {
  projectSlug: string
}

export function ScoresPageContent({ projectSlug }: ScoresPageContentProps) {
  const searchParams = useSearchParams()
  const router = useRouter()
  const { currentProject, hasProject, isLoading: projectLoading } = useProjectOnly()

  const currentTab = searchParams.get('tab') || 'list'
  const page = Number(searchParams.get('page')) || 1
  const limit = Number(searchParams.get('limit')) || 50
  const name = searchParams.get('name') || undefined
  const source = searchParams.get('source') || undefined
  const dataType = searchParams.get('type') || undefined

  const {
    data: scoresResponse,
    isLoading: scoresLoading,
    error: scoresError,
  } = useScoresQuery(currentProject?.id, {
    page,
    limit,
    name,
    source: source as 'code' | 'llm' | 'human' | undefined,
    type: dataType as 'NUMERIC' | 'BOOLEAN' | 'CATEGORICAL' | undefined,
  })

  const handleTabChange = (value: string) => {
    const newParams = new URLSearchParams(searchParams.toString())
    newParams.set('tab', value)
    if (value !== 'list') {
      newParams.delete('page')
      newParams.delete('limit')
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
          />
        </TabsContent>

        <TabsContent value="analytics" className="mt-6">
          <ScoreAnalyticsDashboard projectId={currentProject.id} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
