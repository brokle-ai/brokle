import { useCallback } from 'react'
import { useNavigate } from '@tanstack/react-router'
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
  orgId: string
  projectId: string
  experimentIds: string[]
  baselineId: string | undefined
  viewMode: ComparisonViewMode
}

export function ExperimentCompareView({
  orgId,
  projectId,
  experimentIds,
  baselineId,
  viewMode,
}: ExperimentCompareViewProps) {
  const navigate = useNavigate()

  const updateSearch = useCallback(
    (
      ids: string[],
      baseline: string | undefined,
      view: ComparisonViewMode,
    ) => {
      void navigate({
        to: '/o/$orgId/p/$projectId/experiments/compare',
        params: { orgId, projectId },
        // Compare owns its own search schema; parent list keys
        // (page/limit/q) are not re-serialised when the user is on
        // the compare route.
        search: {
          ids: ids.length > 0 ? ids.join(',') : undefined,
          baseline,
          view: view === 'card' ? undefined : view,
          page: 1,
          limit: 20,
          q: undefined,
        },
        replace: true,
      })
    },
    [navigate, orgId, projectId],
  )

  const handleSelectionChange = useCallback(
    (ids: string[]) => {
      // If the current baseline is no longer in the selection, fall
      // back to the first selected experiment.
      const newBaseline =
        baselineId && ids.includes(baselineId) ? baselineId : ids[0]
      updateSearch(ids, newBaseline, viewMode)
    },
    [baselineId, viewMode, updateSearch],
  )

  const handleBaselineChange = useCallback(
    (id: string) => {
      updateSearch(experimentIds, id, viewMode)
    },
    [experimentIds, viewMode, updateSearch],
  )

  const handleViewModeChange = useCallback(
    (mode: ComparisonViewMode) => {
      updateSearch(experimentIds, baselineId, mode)
    },
    [experimentIds, baselineId, updateSearch],
  )

  const handleShare = useCallback(() => {
    void navigator.clipboard.writeText(window.location.href)
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
          <ComparisonViewToggle
            value={viewMode}
            onChange={handleViewModeChange}
          />
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

      {/* Summary counters — only show when data is loaded and a baseline
          is active. */}
      {!isLoading && !error && scoreRows.length > 0 && baselineId && (
        <ComparisonSummary
          scoreRows={scoreRows}
          baselineId={baselineId}
        />
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
