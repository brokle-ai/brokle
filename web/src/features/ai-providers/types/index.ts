export type AIProvider =
  | 'openai'
  | 'anthropic'
  | 'azure'
  | 'gemini'
  | 'openrouter'
  | 'custom';

export interface ProviderInfo {
  id: AIProvider;
  name: string;
  description: string;
  requiresBaseUrl: boolean;
  configFields?: ProviderConfigField[];
}

export interface ProviderConfigField {
  key: string;
  label: string;
  placeholder: string;
  required: boolean;
  type: 'text' | 'select';
  options?: string[];
}

export interface AIProviderCredential {
  id: string;
  project_id: string;
  name: string; // Unique configuration name (e.g., "OpenAI Production")
  adapter: AIProvider; // API protocol type
  key_preview: string;
  base_url?: string;
  config?: Record<string, unknown>;
  custom_models?: string[];
  headers?: Record<string, string>; // Custom headers (decrypted for editing)
  created_at: string;
  updated_at: string;
}

export interface CreateProviderRequest {
  name: string; // Required unique configuration name
  adapter: AIProvider;
  api_key: string;
  base_url?: string;
  config?: Record<string, unknown>;
  custom_models?: string[];
  headers?: Record<string, string>;
}

export interface UpdateProviderRequest {
  name?: string;
  api_key?: string;
  base_url?: string;
  config?: Record<string, unknown>;
  custom_models?: string[];
  // headers: undefined = don't change, null = clear, object = set new headers
  headers?: Record<string, string> | null;
}

export interface TestConnectionRequest {
  adapter: AIProvider;
  api_key: string;
  base_url?: string;
  config?: Record<string, unknown>;
  headers?: Record<string, string>;
}

export interface TestConnectionResponse {
  success: boolean;
  error?: string;
}

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
};

export const AVAILABLE_PROVIDERS: AIProvider[] = [
  'openai',
  'anthropic',
  'azure',
  'gemini',
  'openrouter',
  'custom',
];

/**
 * Available model from API (based on configured providers)
 */
export interface AvailableModel {
  id: string;
  name: string;
  provider: AIProvider;
  credential_id?: string;   // Present when multiple credentials exist for provider
  credential_name?: string; // Credential display name for UI
  is_custom?: boolean;      // True for user-defined custom models
}

export type ModelsByProvider = Partial<Record<AIProvider, AvailableModel[]>>;

export function getAdapterDisplayName(adapter: AIProvider): string {
  return PROVIDER_INFO[adapter]?.name ?? adapter;
}
