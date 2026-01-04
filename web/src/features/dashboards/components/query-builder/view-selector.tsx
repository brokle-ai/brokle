'use client'

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Label } from '@/components/ui/label'
import type { WidgetViewType, ViewDefinition } from '../../types'

interface ViewSelectorProps {
  value?: WidgetViewType
  onValueChange: (value: WidgetViewType) => void
  viewDefinitions?: Record<WidgetViewType, ViewDefinition>
  disabled?: boolean
  label?: string
}

const VIEW_OPTIONS: Array<{ value: WidgetViewType; label: string; description: string }> = [
  {
    value: 'traces',
    label: 'Traces',
    description: 'Root-level trace metrics and aggregations',
  },
  {
    value: 'spans',
    label: 'Spans',
    description: 'Individual span-level metrics',
  },
  {
    value: 'scores',
    label: 'Quality Scores',
    description: 'Evaluation and quality score metrics',
  },
]

export function ViewSelector({
  value,
  onValueChange,
  viewDefinitions,
  disabled,
  label = 'Data Source',
}: ViewSelectorProps) {
  return (
    <div className="space-y-2">
      <Label className="text-sm font-medium">{label}</Label>
      <Select
        value={value}
        onValueChange={(val) => onValueChange(val as WidgetViewType)}
        disabled={disabled}
      >
        <SelectTrigger className="w-full">
          <SelectValue placeholder="Select a data source..." />
        </SelectTrigger>
        <SelectContent>
          {VIEW_OPTIONS.map((option) => {
            const viewDef = viewDefinitions?.[option.value]
            const measureCount = viewDef?.measures?.length ?? 0
            const dimensionCount = viewDef?.dimensions?.length ?? 0

            return (
              <SelectItem key={option.value} value={option.value}>
                <div className="flex flex-col">
                  <span>{option.label}</span>
                  <span className="text-xs text-muted-foreground">
                    {viewDef
                      ? `${measureCount} measures, ${dimensionCount} dimensions`
                      : option.description}
                  </span>
                </div>
              </SelectItem>
            )
          })}
        </SelectContent>
      </Select>
    </div>
  )
}
