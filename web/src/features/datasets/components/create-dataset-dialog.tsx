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
import { DatasetForm } from './dataset-form'
import { useCreateDatasetMutation } from '../hooks/use-datasets'
import type { CreateDatasetRequest } from '../types'

interface CreateDatasetDialogProps {
  projectId: string
}

export function CreateDatasetDialog({ projectId }: CreateDatasetDialogProps) {
  const [open, setOpen] = useState(false)
  const createMutation = useCreateDatasetMutation(projectId)

  const handleSubmit = async (data: CreateDatasetRequest) => {
    await createMutation.mutateAsync(data)
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create Dataset
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Create Dataset</DialogTitle>
          <DialogDescription>
            Create a new dataset to organize test cases for batch evaluations.
          </DialogDescription>
        </DialogHeader>
        <DatasetForm
          onSubmit={handleSubmit}
          onCancel={() => setOpen(false)}
          isLoading={createMutation.isPending}
        />
      </DialogContent>
    </Dialog>
  )
}
