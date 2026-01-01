'use client'

import { useState, useCallback, useMemo } from 'react'
import { Upload, FileSpreadsheet, ChevronLeft, ChevronRight, Check } from 'lucide-react'
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
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { useImportFromCsvMutation } from '../hooks/use-datasets'
import type { CsvPreview, CSVColumnMapping } from '../types'

interface ImportCsvDialogProps {
  projectId: string
  datasetId: string
  trigger?: React.ReactNode
}

type Step = 'upload' | 'mapping' | 'confirm'

const PREVIEW_ROWS = 5

function parseCSV(content: string, hasHeader: boolean): CsvPreview | null {
  const lines = content.split('\n').filter((line) => line.trim())
  if (lines.length === 0) return null

  const rows: string[][] = []
  let currentRow: string[] = []
  let inQuotes = false
  let currentField = ''

  for (const line of lines) {
    for (let i = 0; i < line.length; i++) {
      const char = line[i]
      const nextChar = line[i + 1]

      if (char === '"' && !inQuotes) {
        inQuotes = true
      } else if (char === '"' && inQuotes) {
        if (nextChar === '"') {
          currentField += '"'
          i++
        } else {
          inQuotes = false
        }
      } else if (char === ',' && !inQuotes) {
        currentRow.push(currentField.trim())
        currentField = ''
      } else {
        currentField += char
      }
    }

    if (!inQuotes) {
      currentRow.push(currentField.trim())
      rows.push(currentRow)
      currentRow = []
      currentField = ''
    } else {
      currentField += '\n'
    }
  }

  if (currentRow.length > 0 || currentField) {
    currentRow.push(currentField.trim())
    rows.push(currentRow)
  }

  if (rows.length === 0) return null

  let headers: string[]
  let dataRows: string[][]

  if (hasHeader && rows.length > 0) {
    headers = rows[0]
    dataRows = rows.slice(1)
  } else {
    const numCols = rows[0]?.length || 0
    headers = Array.from({ length: numCols }, (_, i) => `col_${i}`)
    dataRows = rows
  }

  return {
    headers,
    rows: dataRows,
    rowCount: dataRows.length,
  }
}

