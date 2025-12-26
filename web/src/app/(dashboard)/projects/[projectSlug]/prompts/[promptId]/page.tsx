'use client'

import { useState, useMemo, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { PageHeader } from '@/components/layout/page-header'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Save,
  Loader2,
  History,
  Plus,
  Settings,
  FlaskConical,
} from 'lucide-react'
import { useProjectOnly } from '@/features/projects'
import {
  usePromptQuery,
  useVersionsQuery,
  useProtectedLabelsQuery,
  useUpdatePromptMutation,
  useCreateVersionMutation,
  useSetLabelsMutation,
  PromptEditor,
  LabelBadge,
  VariableList,
  extractVariables,
} from '@/features/prompts'
import { usePlaygroundStore, createMessage } from '@/features/playground'
import type {
  PromptType,
  TextTemplate,
  ChatTemplate,
  CreateVersionRequest,
} from '@/features/prompts'

export default function PromptDetailPage() {
  const params = useParams<{ projectSlug: string; promptId: string }>()
  const router = useRouter()
  const { currentProject } = useProjectOnly()

  const { data: prompt, isLoading: promptLoading } = usePromptQuery(
    currentProject?.id,
    params.promptId
  )
  const { data: versions, isLoading: versionsLoading } = useVersionsQuery(
    currentProject?.id,
    params.promptId
  )
  const { data: protectedLabels } = useProtectedLabelsQuery(currentProject?.id)

  const updateMutation = useUpdatePromptMutation(currentProject?.id || '', params.promptId)
  const createVersionMutation = useCreateVersionMutation(currentProject?.id || '', params.promptId)
  const setLabelsMutation = useSetLabelsMutation(currentProject?.id || '', params.promptId)

  const [isEditing, setIsEditing] = useState(false)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [tagsInput, setTagsInput] = useState('')

  const [editedTemplate, setEditedTemplate] = useState<TextTemplate | ChatTemplate | null>(null)
  const [newCommitMessage, setNewCommitMessage] = useState('')

  useEffect(() => {
    if (prompt && !isEditing) {
      setName(prompt.name)
      setDescription(prompt.description || '')
      setTagsInput(prompt.tags?.join(', ') || '')
    }
  }, [prompt, isEditing])

  const variables = useMemo(() => {
    if (!prompt) return []
    const tmpl = editedTemplate || prompt.template
    return extractVariables(tmpl, prompt.type)
  }, [prompt, editedTemplate])

  const handleSaveMetadata = async () => {
    const tags = tagsInput
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean)

    await updateMutation.mutateAsync({
      name: name.trim(),
      description: description.trim() || undefined,
      tags: tags.length > 0 ? tags : undefined,
    })
    setIsEditing(false)
  }

  const handleCreateVersion = async () => {
    if (!editedTemplate) return

    const request: CreateVersionRequest = {
      template: editedTemplate || prompt!.template,
      commit_message: newCommitMessage.trim() || undefined,
    }

    await createVersionMutation.mutateAsync(request)
    setEditedTemplate(null)
    setNewCommitMessage('')
  }

  const handleSetLabels = async (versionId: string, labels: string[]) => {
    await setLabelsMutation.mutateAsync({ versionId, labels })
  }

  if (promptLoading) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="space-y-6">
            <Skeleton className="h-10 w-48" />
            <Skeleton className="h-[400px]" />
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

  const latestVersion = versions?.[0]

  return (
    <>
      <DashboardHeader />
      <Main>
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
          <Button
            variant="outline"
            onClick={() =>
              router.push(`/projects/${params.projectSlug}/prompts/${params.promptId}/versions`)
            }
          >
            <History className="mr-2 h-4 w-4" />
            History
          </Button>
          <Button
            onClick={() => {
              let messages: ReturnType<typeof createMessage>[] = []

              if (prompt.type === 'chat') {
                // Chat template: convert messages array
                const template = prompt.template as ChatTemplate
                messages = template.messages?.map((m) =>
                  createMessage(
                    (m.role || 'user') as 'system' | 'user' | 'assistant',
                    m.content || ''
                  )
                ) || []
              } else {
                // Text template: convert to single user message
                const template = prompt.template as { content?: string }
                messages = [
                  createMessage('user', template.content || ''),
                ]
              }

              // Create loadedTemplate for change detection (normalized without IDs)
              const loadedTemplate = JSON.stringify(
                messages.map(({ role, content }) => ({ role, content }))
              )

              // Directly populate the store (no sessionStorage, no race conditions)
              usePlaygroundStore.getState().loadFromPrompt({
                messages,
                loadedFromPromptId: prompt.id,
                loadedFromPromptName: prompt.name,
                loadedFromPromptVersionId: prompt.version_id,
                loadedFromPromptVersionNumber: prompt.version,
                loadedTemplate,
              })

              // Navigate to playground (no session ID in URL)
              router.push(`/projects/${params.projectSlug}/playground`)
            }}
          >
            <FlaskConical className="mr-2 h-4 w-4" />
            Try in Playground
          </Button>
        </PageHeader>

        <Tabs defaultValue="template" className="space-y-6">
          <TabsList>
            <TabsTrigger value="template">Template</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>

          <TabsContent value="template" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              <div className="lg:col-span-2 space-y-6">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between">
                    <CardTitle>
                      Template (v{prompt.version})
                      {prompt.is_fallback && (
                        <Badge variant="destructive" className="ml-2">
                          Fallback
                        </Badge>
                      )}
                    </CardTitle>
                    {latestVersion && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          setEditedTemplate(prompt.template)
                        }}
                        disabled={!!editedTemplate}
                      >
                        <Plus className="mr-2 h-4 w-4" />
                        New Version
                      </Button>
                    )}
                  </CardHeader>
                  <CardContent>
                    <PromptEditor
                      type={prompt.type}
                      template={editedTemplate || prompt.template}
                      onChange={(t) => setEditedTemplate(t)}
                      variables={variables}
                      readOnly={!editedTemplate}
                    />
                  </CardContent>
                </Card>

                {editedTemplate && (
                  <Card>
                    <CardHeader>
                      <CardTitle>Create New Version</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div className="space-y-2">
                        <Label>Commit Message</Label>
                        <Textarea
                          value={newCommitMessage}
                          onChange={(e) => setNewCommitMessage(e.target.value)}
                          placeholder="Describe your changes..."
                          rows={2}
                        />
                      </div>
                      <div className="flex gap-2">
                        <Button
                          onClick={handleCreateVersion}
                          disabled={createVersionMutation.isPending}
                        >
                          {createVersionMutation.isPending ? (
                            <>
                              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                              Creating...
                            </>
                          ) : (
                            <>
                              <Save className="mr-2 h-4 w-4" />
                              Save as New Version
                            </>
                          )}
                        </Button>
                        <Button
                          variant="outline"
                          onClick={() => {
                            setEditedTemplate(null)
                            setNewCommitMessage('')
                          }}
                        >
                          Cancel
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                )}
              </div>

              <div className="space-y-6">
                <Card>
                  <CardHeader>
                    <CardTitle>Labels</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="flex flex-wrap gap-2">
                      {prompt.labels.map((label) => (
                        <LabelBadge
                          key={label}
                          label={label}
                          isProtected={protectedLabels?.includes(label)}
                        />
                      ))}
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>Variables</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <VariableList variables={prompt.variables} />
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>Info</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Version</span>
                      <span className="font-mono">v{prompt.version}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Created</span>
                      <span>{new Date(prompt.created_at).toLocaleDateString()}</span>
                    </div>
                    {prompt.commit_message && (
                      <div>
                        <span className="text-muted-foreground">Commit:</span>
                        <p className="mt-1">{prompt.commit_message}</p>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="settings" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Prompt Metadata</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Name</Label>
                  <Input
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    disabled={!isEditing}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Description</Label>
                  <Textarea
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    disabled={!isEditing}
                    rows={2}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Tags</Label>
                  <Input
                    value={tagsInput}
                    onChange={(e) => setTagsInput(e.target.value)}
                    disabled={!isEditing}
                    placeholder="tag1, tag2, tag3"
                  />
                </div>
                <div className="flex gap-2">
                  {isEditing ? (
                    <>
                      <Button
                        onClick={handleSaveMetadata}
                        disabled={updateMutation.isPending}
                      >
                        {updateMutation.isPending ? (
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : (
                          <Save className="mr-2 h-4 w-4" />
                        )}
                        Save
                      </Button>
                      <Button
                        variant="outline"
                        onClick={() => {
                          setIsEditing(false)
                          setName(prompt.name)
                          setDescription(prompt.description || '')
                          setTagsInput(prompt.tags?.join(', ') || '')
                        }}
                      >
                        Cancel
                      </Button>
                    </>
                  ) : (
                    <Button variant="outline" onClick={() => setIsEditing(true)}>
                      <Settings className="mr-2 h-4 w-4" />
                      Edit Metadata
                    </Button>
                  )}
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </Main>
    </>
  )
}
