'use client'

import * as React from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Slider } from '@/components/ui/slider'
import { Loader2, Plus } from 'lucide-react'
import type { ScoreConfig } from '@/features/scores/types'
import type { CreateAnnotationRequest } from '../api/scores-api'

interface AnnotationFormProps {
  scoreConfigs: ScoreConfig[]
  onSubmit: (data: CreateAnnotationRequest) => void
  isSubmitting?: boolean
}

/**
 * AnnotationForm - Dynamic form for creating annotations
 *
 * Features:
 * - Score type selector (dropdown of ScoreConfigs)
 * - Dynamic value input based on selected type:
 *   - NUMERIC: Slider with number input (respects min/max)
 *   - CATEGORICAL: Select dropdown with predefined options
 *   - BOOLEAN: Two-button toggle (Yes/No)
 * - Reason/explanation textarea
 * - Custom score name input when no config selected
 */
export function AnnotationForm({
  scoreConfigs,
  onSubmit,
  isSubmitting = false,
}: AnnotationFormProps) {
  const [selectedConfigId, setSelectedConfigId] = React.useState<string>('')
  const [customName, setCustomName] = React.useState('')
  const [numericValue, setNumericValue] = React.useState<number | null>(null)
  const [stringValue, setStringValue] = React.useState<string | null>(null)
  const [booleanValue, setBooleanValue] = React.useState<boolean | null>(null)
  const [reason, setReason] = React.useState('')

  const selectedConfig = scoreConfigs.find(c => c.id === selectedConfigId)
  const isCustom = selectedConfigId === 'custom'

  // Reset value inputs when config changes
  React.useEffect(() => {
    setNumericValue(null)
    setStringValue(null)
    setBooleanValue(null)
    if (selectedConfig) {
      // Set default value based on config
      if (selectedConfig.type === 'NUMERIC') {
        const min = selectedConfig.min_value ?? 0
        const max = selectedConfig.max_value ?? 10
        setNumericValue(Math.round((min + max) / 2))
      }
    }
  }, [selectedConfigId, selectedConfig])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    const name = isCustom ? customName : (selectedConfig?.name ?? '')
    if (!name) return

    let type = selectedConfig?.type ?? 'NUMERIC'
    if (isCustom) {
      type = 'NUMERIC' // Default for custom scores
    }

    const data: CreateAnnotationRequest = {
      name,
      type,
      reason: reason || null,
    }

    if (type === 'NUMERIC' && numericValue !== null) {
      data.value = numericValue
    } else if (type === 'CATEGORICAL' && stringValue) {
      data.string_value = stringValue
    } else if (type === 'BOOLEAN' && booleanValue !== null) {
      data.value = booleanValue ? 1 : 0
    }

    onSubmit(data)

    // Reset form
    setSelectedConfigId('')
    setCustomName('')
    setNumericValue(null)
    setStringValue(null)
    setBooleanValue(null)
    setReason('')
  }

  const isValid = React.useMemo(() => {
    if (!selectedConfigId) return false
    if (isCustom && !customName.trim()) return false

    if (selectedConfig) {
      switch (selectedConfig.type) {
        case 'NUMERIC':
          return numericValue !== null
        case 'CATEGORICAL':
          return stringValue !== null
        case 'BOOLEAN':
          return booleanValue !== null
      }
    }

    // Custom score requires numeric value
    if (isCustom) return numericValue !== null

    return false
  }, [selectedConfigId, isCustom, customName, selectedConfig, numericValue, stringValue, booleanValue])

  return (
    <form onSubmit={handleSubmit} className='space-y-4'>
      {/* Score Type Selector */}
      <div className='space-y-2'>
        <Label htmlFor='score-type'>Score Type</Label>
        <Select value={selectedConfigId} onValueChange={setSelectedConfigId}>
          <SelectTrigger id='score-type'>
            <SelectValue placeholder='Select a score type...' />
          </SelectTrigger>
          <SelectContent>
            {scoreConfigs.map((config) => (
              <SelectItem key={config.id} value={config.id}>
                <div className='flex items-center gap-2'>
                  <span>{config.name}</span>
                  <span className='text-xs text-muted-foreground'>
                    ({config.type.toLowerCase()})
                  </span>
                </div>
              </SelectItem>
            ))}
            <SelectItem value='custom'>
              <div className='flex items-center gap-2'>
                <Plus className='h-3 w-3' />
                <span>Custom score...</span>
              </div>
            </SelectItem>
          </SelectContent>
        </Select>
        {selectedConfig?.description && (
          <p className='text-xs text-muted-foreground'>{selectedConfig.description}</p>
        )}
      </div>

      {/* Custom Name Input */}
      {isCustom && (
        <div className='space-y-2'>
          <Label htmlFor='custom-name'>Score Name</Label>
          <Input
            id='custom-name'
            value={customName}
            onChange={(e) => setCustomName(e.target.value)}
            placeholder='e.g., helpfulness, accuracy'
          />
        </div>
      )}

      {/* Dynamic Value Input */}
      {(selectedConfig || isCustom) && (
        <div className='space-y-2'>
          <Label>Value</Label>
          {(isCustom || selectedConfig?.type === 'NUMERIC') && (
            <NumericInput
              value={numericValue}
              onChange={setNumericValue}
              min={selectedConfig?.min_value ?? 0}
              max={selectedConfig?.max_value ?? 10}
            />
          )}
          {selectedConfig?.type === 'CATEGORICAL' && (
            <CategoricalInput
              value={stringValue}
              onChange={setStringValue}
              categories={selectedConfig.categories ?? []}
            />
          )}
          {selectedConfig?.type === 'BOOLEAN' && (
            <BooleanInput
              value={booleanValue}
              onChange={setBooleanValue}
            />
          )}
        </div>
      )}

      {/* Reason/Explanation */}
      {selectedConfigId && (
        <div className='space-y-2'>
          <Label htmlFor='reason'>
            Explanation <span className='text-muted-foreground'>(optional)</span>
          </Label>
          <Textarea
            id='reason'
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder='Why did you give this score?'
            rows={2}
          />
        </div>
      )}

      {/* Submit Button */}
      <Button
        type='submit'
        className='w-full'
        disabled={!isValid || isSubmitting}
      >
        {isSubmitting ? (
          <>
            <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            Adding...
          </>
        ) : (
          <>
            <Plus className='mr-2 h-4 w-4' />
            Add Annotation
          </>
        )}
      </Button>
    </form>
  )
}

