'use client'

import { useMemo } from 'react'
import { Bot, AlertCircle, Settings2 } from 'lucide-react'
import { Label } from '@/components/ui/label'
import { Slider } from '@/components/ui/slider'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import {
  useModelsByProvider,
  PROVIDER_INFO,
  type AIProvider,
  type AvailableModel,
} from '@/features/ai-providers'

export interface LLMConfig {
  credential_id: string
  model: string
  temperature: number
}

interface LLMConfigPanelProps {
  orgId: string | undefined
  config: LLMConfig
  onChange: (config: LLMConfig) => void
  disabled?: boolean
}

// Group models by credential for display
interface CredentialGroup {
  credentialId: string
  credentialName: string
  provider: AIProvider
  models: AvailableModel[]
}

/**
 * LLM Configuration Panel for evaluation rules.
 *
 * Features:
 * - Dynamic model selection from configured AI providers
 * - Models grouped by credential/provider
 * - Temperature slider with visual feedback
 * - Loading state while fetching available models
 * - Empty state when no providers configured
 */
export function LLMConfigPanel({
  orgId,
  config,
  onChange,
  disabled = false,
}: LLMConfigPanelProps) {
  const {
    data: allModels,
    modelsByProvider,
    configuredProviders,
    isLoading,
    isError,
  } = useModelsByProvider(orgId)

  // Group models by credential for better organization
  const credentialGroups = useMemo((): CredentialGroup[] => {
    if (!allModels) return []

    const groups = new Map<string, CredentialGroup>()

    allModels.forEach((model) => {
      // Use credential_id if available, otherwise use provider as key
      const key = model.credential_id || model.provider
      const existing = groups.get(key)

      if (existing) {
        existing.models.push(model)
      } else {
        groups.set(key, {
          credentialId: model.credential_id || '',
          credentialName: model.credential_name || PROVIDER_INFO[model.provider]?.name || model.provider,
          provider: model.provider,
          models: [model],
        })
      }
    })

    return Array.from(groups.values())
  }, [allModels])

  // Find currently selected model details
  const selectedModel = useMemo(() => {
    return allModels?.find(
      (m) => m.id === config.model && (m.credential_id === config.credential_id || !config.credential_id)
    )
  }, [allModels, config.model, config.credential_id])

  // Handle model selection - sets both model and credential_id
  const handleModelChange = (modelId: string) => {
    const model = allModels?.find((m) => m.id === modelId)
    if (model) {
      onChange({
        ...config,
        model: model.id,
        credential_id: model.credential_id || '',
      })
    }
  }

  // Handle temperature change
  const handleTemperatureChange = (values: number[]) => {
    onChange({
      ...config,
      temperature: values[0],
    })
  }

  const hasProviders = configuredProviders.length > 0

  return (
    <div className="space-y-4 border rounded-lg p-4">
      <div className="flex items-center gap-2">
        <Bot className="h-5 w-5 text-primary" />
        <h4 className="font-medium">LLM Configuration</h4>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </div>
      ) : isError ? (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Failed to load available models. Please try again.
          </AlertDescription>
        </Alert>
      ) : !hasProviders ? (
        <Alert>
          <Settings2 className="h-4 w-4" />
          <AlertDescription>
            No AI providers configured. Go to{' '}
            <a href="/settings/ai-providers" className="font-medium underline">
              Settings → AI Providers
            </a>{' '}
            to add your API credentials.
          </AlertDescription>
        </Alert>
      ) : (
        <>
          {/* Model Selection */}
          <div className="space-y-2">
            <Label htmlFor="llm-model">Model</Label>
            <Select
              value={config.model || undefined}
              onValueChange={handleModelChange}
              disabled={disabled}
            >
              <SelectTrigger id="llm-model">
                <SelectValue placeholder="Select a model" />
              </SelectTrigger>
              <SelectContent>
                {credentialGroups.map((group) => (
                  <SelectGroup key={group.credentialId || group.provider}>
                    <SelectLabel className="flex items-center gap-2">
                      <ProviderIcon provider={group.provider} />
                      {group.credentialName}
                    </SelectLabel>
                    {group.models.map((model) => (
                      <SelectItem key={model.id} value={model.id}>
                        <div className="flex items-center gap-2">
                          <span>{model.name}</span>
                          {model.is_custom && (
                            <span className="text-xs text-muted-foreground">(custom)</span>
                          )}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectGroup>
                ))}
              </SelectContent>
            </Select>
            {selectedModel && (
              <p className="text-xs text-muted-foreground">
                Using {PROVIDER_INFO[selectedModel.provider]?.name || selectedModel.provider}
                {selectedModel.credential_name && ` (${selectedModel.credential_name})`}
              </p>
            )}
          </div>

          {/* Temperature Slider */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <Label htmlFor="llm-temperature">Temperature</Label>
              <Tooltip>
                <TooltipTrigger asChild>
                  <span className="text-sm font-mono text-muted-foreground cursor-help">
                    {config.temperature.toFixed(1)}
                  </span>
                </TooltipTrigger>
                <TooltipContent>
                  <p className="text-xs max-w-[200px]">
                    Lower values (0-0.3) produce more consistent, deterministic evaluations.
                    Higher values (0.7+) allow for more varied interpretations.
                  </p>
                </TooltipContent>
              </Tooltip>
            </div>
            <div className="flex items-center gap-4">
              <span className="text-xs text-muted-foreground">Precise</span>
              <Slider
                id="llm-temperature"
                min={0}
                max={2}
                step={0.1}
                value={[config.temperature]}
                onValueChange={handleTemperatureChange}
                disabled={disabled}
                className="flex-1"
              />
              <span className="text-xs text-muted-foreground">Creative</span>
            </div>
            <TemperatureIndicator temperature={config.temperature} />
          </div>
        </>
      )}
    </div>
  )
}

// Simple provider icon component
function ProviderIcon({ provider }: { provider: AIProvider }) {
  const iconClass = 'h-3 w-3'

  switch (provider) {
    case 'openai':
      return <span className={cn(iconClass, 'text-green-600')}>●</span>
    case 'anthropic':
      return <span className={cn(iconClass, 'text-orange-600')}>●</span>
    case 'azure':
      return <span className={cn(iconClass, 'text-blue-600')}>●</span>
    case 'gemini':
      return <span className={cn(iconClass, 'text-purple-600')}>●</span>
    case 'openrouter':
      return <span className={cn(iconClass, 'text-pink-600')}>●</span>
    default:
      return <span className={cn(iconClass, 'text-gray-600')}>●</span>
  }
}

// Visual indicator for temperature setting
function TemperatureIndicator({ temperature }: { temperature: number }) {
  const getLabel = () => {
    if (temperature <= 0.3) return { text: 'Consistent scoring', color: 'text-blue-600' }
    if (temperature <= 0.7) return { text: 'Balanced', color: 'text-green-600' }
    if (temperature <= 1.2) return { text: 'Variable', color: 'text-yellow-600' }
    return { text: 'Highly creative', color: 'text-orange-600' }
  }

  const label = getLabel()

  return (
    <p className={cn('text-xs', label.color)}>
      {label.text}
    </p>
  )
}
