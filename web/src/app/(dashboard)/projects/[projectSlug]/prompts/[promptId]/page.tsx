'use client'

import { useCallback, useMemo } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { PageHeader } from '@/components/layout/page-header'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { FlaskConical } from 'lucide-react'
import { useProjectOnly } from '@/features/projects'
import {
  usePromptQuery,
  useVersionsQuery,
  useVersionQuery,
  useProtectedLabelsQuery,
  useSetLabelsMutation,
  usePromptDetailState,
  PromptDetailLayout,
} from '@/features/prompts'
import type { TextTemplate, ChatTemplate, PromptVersion } from '@/features/prompts'
import { usePlaygroundStore, createMessage } from '@/features/playground'

export default function PromptDetailPage() {
  const params = useParams<{ projectSlug: string; promptId: string }>()
  const router = useRouter()
  const { currentProject } = useProjectOnly()

  // Data fetching
  const { data: prompt, isLoading: promptLoading } = usePromptQuery(
    currentProject?.id,
    params.promptId
  )
  const { data: versions = [], isLoading: versionsLoading } = useVersionsQuery(
    currentProject?.id,
    params.promptId
  )
  const { data: protectedLabels = [] } = useProtectedLabelsQuery(
    currentProject?.id
  )

  // Mutations - only labels (editing is now on separate page)
  const setLabelsMutation = useSetLabelsMutation(
    currentProject?.id || '',
    params.promptId
  )

  // URL-based version selection state (lifted from layout for version-scoped actions)
  const { selectedVersionId, setSelectedVersionId } = usePromptDetailState()

  // Fetch selected version data when it differs from the latest version
  const shouldFetchVersion = selectedVersionId && prompt && selectedVersionId !== prompt.version_id
  const { data: fetchedVersion } = useVersionQuery(
    currentProject?.id,
    params.promptId,
    selectedVersionId || '',
    { enabled: !!shouldFetchVersion }
  )

  // Compute the effective version (selected or latest)
  const effectiveVersion = useMemo(() => {
    if (!selectedVersionId || !prompt) return null

    // First try to find in the versions list
    const fromList = versions.find((v) => v.id === selectedVersionId)
    if (fromList) return fromList

    // Fall back to fetched version
    if (fetchedVersion) return fetchedVersion

    return null
  }, [selectedVersionId, prompt, versions, fetchedVersion])

  // Check if version data is still loading (selected but not yet resolved)
  const isVersionDataPending = selectedVersionId && !effectiveVersion

  // Get all available labels from versions
  const availableLabels = Array.from(
    new Set(versions.flatMap((v) => v.labels))
  )

  // Handle label changes for the selected version
  const handleLabelsChange = useCallback(
    async (labels: string[]) => {
      if (!prompt) return
      // Use the selected version ID from URL state (or fall back to latest)
      const versionId = selectedVersionId || prompt.version_id
      await setLabelsMutation.mutateAsync({
        versionId,
        labels,
      })
    },
    [prompt, selectedVersionId, setLabelsMutation]
  )

  // Handle version comparison - now handled by dialog in VersionSidebar
  const handleCompare = useCallback(
    (_fromVersion: number, _toVersion: number) => {
      // Comparison is handled by the VersionDiffDialog within VersionSidebar
    },
    []
  )

  // Handle version restore - navigate to edit page with that version
  const handleRestore = useCallback(
    (version: PromptVersion) => {
      router.push(
        `/projects/${params.projectSlug}/prompts/${params.promptId}/edit?version=${version.id}&restore=true`
      )
    },
    [router, params.projectSlug, params.promptId]
  )

  // Handle Try in Playground - uses selected version data, not always latest
  const handleTryInPlayground = useCallback(() => {
    if (!prompt) return

    // Guard: Don't proceed if version is selected but data hasn't loaded yet
    // This prevents loading the wrong version on deep links with ?version=...
    if (selectedVersionId && !effectiveVersion) return

    // Use effective version or fall back to prompt (latest)
    const versionData = effectiveVersion || {
      id: prompt.version_id,
      version: prompt.version,
      template: prompt.template,
    }

    let messages: ReturnType<typeof createMessage>[] = []

    if (prompt.type === 'chat') {
      const template = versionData.template as ChatTemplate
      messages =
        template.messages?.map((m) =>
          createMessage(
            (m.role || 'user') as 'system' | 'user' | 'assistant',
            m.content || ''
          )
        ) || []
    } else {
      const template = versionData.template as { content?: string }
      messages = [createMessage('user', template.content || '')]
    }

    const loadedTemplate = JSON.stringify(
      messages.map(({ role, content }) => ({ role, content }))
    )

    usePlaygroundStore.getState().loadFromPrompt({
      messages,
      loadedFromPromptId: prompt.id,
      loadedFromPromptName: prompt.name,
      loadedFromPromptVersionId: versionData.id,
      loadedFromPromptVersionNumber: versionData.version,
      loadedTemplate,
    })

    router.push(`/projects/${params.projectSlug}/playground`)
  }, [prompt, selectedVersionId, effectiveVersion, router, params.projectSlug])

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
            title={prompt.name}
            backHref={`/projects/${params.projectSlug}/prompts`}
            description={prompt.description}
            badges={
              <Badge variant={prompt.type === 'chat' ? 'default' : 'secondary'}>
                {prompt.type}
              </Badge>
            }
          >
            <Button onClick={handleTryInPlayground} disabled={!!isVersionDataPending}>
              <FlaskConical className="mr-2 h-4 w-4" />
              Try in Playground
            </Button>
          </PageHeader>

          <div className="flex-1 overflow-hidden">
            <PromptDetailLayout
              prompt={prompt}
              versions={versions}
              versionsLoading={versionsLoading}
              protectedLabels={protectedLabels}
              availableLabels={availableLabels}
              projectId={currentProject?.id || ''}
              projectSlug={params.projectSlug}
              selectedVersionId={selectedVersionId}
              onVersionChange={setSelectedVersionId}
              onLabelsChange={handleLabelsChange}
              onCompare={handleCompare}
              onRestore={handleRestore}
              isLabelsLoading={setLabelsMutation.isPending}
            />
          </div>
        </div>
      </Main>
    </>
  )
}
