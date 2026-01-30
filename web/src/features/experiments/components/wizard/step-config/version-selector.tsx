'use client'

import { useState } from 'react'
import { Check, ChevronsUpDown, GitBranch } from 'lucide-react'
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
import { useVersionsQuery } from '@/features/prompts'

export function VersionSelector() {
  const [open, setOpen] = useState(false)
  const { projectId, state, selectVersion } = useExperimentWizard()
  const { configState } = state

  const promptId = configState.selectedPrompt?.id

  const { data: versions, isLoading } = useVersionsQuery(projectId, promptId, {
    enabled: !!promptId,
  })

  const selectedVersion = configState.selectedVersion

  if (!promptId) {
    return (
      <Button variant="outline" className="w-full justify-between" disabled>
        <span className="text-muted-foreground">Select a prompt first...</span>
        <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
      </Button>
    )
  }

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
          {selectedVersion ? (
            <div className="flex items-center gap-2 truncate">
              <GitBranch className="h-4 w-4 shrink-0" />
              <span className="truncate">v{selectedVersion.version}</span>
              {selectedVersion.labels && selectedVersion.labels.length > 0 && (
                <div className="flex gap-1 ml-auto shrink-0">
                  {selectedVersion.labels.slice(0, 2).map((label) => (
                    <Badge key={label} variant="secondary" className="text-xs">
                      {label}
                    </Badge>
                  ))}
                  {selectedVersion.labels.length > 2 && (
                    <Badge variant="outline" className="text-xs">
                      +{selectedVersion.labels.length - 2}
                    </Badge>
                  )}
                </div>
              )}
            </div>
          ) : (
            <span className="text-muted-foreground">Select version...</span>
          )}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[400px] p-0" align="start">
        <Command>
          <CommandInput placeholder="Search versions..." />
          <CommandList>
            <CommandEmpty>No versions found.</CommandEmpty>
            <CommandGroup>
              {versions?.map((version) => (
                <CommandItem
                  key={version.id}
                  value={`v${version.version} ${version.labels?.join(' ') || ''}`}
                  onSelect={() => {
                    selectVersion(version)
                    setOpen(false)
                  }}
                >
                  <Check
                    className={cn(
                      'mr-2 h-4 w-4',
                      selectedVersion?.id === version.id ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                  <div className="flex flex-col flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">v{version.version}</span>
                      {version.labels && version.labels.length > 0 && (
                        <div className="flex gap-1">
                          {version.labels.map((label) => (
                            <Badge key={label} variant="secondary" className="text-xs">
                              {label}
                            </Badge>
                          ))}
                        </div>
                      )}
                    </div>
                    {version.commit_message && (
                      <span className="text-xs text-muted-foreground truncate">
                        {version.commit_message}
                      </span>
                    )}
                  </div>
                  <span className="text-xs text-muted-foreground ml-2 shrink-0">
                    {new Date(version.created_at).toLocaleDateString()}
                  </span>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
