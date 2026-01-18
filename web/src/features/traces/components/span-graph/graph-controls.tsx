'use client'

import { GitBranch, Atom, Circle, Layers } from 'lucide-react'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { Toggle } from '@/components/ui/toggle'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

/**
 * Layout mode for the graph
 * - dagre: Hierarchical layout (top to bottom)
 * - physics: Force-directed layout with physics simulation
 */
export type LayoutMode = 'dagre' | 'physics'

interface GraphControlsProps {
  layoutMode: LayoutMode
  onLayoutModeChange: (mode: LayoutMode) => void
  showSystemNodes: boolean
  onShowSystemNodesChange: (show: boolean) => void
  groupByStep: boolean
  onGroupByStepChange: (group: boolean) => void
  className?: string
}

/**
 * GraphControls - Control panel for graph visualization options
 *
 * Features:
 * - Layout mode toggle (dagre/physics)
 * - System nodes toggle (__start__, __end__)
 * - Step grouping toggle (parallel execution grouping)
 */
export function GraphControls({
  layoutMode,
  onLayoutModeChange,
  showSystemNodes,
  onShowSystemNodesChange,
  groupByStep,
  onGroupByStepChange,
  className,
}: GraphControlsProps) {
  return (
    <div
      className={cn(
        'absolute top-2 right-2 z-10',
        'flex items-center gap-2',
        'bg-background/80 backdrop-blur-sm',
        'p-1.5 rounded-lg border shadow-sm',
        className
      )}
    >
      {/* Layout Mode Toggle */}
      <ToggleGroup
        type="single"
        value={layoutMode}
        onValueChange={(value) => {
          if (value) onLayoutModeChange(value as LayoutMode)
        }}
        className="gap-0.5"
      >
        <Tooltip>
          <TooltipTrigger asChild>
            <ToggleGroupItem
              value="dagre"
              size="sm"
              className="h-7 w-7 p-0"
              aria-label="Hierarchical layout"
            >
              <GitBranch className="h-3.5 w-3.5" />
            </ToggleGroupItem>
          </TooltipTrigger>
          <TooltipContent side="bottom">
            <p className="text-xs">Hierarchical layout</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <ToggleGroupItem
              value="physics"
              size="sm"
              className="h-7 w-7 p-0"
              aria-label="Force-directed layout"
            >
              <Atom className="h-3.5 w-3.5" />
            </ToggleGroupItem>
          </TooltipTrigger>
          <TooltipContent side="bottom">
            <p className="text-xs">Force-directed layout</p>
          </TooltipContent>
        </Tooltip>
      </ToggleGroup>

      <div className="w-px h-5 bg-border" />

      {/* System Nodes Toggle */}
      <Tooltip>
        <TooltipTrigger asChild>
          <Toggle
            pressed={showSystemNodes}
            onPressedChange={onShowSystemNodesChange}
            size="sm"
            className="h-7 w-7 p-0"
            aria-label="Show start/end nodes"
          >
            <Circle className="h-3.5 w-3.5" />
          </Toggle>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <p className="text-xs">Show start/end nodes</p>
        </TooltipContent>
      </Tooltip>

      {/* Step Grouping Toggle */}
      <Tooltip>
        <TooltipTrigger asChild>
          <Toggle
            pressed={groupByStep}
            onPressedChange={onGroupByStepChange}
            size="sm"
            className="h-7 w-7 p-0"
            aria-label="Group parallel spans"
          >
            <Layers className="h-3.5 w-3.5" />
          </Toggle>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <p className="text-xs">Group parallel spans</p>
        </TooltipContent>
      </Tooltip>
    </div>
  )
}
