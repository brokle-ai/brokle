'use client'

import { Line, LineChart as RechartsLineChart, ResponsiveContainer, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts'
import { BaseChart } from './base-chart'

interface DataPoint {
  [key: string]: string | number
}

interface LineChartProps {
  data: DataPoint[]
  xKey: string
  yKey: string | string[]
  title?: string
  description?: string
  loading?: boolean
  error?: string
  height?: number
  className?: string
  colors?: string[]
  showGrid?: boolean
  showTooltip?: boolean
  showLegend?: boolean
  formatYAxis?: (value: any) => string
  formatTooltip?: (value: any, name: string) => [string, string]
  onDataPointClick?: (data: DataPoint) => void
}

const defaultColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6']

export function LineChart({
  data,
  xKey,
  yKey,
  title,
  description,
  loading = false,
  error,
  height = 300,
  className,
  colors = defaultColors,
  showGrid = true,
  showTooltip = true,
  showLegend = false,
  formatYAxis,
  formatTooltip,
  onDataPointClick
}: LineChartProps) {
  const yKeys = Array.isArray(yKey) ? yKey : [yKey]

  return (
    <BaseChart
      title={title}
      description={description}
      loading={loading}
      error={error}
      height={height}
      className={className}
    >
      <ResponsiveContainer width="100%" height="100%">
        <RechartsLineChart
          data={data}
          onClick={onDataPointClick}
        >
          {showGrid && (
            <CartesianGrid 
              strokeDasharray="3 3" 
              className="stroke-muted" 
            />
          )}
          <XAxis
            dataKey={xKey}
            fontSize={12}
            tickLine={false}
            axisLine={false}
            className="fill-muted-foreground"
          />
          <YAxis
            fontSize={12}
            tickLine={false}
            axisLine={false}
            className="fill-muted-foreground"
            tickFormatter={formatYAxis}
          />
          {showTooltip && (
            <Tooltip
              contentStyle={{
                background: 'hsl(var(--background))',
                border: '1px solid hsl(var(--border))',
                borderRadius: '6px',
              }}
              formatter={formatTooltip}
            />
          )}
          {showLegend && <Legend />}
          {yKeys.map((key, index) => (
            <Line
              key={key}
              type="monotone"
              dataKey={key}
              stroke={colors[index % colors.length]}
              strokeWidth={2}
              dot={{ r: 4 }}
              activeDot={{ r: 6 }}
            />
          ))}
        </RechartsLineChart>
      </ResponsiveContainer>
    </BaseChart>
  )
}