'use client'

import { Switch } from '@/components/ui/switch'
import { Slider } from '@/components/ui/slider'
import { BipolarSlider } from '@/components/ui/bipolar-slider'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import type { ParameterDefinition } from '../types'

interface ParameterControlProps {
  definition: ParameterDefinition
  value: number | undefined
  enabled: boolean
  onValueChange: (value: number) => void
  onEnabledChange: (enabled: boolean) => void
  providerSupported: boolean
  disabled?: boolean
}

export function ParameterControl({
  definition,
  value,
  enabled,
  onValueChange,
  onEnabledChange,
  providerSupported,
  disabled,
}: ParameterControlProps) {
  const displayValue = value ?? definition.defaultValue
  const isInteractable = !disabled && providerSupported

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value
    if (val === '') return
    const num = parseFloat(val)
    if (!isNaN(num)) {
      // Clamp to min/max
      const clamped = Math.min(Math.max(num, definition.min), definition.max)
      onValueChange(clamped)
    }
  }

  return (
    <div className={cn('space-y-2', !providerSupported && 'opacity-50')}>
      {/* Row 1: Label + Input + Toggle */}
      <div className="flex items-center justify-between">
        <Label
          className={cn(
            'text-xs font-medium',
            (!enabled || !providerSupported) && 'text-muted-foreground'
          )}
        >
          {definition.label}
        </Label>
        <div className="flex items-center gap-2">
          <Input
            type="number"
            value={displayValue}
            onChange={handleInputChange}
            disabled={!isInteractable || !enabled}
            className={cn(
              'h-7 w-16 text-xs text-right tabular-nums px-2',
              !enabled && 'opacity-50'
            )}
            min={definition.min}
            max={definition.max}
            step={definition.step}
          />
          <Switch
            checked={enabled && providerSupported}
            onCheckedChange={onEnabledChange}
            disabled={!isInteractable}
            aria-label={`Enable ${definition.label}`}
          />
        </div>
      </div>

      {/* Row 2: Slider (only for slider types) */}
      {definition.type === 'slider' && (
        <Slider
          value={[displayValue]}
          onValueChange={([v]) => onValueChange(v)}
          min={definition.min}
          max={definition.max}
          step={definition.step}
          disabled={!isInteractable || !enabled}
          className={cn(!enabled && 'opacity-50')}
        />
      )}

      {definition.type === 'bipolar-slider' && (
        <BipolarSlider
          value={displayValue}
          onValueChange={onValueChange}
          min={definition.min}
          max={definition.max}
          step={definition.step}
          disabled={!isInteractable || !enabled}
          className={cn(!enabled && 'opacity-50')}
        />
      )}

      {/* No slider for 'number' type - input is already shown above */}

      {!providerSupported && (
        <p className="text-[10px] text-muted-foreground italic">
          Not supported by this provider
        </p>
      )}
    </div>
  )
}
