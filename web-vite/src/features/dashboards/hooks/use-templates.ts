import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  getTemplates,
  getTemplateById,
  createFromTemplate,
} from '../api/templates-api'
import { dashboardQueryKeys } from './use-dashboards-queries'
import type { CreateFromTemplateRequest } from '../types'

export const templateQueryKeys = {
  all: ['dashboard-templates'] as const,
  lists: () => [...templateQueryKeys.all, 'list'] as const,
  list: () => [...templateQueryKeys.lists()] as const,
  details: () => [...templateQueryKeys.all, 'detail'] as const,
  detail: (templateId: string) =>
    [...templateQueryKeys.details(), templateId] as const,
}

export function useTemplatesQuery(options: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: templateQueryKeys.list(),
    queryFn: getTemplates,
    enabled: options.enabled ?? true,
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
  })
}

export function useTemplateQuery(
  templateId: string | undefined,
  options: { enabled?: boolean } = {},
) {
  return useQuery({
    queryKey: templateQueryKeys.detail(templateId ?? ''),
    queryFn: () => {
      if (!templateId) throw new Error('Template ID is required')
      return getTemplateById(templateId)
    },
    enabled: !!templateId && (options.enabled ?? true),
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
  })
}

export function useCreateFromTemplateMutation(projectId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateFromTemplateRequest) =>
      createFromTemplate(projectId, data),
    onSuccess: (newDashboard) => {
      queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.lists() })
      toast.success('Dashboard Created', {
        description: `"${newDashboard.name}" has been created from template.`,
      })
    },
    onError: (error: unknown) => {
      const message =
        error && typeof error === 'object' && 'message' in error
          ? String((error as { message?: unknown }).message ?? '')
          : ''
      toast.error('Failed to Create Dashboard', {
        description:
          message ||
          'Could not create dashboard from template. Please try again.',
      })
    },
  })
}
