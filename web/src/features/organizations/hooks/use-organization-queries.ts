'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createOrganization, updateOrganization } from '../api/organizations-api'
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
      // CRITICAL: Await workspace refetch to ensure data is fresh before navigation
      // Using refetchQueries instead of invalidateQueries ensures the data is actually loaded
      await queryClient.refetchQueries({ queryKey: ['workspace'] })

      // Invalidate organizations list (can be background, less critical)
      queryClient.invalidateQueries({ queryKey: organizationQueryKeys.lists() })

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

// Update organization mutation
export function useUpdateOrganizationMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      orgId,
      data
    }: {
      orgId: string
      data: { name?: string; billing_email?: string }
    }) => {
      return updateOrganization(orgId, data)
    },
    onSuccess: async (updatedOrg: Organization) => {
      // Await workspace refetch to ensure data is fresh
      await queryClient.refetchQueries({ queryKey: ['workspace'] })

      // Invalidate organizations list (can be background)
      queryClient.invalidateQueries({ queryKey: organizationQueryKeys.lists() })

      // Invalidate specific org detail
      queryClient.invalidateQueries({ queryKey: organizationQueryKeys.detail(updatedOrg.id) })

      // Show success toast
      toast.success('Organization Updated!', {
        description: `${updatedOrg.name} has been updated successfully.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Update Organization', {
        description: apiError?.message || 'Could not update organization. Please try again.',
      })
    },
  })
}
