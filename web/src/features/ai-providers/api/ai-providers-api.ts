/**
 * AI Providers API Client
 * CRUD operations and connection testing for AI provider credentials
 */

import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  AIProviderCredential,
  AvailableModel,
  CreateProviderRequest,
  UpdateProviderRequest,
  TestConnectionRequest,
  TestConnectionResponse,
} from '../types'

// Initialize API client with /api base path (Dashboard routes)
const client = new BrokleAPIClient('/api')

/**
 * List all AI provider credentials for a project
 *
 * @param projectId - Project ULID
 * @returns List of configured provider credentials
 *
 * @example
 * ```ts
 * const credentials = await listProviderCredentials('proj_123')
 * ```
 */
export async function listProviderCredentials(
  projectId: string
): Promise<AIProviderCredential[]> {
  const endpoint = `/v1/projects/${projectId}/credentials/ai`
  return client.get<AIProviderCredential[]>(endpoint)
}

/**
 * Get a specific AI provider credential by ID
 *
 * @param projectId - Project ULID
 * @param credentialId - Credential ULID
 * @returns Provider credential details
 *
 * @example
 * ```ts
 * const credential = await getProviderCredential('proj_123', 'cred_456')
 * ```
 */
export async function getProviderCredential(
  projectId: string,
  credentialId: string
): Promise<AIProviderCredential> {
  const endpoint = `/v1/projects/${projectId}/credentials/ai/${credentialId}`
  return client.get<AIProviderCredential>(endpoint)
}

/**
 * Create a new AI provider credential
 *
 * IMPORTANT: The API key is encrypted at rest. Only the key_preview
 * (masked version) is returned in responses.
 *
 * @param projectId - Project ULID
 * @param data - Provider credential data including name and adapter type
 * @returns Created credential
 *
 * @example
 * ```ts
 * const credential = await createProviderCredential('proj_123', {
 *   name: 'OpenAI Production',
 *   adapter: 'openai',
 *   api_key: 'sk-...',
 * })
 * ```
 */
export async function createProviderCredential(
  projectId: string,
  data: CreateProviderRequest
): Promise<AIProviderCredential> {
  const endpoint = `/v1/projects/${projectId}/credentials/ai`
  return client.post<AIProviderCredential, CreateProviderRequest>(endpoint, data)
}

/**
 * Update an existing AI provider credential
 *
 * @param projectId - Project ULID
 * @param credentialId - Credential ULID to update
 * @param data - Fields to update
 * @returns Updated credential
 *
 * @example
 * ```ts
 * const credential = await updateProviderCredential('proj_123', 'cred_456', {
 *   name: 'OpenAI Development',
 *   api_key: 'sk-new-key...',
 * })
 * ```
 */
export async function updateProviderCredential(
  projectId: string,
  credentialId: string,
  data: UpdateProviderRequest
): Promise<AIProviderCredential> {
  const endpoint = `/v1/projects/${projectId}/credentials/ai/${credentialId}`
  return client.patch<AIProviderCredential, UpdateProviderRequest>(endpoint, data)
}

/**
 * Delete an AI provider credential by ID
 *
 * @param projectId - Project ULID
 * @param credentialId - Credential ULID to delete
 *
 * @example
 * ```ts
 * await deleteProviderCredential('proj_123', 'cred_456')
 * ```
 */
export async function deleteProviderCredential(
  projectId: string,
  credentialId: string
): Promise<void> {
  const endpoint = `/v1/projects/${projectId}/credentials/ai/${credentialId}`
  await client.delete<void>(endpoint)
}

/**
 * Test a provider connection without saving credentials
 *
 * Use this to validate API keys before storing them.
 *
 * @param projectId - Project ULID
 * @param data - Connection details to test
 * @returns Success status and optional error message
 *
 * @example
 * ```ts
 * const result = await testProviderConnection('proj_123', {
 *   adapter: 'openai',
 *   api_key: 'sk-...',
 * })
 * if (result.success) {
 *   // Proceed to save
 * } else {
 *   console.error(result.error)
 * }
 * ```
 */
export async function testProviderConnection(
  projectId: string,
  data: TestConnectionRequest
): Promise<TestConnectionResponse> {
  const endpoint = `/v1/projects/${projectId}/credentials/ai/test`
  return client.post<TestConnectionResponse, TestConnectionRequest>(endpoint, data)
}

/**
 * Create a masked preview of an API key
 * Format: first 4 chars + "***" + last 4 chars
 *
 * @param apiKey - Full API key
 * @returns Masked preview string
 *
 * @example
 * ```ts
 * createKeyPreview('sk-abcdefghijklmnop') // "sk-a***mnop"
 * ```
 */
export function createKeyPreview(apiKey: string): string {
  if (apiKey.length <= 8) {
    return apiKey.substring(0, 2) + '***'
  }
  return apiKey.substring(0, 4) + '***' + apiKey.substring(apiKey.length - 4)
}

/**
 * Get available models for a project based on configured providers
 *
 * Returns models from:
 * - Standard providers (openai, anthropic, etc.): default models + custom_models
 * - Custom provider: only custom_models
 *
 * @param projectId - Project ULID
 * @returns List of available models
 *
 * @example
 * ```ts
 * const models = await getAvailableModels('proj_123')
 * // [
 * //   { id: 'gpt-4o', name: 'GPT-4o', provider: 'openai' },
 * //   { id: 'claude-3-5-sonnet', name: 'Claude 3.5 Sonnet', provider: 'anthropic' },
 * // ]
 * ```
 */
export async function getAvailableModels(
  projectId: string
): Promise<AvailableModel[]> {
  const endpoint = `/v1/projects/${projectId}/credentials/ai/models`
  return client.get<AvailableModel[]>(endpoint)
}
