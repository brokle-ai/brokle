'use client'

import * as React from 'react'
import { Loader2, Save, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { ScoreInputField } from './score-input-field'
import { ReasonEditor } from './reason-editor'
import type { Score, ScoreConfig } from '../../types'
import { getScoreTagClasses } from '../../lib/score-colors'
import { cn } from '@/lib/utils'

interface AnnotationFormDialogProps {
  score: Score
  config?: ScoreConfig
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (data: { value?: number | null; string_value?: string | null; reason?: string | null }) => void
  onDelete?: () => void
  isSaving?: boolean
  isDeleting?: boolean
}

/**
 * Dialog for editing an existing annotation
 *
 * Features:
 * - Dynamic input based on score data type
 * - Reason editor with character limit
 * - Delete confirmation
 * - Loading states for save/delete
 */
export function AnnotationFormDialog({
  score,
  config,
  open,
  onOpenChange,
  onSave,
  onDelete,
  isSaving = false,
  isDeleting = false,
}: AnnotationFormDialogProps) {
  // Form state
  const [value, setValue] = React.useState<number | string | boolean | null>(() => {
    if (score.data_type === 'BOOLEAN') {
      return score.value === 1 ? true : score.value === 0 ? false : null
    }
    if (score.data_type === 'CATEGORICAL') {
      return score.string_value ?? null
    }
    return score.value ?? null
  })
  const [reason, setReason] = React.useState(score.reason ?? '')

  // Reset form when score changes
  React.useEffect(() => {
    if (score.data_type === 'BOOLEAN') {
      setValue(score.value === 1 ? true : score.value === 0 ? false : null)
    } else if (score.data_type === 'CATEGORICAL') {
      setValue(score.string_value ?? null)
    } else {
      setValue(score.value ?? null)
    }
    setReason(score.reason ?? '')
  }, [score])

  const { indicatorClass, textClass } = getScoreTagClasses(score.name)

  const handleSave = () => {
    const data: { value?: number | null; string_value?: string | null; reason?: string | null } = {
      reason: reason || null,
    }

    if (score.data_type === 'BOOLEAN') {
      data.value = value === true ? 1 : value === false ? 0 : null
    } else if (score.data_type === 'CATEGORICAL') {
      data.string_value = value as string | null
    } else {
      data.value = value as number | null
    }

    onSave(data)
  }

  const isValid = value !== null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <span className={cn('w-3 h-3 rounded-full', indicatorClass)} />
            <DialogTitle className={textClass}>{score.name}</DialogTitle>
          </div>
          <DialogDescription>
            Edit the value and explanation for this score.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Value Input */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Value</label>
            <ScoreInputField
              dataType={score.data_type}
              value={value}
              onChange={setValue}
              config={config}
            />
          </div>

          {/* Reason Editor */}
          <ReasonEditor
            value={reason}
            onChange={setReason}
            collapsible={false}
            defaultExpanded={true}
          />
        </div>

        <DialogFooter className="flex-row justify-between sm:justify-between">
          {/* Delete Button */}
          {onDelete && (
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="destructive" size="sm" disabled={isDeleting || isSaving}>
                  {isDeleting ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Trash2 className="h-4 w-4" />
                  )}
                  <span className="sr-only">Delete</span>
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete annotation?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This will permanently delete the &quot;{score.name}&quot; annotation.
                    This action cannot be undone.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={onDelete}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  >
                    Delete
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          )}

          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSaving || isDeleting}
            >
              Cancel
            </Button>
            <Button
              onClick={handleSave}
              disabled={!isValid || isSaving || isDeleting}
            >
              {isSaving ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  Save
                </>
              )}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
