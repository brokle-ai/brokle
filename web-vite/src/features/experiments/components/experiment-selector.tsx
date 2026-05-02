import { useMemo, useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
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
import { rawFetch } from '@/lib/api/client'
import {
  experimentListQueryOptions,
  experimentsKeys,
} from '../api/queries'
import type {
  ExperimentListItem,
  ExperimentListResponse,
} from '../api/types'

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
  const [searchInput, setSearchInput] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchInput)
    }, 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  // Paginated dropdown results with optional free-text filter.
  const listQuery = useQuery(
    experimentListQueryOptions(projectId, {
      page: 1,
      limit: 100,
      q: debouncedSearch || undefined,
    }),
  )

  const listRows = listQuery.data?.data ?? []

  // Selected IDs missing from the paginated list response — those that
  // scrolled off the first page / fail the search filter. Fetch them
  // explicitly by ID so the displayed badges always resolve.
  const missingIds = useMemo(() => {
    const known = new Set(listRows.map((e) => e.id))
    return selectedIds.filter((id) => !known.has(id))
  }, [listRows, selectedIds])

  const byIdsQuery = useQuery({
    queryKey: [
      ...experimentsKeys.all,
      'byIds',
      projectId,
      [...missingIds].sort().join(','),
    ] as const,
    queryFn: async () => {
      const qs = new URLSearchParams({
        ids: missingIds.join(','),
        limit: String(Math.max(1, missingIds.length)),
      }).toString()
      const resp = await rawFetch(
        `/api/v1/projects/${projectId}/experiments?${qs}`,
        { method: 'GET' },
      )
      return (await resp.json()) as ExperimentListResponse
    },
    enabled: missingIds.length > 0,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  })

  // Union of both queries — a selected experiment is resolvable iff it
  // appears in either.
  const experimentsMap = useMemo(() => {
    const map = new Map<string, ExperimentListItem>()
    for (const exp of listRows) {
      map.set(exp.id, exp)
    }
    for (const exp of byIdsQuery.data?.data ?? []) {
      map.set(exp.id, exp)
    }
    return map
  }, [listRows, byIdsQuery.data?.data])

  const selectedExperiments = useMemo(() => {
    return selectedIds
      .map((id) => experimentsMap.get(id))
      .filter((exp): exp is ExperimentListItem => exp !== undefined)
  }, [selectedIds, experimentsMap])

  // Dropdown shape: selected-but-not-on-current-page first, then the
  // filtered list results.
  const experiments = useMemo(() => {
    const listIds = new Set(listRows.map((item) => item.id))
    const selectedNotInList = selectedExperiments.filter(
      (exp) => !listIds.has(exp.id),
    )
    return [...selectedNotInList, ...listRows]
  }, [listRows, selectedExperiments])

  const toggleExperiment = (id: string) => {
    if (selectedIds.includes(id)) {
      if (selectedIds.length <= minSelections) return
      onSelectionChange(selectedIds.filter((i) => i !== id))
    } else {
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
      <Popover
        open={open}
        onOpenChange={(isOpen) => {
          setOpen(isOpen)
          if (!isOpen) setSearchInput('')
        }}
      >
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="w-full justify-between"
            disabled={listQuery.isLoading}
          >
            {listQuery.isLoading
              ? 'Loading experiments...'
              : selectedIds.length === 0
                ? 'Select experiments to compare...'
                : `${selectedIds.length} experiment${
                    selectedIds.length > 1 ? 's' : ''
                  } selected`}
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[400px] p-0" align="start">
          <Command shouldFilter={false}>
            <CommandInput
              placeholder="Search experiments..."
              value={searchInput}
              onValueChange={setSearchInput}
            />
            <CommandList>
              <CommandEmpty>
                {listQuery.isLoading || listQuery.isFetching
                  ? 'Searching...'
                  : 'No experiments found.'}
              </CommandEmpty>
              <CommandGroup>
                {experiments.map((experiment) => {
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
                        !canDeselect && isSelected && 'opacity-50',
                      )}
                    >
                      <Check
                        className={cn(
                          'mr-2 h-4 w-4',
                          isSelected ? 'opacity-100' : 'opacity-0',
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
