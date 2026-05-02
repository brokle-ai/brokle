import { useMemo } from 'react'
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import type { CostByModel } from '../api/types'

interface CostByModelChartProps {
  data: CostByModel[]
  className?: string
}

function formatCost(value: number): string {
  if (value < 0.01) return '$0.00'
  if (value < 1) return `$${value.toFixed(3)}`
  if (value < 100) return `$${value.toFixed(2)}`
  return `$${value.toFixed(0)}`
}

function truncateModel(name: string, max = 18): string {
  return name.length <= max ? name : `${name.slice(0, max - 1)}…`
}

export function CostByModelChart({ data, className }: CostByModelChartProps) {
  // Horizontal layout — models on Y, cost on X. Recharts spells this
  // with `layout="vertical"` (the axis labelling is swapped).
  const rows = useMemo(
    () =>
      data.map((d) => ({
        model: truncateModel(d.model),
        fullName: d.model,
        cost: d.cost,
      })),
    [data],
  )

  return (
    <Card className={className}>
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-medium">Cost by model</CardTitle>
      </CardHeader>
      <CardContent>
        {rows.length === 0 ? (
          <div className="flex h-[200px] items-center justify-center text-sm text-muted-foreground">
            No cost data available.
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={Math.max(200, rows.length * 32)}>
            <BarChart
              data={rows}
              layout="vertical"
              margin={{ top: 8, right: 16, bottom: 0, left: 16 }}
            >
              <CartesianGrid
                strokeDasharray="3 3"
                className="stroke-muted"
                horizontal={false}
              />
              <XAxis
                type="number"
                tick={{ fontSize: 11 }}
                stroke="currentColor"
                className="text-muted-foreground"
                tickFormatter={formatCost}
              />
              <YAxis
                type="category"
                dataKey="model"
                tick={{ fontSize: 11 }}
                stroke="currentColor"
                className="text-muted-foreground"
                width={120}
              />
              <Tooltip
                contentStyle={{ fontSize: 12, borderRadius: 8 }}
                formatter={(value) => [
                  formatCost(typeof value === 'number' ? value : 0),
                  'Cost',
                ]}
              />
              <Bar dataKey="cost" fill="#10b981" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
