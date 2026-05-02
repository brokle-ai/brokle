import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Slider } from '@/components/ui/slider'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import type { ScoreType, ScoreConfig } from '../../types'

// ============================================================================
// Score Input Field - Dynamic input based on data type
// ============================================================================

interface ScoreInputFieldProps {
  dataType: ScoreType
  value: number | string | boolean | null
  onChange: (value: number | string | boolean | null) => void
  config?: ScoreConfig
  disabled?: boolean
  className?: string
}

/**
 * Dynamic score input field that renders different inputs based on data type
 *
 * - NUMERIC: Slider (if range <= 10) or number input
 * - CATEGORICAL: Radio buttons (if <= 4 categories) or select dropdown
 * - BOOLEAN: Two-button toggle (Yes/No)
 */
export function ScoreInputField({
  dataType,
  value,
  onChange,
  config,
  disabled = false,
  className,
}: ScoreInputFieldProps) {
  switch (dataType) {
    case 'NUMERIC':
      return (
        <NumericInput
          value={value as number | null}
          onChange={(v) => onChange(v)}
          min={config?.min_value ?? 0}
          max={config?.max_value ?? 10}
          disabled={disabled}
          className={className}
        />
      )
    case 'CATEGORICAL':
      return (
        <CategoricalInput
          value={value as string | null}
          onChange={(v) => onChange(v)}
          categories={config?.categories ?? []}
          disabled={disabled}
          className={className}
        />
      )
    case 'BOOLEAN':
      return (
        <BooleanInput
          value={value as boolean | null}
          onChange={(v) => onChange(v)}
          disabled={disabled}
          className={className}
        />
      )
    default:
      return null
  }
}

// ============================================================================
// Numeric Input
// ============================================================================

interface NumericInputProps {
  value: number | null
  onChange: (value: number | null) => void
  min: number
  max: number
  disabled?: boolean
  className?: string
}

export function NumericInput({
  value,
  onChange,
  min,
  max,
  disabled = false,
  className,
}: NumericInputProps) {
  const range = max - min
  const isFractionalRange = !Number.isInteger(min) || !Number.isInteger(max) || range < 1
  const useSlider = range > 0 && range <= 10 && !isFractionalRange

  if (useSlider) {
    return (
      <div className={cn('space-y-3', className)}>
        <div className="flex items-center gap-4">
          <Slider
            value={[value ?? min]}
            onValueChange={([v]) => onChange(v)}
            min={min}
            max={max}
            step={1}
            disabled={disabled}
            className="flex-1"
          />
          <span className="w-10 text-center font-mono text-sm tabular-nums">
            {value ?? '-'}
          </span>
        </div>
        <div className="flex justify-between text-xs text-muted-foreground">
          <span>{min}</span>
          <span>{max}</span>
        </div>
      </div>
    )
  }

  const step = isFractionalRange ? 'any' : 1

  return (
    <Input
      type="number"
      value={value ?? ''}
      onChange={(e) => onChange(e.target.value ? parseFloat(e.target.value) : null)}
      min={min}
      max={max}
      step={step}
      placeholder={`${min} - ${max}`}
      disabled={disabled}
      className={className}
    />
  )
}

// ============================================================================
// Categorical Input
// ============================================================================

interface CategoricalInputProps {
  value: string | null
  onChange: (value: string | null) => void
  categories: string[]
  disabled?: boolean
  className?: string
}

export function CategoricalInput({
  value,
  onChange,
  categories,
  disabled = false,
  className,
}: CategoricalInputProps) {
  if (categories.length > 0 && categories.length <= 4) {
    return (
      <RadioGroup
        value={value ?? ''}
        onValueChange={onChange}
        disabled={disabled}
        className={cn('flex flex-wrap gap-2', className)}
      >
        {categories.map((category) => (
          <div key={category} className="flex items-center">
            <RadioGroupItem
              value={category}
              id={`cat-${category}`}
              className="peer sr-only"
            />
            <Label
              htmlFor={`cat-${category}`}
              className={cn(
                'px-3 py-1.5 rounded-md border cursor-pointer transition-colors',
                'hover:bg-muted',
                'peer-data-[state=checked]:bg-primary peer-data-[state=checked]:text-primary-foreground peer-data-[state=checked]:border-primary',
                disabled && 'opacity-50 cursor-not-allowed',
              )}
            >
              {category}
            </Label>
          </div>
        ))}
      </RadioGroup>
    )
  }

  return (
    <Select value={value ?? ''} onValueChange={onChange} disabled={disabled}>
      <SelectTrigger className={className}>
        <SelectValue placeholder="Select a category..." />
      </SelectTrigger>
      <SelectContent>
        {categories.length === 0 ? (
          <SelectItem value="_empty" disabled>
            No categories defined
          </SelectItem>
        ) : (
          categories.map((category) => (
            <SelectItem key={category} value={category}>
              {category}
            </SelectItem>
          ))
        )}
      </SelectContent>
    </Select>
  )
}

// ============================================================================
// Boolean Input
// ============================================================================

interface BooleanInputProps {
  value: boolean | null
  onChange: (value: boolean | null) => void
  disabled?: boolean
  className?: string
  labels?: { true: string; false: string }
}

export function BooleanInput({
  value,
  onChange,
  disabled = false,
  className,
  labels = { true: 'Yes', false: 'No' },
}: BooleanInputProps) {
  return (
    <div className={cn('flex gap-2', className)}>
      <Button
        type="button"
        variant={value === true ? 'default' : 'outline'}
        className="flex-1"
        onClick={() => onChange(true)}
        disabled={disabled}
      >
        {labels.true}
      </Button>
      <Button
        type="button"
        variant={value === false ? 'default' : 'outline'}
        className="flex-1"
        onClick={() => onChange(false)}
        disabled={disabled}
      >
        {labels.false}
      </Button>
    </div>
  )
}
