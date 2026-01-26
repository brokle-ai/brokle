'use client'

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { RuleForm } from './rule-form'
import { useUpdateEvaluationRuleMutation } from '../hooks/use-evaluation-rules'
import type { EvaluationRule, UpdateEvaluationRuleRequest } from '../types'

interface EditRuleDialogProps {
  projectId: string
  orgId?: string
  rule: EvaluationRule | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function EditRuleDialog({ projectId, orgId, rule, open, onOpenChange }: EditRuleDialogProps) {
  const updateMutation = useUpdateEvaluationRuleMutation(projectId, rule?.id ?? '')

  const handleSubmit = async (data: UpdateEvaluationRuleRequest) => {
    await updateMutation.mutateAsync(data)
    onOpenChange(false)
  }

  if (!rule) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Evaluation Rule</DialogTitle>
          <DialogDescription>
            Update the configuration for &quot;{rule.name}&quot;.
          </DialogDescription>
        </DialogHeader>
        <RuleForm
          rule={rule}
          onSubmit={handleSubmit}
          onCancel={() => onOpenChange(false)}
          isLoading={updateMutation.isPending}
          orgId={orgId}
        />
      </DialogContent>
    </Dialog>
  )
}
