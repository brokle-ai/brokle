'use client'

import { useState, useMemo } from 'react'
import { Check, ChevronDown, ChevronRight, Plus, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import {
  ProviderIcon,
  ProviderDialog,
  PROVIDER_INFO,
  useAIProvidersQuery,
  useModelsByProvider,
  type AIProvider,
  type AIProviderCredential,
  type AvailableModel,
} from '@/features/ai-providers'

interface ModelSelectorProps {
  value: string | undefined
  credentialId?: string // Currently selected credential ID
  onChange: (model: string, provider: string, credentialId?: string) => void
  disabled?: boolean
  compact?: boolean
  orgId?: string // Organization ID for fetching available models
}

export function ModelSelector({
  value,
  credentialId,
  onChange,
  disabled,
  compact = false,
  orgId,
}: ModelSelectorProps) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [activeProvider, setActiveProvider] = useState<AIProvider | null>(null)
  const [showProviderDialog, setShowProviderDialog] = useState(false)

  const {
    modelsByProvider,
    configuredProviders,
    isLoading: modelsLoading,
  } = useModelsByProvider(orgId)

  // Fetch configured providers for the organization (for the dialog)
  const { data: providerCredentials = [] } = useAIProvidersQuery(orgId || '', {
    enabled: !!orgId,
  })

  const modelDisplayNames = useMemo(() => {
    const lookup: Record<string, string> = {}
    for (const models of Object.values(modelsByProvider)) {
      for (const model of models || []) {
        lookup[model.id] = model.name
      }
    }
    return lookup
  }, [modelsByProvider])

  const getProviderForModel = (modelId: string | undefined): AIProvider | null => {
    if (!modelId) return null
    for (const [provider, models] of Object.entries(modelsByProvider)) {
      if (models?.some((m) => m.id === modelId)) {
        return provider as AIProvider
      }
    }
    return null
  }

  // Set initial active provider based on selected value or first configured provider
  const effectiveActiveProvider = useMemo(() => {
    if (activeProvider && configuredProviders.includes(activeProvider)) {
      return activeProvider
    }
    const providerFromValue = getProviderForModel(value)
    if (providerFromValue && configuredProviders.includes(providerFromValue)) {
      return providerFromValue
    }
    return configuredProviders[0] || null
  }, [activeProvider, value, configuredProviders, modelsByProvider])

  const filteredModels = useMemo(() => {
    if (!effectiveActiveProvider) return []
    const models = modelsByProvider[effectiveActiveProvider] || []
    if (!search.trim()) return models

    const searchLower = search.toLowerCase()
    return models.filter(
      (model) =>
        model.name.toLowerCase().includes(searchLower) ||
        model.id.toLowerCase().includes(searchLower)
    )
  }, [effectiveActiveProvider, modelsByProvider, search])

  // Check if search matches any model in any provider (for auto-switching)
  const searchMatchingProvider = useMemo(() => {
    if (!search.trim()) return null
    const searchLower = search.toLowerCase()

    for (const provider of configuredProviders) {
      const models = modelsByProvider[provider] || []
      if (
        models.some(
          (m) =>
            m.name.toLowerCase().includes(searchLower) ||
            m.id.toLowerCase().includes(searchLower)
        )
      ) {
        return provider
      }
    }
    return null
  }, [search, configuredProviders, modelsByProvider])

  // Auto-switch to provider with matching results when searching
  const displayProvider = useMemo(() => {
    if (search.trim() && filteredModels.length === 0 && searchMatchingProvider) {
      return searchMatchingProvider
    }
    return effectiveActiveProvider
  }, [effectiveActiveProvider, search, filteredModels.length, searchMatchingProvider])

  const displayModels = useMemo(() => {
    if (!displayProvider) return []
    const models = modelsByProvider[displayProvider] || []
    if (!search.trim()) return models

    const searchLower = search.toLowerCase()
    return models.filter(
      (model) =>
        model.name.toLowerCase().includes(searchLower) ||
        model.id.toLowerCase().includes(searchLower)
    )
  }, [displayProvider, modelsByProvider, search])

  const selectedModel = value ? modelDisplayNames[value] || value : undefined
  const selectedProvider = getProviderForModel(value)

  const handleSelectModel = (model: AvailableModel) => {
    // Pass model ID, provider, and credential ID (when present)
    const provider = model.provider || displayProvider || ''
    onChange(model.id, provider, model.credential_id)
    setOpen(false)
    setSearch('')
  }

  if (modelsLoading) {
    return <Skeleton className={cn('rounded-md', compact ? 'h-8 w-[180px]' : 'h-10 w-full')} />
  }

  // No providers configured state
  if (configuredProviders.length === 0 && orgId) {
    return (
      <>
        <Button
          variant="outline"
          onClick={() => setShowProviderDialog(true)}
          disabled={disabled}
          className={cn(
            'justify-start font-normal text-muted-foreground',
            compact ? 'h-8 w-auto max-w-[280px] text-xs px-2' : 'w-full'
          )}
        >
          <Plus className={cn('shrink-0', compact ? 'h-3 w-3 mr-1' : 'h-4 w-4 mr-2')} />
          <span>Add AI Provider</span>
        </Button>
        <ProviderDialog
          orgId={orgId}
          open={showProviderDialog}
          onOpenChange={setShowProviderDialog}
          existingCredentials={providerCredentials}
        />
      </>
    )
  }

  return (
    <>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            disabled={disabled}
            className={cn(
              'justify-between font-normal',
              compact ? 'h-8 w-auto max-w-[280px] text-xs px-2' : 'w-full'
            )}
          >
            <div className="flex items-center gap-2 min-w-0">
              {selectedProvider && (
                <ProviderIcon
                  provider={selectedProvider}
                  className={cn('shrink-0', compact ? 'h-3.5 w-3.5' : 'h-4 w-4')}
                />
              )}
              <span className={cn("truncate", !selectedModel && "text-muted-foreground")}>{selectedModel || 'Select model'}</span>
            </div>
            <ChevronDown className={cn('shrink-0 opacity-50', compact ? 'h-3 w-3' : 'h-4 w-4')} />
          </Button>
        </PopoverTrigger>

        <PopoverContent
          className="w-[520px] p-0"
          align="start"
          side="bottom"
          sideOffset={4}
        >
          {/* Search Input */}
          <div className="flex items-center gap-2 border-b px-3 py-2">
            <Search className="h-4 w-4 shrink-0 text-muted-foreground" />
            <Input
              placeholder="Find a model"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="h-8 border-0 p-0 shadow-none focus-visible:ring-0"
            />
          </div>

          {/* Split Pane */}
          <div className="flex">
            {/* Left Column: Providers */}
            <div className="w-[180px] shrink-0 border-r">
              <ScrollArea className="h-[280px]">
                <div className="p-1">
                  {configuredProviders.map((provider) => {
                    const info = PROVIDER_INFO[provider]
                    const isActive = displayProvider === provider
                    const modelCount = modelsByProvider[provider]?.length || 0

                    return (
                      <button
                        key={provider}
                        onClick={() => {
                          setActiveProvider(provider)
                          setSearch('')
                        }}
                        onMouseEnter={() => setActiveProvider(provider)}
                        className={cn(
                          'flex w-full items-center justify-between gap-2 rounded-sm px-2 py-1.5 text-sm',
                          'hover:bg-accent hover:text-accent-foreground',
                          'focus:bg-accent focus:text-accent-foreground focus:outline-none',
                          isActive && 'bg-accent text-accent-foreground'
                        )}
                      >
                        <div className="flex items-center gap-2">
                          <ProviderIcon provider={provider} className="h-4 w-4" />
                          <span>{info?.name || provider}</span>
                          <span className="text-xs text-muted-foreground">({modelCount})</span>
                        </div>
                        <ChevronRight className="h-3 w-3 opacity-50" />
                      </button>
                    )
                  })}
                </div>

                {/* Add Provider Link */}
                {orgId && (
                  <>
                    <Separator className="my-1" />
                    <div className="p-1">
                      <button
                        onClick={() => {
                          setOpen(false)
                          setShowProviderDialog(true)
                        }}
                        className={cn(
                          'flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm',
                          'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
                          'focus:bg-accent focus:text-accent-foreground focus:outline-none'
                        )}
                      >
                        <Plus className="h-4 w-4" />
                        <span>Add AI provider</span>
                      </button>
                    </div>
                  </>
                )}
              </ScrollArea>
            </div>

            {/* Right Column: Models */}
            <div className="flex-1 min-w-0 overflow-hidden">
              <ScrollArea className="h-[280px]">
                <div className="p-1">
                  {displayModels.length > 0 ? (
                    displayModels.map((model) => {
                      // For selection matching: compare model ID + credential ID when present
                      const isSelected = value === model.id &&
                        (!model.credential_id || credentialId === model.credential_id)
                      // Unique key: combine model ID with credential ID to handle duplicates
                      const modelKey = model.credential_id
                        ? `${model.id}-${model.credential_id}`
                        : model.id
                      // Display name: append credential name when present
                      const displayName = model.credential_name
                        ? `${model.name} (${model.credential_name})`
                        : model.name

                      return (
                        <button
                          key={modelKey}
                          onClick={() => handleSelectModel(model)}
                          className={cn(
                            'flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm overflow-hidden',
                            'hover:bg-accent hover:text-accent-foreground',
                            'focus:bg-accent focus:text-accent-foreground focus:outline-none',
                            isSelected && 'bg-accent'
                          )}
                        >
                          <Check
                            className={cn(
                              'h-4 w-4 shrink-0',
                              isSelected ? 'opacity-100' : 'opacity-0'
                            )}
                          />
                          <ProviderIcon
                            provider={displayProvider!}
                            className="h-4 w-4 shrink-0"
                          />
                          <span className="truncate">{displayName}</span>
                        </button>
                      )
                    })
                  ) : (
                    <div className="px-2 py-6 text-center text-sm text-muted-foreground">
                      {search ? 'No models found' : 'No models available'}
                    </div>
                  )}
                </div>
              </ScrollArea>
            </div>
          </div>
        </PopoverContent>
      </Popover>

      {/* Provider Dialog */}
      {orgId && (
        <ProviderDialog
          orgId={orgId}
          open={showProviderDialog}
          onOpenChange={setShowProviderDialog}
          existingCredentials={providerCredentials}
        />
      )}
    </>
  )
}
