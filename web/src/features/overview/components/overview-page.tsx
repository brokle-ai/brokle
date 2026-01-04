'use client'

import * as React from 'react'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/layout/page-header'
import { TimeRangePicker } from '@/components/shared/time-range-picker'
import { cn } from '@/lib/utils'

import { useProjectOverview, useOverviewTimeRange } from '../hooks'
import { OnboardingChecklist } from './onboarding-checklist'
import { StatsRow } from './stats-row'
import { TraceVolumeChart } from './trace-volume-chart'
import { CostByModelChart } from './cost-by-model-chart'
import { RecentTracesTable } from './recent-traces-table'
import { TopErrorsTable } from './top-errors-table'
import { ScoreOverview } from './score-overview'

interface OverviewPageProps {
  projectId: string
  projectSlug: string
  className?: string
}

export function OverviewPage({ projectId, projectSlug, className }: OverviewPageProps) {
  // URL-persisted time range
  const { timeRange, setTimeRange } = useOverviewTimeRange()

  const {
    data,
    isLoading,
    isRefetching,
    error,
    refetch,
    onboardingProgress,
    isOnboardingDismissed,
    dismissOnboarding,
  } = useProjectOverview(projectId, { timeRange })

  const handleRefresh = () => {
    refetch()
  }

  const errorMessage = error ? (typeof error === 'object' && 'message' in error ? error.message : String(error)) : null

  return (
    <div className={cn('space-y-6', className)}>
      <PageHeader title="Overview">
        <div className="flex items-center gap-2">
          <TimeRangePicker value={timeRange} onChange={setTimeRange} />
          <Button
            variant="outline"
            size="icon"
            onClick={handleRefresh}
            disabled={isLoading || isRefetching}
          >
            <RefreshCw
              className={cn('h-4 w-4', (isLoading || isRefetching) && 'animate-spin')}
            />
          </Button>
        </div>
      </PageHeader>

      {/* Onboarding Checklist - shown for new projects */}
      {!isOnboardingDismissed && (
        <OnboardingChecklist
          checklistStatus={data?.checklist_status ?? null}
          onboardingProgress={onboardingProgress}
          onDismiss={dismissOnboarding}
          projectSlug={projectSlug}
        />
      )}

      {/* Stats Row */}
      <StatsRow
        stats={data?.stats ?? null}
        isLoading={isLoading}
        error={errorMessage}
      />

      {/* Charts Row */}
      <div className="grid gap-4 md:grid-cols-2">
        <TraceVolumeChart
          data={data?.trace_volume ?? []}
          timeRange={timeRange}
          isLoading={isLoading}
          error={errorMessage}
          projectSlug={projectSlug}
        />
        <CostByModelChart
          data={data?.cost_by_model ?? []}
          isLoading={isLoading}
          error={errorMessage}
        />
      </div>

      {/* Tables Row */}
      <div className="grid gap-4 md:grid-cols-2">
        <RecentTracesTable
          data={data?.recent_traces ?? []}
          isLoading={isLoading}
          error={errorMessage}
          projectSlug={projectSlug}
        />
        <TopErrorsTable
          data={data?.top_errors ?? []}
          isLoading={isLoading}
          error={errorMessage}
          projectSlug={projectSlug}
        />
      </div>

      {/* Score Overview - conditional */}
      <ScoreOverview
        data={data?.scores_summary ?? null}
        isLoading={isLoading}
        error={errorMessage}
        projectSlug={projectSlug}
      />
    </div>
  )
}
