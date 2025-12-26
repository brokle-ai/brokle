'use client'

import { useMemo } from 'react'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import type { HeatmapCell } from '../../types'
import { getHeatmapCellColor, getContrastTextColor, SCORE_COLORS } from '../../lib/color-scales'

interface HeatmapProps {
  cells: HeatmapCell[]
  xLabel: string
  yLabel: string
  gridSize?: number
}

export function Heatmap({
  cells,
  xLabel,
  yLabel,
  gridSize = 10,
}: HeatmapProps) {
  const cellMap = useMemo(() => {
    const map = new Map<string, HeatmapCell>()
    cells.forEach((cell) => {
      map.set(`${cell.row}-${cell.col}`, cell)
    })
    return map
  }, [cells])

  const { minValue, maxValue, totalValue } = useMemo(() => {
    if (cells.length === 0) {
      return { minValue: 0, maxValue: 0, totalValue: 0 }
    }
    const values = cells.map((c) => c.value)
    return {
      minValue: Math.min(...values),
      maxValue: Math.max(...values),
      totalValue: values.reduce((sum, v) => sum + v, 0),
    }
  }, [cells])

  const rowLabels = useMemo(() => {
    const labels: string[] = []
    for (let i = 0; i < gridSize; i++) {
      const cell = cells.find((c) => c.row === i)
      labels.push(cell?.row_label ?? i.toString())
    }
    return labels
  }, [cells, gridSize])

  const colLabels = useMemo(() => {
    const labels: string[] = []
    for (let i = 0; i < gridSize; i++) {
      const cell = cells.find((c) => c.col === i)
      labels.push(cell?.col_label ?? i.toString())
    }
    return labels
  }, [cells, gridSize])

  if (cells.length === 0) {
    return (
      <div className="flex items-center justify-center h-[300px] text-muted-foreground">
        No heatmap data available. Select two scores to compare.
      </div>
    )
  }

  return (
    <div className="flex flex-col">
      {/* Y-axis label */}
      <div className="flex items-center gap-2 mb-2">
        <span className="text-xs text-muted-foreground transform -rotate-90 origin-center whitespace-nowrap">
          {yLabel}
        </span>
      </div>

      <div className="flex">
        {/* Y-axis labels */}
        <div className="flex flex-col justify-around pr-2">
          {rowLabels.map((label, i) => (
            <div
              key={i}
              className="text-xs text-muted-foreground text-right h-7 flex items-center justify-end"
              title={label}
            >
              {label.length > 6 ? `${label.slice(0, 4)}..` : label}
            </div>
          ))}
        </div>

        {/* Heatmap grid */}
        <div
          role="grid"
          aria-label={`Comparison heatmap of ${yLabel} vs ${xLabel}`}
          className="grid gap-0.5"
          style={{
            gridTemplateColumns: `repeat(${gridSize}, 1fr)`,
            gridTemplateRows: `repeat(${gridSize}, 1fr)`,
          }}
        >
          {Array.from({ length: gridSize * gridSize }).map((_, index) => {
            const row = Math.floor(index / gridSize)
            const col = index % gridSize
            const cell = cellMap.get(`${row}-${col}`)
            const value = cell?.value ?? 0
            const percentage = totalValue > 0 ? (value / totalValue) * 100 : 0

            const backgroundColor = getHeatmapCellColor(
              value,
              minValue,
              maxValue,
              SCORE_COLORS.primary
            )

            const lightnessMatch = backgroundColor.match(/oklch\((\d+(?:\.\d+)?)%/)
            const lightness = lightnessMatch ? parseFloat(lightnessMatch[1]) : 50
            const textColor = getContrastTextColor(lightness)

            return (
              <TooltipProvider key={index}>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div
                      role="gridcell"
                      tabIndex={0}
                      aria-label={`${yLabel} bin ${cell?.row_label ?? rowLabels[row]}, ${xLabel} bin ${cell?.col_label ?? colLabels[col]}: ${value} items, ${percentage.toFixed(1)}% of total`}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault()
                        }
                      }}
                      className="w-7 h-7 rounded-sm cursor-pointer hover:ring-2 hover:ring-primary hover:ring-offset-1 transition-all flex items-center justify-center focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-1"
                      style={{ backgroundColor }}
                    >
                      {value > 0 && (
                        <span
                          className="text-[10px] font-medium"
                          style={{ color: textColor }}
                        >
                          {value > 99 ? '99+' : value}
                        </span>
                      )}
                    </div>
                  </TooltipTrigger>
                  <TooltipContent>
                    <div className="text-sm">
                      <p className="font-medium">
                        {xLabel}: {cell?.col_label ?? colLabels[col]}
                      </p>
                      <p className="font-medium">
                        {yLabel}: {cell?.row_label ?? rowLabels[row]}
                      </p>
                      <p className="text-muted-foreground">
                        Count: {value.toLocaleString()}
                      </p>
                      <p className="text-muted-foreground">
                        {percentage.toFixed(1)}% of total
                      </p>
                    </div>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            )
          })}
        </div>
      </div>

      {/* X-axis labels */}
      <div className="flex mt-2 ml-8">
        <div
          className="grid gap-0.5 w-full"
          style={{ gridTemplateColumns: `repeat(${gridSize}, 1fr)` }}
        >
          {colLabels.map((label, i) => (
            <div
              key={i}
              className="text-xs text-muted-foreground text-center w-7 truncate"
              title={label}
            >
              {label.length > 4 ? `${label.slice(0, 3)}..` : label}
            </div>
          ))}
        </div>
      </div>

      {/* X-axis label */}
      <div className="text-xs text-muted-foreground text-center mt-1">
        {xLabel}
      </div>

      {/* Legend */}
      <div className="flex items-center justify-center gap-2 mt-4">
        <span className="text-xs text-muted-foreground">Low</span>
        <div className="flex h-3">
          {Array.from({ length: 10 }).map((_, i) => (
            <div
              key={i}
              className="w-4"
              style={{
                backgroundColor: getHeatmapCellColor(
                  i,
                  0,
                  9,
                  SCORE_COLORS.primary
                ),
              }}
            />
          ))}
        </div>
        <span className="text-xs text-muted-foreground">High</span>
      </div>
    </div>
  )
}
