import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { useMemo } from 'react'
import { toast } from 'sonner'
import { rawFetch } from '@/lib/api/client'
import { BrokleError } from '@/lib/api/errors'
import type {
  AIProvider,
  AIProviderCredential,
  AvailableModel,
  CreateProviderRequest,
  ModelsByProvider,
  TestConnectionRequest,
  TestConnectionResponse,
  UpdateProviderRequest,
} from './types'

// TkDodo-style hierarchical query keys. Org-scoped because AI
// credentials are organization-level, not project-level.
export const aiProvidersKeys = {
  all: ['ai-providers'] as const,
  lists: () => [...aiProvidersKeys.all, 'list'] as const,
  list: (orgId: string) => [...aiProvidersKeys.lists(), orgId] as const,
  models: () => [...aiProvidersKeys.all, 'models'] as const,
  modelsFor: (orgId: string) => [...aiProvidersKeys.models(), orgId] as const,
} as const

export const aiProvidersListQueryOptions = (orgId: string) =>
  queryOptions({
    queryKey: aiProvidersKeys.list(orgId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/organizations/${orgId}/credentials/ai`,
        { method: 'GET' },
      )
      return (await resp.json()) as AIProviderCredential[]
    },
    staleTime: 30 * 1000,
  })

export const availableModelsQueryOptions = (orgId: string) =>
  queryOptions({
    queryKey: aiProvidersKeys.modelsFor(orgId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/organizations/${orgId}/credentials/ai/models`,
        { method: 'GET' },
      )
      return (await resp.json()) as AvailableModel[]
    },
    // Models rarely change in a session — 5 min keeps the playground /
    // evaluator model selectors snappy without hammering the backend.
    staleTime: 5 * 60 * 1000,
  })

async function createProvider(
  orgId: string,
  data: CreateProviderRequest,
): Promise<AIProviderCredential> {
  const resp = await rawFetch(
    `/api/v1/organizations/${orgId}/credentials/ai`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as AIProviderCredential
}

async function updateProvider(
  orgId: string,
  credentialId: string,
  data: UpdateProviderRequest,
): Promise<AIProviderCredential> {
  const resp = await rawFetch(
    `/api/v1/organizations/${orgId}/credentials/ai/${credentialId}`,
    {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as AIProviderCredential
}

async function deleteProvider(
  orgId: string,
  credentialId: string,
): Promise<void> {
  await rawFetch(
    `/api/v1/organizations/${orgId}/credentials/ai/${credentialId}`,
    { method: 'DELETE' },
  )
}

async function testConnection(
  orgId: string,
  data: TestConnectionRequest,
): Promise<TestConnectionResponse> {
  const resp = await rawFetch(
    `/api/v1/organizations/${orgId}/credentials/ai/test`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as TestConnectionResponse
}

function errorMessage(err: unknown, fallback: string): string {
  if (err instanceof BrokleError) return err.message
  if (err instanceof Error) return err.message
  return fallback
}

export function useCreateProviderMutation(orgId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateProviderRequest) => createProvider(orgId, data),
    onSuccess: async (credential) => {
      await queryClient.invalidateQueries({
        queryKey: aiProvidersKeys.lists(),
      })
      await queryClient.invalidateQueries({
        queryKey: aiProvidersKeys.models(),
      })
      toast.success('Provider created', {
        description: `${credential.name} has been added.`,
      })
    },
    onError: (err) => {
      toast.error('Failed to create provider', {
        description: errorMessage(err, 'Could not create provider.'),
      })
    },
  })
}

export function useUpdateProviderMutation(orgId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      credentialId,
      data,
    }: {
      credentialId: string
      data: UpdateProviderRequest
    }) => updateProvider(orgId, credentialId, data),
    onSuccess: async (credential) => {
      await queryClient.invalidateQueries({
        queryKey: aiProvidersKeys.lists(),
      })
      await queryClient.invalidateQueries({
        queryKey: aiProvidersKeys.models(),
      })
      toast.success('Provider updated', {
        description: `${credential.name} has been updated.`,
      })
    },
    onError: (err) => {
      toast.error('Failed to update provider', {
        description: errorMessage(err, 'Could not update provider.'),
      })
    },
  })
}

// Delete with optimistic removal — on failure we roll back the cached
// list to its pre-mutation state.
export function useDeleteProviderMutation(orgId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      credentialId,
    }: {
      credentialId: string
      displayName: string
    }) => deleteProvider(orgId, credentialId),
    onMutate: async ({ credentialId }) => {
      await queryClient.cancelQueries({ queryKey: aiProvidersKeys.lists() })
      const previous = queryClient.getQueriesData<AIProviderCredential[]>({
        queryKey: aiProvidersKeys.lists(),
      })
      queryClient.setQueriesData<AIProviderCredential[]>(
        { queryKey: aiProvidersKeys.lists() },
        (old) => (old ? old.filter((c) => c.id !== credentialId) : old),
      )
      return { previous }
    },
    onSuccess: async (_void, vars) => {
      await queryClient.invalidateQueries({
        queryKey: aiProvidersKeys.lists(),
      })
      await queryClient.invalidateQueries({
        queryKey: aiProvidersKeys.models(),
      })
      toast.success('Provider deleted', {
        description: `${vars.displayName} has been removed.`,
      })
    },
    onError: (err, _vars, ctx) => {
      if (ctx?.previous) {
        ctx.previous.forEach(([queryKey, data]) => {
          queryClient.setQueryData(queryKey, data)
        })
      }
      toast.error('Failed to delete provider', {
        description: errorMessage(err, 'Could not delete provider.'),
      })
    },
  })
}

// Test-connection mutation. Unlike save mutations this does not touch
// the cache; it only surfaces success/failure to the caller, which
// renders a result row in the dialog footer.
export function useTestConnectionMutation(orgId: string) {
  return useMutation({
    mutationFn: (data: TestConnectionRequest) => testConnection(orgId, data),
    onSuccess: (result) => {
      if (result.success) {
        toast.success('Connection successful', {
          description: 'The API key is valid.',
        })
      } else {
        toast.error('Connection failed', {
          description: result.error ?? 'Could not connect to the provider.',
        })
      }
    },
    onError: (err) => {
      toast.error('Connection test failed', {
        description: errorMessage(err, 'Could not test connection.'),
      })
    },
  })
}

// Convenience hook for the playground / evaluator model selectors.
// Groups the flat model list by provider and returns the configured
// provider list as a stable array.
export function useModelsByProvider(orgId: string | undefined) {
  const query = useQuery({
    ...availableModelsQueryOptions(orgId ?? ''),
    enabled: !!orgId,
  })

  const modelsByProvider = useMemo((): ModelsByProvider => {
    if (!query.data) return {}
    return query.data.reduce<ModelsByProvider>((acc, model) => {
      const provider = model.provider
      const bucket = acc[provider] ?? []
      bucket.push(model)
      acc[provider] = bucket
      return acc
    }, {})
  }, [query.data])

  const configuredProviders = useMemo(
    () => Object.keys(modelsByProvider) as AIProvider[],
    [modelsByProvider],
  )

  return { ...query, modelsByProvider, configuredProviders }
}
