'use client'

import { useMemo, useState } from 'react'
import { Check, ChevronsUpDown, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
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
import { useExperimentsQuery } from '../hooks/use-experiments'

interface ExperimentSelectorProps {
  projectId: string
  selectedIds: string[]
  onSelectionChange: (ids: string[]) => void
  minSelections?: number
  maxSelections?: number
  className?: string
}

export function ExperimentSelector({
  projectId,
  selectedIds,
  onSelectionChange,
  minSelections = 2,
  maxSelections = 10,
  className,
}: ExperimentSelectorProps) {
  const [open, setOpen] = useState(false)
  const { data: experiments, isLoading } = useExperimentsQuery(projectId)

  const selectedExperiments = useMemo(() => {
    if (!experiments) return []
    return experiments.filter((exp) => selectedIds.includes(exp.id))
  }, [experiments, selectedIds])

  const toggleExperiment = (id: string) => {
    if (selectedIds.includes(id)) {
      // Prevent going below minimum
      if (selectedIds.length <= minSelections) return
      onSelectionChange(selectedIds.filter((i) => i !== id))
    } else {
      // Prevent exceeding maximum
      if (selectedIds.length >= maxSelections) return
      onSelectionChange([...selectedIds, id])
    }
  }

  const removeExperiment = (id: string) => {
    if (selectedIds.length <= minSelections) return
    onSelectionChange(selectedIds.filter((i) => i !== id))
  }

  return (
    <div className={cn('space-y-2', className)}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="w-full justify-between"
            disabled={isLoading}
          >
            {isLoading
              ? 'Loading experiments...'
              : selectedIds.length === 0
                ? 'Select experiments to compare...'
                : `${selectedIds.length} experiment${selectedIds.length > 1 ? 's' : ''} selected`}
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[400px] p-0" align="start">
          <Command>
            <CommandInput placeholder="Search experiments..." />
            <CommandList>
              <CommandEmpty>No experiments found.</CommandEmpty>
              <CommandGroup>
                {experiments?.map((experiment) => {
                  const isSelected = selectedIds.includes(experiment.id)
                  const canSelect =
                    selectedIds.length < maxSelections || isSelected
                  const canDeselect =
                    selectedIds.length > minSelections || !isSelected

                  return (
                    <CommandItem
                      key={experiment.id}
                      value={experiment.name}
                      onSelect={() => toggleExperiment(experiment.id)}
                      disabled={!canSelect && !isSelected}
                      className={cn(
                        !canDeselect && isSelected && 'opacity-50'
                      )}
                    >
                      <Check
                        className={cn(
                          'mr-2 h-4 w-4',
                          isSelected ? 'opacity-100' : 'opacity-0'
                        )}
                      />
                      <div className="flex flex-col">
                        <span>{experiment.name}</span>
                        <span className="text-xs text-muted-foreground">
                          {experiment.status}
                        </span>
                      </div>
                    </CommandItem>
                  )
                })}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      {selectedExperiments.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {selectedExperiments.map((exp) => (
            <Badge key={exp.id} variant="secondary" className="gap-1">
              {exp.name}
              <button
                type="button"
                onClick={() => removeExperiment(exp.id)}
                className="ml-1 hover:text-destructive disabled:opacity-50"
                disabled={selectedIds.length <= minSelections}
                aria-label={`Remove ${exp.name}`}
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}
    </div>
  )
}
