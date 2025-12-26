'use client'

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Heatmap } from './heatmap'
import type { HeatmapCell } from '../../types'

interface HeatmapCardProps {
  heatmap?: HeatmapCell[]
  scoreName: string
  compareScoreName?: string
}

export function HeatmapCard({
  heatmap,
  scoreName,
  compareScoreName,
}: HeatmapCardProps) {
  const hasComparison = !!compareScoreName
  const hasData = heatmap && heatmap.length > 0

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-medium">Score Correlation Heatmap</CardTitle>
        {hasComparison && (
          <CardDescription>
            Distribution of score pairs between {scoreName} and {compareScoreName}
          </CardDescription>
        )}
        {!hasComparison && (
          <CardDescription>
            Select a comparison score to view the correlation heatmap
          </CardDescription>
        )}
      </CardHeader>
      <CardContent>
        {hasData && hasComparison ? (
          <Heatmap
            cells={heatmap}
            xLabel={compareScoreName}
            yLabel={scoreName}
            gridSize={10}
          />
        ) : (
          <div className="flex items-center justify-center h-[300px] text-muted-foreground">
            {hasComparison
              ? 'No overlapping data between the two scores'
              : 'Select a second score to compare'}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
