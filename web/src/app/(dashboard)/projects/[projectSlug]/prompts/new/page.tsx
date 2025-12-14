'use client'

import { useState, useMemo } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ArrowLeft, Save, Loader2 } from 'lucide-react'
import { useProjectOnly } from '@/features/projects'
import {
  useCreatePromptMutation,
  PromptEditor,
  ModelConfigForm,
  LabelManager,
} from '@/features/prompts'
import type {
  PromptType,
  TextTemplate,
  ChatTemplate,
  ModelConfig,
  CreatePromptRequest,
} from '@/features/prompts'

// Extract variables from template
function extractVariables(template: TextTemplate | ChatTemplate, type: PromptType): string[] {
  const variablePattern = /\{\{(\w+)\}\}/g
  const variables = new Set<string>()

  if (type === 'text') {
    const content = (template as TextTemplate).content || ''
    let match
    while ((match = variablePattern.exec(content)) !== null) {
      variables.add(match[1])
    }
  } else {
    const messages = (template as ChatTemplate).messages || []
    for (const msg of messages) {
      if (msg.content) {
        let match
        while ((match = variablePattern.exec(msg.content)) !== null) {
          variables.add(match[1])
        }
      }
    }
  }

  return Array.from(variables)
}

export default function NewPromptPage() {
  const params = useParams<{ projectSlug: string }>()
  const router = useRouter()
  const { currentProject } = useProjectOnly()
  const createMutation = useCreatePromptMutation(currentProject?.id || '')

  // Form state
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [type, setType] = useState<PromptType>('text')
  const [template, setTemplate] = useState<TextTemplate | ChatTemplate>({ content: '' })
  const [config, setConfig] = useState<ModelConfig>({
    model: 'gpt-4o-mini',
    temperature: 0.7,
    max_tokens: 1024,
  })
  const [labels, setLabels] = useState<string[]>([])
  const [commitMessage, setCommitMessage] = useState('')
  const [tagsInput, setTagsInput] = useState('')

  const variables = useMemo(() => extractVariables(template, type), [template, type])

  const handleTypeChange = (newType: PromptType) => {
    setType(newType)
    if (newType === 'text') {
      setTemplate({ content: '' })
    } else {
      setTemplate({ messages: [{ type: 'message', role: 'system', content: '' }] })
    }
  }

  const handleSubmit = async () => {
    if (!currentProject?.id || !name.trim()) return

    const tags = tagsInput
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean)

    const request: CreatePromptRequest = {
      name: name.trim(),
      type,
      description: description.trim() || undefined,
      tags: tags.length > 0 ? tags : undefined,
      template,
      config,
      labels: labels.length > 0 ? labels : undefined,
      commit_message: commitMessage.trim() || undefined,
    }

    try {
      const newPrompt = await createMutation.mutateAsync(request)
      router.push(`/projects/${params.projectSlug}/prompts/${newPrompt.id}`)
    } catch (error) {
      // Error handled by mutation
    }
  }

  return (
    <>
      <DashboardHeader />
      <Main>
        <div className="mb-6 flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => router.back()}
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h2 className="text-2xl font-bold tracking-tight">New Prompt</h2>
            <p className="text-muted-foreground">Create a new prompt template</p>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Basic Info</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Name *</Label>
                  <Input
                    id="name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="my-prompt"
                  />
                  <p className="text-xs text-muted-foreground">
                    Use lowercase letters, numbers, and hyphens
                  </p>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="description">Description</Label>
                  <Textarea
                    id="description"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="What does this prompt do?"
                    rows={2}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="tags">Tags</Label>
                  <Input
                    id="tags"
                    value={tagsInput}
                    onChange={(e) => setTagsInput(e.target.value)}
                    placeholder="tag1, tag2, tag3"
                  />
                  <p className="text-xs text-muted-foreground">
                    Comma-separated list of tags
                  </p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Template</CardTitle>
              </CardHeader>
              <CardContent>
                <PromptEditor
                  type={type}
                  template={template}
                  onChange={setTemplate}
                  onTypeChange={handleTypeChange}
                  variables={variables}
                />
              </CardContent>
            </Card>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Labels</CardTitle>
              </CardHeader>
              <CardContent>
                <LabelManager
                  labels={labels}
                  onChange={setLabels}
                  availableLabels={['production', 'staging', 'development']}
                />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Model Configuration</CardTitle>
              </CardHeader>
              <CardContent>
                <ModelConfigForm config={config} onChange={setConfig} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Commit Message</CardTitle>
              </CardHeader>
              <CardContent>
                <Textarea
                  value={commitMessage}
                  onChange={(e) => setCommitMessage(e.target.value)}
                  placeholder="Initial version"
                  rows={2}
                />
              </CardContent>
            </Card>

            <Button
              onClick={handleSubmit}
              disabled={!name.trim() || createMutation.isPending}
              className="w-full"
              size="lg"
            >
              {createMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  Create Prompt
                </>
              )}
            </Button>
          </div>
        </div>
      </Main>
    </>
  )
}
