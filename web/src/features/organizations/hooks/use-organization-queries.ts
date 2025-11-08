'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createOrganization } from '../api/organizations-api'
import { getCurrentUser } from '@/lib/api'
import { authQueryKeys } from '@/features/authentication'
import { toast } from 'sonner'
import type { Organization } from '../types'

// Query keys for organizations
export const organizationQueryKeys = {
  all: ['organizations'] as const,
  lists: () => [...organizationQueryKeys.all, 'list'] as const,
  detail: (id: string) => [...organizationQueryKeys.all, 'detail', id] as const,
}

// Create organization mutation
export function useCreateOrganizationMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: { name: string; description?: string }) => {
      return createOrganization(data)
    },
    onSuccess: async (newOrganization: Organization) => {
      // Invalidate organizations list
      queryClient.invalidateQueries({ queryKey: organizationQueryKeys.lists() })

      // Invalidate workspace context to include new organization
      queryClient.invalidateQueries({ queryKey: ['workspace'] })

      // Refresh user data (backend sets defaultOrganizationId automatically)
      try {
        const updatedUser = await getCurrentUser()
        queryClient.setQueryData(authQueryKeys.user(), updatedUser)
      } catch (error) {
        if (process.env.NODE_ENV === 'development') {
          console.error('Failed to refresh user after org creation:', error)
        }
      }

      // Show success toast
      toast.success('Organization Created!', {
        description: `${newOrganization.name} is ready to use.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Organization', {
        description: apiError?.message || 'Could not create organization. Please try again.',
      })
    },
  })
}
