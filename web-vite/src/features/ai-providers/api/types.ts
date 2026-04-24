// Wire types for AI provider credentials.
// Shapes mirror internal/transport/http/handlers/credentials/types.go —
// `/api/v1/organizations/{orgId}/credentials/ai` CRUD plus the
// `/test` and `/models` sibling endpoints.
//
// Credentials are organization-scoped (shared across projects) and
// stored encrypted at rest; list/detail responses only surface a
// `key_preview` rather than the raw key.

export type AIProvider =
  | 'openai'
  | 'anthropic'
  | 'azure'
  | 'gemini'
  | 'openrouter'
  | 'custom'

export interface ProviderConfigField {
  key: string
  label: string
  placeholder: string
  required: boolean
  type: 'text' | 'select'
  options?: string[]
}

export interface ProviderInfo {
  id: AIProvider
  name: string
  description: string
  requiresBaseUrl: boolean
  configFields?: ProviderConfigField[]
}

export interface AIProviderCredential {
  id: string
  organization_id: string
  name: string
  adapter: AIProvider
  key_preview: string
  base_url?: string
  config?: Record<string, unknown>
  custom_models?: string[]
  // Headers come back decrypted for edit-mode repopulation; undefined
  // on list responses from backends that don't embed them.
  headers?: Record<string, string>
  created_at: string
  updated_at: string
}

export interface CreateProviderRequest {
  name: string
  adapter: AIProvider
  api_key: string
  base_url?: string
  config?: Record<string, unknown>
  custom_models?: string[]
  headers?: Record<string, string>
}

// Update semantics follow the backend's three-state pattern on
// `headers`: `undefined` = don't touch, empty object = clear, populated
// map = replace. That's why `headers` is typed `Record | null` even
// though we use empty-object-as-clear on the wire (see dialog).
export interface UpdateProviderRequest {
  name?: string
  api_key?: string
  base_url?: string
  config?: Record<string, unknown>
  custom_models?: string[]
  headers?: Record<string, string> | null
}

export interface TestConnectionRequest {
  adapter: AIProvider
  api_key: string
  base_url?: string
  config?: Record<string, unknown>
  headers?: Record<string, string>
}

export interface TestConnectionResponse {
  success: boolean
  error?: string
}

export interface AvailableModel {
  id: string
  name: string
  provider: AIProvider
  credential_id?: string
  credential_name?: string
  is_custom?: boolean
}

export type ModelsByProvider = Partial<Record<AIProvider, AvailableModel[]>>

// Static provider metadata — display names, base-URL requirement,
// adapter-specific config fields (Azure deployment/version, Gemini
// location). The dialog reads this for label/placeholder copy; the
// columns read it for the adapter subtitle in the Provider cell.
export const PROVIDER_INFO: Record<AIProvider, ProviderInfo> = {
  openai: {
    id: 'openai',
    name: 'OpenAI',
    description: 'GPT-4, GPT-3.5, and other OpenAI models',
    requiresBaseUrl: false,
  },
  anthropic: {
    id: 'anthropic',
    name: 'Anthropic',
    description: 'Claude 3 Opus, Sonnet, Haiku and other Anthropic models',
    requiresBaseUrl: false,
  },
  azure: {
    id: 'azure',
    name: 'Azure OpenAI',
    description: 'OpenAI models hosted on Microsoft Azure',
    requiresBaseUrl: true,
    configFields: [
      {
        key: 'deployment_id',
        label: 'Deployment ID',
        placeholder: 'gpt-4-deployment',
        required: true,
        type: 'text',
      },
      {
        key: 'api_version',
        label: 'API Version',
        placeholder: '2024-02-01',
        required: false,
        type: 'text',
      },
    ],
  },
  gemini: {
    id: 'gemini',
    name: 'Google Gemini',
    description: 'Gemini Pro, Ultra and other Google AI models',
    requiresBaseUrl: false,
    configFields: [
      {
        key: 'location',
        label: 'Location',
        placeholder: 'us-central1',
        required: false,
        type: 'text',
      },
    ],
  },
  openrouter: {
    id: 'openrouter',
    name: 'OpenRouter',
    description: 'Access multiple providers through OpenRouter',
    requiresBaseUrl: false,
  },
  custom: {
    id: 'custom',
    name: 'Custom',
    description: 'Self-hosted models (vLLM, Ollama, etc.)',
    requiresBaseUrl: true,
  },
}

export const AVAILABLE_PROVIDERS: AIProvider[] = [
  'openai',
  'anthropic',
  'azure',
  'gemini',
  'openrouter',
  'custom',
]

export function getAdapterDisplayName(adapter: AIProvider): string {
  return PROVIDER_INFO[adapter]?.name ?? adapter
}
