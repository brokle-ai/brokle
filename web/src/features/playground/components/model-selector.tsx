'use client'

import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface ModelSelectorProps {
  value: string | undefined
  onChange: (model: string) => void
  disabled?: boolean
  compact?: boolean
}

const OPENAI_MODELS = [
  'gpt-4o',
  'gpt-4o-mini',
  'gpt-4-turbo',
  'gpt-4',
  'gpt-3.5-turbo',
  'o1',
  'o1-mini',
]

const ANTHROPIC_MODELS = [
  'claude-3-5-sonnet-20241022',
  'claude-3-5-haiku-20241022',
  'claude-3-opus-20240229',
  'claude-3-sonnet-20240229',
  'claude-3-haiku-20240307',
]

// Display-friendly model names for compact view
const MODEL_DISPLAY_NAMES: Record<string, string> = {
  'gpt-4o': 'GPT-4o',
  'gpt-4o-mini': 'GPT-4o Mini',
  'gpt-4-turbo': 'GPT-4 Turbo',
  'gpt-4': 'GPT-4',
  'gpt-3.5-turbo': 'GPT-3.5',
  'o1': 'o1',
  'o1-mini': 'o1 Mini',
  'claude-3-5-sonnet-20241022': 'Claude 3.5 Sonnet',
  'claude-3-5-haiku-20241022': 'Claude 3.5 Haiku',
  'claude-3-opus-20240229': 'Claude 3 Opus',
  'claude-3-sonnet-20240229': 'Claude 3 Sonnet',
  'claude-3-haiku-20240307': 'Claude 3 Haiku',
}

export function ModelSelector({ value, onChange, disabled, compact = false }: ModelSelectorProps) {
  return (
    <Select
      value={value || ''}
      onValueChange={onChange}
      disabled={disabled}
    >
      <SelectTrigger className={compact ? 'w-[160px] h-8 text-xs' : 'w-full'}>
        <SelectValue placeholder="Select model" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>OpenAI</SelectLabel>
          {OPENAI_MODELS.map((model) => (
            <SelectItem key={model} value={model} className="text-sm">
              {MODEL_DISPLAY_NAMES[model] || model}
            </SelectItem>
          ))}
        </SelectGroup>
        <SelectGroup>
          <SelectLabel>Anthropic</SelectLabel>
          {ANTHROPIC_MODELS.map((model) => (
            <SelectItem key={model} value={model} className="text-sm">
              {MODEL_DISPLAY_NAMES[model] || model}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  )
}

export { OPENAI_MODELS, ANTHROPIC_MODELS, MODEL_DISPLAY_NAMES }
