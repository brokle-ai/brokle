'use client'

import * as React from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { BarChart } from '@/components/shared/charts/bar-chart'
import { ArrowRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { CostByModel } from '../types'

interface CostByModelChartProps {
  data: CostByModel[]
  isLoading?: boolean
  error?: string | null
  className?: string
}

function formatCost(value: number): string {
  if (value < 0.01) return '$0.00'
  if (value < 1) return `$${value.toFixed(3)}`
  if (value < 100) return `$${value.toFixed(2)}`
  return `$${value.toFixed(0)}`
}

function truncateModelName(name: string, maxLength: number = 20): string {
  if (name.length <= maxLength) return name
  return name.substring(0, maxLength - 3) + '...'
}

export function CostByModelChart({
  data,
  isLoading,
  error,
  className,
}: CostByModelChartProps) {
  const chartData = React.useMemo(() => {
    return data.map((item) => ({
      model: truncateModelName(item.model),
      cost: item.cost,
      fullName: item.model,
    }))
  }, [data])

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-base font-medium">Cost by Model</CardTitle>
      </CardHeader>
      <CardContent>
        {data.length === 0 && !isLoading ? (
          <div className="h-[200px] flex items-center justify-center text-muted-foreground">
            No cost data available. Costs are calculated from token usage.
          </div>
        ) : (
          <BarChart
            data={chartData}
            xKey="model"
            yKey="cost"
            height={200}
            loading={isLoading}
            error={error ?? undefined}
            colors={['#10b981']}
            formatYAxis={(value) => formatCost(value)}
            formatTooltip={(value, name) => [formatCost(value as number), 'Cost']}
          />
        )}
      </CardContent>
    </Card>
  )
}
