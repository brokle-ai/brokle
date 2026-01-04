'use client'

import { useMemo } from 'react'
import { MultiSelect } from '@/components/shared/forms/multi-select'
import type { DimensionDefinition } from '../../types'

interface DimensionSelectorProps {
  dimensions?: DimensionDefinition[]
  value?: string[]
  onValueChange: (value: string[]) => void
  disabled?: boolean
  label?: string
  maxDisplayed?: number
  maxSelections?: number
}

const COLUMN_TYPE_LABELS: Record<string, string> = {
  string: 'text',
  datetime: 'date/time',
  number: 'numeric',
  boolean: 'yes/no',
  duration: 'duration',
}

export function DimensionSelector({
  dimensions = [],
  value = [],
  onValueChange,
  disabled,
  label = 'Group By',
  maxDisplayed = 3,
  maxSelections,
}: DimensionSelectorProps) {
  const options = useMemo(() => {
    return dimensions.map((dimension) => {
      const typeLabel = COLUMN_TYPE_LABELS[dimension.column_type] ?? dimension.column_type
      const bucketLabel = dimension.bucketable ? ', bucketable' : ''
      return {
        value: dimension.id,
        label: dimension.label,
        description: dimension.description
          ? dimension.description
          : `${typeLabel}${bucketLabel}`,
      }
    })
  }, [dimensions])

  const handleValueChange = (newValue: string[]) => {
    // Enforce max selections if specified
    if (maxSelections && newValue.length > maxSelections) {
      return
    }
    onValueChange(newValue)
  }

  return (
    <MultiSelect
      label={label}
      description={
        maxSelections
          ? `Group data by dimensions (max ${maxSelections})`
          : 'Group data by one or more dimensions'
      }
      options={options}
      value={value}
      onValueChange={handleValueChange}
      placeholder="Select dimensions..."
      searchPlaceholder="Search dimensions..."
      emptyText="No dimensions available for this view."
      maxDisplayed={maxDisplayed}
      disabled={disabled || dimensions.length === 0}
    />
  )
}
