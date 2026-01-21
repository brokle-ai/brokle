'use client'

import * as React from 'react'
import { Save, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'

const COMMIT_MESSAGE_MAX_LENGTH = 500

interface SaveVersionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (commitMessage?: string) => Promise<void>
  isSaving: boolean
}

export function SaveVersionDialog({
  open,
  onOpenChange,
  onSave,
  isSaving,
}: SaveVersionDialogProps) {
  const [commitMessage, setCommitMessage] = React.useState('')

  // Reset commit message when dialog opens
  React.useEffect(() => {
    if (open) {
      setCommitMessage('')
    }
  }, [open])

  const handleSave = async () => {
    await onSave(commitMessage.trim() || undefined)
    onOpenChange(false)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Cmd/Ctrl + Enter to save
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey) && !isSaving) {
      e.preventDefault()
      handleSave()
    }
  }

  const remainingChars = COMMIT_MESSAGE_MAX_LENGTH - commitMessage.length
  const isOverLimit = remainingChars < 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Save New Version</DialogTitle>
          <DialogDescription>
            Create a new version of this prompt with your changes.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3 py-4">
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="commit-message">Commit message</Label>
              <span
                className={`text-xs ${isOverLimit ? 'text-destructive' : 'text-muted-foreground'}`}
              >
                {remainingChars}
              </span>
            </div>
            <Textarea
              id="commit-message"
              value={commitMessage}
              onChange={(e) => setCommitMessage(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Add commit message..."
              className="min-h-[100px] resize-none"
              autoFocus
            />
            <p className="text-xs text-muted-foreground">
              Provide information about the changes made in this version. Helps
              maintain a clear history of prompt iterations.
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isSaving}
          >
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isSaving || isOverLimit}>
            {isSaving ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Save className="h-4 w-4 mr-2" />
            )}
            {isSaving ? 'Saving...' : 'Save Version'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
