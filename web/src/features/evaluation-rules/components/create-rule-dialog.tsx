'use client'

import { useState } from 'react'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { RuleForm } from './rule-form'
import { useCreateEvaluationRuleMutation } from '../hooks/use-evaluation-rules'
import type { CreateEvaluationRuleRequest } from '../types'

interface CreateRuleDialogProps {
  projectId: string
  orgId?: string
}

export function CreateRuleDialog({ projectId, orgId }: CreateRuleDialogProps) {
  const [open, setOpen] = useState(false)
  const createMutation = useCreateEvaluationRuleMutation(projectId)

  const handleSubmit = async (data: CreateEvaluationRuleRequest) => {
    await createMutation.mutateAsync(data)
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create Rule
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Evaluation Rule</DialogTitle>
          <DialogDescription>
            Define rules to automatically score incoming spans using LLM evaluation,
            built-in scorers, or regex patterns.
          </DialogDescription>
        </DialogHeader>
        <RuleForm
          onSubmit={handleSubmit}
          onCancel={() => setOpen(false)}
          isLoading={createMutation.isPending}
          orgId={orgId}
        />
      </DialogContent>
    </Dialog>
  )
}
