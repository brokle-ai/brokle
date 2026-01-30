'use client'

import { useState } from 'react'
import { Check, ChevronsUpDown, FileText } from 'lucide-react'
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
import { usePromptsQuery } from '@/features/prompts'

export function PromptSelector() {
  const [open, setOpen] = useState(false)
  const { projectId, state, selectPrompt } = useExperimentWizard()
  const { configState } = state

  const { data: prompts, isLoading } = usePromptsQuery(projectId)

  const selectedPrompt = configState.selectedPrompt

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
          {selectedPrompt ? (
            <div className="flex items-center gap-2 truncate">
              <FileText className="h-4 w-4 shrink-0" />
              <span className="truncate">{selectedPrompt.name}</span>
              <Badge variant="outline" className="ml-auto shrink-0">
                {selectedPrompt.type}
              </Badge>
            </div>
          ) : (
            <span className="text-muted-foreground">Select a prompt...</span>
          )}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[400px] p-0" align="start">
        <Command>
          <CommandInput placeholder="Search prompts..." />
          <CommandList>
            <CommandEmpty>No prompts found.</CommandEmpty>
            <CommandGroup>
              {prompts?.prompts?.map((prompt) => (
                <CommandItem
                  key={prompt.id}
                  value={prompt.name}
                  onSelect={() => {
                    selectPrompt(prompt)
                    setOpen(false)
                  }}
                >
                  <Check
                    className={cn(
                      'mr-2 h-4 w-4',
                      selectedPrompt?.id === prompt.id ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                  <div className="flex flex-col flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="truncate">{prompt.name}</span>
                      <Badge variant="outline" className="shrink-0">
                        {prompt.type}
                      </Badge>
                    </div>
                    {prompt.description && (
                      <span className="text-xs text-muted-foreground truncate">
                        {prompt.description}
                      </span>
                    )}
                  </div>
                  <span className="text-xs text-muted-foreground ml-2 shrink-0">
                    v{prompt.latest_version}
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
