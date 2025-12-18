'use client'

import { useState } from 'react'
import { Save, Loader2 } from 'lucide-react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { createPrompt, createVersion } from '@/features/prompts/api/prompts-api'
import type { ChatMessage, ModelConfig } from '../types'

// Callback data when prompt is saved
export interface PromptSavedData {
  promptId: string
  promptName: string
  versionId: string
  versionNumber: number
  isNewPrompt: boolean
}

interface SaveAsPromptDialogProps {
  projectId: string
  messages: ChatMessage[]
  config: ModelConfig | null
  loadedFromPromptId: string | null
  loadedFromPromptName: string | null
  loadedFromPromptVersionNumber?: number | null
  disabled?: boolean
  onSuccess?: (data: PromptSavedData) => void
}

type SaveMode = 'new' | 'version'

export function SaveAsPromptDialog({
  projectId,
  messages,
  config,
  loadedFromPromptId,
  loadedFromPromptName,
  loadedFromPromptVersionNumber,
  disabled,
  onSuccess,
}: SaveAsPromptDialogProps) {
  const [open, setOpen] = useState(false)
  const [saveMode, setSaveMode] = useState<SaveMode>(loadedFromPromptId ? 'version' : 'new')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [commitMessage, setCommitMessage] = useState('')

  const queryClient = useQueryClient()

  // Convert playground messages to prompts format (add 'type' field required by backend)
  // Filter out messages with empty content (backend rejects them)
  const toPromptMessages = () =>
    messages
      .filter((m) => m.content.trim().length > 0)
      .map((m) => ({
        type: 'message' as const,
        role: m.role,
        content: m.content,
      }))

  const hasValidMessages = messages.some((m) => m.content.trim().length > 0)

  const createPromptMutation = useMutation({
    mutationFn: async () => {
      const template = {
        messages: toPromptMessages(),
      }

      return createPrompt(projectId, {
        name,
        type: 'chat',
        description: description || undefined,
        template,
        config: config || undefined,
        commit_message: commitMessage || 'Initial version from playground',
      })
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['prompts', projectId] })
      onSuccess?.({
        promptId: data.id,
        promptName: name,
        versionId: data.version_id,
        versionNumber: data.version,
        isNewPrompt: true,
      })
      handleClose()
    },
  })

  const createVersionMutation = useMutation({
    mutationFn: async () => {
      if (!loadedFromPromptId) throw new Error('No prompt to update')

      const template = {
        messages: toPromptMessages(),
      }

      return createVersion(projectId, loadedFromPromptId, {
        template,
        config: config || undefined,
        commit_message: commitMessage || 'Updated from playground',
      })
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['prompts', projectId] })
      if (loadedFromPromptId && loadedFromPromptName) {
        onSuccess?.({
          promptId: loadedFromPromptId,
          promptName: loadedFromPromptName,
          versionId: data.id,
          versionNumber: data.version,
          isNewPrompt: false,
        })
      }
      handleClose()
    },
  })

  const handleClose = () => {
    setOpen(false)
    setName('')
    setDescription('')
    setCommitMessage('')
    setSaveMode(loadedFromPromptId ? 'version' : 'new')
  }

  const handleSave = () => {
    if (saveMode === 'new') {
      createPromptMutation.mutate()
    } else {
      createVersionMutation.mutate()
    }
  }

  const isLoading = createPromptMutation.isPending || createVersionMutation.isPending
  const error = createPromptMutation.error || createVersionMutation.error

  const canSave =
    hasValidMessages &&
    (saveMode === 'version' || (saveMode === 'new' && name.trim().length > 0))

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <Tooltip>
        <TooltipTrigger asChild>
          <DialogTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8" disabled={disabled}>
              <Save className="h-4 w-4" />
            </Button>
          </DialogTrigger>
        </TooltipTrigger>
        <TooltipContent>Save as Prompt</TooltipContent>
      </Tooltip>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Save as Prompt</DialogTitle>
          <DialogDescription>
            Save your current messages as a prompt for reuse and version control.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Mode selection - only show if loaded from existing prompt */}
          {loadedFromPromptId && (
            <div className="space-y-3">
              <Label>Save as</Label>
              <RadioGroup
                value={saveMode}
                onValueChange={(value) => setSaveMode(value as SaveMode)}
              >
                <div className="flex items-start space-x-3">
                  <RadioGroupItem value="version" id="version" />
                  <div className="flex flex-col">
                    <Label htmlFor="version" className="font-normal cursor-pointer">
                      New version of "{loadedFromPromptName}"
                      {loadedFromPromptVersionNumber && (
                        <span className="ml-1 text-xs text-muted-foreground">
                          (based on v{loadedFromPromptVersionNumber})
                        </span>
                      )}
                    </Label>
                    <span className="text-xs text-muted-foreground">
                      Add a new version to the existing prompt
                    </span>
                  </div>
                </div>
                <div className="flex items-start space-x-3">
                  <RadioGroupItem value="new" id="new" />
                  <div className="flex flex-col">
                    <Label htmlFor="new" className="font-normal cursor-pointer">
                      New prompt
                    </Label>
                    <span className="text-xs text-muted-foreground">
                      Create a completely new prompt
                    </span>
                  </div>
                </div>
              </RadioGroup>
            </div>
          )}

          {/* Name (only for new prompts) */}
          {saveMode === 'new' && (
            <div className="space-y-2">
              <Label htmlFor="name">
                Name <span className="text-destructive">*</span>
              </Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="my-prompt"
                autoFocus
              />
              <p className="text-xs text-muted-foreground">
                A unique name for your prompt (letters, numbers, hyphens)
              </p>
            </div>
          )}

          {/* Description (only for new prompts) */}
          {saveMode === 'new' && (
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional description..."
                rows={2}
              />
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="commit">Commit Message</Label>
            <Input
              id="commit"
              value={commitMessage}
              onChange={(e) => setCommitMessage(e.target.value)}
              placeholder={saveMode === 'new' ? 'Initial version' : 'Updated from playground'}
            />
            <p className="text-xs text-muted-foreground">
              Describe the changes for version history
            </p>
          </div>

          {!hasValidMessages && (
            <div className="text-sm text-amber-600 dark:text-amber-500 bg-amber-50 dark:bg-amber-950/50 p-3 rounded-md">
              No messages with content to save. Add content to at least one message before saving.
            </div>
          )}

          {error && (
            <div className="text-sm text-destructive">
              {error instanceof Error ? error.message : 'Failed to save prompt'}
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={isLoading}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={!canSave || isLoading}>
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                {saveMode === 'version' ? 'Save Version' : 'Create Prompt'}
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
