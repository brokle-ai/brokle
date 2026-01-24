'use client'

import { useCallback, useMemo } from 'react'
import { useSearchParams, useRouter, usePathname } from 'next/navigation'
import { Share2, Loader2, GitCompare } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { ExperimentSelector } from './experiment-selector'
import { ScoreComparisonCard } from './score-comparison-card'
import {
  ComparisonViewToggle,
  ComparisonSummary,
  ComparisonTable,
  type ComparisonViewMode,
} from './comparison'
import { useExperimentComparisonQuery } from '../hooks/use-experiment-comparison'

interface ExperimentCompareViewProps {
  projectId: string
}

export function ExperimentCompareView({ projectId }: ExperimentCompareViewProps) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const experimentIds = useMemo(() => {
    const ids = searchParams.get('ids')
    return ids ? ids.split(',').filter(Boolean) : []
  }, [searchParams])

  const baselineId = searchParams.get('baseline') ?? undefined
  const viewMode = (searchParams.get('view') as ComparisonViewMode) ?? 'card'

  const updateURL = useCallback(
    (ids: string[], baseline?: string, view?: ComparisonViewMode) => {
      const params = new URLSearchParams()
      if (ids.length > 0) params.set('ids', ids.join(','))
      if (baseline) params.set('baseline', baseline)
      if (view && view !== 'card') params.set('view', view)
      router.replace(`${pathname}?${params.toString()}`)
    },
    [router, pathname]
  )

  const handleSelectionChange = useCallback(
    (ids: string[]) => {
      // If baseline is removed from selection, clear it
      const newBaseline =
        baselineId && ids.includes(baselineId) ? baselineId : ids[0]
      updateURL(ids, newBaseline, viewMode)
    },
    [baselineId, viewMode, updateURL]
  )

  const handleBaselineChange = useCallback(
    (id: string) => {
      updateURL(experimentIds, id, viewMode)
    },
    [experimentIds, viewMode, updateURL]
  )

  const handleViewModeChange = useCallback(
    (mode: ComparisonViewMode) => {
      updateURL(experimentIds, baselineId, mode)
    },
    [experimentIds, baselineId, updateURL]
  )

  const handleShare = useCallback(() => {
    navigator.clipboard.writeText(window.location.href)
    toast.success('Comparison link copied to clipboard')
  }, [])

  const { scoreRows, experiments, isLoading, error } =
    useExperimentComparisonQuery(projectId, experimentIds, baselineId)

  if (experimentIds.length < 2) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-3">
          <GitCompare className="h-8 w-8 text-muted-foreground" />
          <div>
            <h1 className="text-2xl font-bold">Compare Experiments</h1>
            <p className="text-muted-foreground mt-1">
              Select at least 2 experiments to compare their score metrics
            </p>
          </div>
        </div>

        <ExperimentSelector
          projectId={projectId}
          selectedIds={experimentIds}
          onSelectionChange={handleSelectionChange}
          minSelections={0}
          className="max-w-md"
        />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <GitCompare className="h-8 w-8 text-muted-foreground" />
          <div>
            <h1 className="text-2xl font-bold">Compare Experiments</h1>
            <p className="text-muted-foreground mt-1">
              Comparing {experimentIds.length} experiments
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <ComparisonViewToggle value={viewMode} onChange={handleViewModeChange} />
          <Button variant="outline" size="sm" onClick={handleShare}>
            <Share2 className="h-4 w-4 mr-2" />
            Share
          </Button>
        </div>
      </div>

      <ExperimentSelector
        projectId={projectId}
        selectedIds={experimentIds}
        onSelectionChange={handleSelectionChange}
        className="max-w-md"
      />

      {/* Summary counters - only show when we have data and baseline */}
      {!isLoading && !error && scoreRows.length > 0 && baselineId && (
        <ComparisonSummary scoreRows={scoreRows} baselineId={baselineId} />
      )}

      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      )}

      {error && (
        <div className="text-center py-12 text-destructive">
          Failed to load comparison data. Please try again.
        </div>
      )}

      {!isLoading && !error && scoreRows.length > 0 && (
        viewMode === 'table' ? (
          <ComparisonTable
            scoreRows={scoreRows}
            experiments={experiments}
            experimentIds={experimentIds}
            baselineId={baselineId}
            onBaselineChange={handleBaselineChange}
          />
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {scoreRows.map((row) => (
              <ScoreComparisonCard
                key={row.scoreName}
                row={row}
                experiments={experiments}
                experimentIds={experimentIds}
                baselineId={baselineId}
                onBaselineChange={handleBaselineChange}
              />
            ))}
          </div>
        )
      )}

      {!isLoading && !error && scoreRows.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          <GitCompare className="h-12 w-12 mx-auto mb-4 opacity-50" />
          <p>No scores found for the selected experiments.</p>
          <p className="text-sm mt-2">
            Run evaluations on your experiments to see comparison data.
          </p>
        </div>
      )}
    </div>
  )
}
