'use client'

import { Bar, BarChart as RechartsBarChart, ResponsiveContainer, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts'
import { BaseChart } from './base-chart'

interface DataPoint {
  [key: string]: string | number
}

interface BarChartProps {
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

export function BarChart({
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
}: BarChartProps) {
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
        <RechartsBarChart
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
            <Bar
              key={key}
              dataKey={key}
              fill={colors[index % colors.length]}
              radius={[4, 4, 0, 0]}
            />
          ))}
        </RechartsBarChart>
      </ResponsiveContainer>
    </BaseChart>
  )
}