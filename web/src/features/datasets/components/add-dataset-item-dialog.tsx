'use client'

import { useState } from 'react'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { useCreateDatasetItemMutation } from '../hooks/use-datasets'

interface AddDatasetItemDialogProps {
  projectId: string
  datasetId: string
}

export function AddDatasetItemDialog({ projectId, datasetId }: AddDatasetItemDialogProps) {
  const [open, setOpen] = useState(false)
  const [input, setInput] = useState('{\n  \n}')
  const [expected, setExpected] = useState('')
  const [metadata, setMetadata] = useState('')
  const [error, setError] = useState<string | null>(null)

  const createMutation = useCreateDatasetItemMutation(projectId, datasetId)

  const resetForm = () => {
    setInput('{\n  \n}')
    setExpected('')
    setMetadata('')
    setError(null)
  }

  const parseJson = (value: string, fieldName: string): Record<string, unknown> | null => {
    if (!value.trim()) return null
    try {
      const parsed = JSON.parse(value)
      if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
        throw new Error(`${fieldName} must be a JSON object`)
      }
      return parsed
    } catch (e) {
      if (e instanceof SyntaxError) {
        throw new Error(`Invalid JSON in ${fieldName}: ${e.message}`)
      }
      throw e
    }
  }

  const handleSubmit = async () => {
    setError(null)

    try {
      const inputData = parseJson(input, 'Input')
      if (!inputData) {
        setError('Input is required and must be a valid JSON object')
        return
      }

      const expectedData = parseJson(expected, 'Expected')
      const metadataData = parseJson(metadata, 'Metadata')

      await createMutation.mutateAsync({
        input: inputData,
        expected: expectedData ?? undefined,
        metadata: metadataData ?? undefined,
      })

      resetForm()
      setOpen(false)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'An error occurred')
    }
  }

  return (
    <Dialog open={open} onOpenChange={(isOpen) => {
      setOpen(isOpen)
      if (!isOpen) resetForm()
    }}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Add Item
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Add Dataset Item</DialogTitle>
          <DialogDescription>
            Add a new test case to this dataset. Input and expected values should be valid JSON objects.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <Label htmlFor="input">
              Input <span className="text-destructive">*</span>
            </Label>
            <Textarea
              id="input"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder='{"prompt": "Hello, how can I help you?"}'
              className="font-mono text-sm h-32"
            />
            <p className="text-xs text-muted-foreground">
              The input data for this test case (JSON object)
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="expected">Expected (Optional)</Label>
            <Textarea
              id="expected"
              value={expected}
              onChange={(e) => setExpected(e.target.value)}
              placeholder='{"response": "Expected response..."}'
              className="font-mono text-sm h-24"
            />
            <p className="text-xs text-muted-foreground">
              The expected output for evaluation comparison (JSON object)
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="metadata">Metadata (Optional)</Label>
            <Textarea
              id="metadata"
              value={metadata}
              onChange={(e) => setMetadata(e.target.value)}
              placeholder='{"category": "greeting", "priority": 1}'
              className="font-mono text-sm h-20"
            />
            <p className="text-xs text-muted-foreground">
              Additional metadata for this test case (JSON object)
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => setOpen(false)}
            disabled={createMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={createMutation.isPending}
          >
            {createMutation.isPending ? 'Adding...' : 'Add Item'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
