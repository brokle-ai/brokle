import { useMemo } from 'react'
import { format, parseISO } from 'date-fns'
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { TimeSeriesPoint } from '../../api/types'
import { CHART_COLORS } from '../../lib/color-scales'

interface TimelineChartCardProps {
  timeSeries: TimeSeriesPoint[]
  compareTimeSeries?: TimeSeriesPoint[]
  scoreName: string
  compareScoreName?: string
  mean?: number
  compareMean?: number
}

interface ChartDataPoint {
  date: string
  formattedDate: string
  value: number | null
  compareValue: number | null
  count: number
  compareCount: number
}

export function TimelineChartCard({
  timeSeries,
  compareTimeSeries,
  scoreName,
  compareScoreName,
  mean,
  compareMean,
}: TimelineChartCardProps) {
  const chartData = useMemo(() => {
    const dateMap = new Map<string, ChartDataPoint>()

    timeSeries.forEach((point) => {
      const dateKey = point.timestamp.split('T')[0]
      dateMap.set(dateKey, {
        date: dateKey,
        formattedDate: format(parseISO(point.timestamp), 'MMM d'),
        value: point.avg_value,
        compareValue: null,
        count: point.count,
        compareCount: 0,
      })
    })

    if (compareTimeSeries) {
      compareTimeSeries.forEach((point) => {
        const dateKey = point.timestamp.split('T')[0]
        const existing = dateMap.get(dateKey)
        if (existing) {
          existing.compareValue = point.avg_value
          existing.compareCount = point.count
        } else {
          dateMap.set(dateKey, {
            date: dateKey,
            formattedDate: format(parseISO(point.timestamp), 'MMM d'),
            value: null,
            compareValue: point.avg_value,
            count: 0,
            compareCount: point.count,
          })
        }
      })
    }

    return Array.from(dateMap.values()).sort((a, b) =>
      a.date.localeCompare(b.date),
    )
  }, [timeSeries, compareTimeSeries])

  if (chartData.length === 0) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-medium">
            Score Over Time
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex h-[300px] items-center justify-center text-muted-foreground">
            No time series data available
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-medium">
          Score Over Time
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-[300px]">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart
              data={chartData}
              margin={{ top: 5, right: 20, left: 10, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis
                dataKey="formattedDate"
                tick={{ fontSize: 12 }}
                tickLine={false}
                axisLine={false}
                className="text-muted-foreground"
              />
              <YAxis
                tick={{ fontSize: 12 }}
                tickLine={false}
                axisLine={false}
                className="text-muted-foreground"
                domain={['auto', 'auto']}
              />
              <Tooltip
                content={({ active, payload, label }) => {
                  if (!active || !payload?.length) return null
                  return (
                    <div className="rounded-lg border bg-popover p-3 shadow-lg">
                      <p className="mb-2 text-sm font-medium">{label}</p>
                      {payload.map((entry, index) => (
                        <div
                          key={index}
                          className="flex items-center gap-2 text-sm"
                        >
                          <div
                            className="h-3 w-3 rounded-full"
                            style={{ backgroundColor: entry.color }}
                          />
                          <span className="text-muted-foreground">
                            {entry.name}:
                          </span>
                          <span className="font-medium">
                            {typeof entry.value === 'number'
                              ? entry.value.toFixed(4)
                              : 'N/A'}
                          </span>
                        </div>
                      ))}
                    </div>
                  )
                }}
              />
              <Legend />
              <Line
                type="monotone"
                dataKey="value"
                name={scoreName}
                stroke={CHART_COLORS.series1}
                strokeWidth={2}
                dot={{ r: 3 }}
                activeDot={{ r: 5 }}
                connectNulls
              />
              {compareScoreName && (
                <Line
                  type="monotone"
                  dataKey="compareValue"
                  name={compareScoreName}
                  stroke={CHART_COLORS.series2}
                  strokeWidth={2}
                  dot={{ r: 3 }}
                  activeDot={{ r: 5 }}
                  connectNulls
                />
              )}
              {mean !== undefined && (
                <ReferenceLine
                  y={mean}
                  stroke={CHART_COLORS.series1}
                  strokeDasharray="5 5"
                  strokeOpacity={0.5}
                  label={{
                    value: `Mean: ${mean.toFixed(2)}`,
                    position: 'insideTopRight',
                    fontSize: 10,
                    fill: CHART_COLORS.series1,
                  }}
                />
              )}
              {compareMean !== undefined && compareScoreName && (
                <ReferenceLine
                  y={compareMean}
                  stroke={CHART_COLORS.series2}
                  strokeDasharray="5 5"
                  strokeOpacity={0.5}
                />
              )}
            </LineChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}
