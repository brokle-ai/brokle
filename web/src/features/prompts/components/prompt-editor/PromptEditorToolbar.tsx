'use client'

import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Save, X } from 'lucide-react'

interface PromptEditorToolbarProps {
  name: string
  type: 'text' | 'chat'
  isDirty: boolean
  isSaving: boolean
  isEditMode: boolean
  onSave: () => void
  onCancel: () => void
}

export function PromptEditorToolbar({
  name,
  type,
  isDirty,
  isSaving,
  isEditMode,
  onSave,
  onCancel,
}: PromptEditorToolbarProps) {
  return (
    <div className="flex items-center justify-between border-b pb-4">
      <div className="flex items-center gap-3">
        <h2 className="text-2xl font-bold tracking-tight">{name || 'New Prompt'}</h2>
        <Badge variant={type === 'text' ? 'default' : 'secondary'}>
          {type}
        </Badge>
        {isDirty && (
          <Badge variant="outline" className="text-amber-600">
            Unsaved changes
          </Badge>
        )}
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" onClick={onCancel} disabled={isSaving}>
          <X className="mr-2 h-4 w-4" />
          Cancel
        </Button>
        <Button onClick={onSave} disabled={!isDirty || isSaving}>
          {isSaving ? (
            'Saving...'
          ) : (
            <>
              <Save className="mr-2 h-4 w-4" />
              {isEditMode ? 'Save Version' : 'Create Prompt'}
            </>
          )}
        </Button>
      </div>
    </div>
  )
}
