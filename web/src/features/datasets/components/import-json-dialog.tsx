'use client'

import { useState, useCallback } from 'react'
import { Upload, FileJson } from 'lucide-react'
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
import { Switch } from '@/components/ui/switch'
import { useImportFromJsonMutation } from '../hooks/use-datasets'

interface ImportJsonDialogProps {
  projectId: string
  datasetId: string
  trigger?: React.ReactNode
}

export function ImportJsonDialog({ projectId, datasetId, trigger }: ImportJsonDialogProps) {
  const [open, setOpen] = useState(false)
  const [jsonContent, setJsonContent] = useState('')
  const [deduplicate, setDeduplicate] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [preview, setPreview] = useState<Record<string, unknown>[] | null>(null)

  const importMutation = useImportFromJsonMutation(projectId, datasetId)

  const resetForm = useCallback(() => {
    setJsonContent('')
    setDeduplicate(true)
    setError(null)
    setPreview(null)
  }, [])

  const parseAndValidate = useCallback((content: string): Record<string, unknown>[] | null => {
    if (!content.trim()) {
      setError('Please enter JSON content')
      setPreview(null)
      return null
    }

    try {
      const trimmed = content.trim()
      let items: Record<string, unknown>[]

      // Try parsing as JSON array first
      try {
        const parsed = JSON.parse(trimmed)
        items = Array.isArray(parsed) ? parsed : [parsed]
      } catch {
        // Try parsing as JSONL (one JSON object per line)
        items = trimmed
          .split('\n')
          .filter((line) => line.trim())
          .map((line) => JSON.parse(line.trim()))
      }

      // Validate items have required structure
      for (let i = 0; i < items.length; i++) {
        const item = items[i]
        if (typeof item !== 'object' || item === null || Array.isArray(item)) {
          throw new Error(`Item ${i + 1} must be a JSON object`)
        }
        // Check if item has input field or any fields that could be mapped
        if (!item.input && Object.keys(item).length === 0) {
          throw new Error(`Item ${i + 1} is empty`)
        }
      }

      setError(null)
      setPreview(items.slice(0, 3))
      return items
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Invalid JSON format'
      setError(message)
      setPreview(null)
      return null
    }
  }, [])

  const handleContentChange = useCallback((value: string) => {
    setJsonContent(value)
    if (value.trim()) {
      parseAndValidate(value)
    } else {
      setError(null)
      setPreview(null)
    }
  }, [parseAndValidate])

  const handleSubmit = async () => {
    const items = parseAndValidate(jsonContent)
    if (!items) return

    try {
      await importMutation.mutateAsync({
        items,
        deduplicate,
      })
      resetForm()
      setOpen(false)
    } catch {
      // Mutation's onError handles toast notification
    }
  }

  const handleFileUpload = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (e) => {
      const content = e.target?.result as string
      setJsonContent(content)
      parseAndValidate(content)
    }
    reader.onerror = () => {
      setError('Failed to read file')
    }
    reader.readAsText(file)

    event.target.value = ''
  }, [parseAndValidate])

  return (
    <Dialog open={open} onOpenChange={(isOpen) => {
      setOpen(isOpen)
      if (!isOpen) resetForm()
    }}>
      <DialogTrigger asChild>
        {trigger ?? (
          <Button variant="outline">
            <FileJson className="mr-2 h-4 w-4" />
            Import JSON
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[700px]">
        <DialogHeader>
          <DialogTitle>Import from JSON</DialogTitle>
          <DialogDescription>
            Paste JSON content or upload a JSON/JSONL file to import items into this dataset.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="json-content">JSON Content</Label>
              <label className="cursor-pointer">
                <input
                  type="file"
                  accept=".json,.jsonl"
                  onChange={handleFileUpload}
                  className="hidden"
                />
                <Button variant="ghost" size="sm" asChild>
                  <span>
                    <Upload className="mr-2 h-4 w-4" />
                    Upload File
                  </span>
                </Button>
              </label>
            </div>
            <Textarea
              id="json-content"
              value={jsonContent}
              onChange={(e) => handleContentChange(e.target.value)}
              placeholder={`[
  {"input": {"prompt": "Hello"}, "expected": {"response": "Hi there!"}},
  {"input": {"prompt": "Goodbye"}, "expected": {"response": "See you!"}}
]

Or JSONL format (one object per line):
{"input": {"prompt": "Hello"}, "expected": {"response": "Hi there!"}}
{"input": {"prompt": "Goodbye"}, "expected": {"response": "See you!"}}`}
              className="font-mono text-sm h-48"
            />
            <p className="text-xs text-muted-foreground">
              Each item should have an &quot;input&quot; field. &quot;expected&quot; and &quot;metadata&quot; fields are optional.
            </p>
          </div>

          {preview && preview.length > 0 && (
            <div className="space-y-2">
              <Label>Preview (first {preview.length} items)</Label>
              <div className="rounded-md border bg-muted/50 p-3 max-h-32 overflow-auto">
                <pre className="text-xs font-mono">
                  {JSON.stringify(preview, null, 2)}
                </pre>
              </div>
            </div>
          )}

          <div className="flex items-center justify-between rounded-lg border p-3">
            <div className="space-y-0.5">
              <Label htmlFor="deduplicate">Skip duplicates</Label>
              <p className="text-xs text-muted-foreground">
                Skip items with identical content to existing items
              </p>
            </div>
            <Switch
              id="deduplicate"
              checked={deduplicate}
              onCheckedChange={setDeduplicate}
            />
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => setOpen(false)}
            disabled={importMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={importMutation.isPending || !preview}
          >
            {importMutation.isPending ? 'Importing...' : 'Import Items'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
