'use client'

import { useCallback, useMemo } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { PageHeader } from '@/components/layout/page-header'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { useProjectOnly } from '@/features/projects'
import {
  usePromptQuery,
  useVersionsQuery,
  useVersionQuery,
  useCreateVersionMutation,
} from '@/features/prompts'
import { PromptEditLayout } from '@/features/prompts/components/prompt-edit'
import { usePromptEditState } from '@/features/prompts/hooks/use-prompt-edit-state'
import type { CreateVersionRequest } from '@/features/prompts'

export default function PromptEditPage() {
  const params = useParams<{ projectSlug: string; promptId: string }>()
  const router = useRouter()
  const { currentProject } = useProjectOnly()

  // URL state for edit page
  const { sourceVersionId, isRestoreFlow, setSourceVersionId } = usePromptEditState()

  // Data fetching
  const { data: prompt, isLoading: promptLoading } = usePromptQuery(
    currentProject?.id,
    params.promptId
  )
  const { data: versions = [], isLoading: versionsLoading } = useVersionsQuery(
    currentProject?.id,
    params.promptId
  )

  // Fetch source version if specified and different from current
  const shouldFetchVersion = sourceVersionId && prompt && sourceVersionId !== prompt.version_id
  const { data: fetchedVersion } = useVersionQuery(
    currentProject?.id,
    params.promptId,
    sourceVersionId || '',
    { enabled: !!shouldFetchVersion }
  )

  // Find source version from versions list or use fetched version
  const sourceVersion = useMemo(() => {
    if (!sourceVersionId) {
      // Default to latest version from list
      return versions.length > 0 ? versions[0] : null
    }

    // First try to find in the versions list
    const fromList = versions.find((v) => v.id === sourceVersionId)
    if (fromList) return fromList

    // Fall back to fetched version
    if (fetchedVersion) return fetchedVersion

    return null
  }, [sourceVersionId, versions, fetchedVersion])

  // Mutations
  const createVersionMutation = useCreateVersionMutation(
    currentProject?.id || '',
    params.promptId
  )

  // Handle save - create new version and redirect back to details
  const handleSave = useCallback(
    async (data: CreateVersionRequest) => {
      await createVersionMutation.mutateAsync(data)
      // Navigate back to details page
      router.push(`/projects/${params.projectSlug}/prompts/${params.promptId}`)
    },
    [createVersionMutation, router, params.projectSlug, params.promptId]
  )

  // Handle cancel - navigate back to details
  const handleCancel = useCallback(() => {
    router.push(`/projects/${params.projectSlug}/prompts/${params.promptId}`)
  }, [router, params.projectSlug, params.promptId])

  // Handle version selection from sidebar - update source version
  const handleVersionSelect = useCallback(
    (version: { id: string }) => {
      setSourceVersionId(version.id)
    },
    [setSourceVersionId]
  )

  // Loading state
  if (promptLoading) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="space-y-6">
            <Skeleton className="h-10 w-48" />
            <Skeleton className="h-[600px]" />
          </div>
        </Main>
      </>
    )
  }

  // Not found state
  if (!prompt) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="flex flex-col items-center justify-center py-12">
            <p className="text-muted-foreground">Prompt not found</p>
            <Button variant="link" onClick={() => router.back()}>
              Go back
            </Button>
          </div>
        </Main>
      </>
    )
  }

  return (
    <>
      <DashboardHeader />
      <Main fixed>
        <div className="flex h-full flex-col">
          <PageHeader
            title={`Edit: ${prompt.name}`}
            backHref={`/projects/${params.projectSlug}/prompts/${params.promptId}`}
            badges={
              <Badge variant={prompt.type === 'chat' ? 'default' : 'secondary'}>
                {prompt.type}
              </Badge>
            }
          />

          <div className="flex-1 overflow-hidden">
            <PromptEditLayout
              prompt={prompt}
              versions={versions}
              versionsLoading={versionsLoading}
              sourceVersion={sourceVersion}
              isRestoreFlow={isRestoreFlow}
              onSave={handleSave}
              onCancel={handleCancel}
              onVersionSelect={handleVersionSelect}
              isSaving={createVersionMutation.isPending}
            />
          </div>
        </div>
      </Main>
    </>
  )
}
