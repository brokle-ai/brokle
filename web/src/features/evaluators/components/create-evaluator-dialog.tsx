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
import { EvaluatorForm } from './evaluator-form'
import { useCreateEvaluatorMutation } from '../hooks/use-evaluators'
import type { CreateEvaluatorRequest } from '../types'

interface CreateEvaluatorDialogProps {
  projectId: string
  orgId?: string
}

export function CreateEvaluatorDialog({ projectId, orgId }: CreateEvaluatorDialogProps) {
  const [open, setOpen] = useState(false)
  const createMutation = useCreateEvaluatorMutation(projectId)

  const handleSubmit = async (data: CreateEvaluatorRequest) => {
    await createMutation.mutateAsync(data)
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create Evaluator
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Evaluator</DialogTitle>
          <DialogDescription>
            Define evaluators to automatically score incoming spans using LLM evaluation,
            built-in scorers, or regex patterns.
          </DialogDescription>
        </DialogHeader>
        <EvaluatorForm
          onSubmit={handleSubmit}
          onCancel={() => setOpen(false)}
          isLoading={createMutation.isPending}
          orgId={orgId}
        />
      </DialogContent>
    </Dialog>
  )
}
