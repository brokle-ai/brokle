'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createProject } from '@/features/organizations/api/organizations-api'
import { updateProject } from '../api/projects-api'
import { toast } from 'sonner'
import type { Project } from '@/features/organizations/types'
import type { UpdateProjectRequest, Project as APIProject } from '../api/projects-api'

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

// Update project mutation
export function useUpdateProjectMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      projectId,
      data
    }: {
      projectId: string
      data: UpdateProjectRequest
    }) => {
      return updateProject(projectId, data)
    },
    onSuccess: (updatedProject: APIProject) => {
      // Invalidate workspace context (includes current project)
      queryClient.invalidateQueries({ queryKey: ['workspace'] })

      // Invalidate projects list
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.lists() })

      // Invalidate specific project detail
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.detail(updatedProject.id) })

      // Show success toast
      toast.success('Project Updated!', {
        description: `${updatedProject.name} has been updated successfully.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Update Project', {
        description: apiError?.message || 'Could not update project settings. Please try again.',
      })
    },
  })
}
