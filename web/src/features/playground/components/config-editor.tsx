'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import type { ModelConfig, ParameterKey } from '../types'
import { PARAMETER_DEFINITIONS, PROVIDER_PARAMETER_SUPPORT } from '../types'
import { ModelSelector } from './model-selector'
import { ParameterControl } from './parameter-control'

interface ConfigEditorProps {
  config: ModelConfig | null
  onChange: (config: ModelConfig) => void
  disabled?: boolean
  hideModelSelector?: boolean
  projectId?: string
}

export function ConfigEditor({ config, onChange, disabled, hideModelSelector = false, projectId }: ConfigEditorProps) {
  const currentConfig = config || {}

  const handleChange = (field: string, value: number | boolean | string[] | undefined) => {
    const newConfig: ModelConfig = { ...currentConfig, [field]: value }

    // When enabling a parameter, initialize with default value if none exists
    if (typeof value === 'boolean' && value === true && field.endsWith('_enabled')) {
      const paramKey = field.replace('_enabled', '') as ParameterKey
      const def = PARAMETER_DEFINITIONS.find((d) => d.key === paramKey)
      if (def && newConfig[paramKey] === undefined) {
        ;(newConfig as Record<string, number | boolean | undefined>)[paramKey] = def.defaultValue
      }
    }

    onChange(newConfig)
  }

  const provider = currentConfig.provider || 'openai'
  const supportedParams = PROVIDER_PARAMETER_SUPPORT[provider] || PROVIDER_PARAMETER_SUPPORT.openai

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm font-medium">Model Configuration</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {!hideModelSelector && (
          <div className="space-y-2">
            <Label>
              Model <span className="text-destructive">*</span>
            </Label>
            <ModelSelector
              value={currentConfig.model}
              credentialId={currentConfig.credential_id}
              onChange={(model, provider, credentialId) => {
                onChange({ ...currentConfig, model, provider, credential_id: credentialId })
              }}
              disabled={disabled}
              projectId={projectId}
            />
            <p className="text-xs text-muted-foreground">
              Select the model for execution
            </p>
          </div>
        )}

        {/* Parameter Controls with Enable/Disable Toggles */}
        {PARAMETER_DEFINITIONS.map((def) => {
          const providerSupported = supportedParams.includes(def.key)
          const enabledKey = `${def.key}_enabled` as keyof ModelConfig

          return (
            <ParameterControl
              key={def.key}
              definition={def}
              value={currentConfig[def.key] as number | undefined}
              enabled={(currentConfig[enabledKey] as boolean) ?? false}
              onValueChange={(v) => handleChange(def.key, v)}
              onEnabledChange={(e) => handleChange(enabledKey, e)}
              providerSupported={providerSupported}
              disabled={disabled}
            />
          )
        })}

        {/* Stop Sequences */}
        <div className="space-y-2">
          <Label className="text-xs">Stop Sequences</Label>
          <Input
            value={currentConfig.stop?.join(', ') || ''}
            onChange={(e) =>
              handleChange(
                'stop',
                e.target.value ? e.target.value.split(',').map((s) => s.trim()) : undefined
              )
            }
            placeholder="e.g., END, STOP"
            disabled={disabled}
            className="h-8 text-xs"
          />
          <p className="text-xs text-muted-foreground">
            Comma-separated sequences to stop generation
          </p>
        </div>
      </CardContent>
    </Card>
  )
}
