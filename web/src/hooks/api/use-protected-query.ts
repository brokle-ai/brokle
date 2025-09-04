'use client'

import { 
  useQuery, 
  useMutation,
  UseQueryOptions, 
  UseMutationOptions,
  QueryKey,
  QueryFunction,
  MutationFunction
} from '@tanstack/react-query'
import { useAuth } from '@/hooks/auth/use-auth'
import type { APIError } from '@/lib/api/core/types'

// Protected query hook that automatically handles auth state
export function useProtectedQuery<
  TQueryFnData = unknown,
  TError = APIError,
  TData = TQueryFnData,
  TQueryKey extends QueryKey = QueryKey
>(
  queryKey: TQueryKey,
  queryFn: QueryFunction<TQueryFnData, TQueryKey>,
  options?: Omit<UseQueryOptions<TQueryFnData, TError, TData, TQueryKey>, 'queryKey' | 'queryFn'>
) {
  const { isAuthenticated, isLoading: authLoading } = useAuth()

  return useQuery({
    queryKey,
    queryFn,
    enabled: isAuthenticated && !authLoading && (options?.enabled !== false),
    retry: (failureCount, error) => {
      // Don't retry on auth errors
      if ((error as APIError)?.statusCode === 401) {
        return false
      }
      return failureCount < 3
    },
    staleTime: 5 * 60 * 1000, // 5 minutes default
    ...options,
  })
}

// Protected mutation hook with automatic error handling
export function useProtectedMutation<
  TData = unknown,
  TError = APIError,
  TVariables = void,
  TContext = unknown
>(
  mutationFn: MutationFunction<TData, TVariables>,
  options?: UseMutationOptions<TData, TError, TVariables, TContext>
) {
  const { isAuthenticated } = useAuth()

  return useMutation({
    mutationFn: async (variables: TVariables) => {
      if (!isAuthenticated) {
        throw new Error('User must be authenticated to perform this action')
      }
      return mutationFn(variables)
    },
    retry: (failureCount, error) => {
      // Don't retry on auth errors or client errors
      const apiError = error as APIError
      if (apiError?.statusCode === 401 || (apiError?.statusCode >= 400 && apiError?.statusCode < 500)) {
        return false
      }
      return failureCount < 2
    },
    ...options,
  })
}

// Optimistic update helper
export function useOptimisticMutation<
  TData = unknown,
  TError = APIError,
  TVariables = void,
  TContext = unknown
>(
  mutationFn: MutationFunction<TData, TVariables>,
  options: UseMutationOptions<TData, TError, TVariables, TContext> & {
    optimisticUpdate?: (variables: TVariables) => void
    onRollback?: (context: TContext) => void
  }
) {
  const { optimisticUpdate, onRollback, ...mutationOptions } = options

  return useProtectedMutation(mutationFn, {
    ...mutationOptions,
    onMutate: async (variables: TVariables) => {
      // Apply optimistic update
      optimisticUpdate?.(variables)
      
      // Call original onMutate if provided
      const context = await mutationOptions.onMutate?.(variables)
      return context
    },
    onError: (error, variables, context) => {
      // Rollback optimistic update
      if (context) {
        onRollback?.(context)
      }
      
      // Call original onError if provided
      mutationOptions.onError?.(error, variables, context)
    },
  })
}

// Query with automatic invalidation on auth changes
export function useAutoRefreshQuery<
  TQueryFnData = unknown,
  TError = APIError,
  TData = TQueryFnData,
  TQueryKey extends QueryKey = QueryKey
>(
  queryKey: TQueryKey,
  queryFn: QueryFunction<TQueryFnData, TQueryKey>,
  options?: Omit<UseQueryOptions<TQueryFnData, TError, TData, TQueryKey>, 'queryKey' | 'queryFn'> & {
    refreshOnAuth?: boolean
    refreshInterval?: number
  }
) {
  const { isAuthenticated, user } = useAuth()
  const { refreshOnAuth = true, refreshInterval, ...queryOptions } = options || {}

  // Include user ID in query key to automatically refresh when user changes
  const enhancedQueryKey = refreshOnAuth && user 
    ? [...queryKey, { userId: user.id }] as TQueryKey
    : queryKey

  return useProtectedQuery(
    enhancedQueryKey,
    queryFn,
    {
      ...queryOptions,
      refetchInterval: refreshInterval,
      refetchIntervalInBackground: false,
    }
  )
}

// Paginated query helper
export function usePaginatedQuery<TData = unknown, TError = APIError>(
  queryKey: QueryKey,
  queryFn: (page: number, limit: number) => Promise<{
    data: TData[]
    pagination: {
      page: number
      limit: number
      total: number
      pages: number
      hasNext: boolean
      hasPrev: boolean
    }
  }>,
  options?: {
    initialPage?: number
    pageSize?: number
    enabled?: boolean
  }
) {
  const { initialPage = 1, pageSize = 20, enabled = true } = options || {}

  return useProtectedQuery(
    [...queryKey, { page: initialPage, limit: pageSize }],
    () => queryFn(initialPage, pageSize),
    {
      enabled,
      keepPreviousData: true, // Keep previous data while loading new page
    }
  )
}