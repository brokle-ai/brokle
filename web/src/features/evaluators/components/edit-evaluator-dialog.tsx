'use client'

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { EvaluatorForm } from './evaluator-form'
import { useUpdateEvaluatorMutation } from '../hooks/use-evaluators'
import type { Evaluator, UpdateEvaluatorRequest } from '../types'

interface EditEvaluatorDialogProps {
  projectId: string
  orgId?: string
  evaluator: Evaluator | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function EditEvaluatorDialog({ projectId, orgId, evaluator, open, onOpenChange }: EditEvaluatorDialogProps) {
  const updateMutation = useUpdateEvaluatorMutation(projectId, evaluator?.id ?? '')

  const handleSubmit = async (data: UpdateEvaluatorRequest) => {
    await updateMutation.mutateAsync(data)
    onOpenChange(false)
  }

  if (!evaluator) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Evaluator</DialogTitle>
          <DialogDescription>
            Update the configuration for &quot;{evaluator.name}&quot;.
          </DialogDescription>
        </DialogHeader>
        <EvaluatorForm
          evaluator={evaluator}
          onSubmit={handleSubmit}
          onCancel={() => onOpenChange(false)}
          isLoading={updateMutation.isPending}
          orgId={orgId}
        />
      </DialogContent>
    </Dialog>
  )
}
