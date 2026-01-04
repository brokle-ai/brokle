'use client'

import { useState, useRef, useCallback } from 'react'
import { AlertCircle, CheckCircle, FileJson, Upload } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { cn } from '@/lib/utils'
import type { DashboardExport, DashboardImportRequest } from '../types'

interface ImportDashboardDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onImport: (data: DashboardImportRequest) => void
  isPending?: boolean
}

interface ValidationResult {
  isValid: boolean
  error?: string
  data?: DashboardExport
}

function validateDashboardExport(json: unknown): ValidationResult {
  if (!json || typeof json !== 'object') {
    return { isValid: false, error: 'Invalid JSON: expected an object' }
  }

  const data = json as Record<string, unknown>

  // Check required fields
  if (!data.version || typeof data.version !== 'string') {
    return { isValid: false, error: 'Missing or invalid "version" field' }
  }

  if (!data.name || typeof data.name !== 'string') {
    return { isValid: false, error: 'Missing or invalid "name" field' }
  }

  if (!data.config || typeof data.config !== 'object') {
    return { isValid: false, error: 'Missing or invalid "config" field' }
  }

  if (!Array.isArray(data.layout)) {
    return { isValid: false, error: 'Missing or invalid "layout" field' }
  }

  // Validate layout items
  for (let i = 0; i < data.layout.length; i++) {
    const item = data.layout[i] as Record<string, unknown>
    if (!item.widget_id || typeof item.widget_id !== 'string') {
      return { isValid: false, error: `Layout item ${i} missing "widget_id"` }
    }
    if (typeof item.x !== 'number' || typeof item.y !== 'number') {
      return { isValid: false, error: `Layout item ${i} has invalid position` }
    }
    if (typeof item.w !== 'number' || typeof item.h !== 'number') {
      return { isValid: false, error: `Layout item ${i} has invalid size` }
    }
  }

  return { isValid: true, data: data as unknown as DashboardExport }
}

export function ImportDashboardDialog({
  open,
  onOpenChange,
  onImport,
  isPending,
}: ImportDashboardDialogProps) {
  const [activeTab, setActiveTab] = useState<'file' | 'paste'>('file')
  const [jsonText, setJsonText] = useState('')
  const [overrideName, setOverrideName] = useState('')
  const [validation, setValidation] = useState<ValidationResult | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleReset = useCallback(() => {
    setJsonText('')
    setOverrideName('')
    setValidation(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }, [])

  const handleValidate = useCallback((text: string) => {
    if (!text.trim()) {
      setValidation(null)
      return
    }

    try {
      const parsed = JSON.parse(text)
      const result = validateDashboardExport(parsed)
      setValidation(result)

      // Auto-populate override name if empty
      if (result.isValid && result.data && !overrideName) {
        setOverrideName(result.data.name)
      }
    } catch {
      setValidation({ isValid: false, error: 'Invalid JSON syntax' })
    }
  }, [overrideName])

  const handleTextChange = useCallback(
    (text: string) => {
      setJsonText(text)
      handleValidate(text)
    },
    [handleValidate]
  )

  const handleFileUpload = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0]
      if (!file) return

      const reader = new FileReader()
      reader.onload = (e) => {
        const text = e.target?.result as string
        setJsonText(text)
        handleValidate(text)
      }
      reader.onerror = () => {
        setValidation({ isValid: false, error: 'Failed to read file' })
      }
      reader.readAsText(file)
    },
    [handleValidate]
  )

  const handleDrop = useCallback(
    (event: React.DragEvent<HTMLDivElement>) => {
      event.preventDefault()
      const file = event.dataTransfer.files?.[0]
      if (!file) return

      if (!file.name.endsWith('.json')) {
        setValidation({ isValid: false, error: 'Only JSON files are supported' })
        return
      }

      const reader = new FileReader()
      reader.onload = (e) => {
        const text = e.target?.result as string
        setJsonText(text)
        handleValidate(text)
      }
      reader.readAsText(file)
    },
    [handleValidate]
  )

  const handleDragOver = useCallback((event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault()
  }, [])

  const handleImport = useCallback(() => {
    if (!validation?.isValid || !validation.data) return

    const importRequest: DashboardImportRequest = {
      data: validation.data,
      name: overrideName || undefined,
    }

    onImport(importRequest)
  }, [validation, overrideName, onImport])

  const handleClose = useCallback(
    (open: boolean) => {
      if (!open) {
        handleReset()
      }
      onOpenChange(open)
    },
    [handleReset, onOpenChange]
  )

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Import Dashboard</DialogTitle>
          <DialogDescription>
            Import a dashboard configuration from a JSON file or paste the JSON directly.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'file' | 'paste')}>
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="file">Upload File</TabsTrigger>
              <TabsTrigger value="paste">Paste JSON</TabsTrigger>
            </TabsList>

            <TabsContent value="file" className="space-y-4">
              <div
                onDrop={handleDrop}
                onDragOver={handleDragOver}
                className={cn(
                  'flex flex-col items-center justify-center gap-3 rounded-lg border-2 border-dashed p-8 transition-colors',
                  'hover:border-primary/50 hover:bg-accent/50',
                  validation?.isValid && 'border-green-500 bg-green-50 dark:bg-green-950'
                )}
              >
                {validation?.isValid ? (
                  <CheckCircle className="h-10 w-10 text-green-500" />
                ) : (
                  <FileJson className="h-10 w-10 text-muted-foreground" />
                )}
                <div className="text-center">
                  <p className="text-sm font-medium">
                    {validation?.isValid
                      ? validation.data?.name
                      : 'Drop JSON file here or click to upload'}
                  </p>
                  {validation?.isValid && validation.data && (
                    <p className="text-xs text-muted-foreground mt-1">
                      {validation.data.config.widgets?.length ?? 0} widgets •{' '}
                      Exported {new Date(validation.data.exported_at).toLocaleDateString()}
                    </p>
                  )}
                </div>
                <Input
                  ref={fileInputRef}
                  type="file"
                  accept=".json"
                  onChange={handleFileUpload}
                  className="hidden"
                  id="dashboard-file-upload"
                />
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => fileInputRef.current?.click()}
                >
                  <Upload className="h-4 w-4 mr-1.5" />
                  Select File
                </Button>
              </div>
            </TabsContent>

            <TabsContent value="paste" className="space-y-4">
              <Textarea
                value={jsonText}
                onChange={(e) => handleTextChange(e.target.value)}
                placeholder="Paste your dashboard JSON here..."
                className="min-h-[200px] font-mono text-sm"
              />
            </TabsContent>
          </Tabs>

          {/* Validation message */}
          {validation && !validation.isValid && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{validation.error}</AlertDescription>
            </Alert>
          )}

          {/* Import options */}
          {validation?.isValid && (
            <div className="space-y-4 pt-2">
              <div className="space-y-2">
                <Label htmlFor="override-name">Dashboard Name</Label>
                <Input
                  id="override-name"
                  value={overrideName}
                  onChange={(e) => setOverrideName(e.target.value)}
                  placeholder={validation.data?.name}
                />
                <p className="text-xs text-muted-foreground">
                  Override the imported dashboard name (optional)
                </p>
              </div>

            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => handleClose(false)}>
            Cancel
          </Button>
          <Button
            onClick={handleImport}
            disabled={!validation?.isValid || isPending}
          >
            {isPending ? 'Importing...' : 'Import Dashboard'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
