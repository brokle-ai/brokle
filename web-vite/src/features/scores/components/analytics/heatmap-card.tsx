import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import type { HeatmapCell } from '../../api/types'
import { Heatmap } from './heatmap'

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
        <CardTitle className="text-base font-medium">
          Score Correlation Heatmap
        </CardTitle>
        {hasComparison ? (
          <CardDescription>
            Distribution of score pairs between {scoreName} and{' '}
            {compareScoreName}
          </CardDescription>
        ) : (
          <CardDescription>
            Select a comparison score to view the correlation heatmap
          </CardDescription>
        )}
      </CardHeader>
      <CardContent>
        {hasData && hasComparison && compareScoreName ? (
          <Heatmap
            cells={heatmap}
            xLabel={compareScoreName}
            yLabel={scoreName}
            gridSize={10}
          />
        ) : (
          <div className="flex h-[300px] items-center justify-center text-muted-foreground">
            {hasComparison
              ? 'No overlapping data between the two scores'
              : 'Select a second score to compare'}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
