'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  getTemplates,
  getTemplateById,
  createFromTemplate,
} from '../api/templates-api'
import { dashboardQueryKeys } from './use-dashboards-queries'
import type { CreateFromTemplateRequest } from '../types'

/**
 * Query keys for template queries
 */
export const templateQueryKeys = {
  all: ['dashboard-templates'] as const,

  // Lists
  lists: () => [...templateQueryKeys.all, 'list'] as const,
  list: () => [...templateQueryKeys.lists()] as const,

  // Details
  details: () => [...templateQueryKeys.all, 'detail'] as const,
  detail: (templateId: string) =>
    [...templateQueryKeys.details(), templateId] as const,
}

/**
 * Query hook to list all available dashboard templates
 */
export function useTemplatesQuery(options: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: templateQueryKeys.list(),
    queryFn: getTemplates,
    enabled: options.enabled ?? true,
    staleTime: 5 * 60 * 1000, // 5 minutes - templates rarely change
    gcTime: 30 * 60 * 1000, // 30 minutes
  })
}

/**
 * Query hook to get a single template by ID
 */
export function useTemplateQuery(
  templateId: string | undefined,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: templateQueryKeys.detail(templateId || ''),
    queryFn: async () => {
      if (!templateId) {
        throw new Error('Template ID is required')
      }
      return getTemplateById(templateId)
    },
    enabled: !!templateId && (options.enabled ?? true),
    staleTime: 5 * 60 * 1000,
    gcTime: 30 * 60 * 1000,
  })
}

/**
 * Mutation hook to create a dashboard from a template
 */
export function useCreateFromTemplateMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateFromTemplateRequest) => {
      return createFromTemplate(projectId, data)
    },
    onSuccess: (newDashboard) => {
      queryClient.invalidateQueries({
        queryKey: dashboardQueryKeys.lists(),
      })
      toast.success('Dashboard Created', {
        description: `"${newDashboard.name}" has been created from template.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Dashboard', {
        description:
          apiError?.message ||
          'Could not create dashboard from template. Please try again.',
      })
    },
  })
}
