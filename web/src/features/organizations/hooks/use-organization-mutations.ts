'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createOrganization } from '../api/organizations-api'
import { toast } from 'sonner'
import type { CreateOrganizationData } from '../types'

export function useCreateOrganizationMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateOrganizationData) => {
      return createOrganization(data)
    },
    onSuccess: (organization) => {
      queryClient.invalidateQueries({ queryKey: ['organizations'] })

      toast.success('Organization Created!', {
        description: `${organization.name} has been created successfully.`,
      })
    },
    onError: (error: any) => {
      toast.error('Failed to Create Organization', {
        description: error?.message || 'Failed to create organization',
      })
    },
  })
}