export function ImportCsvDialog({ projectId, datasetId, trigger }: ImportCsvDialogProps) {
  const [open, setOpen] = useState(false)
  const [step, setStep] = useState<Step>('upload')
  const [csvContent, setCsvContent] = useState('')
  const [hasHeader, setHasHeader] = useState(true)
  const [deduplicate, setDeduplicate] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [preview, setPreview] = useState<CsvPreview | null>(null)
  const [columnMapping, setColumnMapping] = useState<CSVColumnMapping>({
    input_column: '',
    expected_column: '',
    metadata_columns: [],
  })

  const importMutation = useImportFromCsvMutation(projectId, datasetId)

  const resetForm = useCallback(() => {
    setStep('upload')
    setCsvContent('')
    setHasHeader(true)
    setDeduplicate(true)
    setError(null)
    setPreview(null)
    setColumnMapping({
      input_column: '',
      expected_column: '',
      metadata_columns: [],
    })
  }, [])

  const parseAndValidate = useCallback((content: string, header: boolean) => {
    if (!content.trim()) {
      setError('Please provide CSV content')
      setPreview(null)
      return false
    }

    const parsed = parseCSV(content, header)
    if (!parsed || parsed.rowCount === 0) {
      setError('Could not parse CSV content or no data rows found')
      setPreview(null)
      return false
    }

    if (parsed.headers.length === 0) {
      setError('CSV must have at least one column')
      setPreview(null)
      return false
    }

    setError(null)
    setPreview(parsed)

    // Auto-select first column as input if not set
    if (!columnMapping.input_column && parsed.headers.length > 0) {
      setColumnMapping((prev) => ({
        ...prev,
        input_column: parsed.headers[0],
      }))
    }

    return true
  }, [columnMapping.input_column])

  const handleFileUpload = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    if (!file.name.endsWith('.csv')) {
      setError('Please upload a CSV file')
      return
    }

    const reader = new FileReader()
    reader.onload = (e) => {
      const content = e.target?.result as string
      setCsvContent(content)
      parseAndValidate(content, hasHeader)
    }
    reader.onerror = () => {
      setError('Failed to read file')
    }
    reader.readAsText(file)

    event.target.value = ''
  }, [hasHeader, parseAndValidate])

  const handleHasHeaderChange = useCallback((checked: boolean) => {
    setHasHeader(checked)
    if (csvContent) {
      parseAndValidate(csvContent, checked)
    }
  }, [csvContent, parseAndValidate])

  const handleInputColumnChange = useCallback((value: string) => {
    setColumnMapping((prev) => ({
      ...prev,
      input_column: value,
    }))
  }, [])

  const handleExpectedColumnChange = useCallback((value: string) => {
    setColumnMapping((prev) => ({
      ...prev,
      expected_column: value === 'none' ? '' : value,
    }))
  }, [])

  const handleMetadataColumnToggle = useCallback((column: string) => {
    setColumnMapping((prev) => {
      const current = prev.metadata_columns || []
      const isSelected = current.includes(column)
      return {
        ...prev,
        metadata_columns: isSelected
          ? current.filter((c) => c !== column)
          : [...current, column],
      }
    })
  }, [])

  const availableForExpected = useMemo(() => {
    if (!preview) return []
    return preview.headers.filter(
      (h) => h !== columnMapping.input_column && !columnMapping.metadata_columns?.includes(h)
    )
  }, [preview, columnMapping.input_column, columnMapping.metadata_columns])

  const availableForMetadata = useMemo(() => {
    if (!preview) return []
    return preview.headers.filter(
      (h) => h !== columnMapping.input_column && h !== columnMapping.expected_column
    )
  }, [preview, columnMapping.input_column, columnMapping.expected_column])

  const canProceedToMapping = preview && preview.rowCount > 0
  const canProceedToConfirm = columnMapping.input_column !== ''
  const canImport = canProceedToConfirm

  const handleNext = () => {
    if (step === 'upload' && canProceedToMapping) {
      setStep('mapping')
    } else if (step === 'mapping' && canProceedToConfirm) {
      setStep('confirm')
    }
  }

  const handleBack = () => {
    if (step === 'mapping') {
      setStep('upload')
    } else if (step === 'confirm') {
      setStep('mapping')
    }
  }

  const handleSubmit = async () => {
    if (!preview || !canImport) return

    try {
      await importMutation.mutateAsync({
        content: csvContent,
        column_mapping: {
          input_column: columnMapping.input_column,
          expected_column: columnMapping.expected_column || undefined,
          metadata_columns: columnMapping.metadata_columns?.length
            ? columnMapping.metadata_columns
            : undefined,
        },
        has_header: hasHeader,
        deduplicate,
      })
      resetForm()
      setOpen(false)
    } catch {
      // Mutation's onError handles toast notification
    }
  }

  const renderStepIndicator = () => (
    <div className="flex items-center justify-center gap-2 mb-4">
      {(['upload', 'mapping', 'confirm'] as const).map((s, i) => (
        <div key={s} className="flex items-center">
          <div
            className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
              step === s
                ? 'bg-primary text-primary-foreground'
                : (step === 'mapping' && s === 'upload') ||
                  (step === 'confirm' && (s === 'upload' || s === 'mapping'))
                  ? 'bg-primary/20 text-primary'
                  : 'bg-muted text-muted-foreground'
            }`}
          >
            {(step === 'mapping' && s === 'upload') ||
            (step === 'confirm' && (s === 'upload' || s === 'mapping')) ? (
              <Check className="h-4 w-4" />
            ) : (
              i + 1
            )}
          </div>
          {i < 2 && <div className="w-12 h-0.5 bg-muted mx-1" />}
        </div>
      ))}
    </div>
  )

  const renderUploadStep = () => (
    <div className="space-y-4">
      <div className="border-2 border-dashed rounded-lg p-8 text-center">
        <FileSpreadsheet className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
        <div className="space-y-2">
          <label className="cursor-pointer">
            <input
              type="file"
              accept=".csv"
              onChange={handleFileUpload}
              className="hidden"
            />
            <Button variant="outline" asChild>
              <span>
                <Upload className="mr-2 h-4 w-4" />
                Choose CSV File
              </span>
            </Button>
          </label>
          <p className="text-sm text-muted-foreground">
            Upload a CSV file to import items into this dataset
          </p>
        </div>
      </div>

      <div className="flex items-center justify-between rounded-lg border p-3">
        <div className="space-y-0.5">
          <Label htmlFor="has-header">First row is header</Label>
          <p className="text-xs text-muted-foreground">
            Use first row as column names
          </p>
        </div>
        <Switch
          id="has-header"
          checked={hasHeader}
          onCheckedChange={handleHasHeaderChange}
        />
      </div>

      {preview && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <Label>Preview ({preview.rowCount} rows, {preview.headers.length} columns)</Label>
          </div>
          <div className="rounded-md border overflow-auto max-h-48">
            <table className="w-full text-sm">
              <thead className="bg-muted/50">
                <tr>
                  {preview.headers.map((header, i) => (
                    <th key={i} className="px-3 py-2 text-left font-medium">
                      {header}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {preview.rows.slice(0, PREVIEW_ROWS).map((row, i) => (
                  <tr key={i} className="border-t">
                    {row.map((cell, j) => (
                      <td key={j} className="px-3 py-2 font-mono text-xs truncate max-w-[200px]">
                        {cell || <span className="text-muted-foreground italic">empty</span>}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {preview.rowCount > PREVIEW_ROWS && (
            <p className="text-xs text-muted-foreground text-center">
              Showing first {PREVIEW_ROWS} of {preview.rowCount} rows
            </p>
          )}
        </div>
      )}
    </div>
  )

  const renderMappingStep = () => (
    <div className="space-y-4">
      <div className="space-y-3">
        <div className="space-y-2">
          <Label htmlFor="input-column">Input Column *</Label>
          <Select value={columnMapping.input_column} onValueChange={handleInputColumnChange}>
            <SelectTrigger id="input-column">
              <SelectValue placeholder="Select input column" />
            </SelectTrigger>
            <SelectContent>
              {preview?.headers.map((header) => (
                <SelectItem key={header} value={header}>
                  {header}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="text-xs text-muted-foreground">
            This column will be used as the input data for each item
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="expected-column">Expected Output Column (Optional)</Label>
          <Select
            value={columnMapping.expected_column || 'none'}
            onValueChange={handleExpectedColumnChange}
          >
            <SelectTrigger id="expected-column">
              <SelectValue placeholder="Select expected column" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="none">None</SelectItem>
              {availableForExpected.map((header) => (
                <SelectItem key={header} value={header}>
                  {header}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="text-xs text-muted-foreground">
            This column will be used as the expected output/ground truth
          </p>
        </div>

        <div className="space-y-2">
          <Label>Metadata Columns (Optional)</Label>
          <div className="flex flex-wrap gap-2 p-3 border rounded-md min-h-[60px]">
            {availableForMetadata.length > 0 ? (
              availableForMetadata.map((header) => {
                const isSelected = columnMapping.metadata_columns?.includes(header)
                return (
                  <Badge
                    key={header}
                    variant={isSelected ? 'default' : 'outline'}
                    className="cursor-pointer"
                    onClick={() => handleMetadataColumnToggle(header)}
                  >
                    {header}
                    {isSelected && <Check className="ml-1 h-3 w-3" />}
                  </Badge>
                )
              })
            ) : (
              <span className="text-sm text-muted-foreground">
                No columns available for metadata
              </span>
            )}
          </div>
          <p className="text-xs text-muted-foreground">
            Click columns to include as metadata
          </p>
        </div>
      </div>

      {preview && columnMapping.input_column && (
        <div className="space-y-2">
          <Label>Mapped Data Preview</Label>
          <div className="rounded-md border bg-muted/50 p-3 max-h-32 overflow-auto">
            <pre className="text-xs font-mono">
              {JSON.stringify(
                preview.rows.slice(0, 2).map((row) => {
                  const inputIdx = preview.headers.indexOf(columnMapping.input_column)
                  const expectedIdx = columnMapping.expected_column
                    ? preview.headers.indexOf(columnMapping.expected_column)
                    : -1

                  const item: Record<string, unknown> = {
                    input: { value: row[inputIdx] || '' },
                  }

                  if (expectedIdx >= 0) {
                    item.expected = { value: row[expectedIdx] || '' }
                  }

                  if (columnMapping.metadata_columns?.length) {
                    const metadata: Record<string, string> = {}
                    for (const col of columnMapping.metadata_columns) {
                      const idx = preview.headers.indexOf(col)
                      if (idx >= 0) {
                        metadata[col] = row[idx] || ''
                      }
                    }
                    if (Object.keys(metadata).length > 0) {
                      item.metadata = metadata
                    }
                  }

                  return item
                }),
                null,
                2
              )}
            </pre>
          </div>
        </div>
      )}
    </div>
  )

  const renderConfirmStep = () => (
    <div className="space-y-4">
      <Alert>
        <AlertDescription>
          Ready to import <strong>{preview?.rowCount}</strong> items into this dataset.
        </AlertDescription>
      </Alert>

      <div className="space-y-3 text-sm">
        <div className="flex justify-between py-2 border-b">
          <span className="text-muted-foreground">Total Rows</span>
          <span className="font-medium">{preview?.rowCount}</span>
        </div>
        <div className="flex justify-between py-2 border-b">
          <span className="text-muted-foreground">Input Column</span>
          <Badge variant="outline">{columnMapping.input_column}</Badge>
        </div>
        {columnMapping.expected_column && (
          <div className="flex justify-between py-2 border-b">
            <span className="text-muted-foreground">Expected Column</span>
            <Badge variant="outline">{columnMapping.expected_column}</Badge>
          </div>
        )}
        {columnMapping.metadata_columns && columnMapping.metadata_columns.length > 0 && (
          <div className="flex justify-between py-2 border-b">
            <span className="text-muted-foreground">Metadata Columns</span>
            <div className="flex gap-1 flex-wrap justify-end">
              {columnMapping.metadata_columns.map((col) => (
                <Badge key={col} variant="outline">{col}</Badge>
              ))}
            </div>
          </div>
        )}
      </div>

      <div className="flex items-center justify-between rounded-lg border p-3">
        <div className="space-y-0.5">
          <Label htmlFor="deduplicate-confirm">Skip duplicates</Label>
          <p className="text-xs text-muted-foreground">
            Skip items with identical content to existing items
          </p>
        </div>
        <Switch
          id="deduplicate-confirm"
          checked={deduplicate}
          onCheckedChange={setDeduplicate}
        />
      </div>
    </div>
  )

  return (
    <Dialog open={open} onOpenChange={(isOpen) => {
      setOpen(isOpen)
      if (!isOpen) resetForm()
    }}>
      <DialogTrigger asChild>
        {trigger ?? (
          <Button variant="outline">
            <FileSpreadsheet className="mr-2 h-4 w-4" />
            Import CSV
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Import from CSV</DialogTitle>
          <DialogDescription>
            {step === 'upload' && 'Upload a CSV file to import items into this dataset.'}
            {step === 'mapping' && 'Map CSV columns to dataset item fields.'}
            {step === 'confirm' && 'Review and confirm your import.'}
          </DialogDescription>
        </DialogHeader>

        {renderStepIndicator()}

        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <div className="py-2">
          {step === 'upload' && renderUploadStep()}
          {step === 'mapping' && renderMappingStep()}
          {step === 'confirm' && renderConfirmStep()}
        </div>

        <DialogFooter className="flex justify-between">
          <div>
            {step !== 'upload' && (
              <Button variant="ghost" onClick={handleBack} disabled={importMutation.isPending}>
                <ChevronLeft className="mr-1 h-4 w-4" />
                Back
              </Button>
            )}
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => setOpen(false)}
              disabled={importMutation.isPending}
            >
              Cancel
            </Button>
            {step !== 'confirm' ? (
              <Button
                onClick={handleNext}
                disabled={
                  (step === 'upload' && !canProceedToMapping) ||
                  (step === 'mapping' && !canProceedToConfirm)
                }
              >
                Next
                <ChevronRight className="ml-1 h-4 w-4" />
              </Button>
            ) : (
              <Button onClick={handleSubmit} disabled={importMutation.isPending || !canImport}>
                {importMutation.isPending ? 'Importing...' : 'Import Items'}
              </Button>
            )}
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
