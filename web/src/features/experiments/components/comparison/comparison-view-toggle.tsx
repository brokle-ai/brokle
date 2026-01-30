'use client'

import { LayoutGrid, Table2 } from 'lucide-react'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

export type ComparisonViewMode = 'card' | 'table'

interface ComparisonViewToggleProps {
  value: ComparisonViewMode
  onChange: (value: ComparisonViewMode) => void
  className?: string
}

/**
 * Toggle between card grid and table views
 * Based on Langfuse pattern: StatisticsCard layout options
 */
export function ComparisonViewToggle({
  value,
  onChange,
  className,
}: ComparisonViewToggleProps) {
  return (
    <TooltipProvider>
      <ToggleGroup
        type="single"
        value={value}
        onValueChange={(v) => {
          if (v) onChange(v as ComparisonViewMode)
        }}
        className={className}
      >
        <Tooltip>
          <TooltipTrigger asChild>
            <ToggleGroupItem value="card" aria-label="Card view" size="sm">
              <LayoutGrid className="h-4 w-4" />
            </ToggleGroupItem>
          </TooltipTrigger>
          <TooltipContent>Card view</TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <ToggleGroupItem value="table" aria-label="Table view" size="sm">
              <Table2 className="h-4 w-4" />
            </ToggleGroupItem>
          </TooltipTrigger>
          <TooltipContent>Table view</TooltipContent>
        </Tooltip>
      </ToggleGroup>
    </TooltipProvider>
  )
}
