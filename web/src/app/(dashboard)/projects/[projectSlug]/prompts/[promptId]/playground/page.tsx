'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { ArrowLeft, Settings } from 'lucide-react'
import { useProjectOnly } from '@/features/projects'
import {
  usePromptQuery,
  useVersionsQuery,
  useVersionQuery,
  useExecutePromptMutation,
  PromptPlayground,
  LabelBadge,
} from '@/features/prompts'
import type { ExecutePromptRequest, ExecutePromptResponse } from '@/features/prompts'

export default function PromptPlaygroundPage() {
  const params = useParams<{ projectSlug: string; promptId: string }>()
  const router = useRouter()
  const { currentProject } = useProjectOnly()

  const [selectedVersionId, setSelectedVersionId] = useState<string>('')

  // Queries
  const { data: prompt, isLoading: promptLoading } = usePromptQuery(
    currentProject?.id,
    params.promptId
  )
  const { data: versions, isLoading: versionsLoading } = useVersionsQuery(
    currentProject?.id,
    params.promptId
  )

  // Get selected version (default to latest)
  const effectiveVersionId = selectedVersionId || versions?.[0]?.id || ''
  const { data: selectedVersion, isLoading: versionLoading } = useVersionQuery(
    currentProject?.id,
    params.promptId,
    effectiveVersionId,
    { enabled: !!effectiveVersionId }
  )

  // Mutation
  const executeMutation = useExecutePromptMutation(currentProject?.id || '', params.promptId)

  const handleExecute = async (request: ExecutePromptRequest): Promise<ExecutePromptResponse> => {
    return executeMutation.mutateAsync({
      versionId: effectiveVersionId,
      request,
    })
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
        {/* Header */}
        <div className="mb-6 flex items-start justify-between">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="icon" onClick={() => router.back()}>
              <ArrowLeft className="h-4 w-4" />
            </Button>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-2xl font-bold tracking-tight">
                  Playground: {prompt.name}
                </h2>
                <Badge variant={prompt.type === 'chat' ? 'default' : 'secondary'}>
                  {prompt.type}
                </Badge>
              </div>
              <p className="text-muted-foreground">
                Test your prompt with different variables and model settings
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">Version:</span>
              <Select value={effectiveVersionId} onValueChange={setSelectedVersionId}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="Select version" />
                </SelectTrigger>
                <SelectContent>
                  {versions?.map((v) => (
                    <SelectItem key={v.id} value={v.id}>
                      <div className="flex items-center gap-2">
                        <span className="font-mono">v{v.version}</span>
                        {v.labels.map((label) => (
                          <Badge
                            key={label}
                            variant="outline"
                            className="text-xs"
                          >
                            {label}
                          </Badge>
                        ))}
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Button
              variant="outline"
              onClick={() =>
                router.push(
                  `/projects/${params.projectSlug}/prompts/${params.promptId}`
                )
              }
            >
              <Settings className="mr-2 h-4 w-4" />
              Edit Prompt
            </Button>
          </div>
        </div>

        {/* Labels */}
        {selectedVersion && selectedVersion.labels.length > 0 && (
          <div className="mb-6 flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Labels:</span>
            {selectedVersion.labels.map((label) => (
              <LabelBadge key={label} label={label} />
            ))}
          </div>
        )}

        {/* Playground */}
        {versionLoading ? (
          <Skeleton className="h-[500px]" />
        ) : selectedVersion ? (
          <PromptPlayground
            version={selectedVersion}
            promptType={prompt.type}
            onExecute={handleExecute}
            isExecuting={executeMutation.isPending}
          />
        ) : (
          <div className="flex flex-col items-center justify-center py-12">
            <p className="text-muted-foreground">No version selected</p>
          </div>
        )}
      </Main>
    </>
  )
}
