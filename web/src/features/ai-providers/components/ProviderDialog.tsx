'use client'

import { useState, useEffect } from 'react'
import { Loader2, CheckCircle2, XCircle, Eye, EyeOff } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  useCreateProviderMutation,
  useUpdateProviderMutation,
  useTestConnectionMutation,
} from '../hooks/use-ai-providers'
import type { AIProvider, AIProviderCredential, CreateProviderRequest, UpdateProviderRequest, TestConnectionRequest } from '../types'
import { PROVIDER_INFO, AVAILABLE_PROVIDERS } from '../types'
import { ProviderIcon } from './ProviderIcon'

interface ProviderDialogProps {
  orgId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  existingCredential?: AIProviderCredential
  existingCredentials: AIProviderCredential[]
}

export function ProviderDialog({
  orgId,
  open,
  onOpenChange,
  existingCredential,
  existingCredentials,
}: ProviderDialogProps) {
  const isEdit = !!existingCredential

  const createMutation = useCreateProviderMutation(orgId)
  const updateMutation = useUpdateProviderMutation(orgId)
  const testMutation = useTestConnectionMutation(orgId)

  // Form state - using 'adapter' for API protocol type, 'name' for configuration name
  const [adapter, setAdapter] = useState<AIProvider | ''>(existingCredential?.adapter || '')
  const [name, setName] = useState(existingCredential?.name || '')
  const [apiKey, setApiKey] = useState('')
  const [baseUrl, setBaseUrl] = useState(existingCredential?.base_url || '')
  const [config, setConfig] = useState<Record<string, string>>(() => {
    const existingConfig = existingCredential?.config as Record<string, string> | undefined
    return existingConfig || {}
  })
  const [headers, setHeaders] = useState('')
  const [customModels, setCustomModels] = useState(() => {
    return existingCredential?.custom_models?.join(', ') || ''
  })
  const [showApiKey, setShowApiKey] = useState(false)
  const [testResult, setTestResult] = useState<'success' | 'error' | null>(null)

  // Get adapter info (null when no adapter selected)
  const adapterInfo = adapter ? PROVIDER_INFO[adapter] : null

  // Reset form when dialog opens/closes
  useEffect(() => {
    if (open && existingCredential) {
      // Edit mode: populate from existing credential
      setAdapter(existingCredential.adapter)
      setName(existingCredential.name)
      setApiKey('')
      setBaseUrl(existingCredential.base_url || '')
      setConfig((existingCredential.config as Record<string, string>) || {})
      // Show existing headers for editing (decrypted from backend)
      setHeaders(
        existingCredential.headers && Object.keys(existingCredential.headers).length > 0
          ? JSON.stringify(existingCredential.headers, null, 2)
          : ''
      )
      setCustomModels(existingCredential.custom_models?.join(', ') || '')
    } else if (open && !existingCredential) {
      // Add mode: reset to defaults
      setAdapter(AVAILABLE_PROVIDERS[0]) // Default to first adapter
      setName('')
      setApiKey('')
      setBaseUrl('')
      setConfig({})
      setHeaders('')
      setCustomModels('')
    }
    setTestResult(null)
    setShowApiKey(false)
  }, [open, existingCredential])

  // All adapters are available for selection (multiple configs allowed per adapter)
  const availableForSelection = AVAILABLE_PROVIDERS

  const handleAdapterChange = (newAdapter: AIProvider) => {
    setAdapter(newAdapter)
    setConfig({})
    setTestResult(null)
    setCustomModels('')
    // Reset base_url if not required by the new adapter
    const newAdapterInfo = PROVIDER_INFO[newAdapter]
    if (!newAdapterInfo.requiresBaseUrl) {
      setBaseUrl('')
    }
  }

  const handleConfigChange = (key: string, value: string) => {
    setConfig((prev) => ({ ...prev, [key]: value }))
  }

  const buildCreateRequest = (): CreateProviderRequest => {
    const request: CreateProviderRequest = {
      name: name.trim(),
      adapter: adapter as AIProvider, // Safe because we validate in isFormValid
      api_key: apiKey,
    }

    if (baseUrl) {
      request.base_url = baseUrl
    }

    if (Object.keys(config).length > 0) {
      request.config = config
    }

    if (headers.trim()) {
      try {
        request.headers = JSON.parse(headers)
      } catch {
        // Invalid JSON, ignore headers
      }
    }

    // Parse custom models from comma-separated string
    if (customModels.trim()) {
      request.custom_models = customModels
        .split(',')
        .map((m) => m.trim())
        .filter(Boolean)
    }

    return request
  }

  const buildUpdateRequest = (): UpdateProviderRequest => {
    const request: UpdateProviderRequest = {}

    // Only include fields that have changed or been set
    if (name.trim() !== existingCredential?.name) {
      request.name = name.trim()
    }

    if (apiKey) {
      request.api_key = apiKey
    }

    if (baseUrl !== (existingCredential?.base_url || '')) {
      request.base_url = baseUrl || undefined
    }

    if (Object.keys(config).length > 0) {
      request.config = config
    }

    // Handle headers: detect changes and support clearing
    const existingHeaders = existingCredential?.headers
    const hadHeaders = existingHeaders && Object.keys(existingHeaders).length > 0

    if (headers.trim()) {
      try {
        const parsedHeaders = JSON.parse(headers)
        // Only include if changed from existing
        if (JSON.stringify(parsedHeaders) !== JSON.stringify(existingHeaders || {})) {
          request.headers = parsedHeaders
        }
      } catch {
        // Invalid JSON, ignore headers
      }
    } else if (hadHeaders) {
      // Headers were cleared - send empty object to trigger clearing on backend
      // Backend will see non-nil pointer with empty map and clear headers
      request.headers = {}
    }

    // Parse custom models from comma-separated string
    const parsedModels = customModels.trim()
      ? customModels.split(',').map((m) => m.trim()).filter(Boolean)
      : undefined
    if (parsedModels) {
      request.custom_models = parsedModels
    }

    return request
  }

  const buildTestRequest = (): TestConnectionRequest => {
    const request: TestConnectionRequest = {
      adapter: adapter as AIProvider,
      api_key: apiKey,
    }

    if (baseUrl) {
      request.base_url = baseUrl
    }

    if (Object.keys(config).length > 0) {
      request.config = config
    }

    if (headers.trim()) {
      try {
        request.headers = JSON.parse(headers)
      } catch {
        // Invalid JSON, ignore headers
      }
    }

    return request
  }

  const handleTestConnection = async () => {
    if (!apiKey) return

    setTestResult(null)
    const request = buildTestRequest()

    try {
      const result = await testMutation.mutateAsync(request)
      setTestResult(result.success ? 'success' : 'error')
    } catch {
      setTestResult('error')
    }
  }

  const handleSave = async () => {
    try {
      if (isEdit && existingCredential) {
        // Update existing credential
        await updateMutation.mutateAsync({
          credentialId: existingCredential.id,
          data: buildUpdateRequest(),
        })
      } else {
        // Create new credential
        await createMutation.mutateAsync(buildCreateRequest())
      }
      onOpenChange(false)
    } catch {
      // Error handled by mutation
    }
  }

  // Check if configuration name already exists (for uniqueness validation)
  const isNameTaken = () => {
    if (!name.trim()) return false
    return existingCredentials.some(
      (c) => c.name.toLowerCase() === name.trim().toLowerCase() && c.id !== existingCredential?.id
    )
  }

  const isFormValid = () => {
    if (!adapter) return false // Must select an adapter
    if (!adapterInfo) return false
    // Name is always required and must be unique
    if (!name.trim()) return false
    if (isNameTaken()) return false
    // API key required for new providers, optional for edits
    if (!isEdit && !apiKey) return false
    if (adapterInfo.requiresBaseUrl && !baseUrl) return false
    // Custom adapter requires at least one model defined
    if (adapter === 'custom' && !customModels.trim()) return false
    // Check required config fields (e.g., Azure deployment_id)
    if (adapterInfo.configFields) {
      for (const field of adapterInfo.configFields) {
        if (field.required && !config[field.key]?.trim()) {
          return false
        }
      }
    }
    return true
  }

  const isSaving = createMutation.isPending || updateMutation.isPending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>
            {isEdit ? `Edit ${existingCredential?.name || 'Provider'}` : 'Add AI Provider'}
          </DialogTitle>
          <DialogDescription>
            {isEdit
              ? 'Update the API credentials for this configuration.'
              : 'Configure API credentials to enable AI features.'}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Configuration Name - Always required */}
          <div className="space-y-2">
            <Label htmlFor="name">Configuration Name *</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., OpenAI Production, Claude Dev"
              className={isNameTaken() ? 'border-destructive' : ''}
            />
            {isNameTaken() ? (
              <p className="text-xs text-destructive">
                A configuration with this name already exists
              </p>
            ) : (
              <p className="text-xs text-muted-foreground">
                A unique name to identify this configuration
              </p>
            )}
          </div>

          {/* Adapter Selection */}
          {!isEdit && (
            <div className="space-y-2">
              <Label htmlFor="adapter">Provider Type *</Label>
              <Select value={adapter} onValueChange={handleAdapterChange}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a provider type" />
                </SelectTrigger>
                <SelectContent>
                  {availableForSelection.map((p) => (
                    <SelectItem key={p} value={p}>
                      <div className="flex items-center gap-2">
                        <ProviderIcon provider={p} className="h-4 w-4" />
                        <span>{PROVIDER_INFO[p].name}</span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {adapterInfo && (
                <p className="text-xs text-muted-foreground">{adapterInfo.description}</p>
              )}
            </div>
          )}

          {/* API Key */}
          <div className="space-y-2">
            <Label htmlFor="apiKey">
              API Key {isEdit ? '(leave blank to keep current)' : '*'}
            </Label>
            <div className="relative">
              <Input
                id="apiKey"
                type={showApiKey ? 'text' : 'password'}
                autoComplete="new-password"
                value={apiKey}
                onChange={(e) => {
                  setApiKey(e.target.value)
                  setTestResult(null)
                }}
                placeholder={isEdit ? '••••••••••••••••' : 'Enter API key'}
                className="pr-10"
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="absolute right-0 top-0 h-full px-3 hover:bg-transparent"
                onClick={() => setShowApiKey(!showApiKey)}
              >
                {showApiKey ? (
                  <EyeOff className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <Eye className="h-4 w-4 text-muted-foreground" />
                )}
              </Button>
            </div>
          </div>

          {/* Base URL (required for some adapters) */}
          <div className="space-y-2">
            <Label htmlFor="baseUrl">
              Base URL {adapterInfo?.requiresBaseUrl ? '*' : '(Optional)'}
            </Label>
            <Input
              id="baseUrl"
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
              placeholder={
                adapter === 'azure'
                  ? 'https://your-resource.openai.azure.com'
                  : adapter === 'custom'
                  ? 'http://localhost:11434/v1'
                  : 'Leave blank for default'
              }
            />
            {!adapterInfo?.requiresBaseUrl && (
              <p className="text-xs text-muted-foreground">
                Override the default API endpoint (for proxies or self-hosted)
              </p>
            )}
          </div>

          {/* Adapter-specific config fields */}
          {adapterInfo?.configFields?.map((field) => (
            <div key={field.key} className="space-y-2">
              <Label htmlFor={field.key}>
                {field.label} {field.required && '*'}
              </Label>
              <Input
                id={field.key}
                value={config[field.key] || ''}
                onChange={(e) => handleConfigChange(field.key, e.target.value)}
                placeholder={field.placeholder}
              />
            </div>
          ))}

          {/* Custom Models */}
          {adapter && (
            <div className="space-y-2">
              <Label htmlFor="customModels">
                {adapter === 'custom' ? 'Available Models *' : 'Custom Models (Optional)'}
              </Label>
              <Input
                id="customModels"
                value={customModels}
                onChange={(e) => setCustomModels(e.target.value)}
                placeholder={
                  adapter === 'custom'
                    ? 'llama-3.1, mistral-7b, codellama'
                    : 'ft:gpt-4o:my-org, my-fine-tuned-model'
                }
              />
              <p className="text-xs text-muted-foreground">
                {adapter === 'custom'
                  ? 'Comma-separated list of model IDs available on this provider'
                  : 'Add fine-tuned or private models (comma-separated)'}
              </p>
            </div>
          )}

          {/* Custom Headers (advanced, for custom adapter) */}
          {adapter === 'custom' && (
            <div className="space-y-2">
              <Label htmlFor="headers">Custom Headers (JSON)</Label>
              <Textarea
                id="headers"
                value={headers}
                onChange={(e) => setHeaders(e.target.value)}
                placeholder='{"X-Custom-Header": "value"}'
                rows={2}
                className="font-mono text-sm"
              />
              <p className="text-xs text-muted-foreground">
                Additional HTTP headers for authentication (stored encrypted)
              </p>
            </div>
          )}

          {/* Test Connection Result */}
          {testResult && (
            <div
              className={`flex items-center gap-2 p-3 rounded-md ${
                testResult === 'success'
                  ? 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300'
                  : 'bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300'
              }`}
            >
              {testResult === 'success' ? (
                <>
                  <CheckCircle2 className="h-4 w-4" />
                  <span>Connection successful! The API key is valid.</span>
                </>
              ) : (
                <>
                  <XCircle className="h-4 w-4" />
                  <span>Connection failed. Please check your credentials.</span>
                </>
              )}
            </div>
          )}
        </div>

        <DialogFooter className="flex-col sm:flex-row gap-2">
          <Button
            variant="outline"
            onClick={handleTestConnection}
            disabled={!adapter || !apiKey || testMutation.isPending}
            className="sm:mr-auto"
          >
            {testMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Testing...
              </>
            ) : (
              'Test Connection'
            )}
          </Button>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isSaving}>
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={!isFormValid() || isSaving}
          >
            {isSaving ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : isEdit ? (
              'Update Configuration'
            ) : (
              'Add Configuration'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
