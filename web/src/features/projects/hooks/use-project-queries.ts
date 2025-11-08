'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createProject } from '@/features/organizations/api/organizations-api'
import { toast } from 'sonner'
import type { Project } from '@/features/organizations/types'

// Query keys for projects
export const projectQueryKeys = {
  all: ['projects'] as const,
  lists: () => [...projectQueryKeys.all, 'list'] as const,
  detail: (id: string) => [...projectQueryKeys.all, 'detail', id] as const,
}

// Create project mutation
export function useCreateProjectMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: { organizationId: string; name: string }) => {
      return createProject(data.organizationId, { name: data.name })
    },
    onSuccess: async (newProject: Project, variables) => {
      // Invalidate organization-specific projects (matches useOrganizationProjects key)
      queryClient.invalidateQueries({
        queryKey: ['organizations', variables.organizationId, 'projects'],
      })

      // Invalidate general projects list
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.lists() })

      // Invalidate workspace context to include new project
      queryClient.invalidateQueries({ queryKey: ['workspace'] })

      // Show success toast
      toast.success('Project Created!', {
        description: `${newProject.name} is ready to use.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Project', {
        description:
          apiError?.message || 'Could not create project. Please try again.',
      })
    },
  })
}
