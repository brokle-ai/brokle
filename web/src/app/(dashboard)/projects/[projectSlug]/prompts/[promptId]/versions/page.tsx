'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { PageHeader } from '@/components/layout/page-header'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Play, Settings } from 'lucide-react'
import { useProjectOnly } from '@/features/projects'
import {
  usePromptQuery,
  useVersionsQuery,
  useVersionDiffQuery,
  useProtectedLabelsQuery,
  useSetLabelsMutation,
  VersionHistory,
  VersionDiff,
  LabelSelector,
} from '@/features/prompts'
import type { PromptVersion } from '@/features/prompts'

export default function PromptVersionsPage() {
  const params = useParams<{ projectSlug: string; promptId: string }>()
  const router = useRouter()
  const { currentProject } = useProjectOnly()

  const [selectedVersion, setSelectedVersion] = useState<PromptVersion | null>(null)
  const [compareVersions, setCompareVersions] = useState<{
    from: number
    to: number
  } | null>(null)
  const [editingLabels, setEditingLabels] = useState<PromptVersion | null>(null)

  const { data: prompt, isLoading: promptLoading } = usePromptQuery(
    currentProject?.id,
    params.promptId
  )
  const { data: versions, isLoading: versionsLoading } = useVersionsQuery(
    currentProject?.id,
    params.promptId
  )
  const { data: protectedLabels } = useProtectedLabelsQuery(currentProject?.id)
  const { data: diffData, isLoading: diffLoading } = useVersionDiffQuery(
    currentProject?.id,
    params.promptId,
    compareVersions?.from,
    compareVersions?.to,
    { enabled: !!compareVersions }
  )

  const setLabelsMutation = useSetLabelsMutation(currentProject?.id || '', params.promptId)

  const handleVersionSelect = (version: PromptVersion) => {
    setSelectedVersion(version)
  }

  const handleCompare = (from: number, to: number) => {
    setCompareVersions({ from, to })
  }

  const handleSaveLabels = async (labels: string[]) => {
    if (!editingLabels) return
    await setLabelsMutation.mutateAsync({
      versionId: editingLabels.id,
      labels,
    })
    setEditingLabels(null)
  }

  if (promptLoading || versionsLoading) {
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
      <Main>
        <PageHeader
          title={`Version History: ${prompt.name}`}
          backHref={`/projects/${params.projectSlug}/prompts/${params.promptId}`}
          badges={
            <Badge variant={prompt.type === 'chat' ? 'default' : 'secondary'}>
              {prompt.type}
            </Badge>
          }
        >
          <Button
            variant="outline"
            onClick={() =>
              router.push(
                `/projects/${params.projectSlug}/prompts/${params.promptId}/playground`
              )
            }
          >
            <Play className="mr-2 h-4 w-4" />
            Playground
          </Button>
          <Button
            onClick={() =>
              router.push(
                `/projects/${params.projectSlug}/prompts/${params.promptId}`
              )
            }
          >
            <Settings className="mr-2 h-4 w-4" />
            Edit Prompt
          </Button>
        </PageHeader>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div>
            <h3 className="text-lg font-semibold mb-4">All Versions</h3>
            <VersionHistory
              versions={versions || []}
              protectedLabels={protectedLabels || []}
              selectedVersionId={selectedVersion?.id}
              onVersionSelect={handleVersionSelect}
              onCompare={handleCompare}
            />
          </div>

          <div className="space-y-6">
            {selectedVersion && (
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-semibold">
                    Version {selectedVersion.version} Details
                  </h3>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setEditingLabels(selectedVersion)}
                  >
                    Edit Labels
                  </Button>
                </div>
                <div className="rounded-md border p-4 space-y-4">
                  <div>
                    <h4 className="text-sm font-medium text-muted-foreground mb-2">
                      Template
                    </h4>
                    <pre className="whitespace-pre-wrap rounded-md bg-muted p-4 font-mono text-sm overflow-auto max-h-[400px]">
                      {JSON.stringify(selectedVersion.template, null, 2)}
                    </pre>
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-muted-foreground mb-2">
                      Variables
                    </h4>
                    <div className="flex flex-wrap gap-1">
                      {selectedVersion.variables.length > 0 ? (
                        selectedVersion.variables.map((v) => (
                          <Badge key={v} variant="outline" className="font-mono">
                            {`{{${v}}}`}
                          </Badge>
                        ))
                      ) : (
                        <span className="text-sm text-muted-foreground italic">
                          No variables
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            )}

            {compareVersions && (
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-semibold">Comparing Versions</h3>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setCompareVersions(null)}
                  >
                    Clear
                  </Button>
                </div>
                {diffLoading ? (
                  <Skeleton className="h-[300px]" />
                ) : diffData ? (
                  <VersionDiff diff={diffData} />
                ) : (
                  <p className="text-muted-foreground">Failed to load diff</p>
                )}
              </div>
            )}
          </div>
        </div>

        <Dialog open={!!editingLabels} onOpenChange={() => setEditingLabels(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                Edit Labels for Version {editingLabels?.version}
              </DialogTitle>
            </DialogHeader>
            {editingLabels && (
              <LabelSelector
                labels={editingLabels.labels}
                protectedLabels={protectedLabels || []}
                availableLabels={['production', 'staging', 'development']}
                onChange={handleSaveLabels}
                isLoading={setLabelsMutation.isPending}
              />
            )}
          </DialogContent>
        </Dialog>
      </Main>
    </>
  )
}