// ============================================================================
// Value Input Components
// ============================================================================

interface NumericInputProps {
  value: number | null
  onChange: (value: number | null) => void
  min: number
  max: number
}

function NumericInput({ value, onChange, min, max }: NumericInputProps) {
  const range = max - min
  const useSlider = range <= 10

  if (useSlider) {
    return (
      <div className='space-y-3'>
        <div className='flex items-center gap-4'>
          <Slider
            value={[value ?? min]}
            onValueChange={([v]) => onChange(v)}
            min={min}
            max={max}
            step={1}
            className='flex-1'
          />
          <span className='w-8 text-center font-mono text-sm'>
            {value ?? '-'}
          </span>
        </div>
        <div className='flex justify-between text-xs text-muted-foreground'>
          <span>{min}</span>
          <span>{max}</span>
        </div>
      </div>
    )
  }

  return (
    <Input
      type='number'
      value={value ?? ''}
      onChange={(e) => onChange(e.target.value ? parseFloat(e.target.value) : null)}
      min={min}
      max={max}
      step='0.1'
      placeholder={`${min} - ${max}`}
    />
  )
}

interface CategoricalInputProps {
  value: string | null
  onChange: (value: string | null) => void
  categories: string[]
}

function CategoricalInput({ value, onChange, categories }: CategoricalInputProps) {
  // For small number of categories, use radio buttons
  if (categories.length <= 4) {
    return (
      <RadioGroup
        value={value ?? ''}
        onValueChange={onChange}
        className='flex flex-wrap gap-2'
      >
        {categories.map((category) => (
          <div key={category} className='flex items-center'>
            <RadioGroupItem
              value={category}
              id={`cat-${category}`}
              className='peer sr-only'
            />
            <Label
              htmlFor={`cat-${category}`}
              className='px-3 py-1.5 rounded-md border cursor-pointer transition-colors peer-data-[state=checked]:bg-primary peer-data-[state=checked]:text-primary-foreground peer-data-[state=checked]:border-primary hover:bg-muted'
            >
              {category}
            </Label>
          </div>
        ))}
      </RadioGroup>
    )
  }

  // For larger lists, use a select
  return (
    <Select value={value ?? ''} onValueChange={onChange}>
      <SelectTrigger>
        <SelectValue placeholder='Select a category...' />
      </SelectTrigger>
      <SelectContent>
        {categories.map((category) => (
          <SelectItem key={category} value={category}>
            {category}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}

interface BooleanInputProps {
  value: boolean | null
  onChange: (value: boolean | null) => void
}

function BooleanInput({ value, onChange }: BooleanInputProps) {
  return (
    <div className='flex gap-2'>
      <Button
        type='button'
        variant={value === true ? 'default' : 'outline'}
        className='flex-1'
        onClick={() => onChange(true)}
      >
        Yes
      </Button>
      <Button
        type='button'
        variant={value === false ? 'default' : 'outline'}
        className='flex-1'
        onClick={() => onChange(false)}
      >
        No
      </Button>
    </div>
  )
}
