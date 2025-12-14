'use client'

import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Slider } from '@/components/ui/slider'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Plus, X } from 'lucide-react'
import type { ModelConfig } from '../../types'

interface ModelConfigFormProps {
  config: ModelConfig
  onChange: (config: ModelConfig) => void
  disabled?: boolean
}

const MODELS = [
  { value: 'gpt-4o', label: 'GPT-4o', provider: 'openai' },
  { value: 'gpt-4o-mini', label: 'GPT-4o Mini', provider: 'openai' },
  { value: 'gpt-4-turbo', label: 'GPT-4 Turbo', provider: 'openai' },
  { value: 'gpt-3.5-turbo', label: 'GPT-3.5 Turbo', provider: 'openai' },
  { value: 'claude-3-5-sonnet-20241022', label: 'Claude 3.5 Sonnet', provider: 'anthropic' },
  { value: 'claude-3-haiku-20240307', label: 'Claude 3 Haiku', provider: 'anthropic' },
  { value: 'claude-3-opus-20240229', label: 'Claude 3 Opus', provider: 'anthropic' },
]

export function ModelConfigForm({ config, onChange, disabled }: ModelConfigFormProps) {
  const handleStopSequenceAdd = () => {
    const stop = config.stop || []
    onChange({ ...config, stop: [...stop, ''] })
  }

  const handleStopSequenceChange = (index: number, value: string) => {
    const stop: string[] = [...(config.stop || [])]
    stop[index] = value
    onChange({ ...config, stop })
  }

  const handleStopSequenceRemove = (index: number) => {
    const stop = (config.stop || []).filter((_: string, i: number) => i !== index)
    onChange({ ...config, stop: stop.length > 0 ? stop : undefined })
  }

  return (
    <div className="space-y-6">
      {/* Model Selection */}
      <div className="space-y-2">
        <Label>Model</Label>
        <Select
          value={config.model || ''}
          onValueChange={(model) => onChange({ ...config, model })}
          disabled={disabled}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select a model" />
          </SelectTrigger>
          <SelectContent>
            <div className="text-xs text-muted-foreground px-2 py-1">OpenAI</div>
            {MODELS.filter((m) => m.provider === 'openai').map((model) => (
              <SelectItem key={model.value} value={model.value}>
                {model.label}
              </SelectItem>
            ))}
            <div className="text-xs text-muted-foreground px-2 py-1 mt-2">Anthropic</div>
            {MODELS.filter((m) => m.provider === 'anthropic').map((model) => (
              <SelectItem key={model.value} value={model.value}>
                {model.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Temperature */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label>Temperature</Label>
          <span className="text-sm text-muted-foreground">
            {config.temperature ?? 0.7}
          </span>
        </div>
        <Slider
          value={[config.temperature ?? 0.7]}
          onValueChange={([value]) => onChange({ ...config, temperature: value })}
          min={0}
          max={2}
          step={0.1}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Higher values make output more random, lower values more deterministic
        </p>
      </div>

      {/* Max Tokens */}
      <div className="space-y-2">
        <Label>Max Tokens</Label>
        <Input
          type="number"
          value={config.max_tokens ?? ''}
          onChange={(e) =>
            onChange({
              ...config,
              max_tokens: e.target.value ? parseInt(e.target.value) : undefined,
            })
          }
          placeholder="4096"
          min={1}
          max={128000}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Maximum number of tokens to generate
        </p>
      </div>

      {/* Top P */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label>Top P</Label>
          <span className="text-sm text-muted-foreground">
            {config.top_p ?? 1}
          </span>
        </div>
        <Slider
          value={[config.top_p ?? 1]}
          onValueChange={([value]) => onChange({ ...config, top_p: value })}
          min={0}
          max={1}
          step={0.05}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Nucleus sampling threshold
        </p>
      </div>

      {/* Frequency Penalty */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label>Frequency Penalty</Label>
          <span className="text-sm text-muted-foreground">
            {config.frequency_penalty ?? 0}
          </span>
        </div>
        <Slider
          value={[config.frequency_penalty ?? 0]}
          onValueChange={([value]) =>
            onChange({ ...config, frequency_penalty: value })
          }
          min={-2}
          max={2}
          step={0.1}
          disabled={disabled}
        />
      </div>

      {/* Presence Penalty */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label>Presence Penalty</Label>
          <span className="text-sm text-muted-foreground">
            {config.presence_penalty ?? 0}
          </span>
        </div>
        <Slider
          value={[config.presence_penalty ?? 0]}
          onValueChange={([value]) =>
            onChange({ ...config, presence_penalty: value })
          }
          min={-2}
          max={2}
          step={0.1}
          disabled={disabled}
        />
      </div>

      {/* Stop Sequences */}
      <div className="space-y-2">
        <Label>Stop Sequences</Label>
        <div className="space-y-2">
          {(config.stop || []).map((seq, index) => (
            <div key={index} className="flex gap-2">
              <Input
                value={seq}
                onChange={(e) => handleStopSequenceChange(index, e.target.value)}
                placeholder="Stop sequence..."
                disabled={disabled}
              />
              <Button
                variant="ghost"
                size="icon"
                onClick={() => handleStopSequenceRemove(index)}
                disabled={disabled}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          ))}
          <Button
            variant="outline"
            size="sm"
            onClick={handleStopSequenceAdd}
            disabled={disabled || (config.stop?.length || 0) >= 4}
          >
            <Plus className="mr-2 h-4 w-4" />
            Add stop sequence
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">
          Up to 4 sequences where the API will stop generating
        </p>
      </div>
    </div>
  )
}
