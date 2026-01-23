'use client'

import { AlignJustify, AlignCenter, AlignLeft } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import type { RowHeight } from './types'
import { ROW_HEIGHT_LABELS } from './types'

interface RowHeightSelectorProps {
  value: RowHeight
  onChange: (value: RowHeight) => void
  className?: string
}

const ROW_HEIGHT_ICONS: Record<RowHeight, React.ReactNode> = {
  small: <AlignLeft className="h-4 w-4 rotate-90" />,
  medium: <AlignCenter className="h-4 w-4 rotate-90" />,
  large: <AlignJustify className="h-4 w-4 rotate-90" />,
}

const ROW_HEIGHT_DESCRIPTIONS: Record<RowHeight, string> = {
  small: 'Compact rows - single line',
  medium: 'Standard rows - preview content',
  large: 'Expanded rows - full content',
}

export function RowHeightSelector({ value, onChange, className }: RowHeightSelectorProps) {
  return (
    <ToggleGroup
      type="single"
      value={value}
      onValueChange={(v) => v && onChange(v as RowHeight)}
      className={cn('border rounded-md', className)}
    >
      {(['small', 'medium', 'large'] as const).map((height) => (
        <Tooltip key={height}>
          <TooltipTrigger asChild>
            <ToggleGroupItem
              value={height}
              aria-label={ROW_HEIGHT_DESCRIPTIONS[height]}
              className="h-8 w-8 p-0"
            >
              <span className="text-xs font-medium">{ROW_HEIGHT_LABELS[height]}</span>
            </ToggleGroupItem>
          </TooltipTrigger>
          <TooltipContent side="bottom">
            {ROW_HEIGHT_DESCRIPTIONS[height]}
          </TooltipContent>
        </Tooltip>
      ))}
    </ToggleGroup>
  )
}

// Alternative button style (for use in toolbars)
interface RowHeightButtonProps {
  value: RowHeight
  onClick: () => void
  className?: string
}

export function RowHeightButton({ value, onClick, className }: RowHeightButtonProps) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          onClick={onClick}
          className={cn('h-8 px-2 gap-1', className)}
        >
          {ROW_HEIGHT_ICONS[value]}
          <span className="text-xs">{ROW_HEIGHT_LABELS[value]}</span>
        </Button>
      </TooltipTrigger>
      <TooltipContent side="bottom">
        Row height: {ROW_HEIGHT_DESCRIPTIONS[value]}
      </TooltipContent>
    </Tooltip>
  )
}
