'use client'

import { useState } from 'react'
import { Loader2, Save } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { useCreateSessionMutation } from '../hooks/use-playground-queries'
import { usePlaygroundStore } from '../stores/playground-store'

interface SaveSessionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectId: string
  // Optional sessionId for updating existing saved sessions
  sessionId?: string
  // Called after successful save with new session ID
  onSuccess?: (sessionId: string) => void
}

export function SaveSessionDialog({
  open,
  onOpenChange,
  projectId,
  sessionId,
  onSuccess,
}: SaveSessionDialogProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [tagsInput, setTagsInput] = useState('')

  const createSessionMutation = useCreateSessionMutation(projectId)
  const windows = usePlaygroundStore((s) => s.windows)

  const handleSave = async () => {
    if (!name.trim()) return

    const tags = tagsInput
      .split(',')
      .map((t) => t.trim())
      .filter((t) => t.length > 0)

    const windowsPayload = windows.map((w) => ({
      template: { messages: w.messages },
      variables: w.variables,
      config: w.config || undefined,
      loadedFromPromptId: w.loadedFromPromptId || undefined,
      loadedFromPromptName: w.loadedFromPromptName || undefined,
      loadedFromPromptVersionId: w.loadedFromPromptVersionId || undefined,
      loadedFromPromptVersionNumber: w.loadedFromPromptVersionNumber || undefined,
      loadedTemplate: w.loadedTemplate || undefined,
    }))

    try {
      const savedSession = await createSessionMutation.mutateAsync({
        name: name.trim(),
        description: description.trim() || undefined,
        tags,
        windows: windowsPayload,
      })

      setName('')
      setDescription('')
      setTagsInput('')
      onOpenChange(false)
      onSuccess?.(savedSession.id)
    } catch {
      // Handled by mutation
    }
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      setName('')
      setDescription('')
      setTagsInput('')
    }
    onOpenChange(newOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Save Playground Session</DialogTitle>
          <DialogDescription>
            Save this session for easy access later. Saved sessions persist until deleted.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="session-name">Name *</Label>
            <Input
              id="session-name"
              placeholder="My experiment"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={200}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="session-description">Description</Label>
            <Textarea
              id="session-description"
              placeholder="Optional description..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="session-tags">Tags</Label>
            <Input
              id="session-tags"
              placeholder="comma, separated, tags"
              value={tagsInput}
              onChange={(e) => setTagsInput(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Optional tags to help organize your sessions (max 10)
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={createSessionMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={!name.trim() || createSessionMutation.isPending}
          >
            {createSessionMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                Save Session
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
