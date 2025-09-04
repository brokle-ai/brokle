'use client'

import { PieChart as RechartsPieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts'
import { BaseChart } from './base-chart'

interface DataPoint {
  name: string
  value: number
  [key: string]: string | number
}

interface PieChartProps {
  data: DataPoint[]
  title?: string
  description?: string
  loading?: boolean
  error?: string
  height?: number
  className?: string
  colors?: string[]
  showTooltip?: boolean
  showLegend?: boolean
  showLabels?: boolean
  innerRadius?: number
  outerRadius?: number
  formatTooltip?: (value: any, name: string) => [string, string]
  onDataPointClick?: (data: DataPoint) => void
}

const defaultColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#84cc16', '#f97316']

export function PieChart({
  data,
  title,
  description,
  loading = false,
  error,
  height = 300,
  className,
  colors = defaultColors,
  showTooltip = true,
  showLegend = true,
  showLabels = false,
  innerRadius = 0,
  outerRadius = 80,
  formatTooltip,
  onDataPointClick
}: PieChartProps) {
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
        <RechartsPieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            labelLine={false}
            label={showLabels ? ({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%` : false}
            outerRadius={outerRadius}
            innerRadius={innerRadius}
            fill="#8884d8"
            dataKey="value"
            onClick={onDataPointClick}
          >
            {data.map((entry, index) => (
              <Cell 
                key={`cell-${index}`} 
                fill={colors[index % colors.length]} 
              />
            ))}
          </Pie>
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
        </RechartsPieChart>
      </ResponsiveContainer>
    </BaseChart>
  )
}