'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Slider } from '@/components/ui/slider'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import type { ModelConfig } from '../types'
import { ModelSelector } from './model-selector'

interface ConfigEditorProps {
  config: ModelConfig | null
  onChange: (config: ModelConfig) => void
  disabled?: boolean
  hideModelSelector?: boolean
}

const PRESETS = {
  creative: {
    temperature: 1.0,
    top_p: 1.0,
    frequency_penalty: 0.5,
    presence_penalty: 0.5,
  },
  balanced: {
    temperature: 0.7,
    top_p: 0.9,
    frequency_penalty: 0.0,
    presence_penalty: 0.0,
  },
  precise: {
    temperature: 0.3,
    top_p: 0.5,
    frequency_penalty: 0.0,
    presence_penalty: 0.0,
  },
}

export function ConfigEditor({ config, onChange, disabled, hideModelSelector = false }: ConfigEditorProps) {
  const currentConfig = config || {}

  const handleChange = (field: keyof ModelConfig, value: any) => {
    onChange({ ...currentConfig, [field]: value })
  }

  const applyPreset = (preset: keyof typeof PRESETS) => {
    onChange({ ...currentConfig, ...PRESETS[preset] })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm font-medium">Model Configuration</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="basic" className="w-full">
          <TabsList className="w-full">
            <TabsTrigger value="basic" className="flex-1">
              Basic
            </TabsTrigger>
            <TabsTrigger value="advanced" className="flex-1">
              Advanced
            </TabsTrigger>
            <TabsTrigger value="presets" className="flex-1">
              Presets
            </TabsTrigger>
          </TabsList>

          <TabsContent value="basic" className="space-y-4 mt-4">
            {!hideModelSelector && (
              <div className="space-y-2">
                <Label>
                  Model <span className="text-destructive">*</span>
                </Label>
                <ModelSelector
                  value={currentConfig.model}
                  onChange={(value) => handleChange('model', value)}
                  disabled={disabled}
                />
                <p className="text-xs text-muted-foreground">
                  Select the model for execution
                </p>
              </div>
            )}

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>Temperature</Label>
                <span className="text-sm text-muted-foreground">
                  {currentConfig.temperature?.toFixed(1) ?? '0.7'}
                </span>
              </div>
              <Slider
                value={[currentConfig.temperature ?? 0.7]}
                onValueChange={([value]) => handleChange('temperature', value)}
                min={0}
                max={2}
                step={0.1}
                disabled={disabled}
                className="w-full"
              />
              <p className="text-xs text-muted-foreground">
                Controls randomness: 0 is focused, 2 is creative
              </p>
            </div>

            <div className="space-y-2">
              <Label>Max Tokens</Label>
              <Input
                type="number"
                value={currentConfig.max_tokens || ''}
                onChange={(e) =>
                  handleChange('max_tokens', e.target.value ? parseInt(e.target.value) : undefined)
                }
                placeholder="Default (e.g., 4096)"
                disabled={disabled}
                min={1}
                max={128000}
              />
              <p className="text-xs text-muted-foreground">
                Maximum length of generated response
              </p>
            </div>
          </TabsContent>

          <TabsContent value="advanced" className="space-y-4 mt-4">
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>Top P</Label>
                <span className="text-sm text-muted-foreground">
                  {currentConfig.top_p?.toFixed(2) ?? '1.00'}
                </span>
              </div>
              <Slider
                value={[currentConfig.top_p ?? 1.0]}
                onValueChange={([value]) => handleChange('top_p', value)}
                min={0}
                max={1}
                step={0.05}
                disabled={disabled}
              />
              <p className="text-xs text-muted-foreground">
                Nucleus sampling threshold
              </p>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>Frequency Penalty</Label>
                <span className="text-sm text-muted-foreground">
                  {currentConfig.frequency_penalty?.toFixed(1) ?? '0.0'}
                </span>
              </div>
              <Slider
                value={[currentConfig.frequency_penalty ?? 0.0]}
                onValueChange={([value]) => handleChange('frequency_penalty', value)}
                min={-2}
                max={2}
                step={0.1}
                disabled={disabled}
              />
              <p className="text-xs text-muted-foreground">
                Reduce repetition of tokens based on frequency
              </p>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>Presence Penalty</Label>
                <span className="text-sm text-muted-foreground">
                  {currentConfig.presence_penalty?.toFixed(1) ?? '0.0'}
                </span>
              </div>
              <Slider
                value={[currentConfig.presence_penalty ?? 0.0]}
                onValueChange={([value]) => handleChange('presence_penalty', value)}
                min={-2}
                max={2}
                step={0.1}
                disabled={disabled}
              />
              <p className="text-xs text-muted-foreground">
                Reduce repetition of tokens based on presence
              </p>
            </div>

            <div className="space-y-2">
              <Label>Stop Sequences</Label>
              <Input
                value={currentConfig.stop?.join(', ') || ''}
                onChange={(e) =>
                  handleChange(
                    'stop',
                    e.target.value ? e.target.value.split(',').map((s) => s.trim()) : []
                  )
                }
                placeholder="e.g., END, STOP"
                disabled={disabled}
              />
              <p className="text-xs text-muted-foreground">
                Comma-separated sequences to stop generation
              </p>
            </div>
          </TabsContent>

          <TabsContent value="presets" className="space-y-3 mt-4">
            <p className="text-sm text-muted-foreground">
              Apply preset configurations for common use cases
            </p>
            <div className="grid gap-2">
              <Button
                variant="outline"
                onClick={() => applyPreset('creative')}
                disabled={disabled}
                className="justify-start"
              >
                <div className="text-left">
                  <div className="font-medium">Creative</div>
                  <div className="text-xs text-muted-foreground">
                    High temperature (1.0) for diverse outputs
                  </div>
                </div>
              </Button>
              <Button
                variant="outline"
                onClick={() => applyPreset('balanced')}
                disabled={disabled}
                className="justify-start"
              >
                <div className="text-left">
                  <div className="font-medium">Balanced</div>
                  <div className="text-xs text-muted-foreground">
                    Moderate temperature (0.7) for general use
                  </div>
                </div>
              </Button>
              <Button
                variant="outline"
                onClick={() => applyPreset('precise')}
                disabled={disabled}
                className="justify-start"
              >
                <div className="text-left">
                  <div className="font-medium">Precise</div>
                  <div className="text-xs text-muted-foreground">
                    Low temperature (0.3) for focused outputs
                  </div>
                </div>
              </Button>
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}
