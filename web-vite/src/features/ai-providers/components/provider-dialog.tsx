import { useEffect, useState } from 'react'
import { CheckCircle2, Eye, EyeOff, Loader2, XCircle } from 'lucide-react'
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
  useTestConnectionMutation,
  useUpdateProviderMutation,
} from '../api/queries'
import type {
  AIProvider,
  AIProviderCredential,
  CreateProviderRequest,
  TestConnectionRequest,
  UpdateProviderRequest,
} from '../api/types'
import { AVAILABLE_PROVIDERS, PROVIDER_INFO } from '../api/types'
import { ProviderIcon } from './provider-icon'

interface ProviderDialogProps {
  orgId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  // `existingCredential` toggles edit mode. When provided, the form
  // pre-populates from it and the adapter select is hidden (adapter
  // type is immutable post-creation).
  existingCredential?: AIProviderCredential
  existingCredentials: AIProviderCredential[]
}

// Add / Edit provider dialog. Form state lives locally (not in react-
// hook-form) because the fields are conditional on adapter and the
// three-state headers-update pattern (unset / clear / replace) doesn't
// map cleanly to RHF's dirty tracking. Test-connection uses the
// current form values without committing them.
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

  // Form state. `adapter` is the API protocol (openai/anthropic/…);
  // `name` is the user-facing config label, unique within an org.
  const [adapter, setAdapter] = useState<AIProvider | ''>(
    existingCredential?.adapter ?? '',
  )
  const [name, setName] = useState(existingCredential?.name ?? '')
  const [apiKey, setApiKey] = useState('')
  const [baseUrl, setBaseUrl] = useState(existingCredential?.base_url ?? '')
  const [config, setConfig] = useState<Record<string, string>>(() => {
    const existing = existingCredential?.config as
      | Record<string, string>
      | undefined
    return existing ?? {}
  })
  const [headers, setHeaders] = useState('')
  const [customModels, setCustomModels] = useState(
    existingCredential?.custom_models?.join(', ') ?? '',
  )
  const [showApiKey, setShowApiKey] = useState(false)
  const [testResult, setTestResult] = useState<'success' | 'error' | null>(null)

  const adapterInfo = adapter ? PROVIDER_INFO[adapter] : null

  // Reset form when the dialog opens. Edit-mode repopulates from the
  // credential; add-mode defaults the adapter to the first entry so
  // the Test button has a reasonable state to read from.
  useEffect(() => {
    if (!open) return
    if (existingCredential) {
      setAdapter(existingCredential.adapter)
      setName(existingCredential.name)
      setApiKey('')
      setBaseUrl(existingCredential.base_url ?? '')
      setConfig((existingCredential.config as Record<string, string>) ?? {})
      setHeaders(
        existingCredential.headers &&
          Object.keys(existingCredential.headers).length > 0
          ? JSON.stringify(existingCredential.headers, null, 2)
          : '',
      )
      setCustomModels(existingCredential.custom_models?.join(', ') ?? '')
    } else {
      setAdapter(AVAILABLE_PROVIDERS[0])
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

  const handleAdapterChange = (next: AIProvider) => {
    setAdapter(next)
    setConfig({})
    setTestResult(null)
    setCustomModels('')
    if (!PROVIDER_INFO[next].requiresBaseUrl) setBaseUrl('')
  }

  const handleConfigChange = (key: string, value: string) => {
    setConfig((prev) => ({ ...prev, [key]: value }))
  }

  const parseCustomModels = (): string[] | undefined => {
    if (!customModels.trim()) return undefined
    return customModels
      .split(',')
      .map((m) => m.trim())
      .filter(Boolean)
  }

  const parseHeaders = (): Record<string, string> | null => {
    if (!headers.trim()) return null
    try {
      return JSON.parse(headers) as Record<string, string>
    } catch {
      return null
    }
  }

  const buildCreateRequest = (): CreateProviderRequest => {
    const req: CreateProviderRequest = {
      name: name.trim(),
      adapter: adapter as AIProvider,
      api_key: apiKey,
    }
    if (baseUrl) req.base_url = baseUrl
    if (Object.keys(config).length > 0) req.config = config
    const parsedHeaders = parseHeaders()
    if (parsedHeaders) req.headers = parsedHeaders
    const parsedModels = parseCustomModels()
    if (parsedModels) req.custom_models = parsedModels
    return req
  }

  // Three-state headers semantics — undefined = leave alone, empty
  // object = clear, populated = replace. The backend ignores `null`
  // distinctly from `undefined`, so we never send `null` on the wire.
  const buildUpdateRequest = (): UpdateProviderRequest => {
    const req: UpdateProviderRequest = {}
    if (name.trim() !== existingCredential?.name) req.name = name.trim()
    if (apiKey) req.api_key = apiKey
    if (baseUrl !== (existingCredential?.base_url ?? '')) {
      req.base_url = baseUrl || undefined
    }
    if (Object.keys(config).length > 0) req.config = config

    const existingHeaders = existingCredential?.headers
    const hadHeaders =
      existingHeaders && Object.keys(existingHeaders).length > 0
    if (headers.trim()) {
      const parsed = parseHeaders()
      if (
        parsed &&
        JSON.stringify(parsed) !== JSON.stringify(existingHeaders ?? {})
      ) {
        req.headers = parsed
      }
    } else if (hadHeaders) {
      req.headers = {}
    }

    const parsedModels = parseCustomModels()
    if (parsedModels) req.custom_models = parsedModels
    return req
  }

  const buildTestRequest = (): TestConnectionRequest => {
    const req: TestConnectionRequest = {
      adapter: adapter as AIProvider,
      api_key: apiKey,
    }
    if (baseUrl) req.base_url = baseUrl
    if (Object.keys(config).length > 0) req.config = config
    const parsed = parseHeaders()
    if (parsed) req.headers = parsed
    return req
  }

  const handleTest = async () => {
    if (!apiKey) return
    setTestResult(null)
    try {
      const result = await testMutation.mutateAsync(buildTestRequest())
      setTestResult(result.success ? 'success' : 'error')
    } catch {
      setTestResult('error')
    }
  }

  const handleSave = async () => {
    try {
      if (isEdit && existingCredential) {
        await updateMutation.mutateAsync({
          credentialId: existingCredential.id,
          data: buildUpdateRequest(),
        })
      } else {
        await createMutation.mutateAsync(buildCreateRequest())
      }
      onOpenChange(false)
    } catch {
      // Mutation onError surfaces the toast.
    }
  }

  const isNameTaken = () => {
    if (!name.trim()) return false
    return existingCredentials.some(
      (c) =>
        c.name.toLowerCase() === name.trim().toLowerCase() &&
        c.id !== existingCredential?.id,
    )
  }

  const isFormValid = () => {
    if (!adapter || !adapterInfo) return false
    if (!name.trim()) return false
    if (isNameTaken()) return false
    if (!isEdit && !apiKey) return false
    if (adapterInfo.requiresBaseUrl && !baseUrl) return false
    if (adapter === 'custom' && !customModels.trim()) return false
    if (adapterInfo.configFields) {
      for (const field of adapterInfo.configFields) {
        if (field.required && !config[field.key]?.trim()) return false
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
            {isEdit
              ? `Edit ${existingCredential?.name ?? 'Provider'}`
              : 'Add AI provider'}
          </DialogTitle>
          <DialogDescription>
            {isEdit
              ? 'Update the API credentials for this configuration.'
              : 'Configure API credentials to enable AI features.'}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="provider-name">Configuration name *</Label>
            <Input
              id="provider-name"
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
                Unique name to identify this configuration
              </p>
            )}
          </div>

          {!isEdit && (
            <div className="space-y-2">
              <Label htmlFor="provider-adapter">Provider type *</Label>
              <Select
                value={adapter}
                onValueChange={(v) => handleAdapterChange(v as AIProvider)}
              >
                <SelectTrigger id="provider-adapter">
                  <SelectValue placeholder="Select a provider type" />
                </SelectTrigger>
                <SelectContent>
                  {AVAILABLE_PROVIDERS.map((p) => (
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
                <p className="text-xs text-muted-foreground">
                  {adapterInfo.description}
                </p>
              )}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="provider-api-key">
              API key {isEdit ? '(leave blank to keep current)' : '*'}
            </Label>
            <div className="relative">
              <Input
                id="provider-api-key"
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
                onClick={() => setShowApiKey((v) => !v)}
              >
                {showApiKey ? (
                  <EyeOff className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <Eye className="h-4 w-4 text-muted-foreground" />
                )}
              </Button>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="provider-base-url">
              Base URL {adapterInfo?.requiresBaseUrl ? '*' : '(optional)'}
            </Label>
            <Input
              id="provider-base-url"
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
                Override the default API endpoint (proxies or self-hosted)
              </p>
            )}
          </div>

          {adapterInfo?.configFields?.map((field) => (
            <div key={field.key} className="space-y-2">
              <Label htmlFor={`provider-${field.key}`}>
                {field.label} {field.required && '*'}
              </Label>
              <Input
                id={`provider-${field.key}`}
                value={config[field.key] ?? ''}
                onChange={(e) =>
                  handleConfigChange(field.key, e.target.value)
                }
                placeholder={field.placeholder}
              />
            </div>
          ))}

          {adapter && (
            <div className="space-y-2">
              <Label htmlFor="provider-custom-models">
                {adapter === 'custom'
                  ? 'Available models *'
                  : 'Custom models (optional)'}
              </Label>
              <Input
                id="provider-custom-models"
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
                  ? 'Comma-separated model IDs exposed by this provider'
                  : 'Add fine-tuned or private models (comma-separated)'}
              </p>
            </div>
          )}

          {adapter === 'custom' && (
            <div className="space-y-2">
              <Label htmlFor="provider-headers">Custom headers (JSON)</Label>
              <Textarea
                id="provider-headers"
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

          {testResult && (
            <div
              className={`flex items-center gap-2 rounded-md p-3 ${
                testResult === 'success'
                  ? 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300'
                  : 'bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300'
              }`}
            >
              {testResult === 'success' ? (
                <>
                  <CheckCircle2 className="h-4 w-4" />
                  <span>Connection successful. The API key is valid.</span>
                </>
              ) : (
                <>
                  <XCircle className="h-4 w-4" />
                  <span>Connection failed. Check your credentials.</span>
                </>
              )}
            </div>
          )}
        </div>

        <DialogFooter className="flex-col gap-2 sm:flex-row">
          <Button
            variant="outline"
            onClick={handleTest}
            disabled={!adapter || !apiKey || testMutation.isPending}
            className="sm:mr-auto"
          >
            {testMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Testing…
              </>
            ) : (
              'Test connection'
            )}
          </Button>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isSaving}
          >
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={!isFormValid() || isSaving}>
            {isSaving ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving…
              </>
            ) : isEdit ? (
              'Update configuration'
            ) : (
              'Add configuration'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
