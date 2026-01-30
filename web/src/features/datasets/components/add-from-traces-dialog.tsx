'use client'

import { useState, useCallback, useEffect, useMemo } from 'react'
import { Database, X, Check, ChevronsUpDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { useImportFromTracesMutation } from '../hooks/use-datasets'
import { useDatasetsQuery } from '../hooks/use-datasets'
import type { Dataset } from '../types'

interface AddFromTracesDialogProps {
  projectId: string
  traceIds: string[]
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AddFromTracesDialog({
  projectId,
  traceIds,
  open,
  onOpenChange,
}: AddFromTracesDialogProps) {
  const [selectedDatasetId, setSelectedDatasetId] = useState<string | null>(null)
  const [deduplicate, setDeduplicate] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [datasetPopoverOpen, setDatasetPopoverOpen] = useState(false)
  const [searchInput, setSearchInput] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchInput)
    }, 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  const { data: datasetsResponse, isLoading: datasetsLoading, isFetching: datasetsFetching } = useDatasetsQuery(projectId, {
    limit: 100,
    search: debouncedSearch || undefined,
  })

  // Cache of selected dataset object (persists across search changes)
  const [selectedDatasetCache, setSelectedDatasetCache] = useState<Dataset | null>(null)

  // Populate cache when selected dataset appears in API response
  useEffect(() => {
    const items = datasetsResponse?.data ?? []

    if (selectedDatasetId) {
      // Try to find selected dataset in current results
      const found = items.find((d: Dataset) => d.id === selectedDatasetId)
      if (found && (!selectedDatasetCache || selectedDatasetCache.id !== found.id)) {
        setSelectedDatasetCache(found)
      }
    } else {
      // Clear cache when nothing is selected
      setSelectedDatasetCache(null)
    }
  }, [datasetsResponse?.data, selectedDatasetId, selectedDatasetCache])

  // Get selected dataset from cache (always available regardless of search)
  const selectedDataset = selectedDatasetCache

  // Merge API results with cached selection if not in current results
  const datasets = useMemo(() => {
    const items = datasetsResponse?.data ?? []
    if (selectedDatasetCache) {
      const selectedInResults = items.find((d: Dataset) => d.id === selectedDatasetCache.id)
      if (!selectedInResults) {
        // Prepend cached selection to results
        return [selectedDatasetCache, ...items]
      }
    }
    return items
  }, [datasetsResponse?.data, selectedDatasetCache])

  const importMutation = useImportFromTracesMutation(projectId, selectedDatasetId ?? '')

  const resetForm = useCallback(() => {
    setSelectedDatasetId(null)
    setSelectedDatasetCache(null)
    setDeduplicate(true)
    setError(null)
  }, [])

  const handleSubmit = async () => {
    if (!selectedDatasetId) {
      setError('Please select a dataset')
      return
    }

    if (traceIds.length === 0) {
      setError('No traces selected')
      return
    }

    setError(null)

    try {
      await importMutation.mutateAsync({
        trace_ids: traceIds,
        deduplicate,
      })
      resetForm()
      onOpenChange(false)
    } catch {
      // Error is handled by mutation
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        onOpenChange(isOpen)
        if (!isOpen) resetForm()
      }}
    >
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Add to Dataset</DialogTitle>
          <DialogDescription>
            Create dataset items from {traceIds.length} selected trace{traceIds.length !== 1 ? 's' : ''}.
            Input/output data will be extracted from the traces.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <Label>Selected Traces</Label>
            <ScrollArea className="h-20 rounded-md border p-2">
              <div className="flex flex-wrap gap-1">
                {traceIds.map((traceId) => (
                  <Badge key={traceId} variant="secondary" className="font-mono text-xs">
                    {traceId.slice(0, 8)}...
                  </Badge>
                ))}
              </div>
            </ScrollArea>
          </div>

          <div className="space-y-2">
            <Label>Target Dataset</Label>
            <Popover
              open={datasetPopoverOpen}
              onOpenChange={(isOpen) => {
                setDatasetPopoverOpen(isOpen)
                if (!isOpen) setSearchInput('')
              }}
            >
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  role="combobox"
                  aria-expanded={datasetPopoverOpen}
                  className="w-full justify-between"
                  disabled={datasetsLoading}
                >
                  {selectedDataset ? selectedDataset.name : 'Select a dataset...'}
                  <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[400px] p-0" align="start">
                <Command shouldFilter={false}>
                  <CommandInput
                    placeholder="Search datasets..."
                    value={searchInput}
                    onValueChange={setSearchInput}
                  />
                  <CommandList>
                    <CommandEmpty>
                      {datasetsLoading || datasetsFetching ? 'Searching...' : 'No datasets found.'}
                    </CommandEmpty>
                    <CommandGroup>
                      {datasets?.map((dataset: Dataset) => (
                        <CommandItem
                          key={dataset.id}
                          value={dataset.name}
                          onSelect={() => {
                            setSelectedDatasetId(dataset.id)
                            setDatasetPopoverOpen(false)
                          }}
                        >
                          <Check
                            className={cn(
                              'mr-2 h-4 w-4',
                              selectedDatasetId === dataset.id
                                ? 'opacity-100'
                                : 'opacity-0'
                            )}
                          />
                          <div className="flex flex-col">
                            <span>{dataset.name}</span>
                            {dataset.description && (
                              <span className="text-xs text-muted-foreground">
                                {dataset.description}
                              </span>
                            )}
                          </div>
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
            <p className="text-xs text-muted-foreground">
              Choose which dataset to add the trace data to
            </p>
          </div>

          <div className="flex items-center justify-between rounded-lg border p-3">
            <div className="space-y-0.5">
              <Label htmlFor="deduplicate-traces">Skip duplicates</Label>
              <p className="text-xs text-muted-foreground">
                Skip traces that already exist in the dataset
              </p>
            </div>
            <Switch
              id="deduplicate-traces"
              checked={deduplicate}
              onCheckedChange={setDeduplicate}
            />
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={importMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={importMutation.isPending || !selectedDatasetId}
          >
            {importMutation.isPending ? 'Adding...' : `Add ${traceIds.length} Trace${traceIds.length !== 1 ? 's' : ''}`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// Simplified version for single trace - can be used in row actions
interface AddTraceToDatasetDialogProps {
  projectId: string
  traceId: string
  traceName?: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AddTraceToDatasetDialog({
  projectId,
  traceId,
  traceName,
  open,
  onOpenChange,
}: AddTraceToDatasetDialogProps) {
  return (
    <AddFromTracesDialog
      projectId={projectId}
      traceIds={[traceId]}
      open={open}
      onOpenChange={onOpenChange}
    />
  )
}
