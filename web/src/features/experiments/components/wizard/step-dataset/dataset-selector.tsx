'use client'

import { useState } from 'react'
import { Check, ChevronsUpDown, Database } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
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
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { useExperimentWizard } from '../../../context/experiment-wizard-context'
import { useDatasetsQuery } from '@/features/datasets'

export function DatasetSelector() {
  const [open, setOpen] = useState(false)
  const { projectId, state, selectDataset } = useExperimentWizard()
  const { datasetState } = state

  const { data: datasetsResponse, isLoading } = useDatasetsQuery(projectId)
  const datasets = datasetsResponse?.datasets ?? []

  const selectedDataset = datasetState.selectedDataset

  if (isLoading) {
    return <Skeleton className="h-10 w-full" />
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-full justify-between"
        >
          {selectedDataset ? (
            <div className="flex items-center gap-2 truncate">
              <Database className="h-4 w-4 shrink-0" />
              <span className="truncate">{selectedDataset.name}</span>
              {selectedDataset.item_count !== undefined && (
                <Badge variant="outline" className="ml-auto shrink-0">
                  {selectedDataset.item_count} items
                </Badge>
              )}
            </div>
          ) : (
            <span className="text-muted-foreground">Select a dataset...</span>
          )}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[400px] p-0" align="start">
        <Command>
          <CommandInput placeholder="Search datasets..." />
          <CommandList>
            <CommandEmpty>No datasets found.</CommandEmpty>
            <CommandGroup>
              {datasets?.map((dataset) => (
                <CommandItem
                  key={dataset.id}
                  value={dataset.name}
                  onSelect={() => {
                    selectDataset({
                      id: dataset.id,
                      name: dataset.name,
                      description: dataset.description,
                      item_count: 'item_count' in dataset ? (dataset as { item_count?: number }).item_count : undefined,
                      created_at: dataset.created_at,
                      updated_at: dataset.updated_at,
                    })
                    setOpen(false)
                  }}
                >
                  <Check
                    className={cn(
                      'mr-2 h-4 w-4',
                      selectedDataset?.id === dataset.id ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                  <div className="flex flex-col flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="truncate">{dataset.name}</span>
                    </div>
                    {dataset.description && (
                      <span className="text-xs text-muted-foreground truncate">
                        {dataset.description}
                      </span>
                    )}
                  </div>
                  {'item_count' in dataset && (
                    <span className="text-xs text-muted-foreground ml-2 shrink-0">
                      {(dataset as { item_count?: number }).item_count ?? 0} items
                    </span>
                  )}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
