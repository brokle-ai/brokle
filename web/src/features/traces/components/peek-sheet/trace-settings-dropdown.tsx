'use client'

import * as React from 'react'
import { Settings2, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuCheckboxItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

/**
 * Observation levels matching OTEL conventions
 * Used for filtering spans by severity
 */
export type ObservationLevel = 'all' | 'default' | 'warning' | 'error'

export const OBSERVATION_LEVELS: { value: ObservationLevel; label: string }[] = [
  { value: 'all', label: 'Show All' },
  { value: 'default', label: 'Default+' },
  { value: 'warning', label: 'Warning+' },
  { value: 'error', label: 'Error Only' },
]

/**
 * Settings state type for trace visualization
 */
export interface TraceDisplaySettings {
  /** Show duration metric on spans */
  showDuration: boolean
  /** Show cost and token metrics on spans */
  showCostTokens: boolean
  /** Color-code metrics using heatmap */
  colorCodeMetrics: boolean
  /** Minimum observation level to display */
  minLevel: ObservationLevel
}

/**
 * Default settings
 */
export const DEFAULT_TRACE_SETTINGS: TraceDisplaySettings = {
  showDuration: true,
  showCostTokens: true,
  colorCodeMetrics: false,
  minLevel: 'all',
}

interface TraceSettingsDropdownProps {
  settings: TraceDisplaySettings
  onSettingsChange: (settings: TraceDisplaySettings) => void
  className?: string
}

/**
 * TraceSettingsDropdown - View options menu for trace visualization
 *
 * Provides toggles for:
 * - Show Duration
 * - Show Cost/Tokens
 * - Color Code Metrics (heatmap)
 * - Min Level filter (DEBUG/DEFAULT/WARNING/ERROR)
 *
 * Settings are session-only (not persisted to localStorage)
 */
export function TraceSettingsDropdown({
  settings,
  onSettingsChange,
  className,
}: TraceSettingsDropdownProps) {
  const updateSetting = <K extends keyof TraceDisplaySettings>(
    key: K,
    value: TraceDisplaySettings[K]
  ) => {
    onSettingsChange({ ...settings, [key]: value })
  }

  // Color code metrics should be disabled if both duration and cost/tokens are hidden
  const canColorCode = settings.showDuration || settings.showCostTokens

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          size='icon'
          className={className}
          title='View settings'
        >
          <Settings2 className='h-4 w-4' />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-48'>
        <DropdownMenuLabel>View Options</DropdownMenuLabel>
        <DropdownMenuSeparator />

        <DropdownMenuCheckboxItem
          checked={settings.showDuration}
          onCheckedChange={(checked) => updateSetting('showDuration', checked)}
        >
          Show Duration
        </DropdownMenuCheckboxItem>

        <DropdownMenuCheckboxItem
          checked={settings.showCostTokens}
          onCheckedChange={(checked) => updateSetting('showCostTokens', checked)}
        >
          Show Cost/Tokens
        </DropdownMenuCheckboxItem>

        <DropdownMenuCheckboxItem
          checked={settings.colorCodeMetrics}
          onCheckedChange={(checked) => updateSetting('colorCodeMetrics', checked)}
          disabled={!canColorCode}
        >
          Color Code Metrics
        </DropdownMenuCheckboxItem>

        <DropdownMenuSeparator />

        <DropdownMenuSub>
          <DropdownMenuSubTrigger>Min Level</DropdownMenuSubTrigger>
          <DropdownMenuSubContent>
            {OBSERVATION_LEVELS.map((level) => (
              <DropdownMenuCheckboxItem
                key={level.value}
                checked={settings.minLevel === level.value}
                onCheckedChange={() => updateSetting('minLevel', level.value)}
              >
                {level.label}
              </DropdownMenuCheckboxItem>
            ))}
          </DropdownMenuSubContent>
        </DropdownMenuSub>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

/**
 * Hook for managing trace display settings
 * Settings are stored in React state (session-only, not persisted)
 */
export function useTraceDisplaySettings(
  initialSettings: Partial<TraceDisplaySettings> = {}
): [TraceDisplaySettings, (settings: TraceDisplaySettings) => void] {
  const [settings, setSettings] = React.useState<TraceDisplaySettings>({
    ...DEFAULT_TRACE_SETTINGS,
    ...initialSettings,
  })

  return [settings, setSettings]
}

/**
 * Filter function to apply level filtering to spans
 */
export function filterByLevel<T extends { level?: string | null }>(
  items: T[],
  minLevel: ObservationLevel
): T[] {
  if (minLevel === 'all') return items

  const levelPriority: Record<string, number> = {
    debug: 0,
    default: 1,
    info: 1,
    warning: 2,
    warn: 2,
    error: 3,
  }

  const minPriority =
    minLevel === 'default' ? 1 : minLevel === 'warning' ? 2 : minLevel === 'error' ? 3 : 0

  return items.filter((item) => {
    const itemLevel = (item.level || 'default').toLowerCase()
    const itemPriority = levelPriority[itemLevel] ?? 1
    return itemPriority >= minPriority
  })
}
