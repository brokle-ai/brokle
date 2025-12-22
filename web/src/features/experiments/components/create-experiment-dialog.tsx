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
import { ExperimentForm } from './experiment-form'
import { useCreateExperimentMutation } from '../hooks/use-experiments'
import type { CreateExperimentRequest } from '../types'

interface CreateExperimentDialogProps {
  projectId: string
}

export function CreateExperimentDialog({ projectId }: CreateExperimentDialogProps) {
  const [open, setOpen] = useState(false)
  const createMutation = useCreateExperimentMutation(projectId)

  const handleSubmit = async (data: CreateExperimentRequest) => {
    await createMutation.mutateAsync(data)
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create Experiment
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Create Experiment</DialogTitle>
          <DialogDescription>
            Create a new experiment to compare model outputs and track evaluations.
            You can also create experiments programmatically via the SDK.
          </DialogDescription>
        </DialogHeader>
        <ExperimentForm
          onSubmit={handleSubmit}
          onCancel={() => setOpen(false)}
          isLoading={createMutation.isPending}
        />
      </DialogContent>
    </Dialog>
  )
}
