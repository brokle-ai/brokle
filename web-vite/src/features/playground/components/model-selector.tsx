import { useState } from 'react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { ChevronDown, Sparkles } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ModelSelectorProps {
  value: string | undefined
  credentialId?: string
  onChange: (model: string, provider: string, credentialId?: string) => void
  disabled?: boolean
  compact?: boolean
  orgId?: string
}

// Simplified model selector for web-vite. The original web/ version
// depends on a full `features/ai-providers` module (configured
// credentials, model catalog) that hasn't been ported yet — so this
// variant accepts a freeform model ID + a provider dropdown. Once the
// AI providers feature lands in web-vite, swap this component for the
// richer split-pane picker.
const PROVIDERS = [
  { id: 'openai', label: 'OpenAI' },
  { id: 'anthropic', label: 'Anthropic' },
  { id: 'azure', label: 'Azure' },
  { id: 'gemini', label: 'Gemini' },
  { id: 'openrouter', label: 'OpenRouter' },
  { id: 'custom', label: 'Custom' },
] as const

type ProviderId = (typeof PROVIDERS)[number]['id']

export function ModelSelector({
  value,
  credentialId,
  onChange,
  disabled,
  compact = false,
}: ModelSelectorProps) {
  const [open, setOpen] = useState(false)
  const [provider, setProvider] = useState<ProviderId>('openai')
  const [model, setModel] = useState(value ?? '')
  const [cred, setCred] = useState(credentialId ?? '')

  const handleApply = () => {
    if (!model.trim()) return
    onChange(model.trim(), provider, cred.trim() || undefined)
    setOpen(false)
  }

  const display = value ? value : 'Select model'

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn(
            'justify-between font-normal',
            compact ? 'h-8 w-auto max-w-[280px] text-xs px-2' : 'w-full',
          )}
        >
          <div className="flex items-center gap-2 min-w-0">
            <Sparkles
              className={cn(
                'shrink-0 text-muted-foreground',
                compact ? 'h-3.5 w-3.5' : 'h-4 w-4',
              )}
            />
            <span
              className={cn('truncate', !value && 'text-muted-foreground')}
            >
              {display}
            </span>
          </div>
          <ChevronDown
            className={cn(
              'shrink-0 opacity-50',
              compact ? 'h-3 w-3' : 'h-4 w-4',
            )}
          />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80" align="start">
        <div className="space-y-3">
          <div className="space-y-2">
            <Label className="text-xs">Provider</Label>
            <Select
              value={provider}
              onValueChange={(v) => setProvider(v as ProviderId)}
            >
              <SelectTrigger className="h-8 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {PROVIDERS.map((p) => (
                  <SelectItem key={p.id} value={p.id} className="text-xs">
                    {p.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label className="text-xs" htmlFor="model-id">
              Model ID
            </Label>
            <Input
              id="model-id"
              value={model}
              onChange={(e) => setModel(e.target.value)}
              placeholder="e.g. gpt-4o-mini"
              className="h-8 text-xs"
            />
            <p className="text-[10px] text-muted-foreground">
              Enter the provider model identifier the backend should route to.
            </p>
          </div>

          <div className="space-y-2">
            <Label className="text-xs" htmlFor="credential-id">
              Credential ID{' '}
              <span className="font-normal text-muted-foreground">
                (optional)
              </span>
            </Label>
            <Input
              id="credential-id"
              value={cred}
              onChange={(e) => setCred(e.target.value)}
              placeholder="cred_..."
              className="h-8 text-xs"
            />
          </div>

          <div className="flex justify-end gap-2 pt-1">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button size="sm" onClick={handleApply} disabled={!model.trim()}>
              Apply
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}
