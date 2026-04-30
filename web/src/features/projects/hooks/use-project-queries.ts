'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createProject } from '@/features/organizations/api/organizations-api'
import { updateProject } from '../api/projects-api'
import { toast } from 'sonner'
import type { Project } from '@/features/organizations/types'
import type { UpdateProjectRequest, Project as APIProject } from '../api/projects-api'

// Workspace cache (`['workspace']`) is the single source of truth for
// the user's project tree (Langfuse session-bootstrap pattern). Every
// project-affecting mutation invalidates it; consumers re-render off
// the refreshed bootstrap. No per-list / per-detail query keys exist
// because no `useQuery` reads them.

// Create project mutation
export function useCreateProjectMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: { organizationId: string; name: string }) => {
      return createProject(data.organizationId, { name: data.name })
    },
    onSuccess: async (newProject: Project) => {
      queryClient.invalidateQueries({ queryKey: ['workspace'] })

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
      queryClient.invalidateQueries({ queryKey: ['workspace'] })

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
