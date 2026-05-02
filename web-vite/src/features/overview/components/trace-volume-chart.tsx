import { useMemo } from 'react'
import {
  CartesianGrid,
  Line,
  LineChart,
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
import type { TimeSeriesPoint } from '../api/types'

interface TraceVolumeChartProps {
  data: TimeSeriesPoint[]
  className?: string
}

function formatHour(timestamp: string): string {
  const d = new Date(timestamp)
  if (Number.isNaN(d.getTime())) return timestamp
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

export function TraceVolumeChart({ data, className }: TraceVolumeChartProps) {
  const points = useMemo(
    () =>
      data.map((p) => ({
        time: formatHour(p.timestamp),
        traces: p.value,
      })),
    [data],
  )

  return (
    <Card className={className}>
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-medium">Trace volume</CardTitle>
      </CardHeader>
      <CardContent>
        {points.length === 0 ? (
          <div className="flex h-[200px] items-center justify-center text-sm text-muted-foreground">
            No trace data yet.
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={200}>
            <LineChart
              data={points}
              margin={{ top: 8, right: 8, bottom: 0, left: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis
                dataKey="time"
                tick={{ fontSize: 11 }}
                stroke="currentColor"
                className="text-muted-foreground"
              />
              <YAxis
                tick={{ fontSize: 11 }}
                stroke="currentColor"
                className="text-muted-foreground"
                allowDecimals={false}
              />
              <Tooltip
                contentStyle={{
                  fontSize: 12,
                  borderRadius: 8,
                }}
              />
              <Line
                type="monotone"
                dataKey="traces"
                stroke="#3b82f6"
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
