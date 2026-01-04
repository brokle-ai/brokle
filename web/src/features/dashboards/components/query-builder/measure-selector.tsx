'use client'

import { useMemo } from 'react'
import { MultiSelect } from '@/components/shared/forms/multi-select'
import type { MeasureDefinition } from '../../types'

interface MeasureSelectorProps {
  measures?: MeasureDefinition[]
  value?: string[]
  onValueChange: (value: string[]) => void
  disabled?: boolean
  label?: string
  maxDisplayed?: number
}

const UNIT_LABELS: Record<string, string> = {
  count: 'count',
  ms: 'milliseconds',
  tokens: 'tokens',
  USD: 'USD',
  percentage: '%',
}

export function MeasureSelector({
  measures = [],
  value = [],
  onValueChange,
  disabled,
  label = 'Measures',
  maxDisplayed = 3,
}: MeasureSelectorProps) {
  const options = useMemo(() => {
    return measures.map((measure) => {
      const unitLabel = measure.unit ? ` (${UNIT_LABELS[measure.unit] ?? measure.unit})` : ''
      return {
        value: measure.id,
        label: measure.label,
        description: measure.description
          ? `${measure.description}${unitLabel}`
          : `${measure.type}${unitLabel}`,
      }
    })
  }, [measures])

  return (
    <MultiSelect
      label={label}
      description="Select one or more metrics to display"
      options={options}
      value={value}
      onValueChange={onValueChange}
      placeholder="Select measures..."
      searchPlaceholder="Search measures..."
      emptyText="No measures available for this view."
      maxDisplayed={maxDisplayed}
      disabled={disabled || measures.length === 0}
    />
  )
}
