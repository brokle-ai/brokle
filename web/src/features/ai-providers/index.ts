// Public exports for ai-providers feature

// Types
export type {
  AIProvider,
  ProviderInfo,
  ProviderConfigField,
  AIProviderCredential,
  CreateProviderRequest,
  UpdateProviderRequest,
  TestConnectionRequest,
  TestConnectionResponse,
  AvailableModel,
  ModelsByProvider,
} from './types'
export { PROVIDER_INFO, AVAILABLE_PROVIDERS, getAdapterDisplayName } from './types'

// API
export {
  listProviderCredentials,
  getProviderCredential,
  createProviderCredential,
  updateProviderCredential,
  deleteProviderCredential,
  testProviderConnection,
  createKeyPreview,
  getAvailableModels,
} from './api/ai-providers-api'

// Hooks
export {
  aiProviderQueryKeys,
  useAIProvidersQuery,
  useCreateProviderMutation,
  useUpdateProviderMutation,
  useDeleteProviderMutation,
  useTestConnectionMutation,
  useAvailableModelsQuery,
  useModelsByProvider,
} from './hooks/use-ai-providers'

// Components
export { AIProvidersSettings } from './components/AIProvidersSettings'
export { ProviderDialog } from './components/ProviderDialog'
export { ProviderIcon } from './components/ProviderIcon'
