'use client'

import React from 'react'
import { useRouter } from 'next/navigation'
import { FolderOpen, Check, ChevronsUpDown, Plus } from 'lucide-react'
import { cn } from '@/lib/utils'
import { buildProjectUrl } from '@/lib/utils/slug-utils'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { useAuth } from '@/hooks/auth/use-auth'

interface Project {
  id: string
  name: string
  slug: string
  description?: string
}

interface ProjectSwitcherProps {
  currentProject?: Project | null
  projects?: Project[]
}

export function ProjectSwitcher({ currentProject, projects = [] }: ProjectSwitcherProps) {
  const [open, setOpen] = React.useState(false)
  const router = useRouter()
  const { user } = useAuth()

  const handleSelect = (project: any) => {
    if (project.id !== currentProject?.id) {
      const projectUrl = buildProjectUrl(project.name, project.id)
      router.push(projectUrl)
    }
    setOpen(false)
  }

  if (!user) {
    return null
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          aria-label="Select project"
          className="w-[200px] justify-between"
        >
          <div className="flex items-center gap-2">
            <FolderOpen className="h-4 w-4" />
            <span className="truncate">
              {currentProject?.name || 'Select project'}
            </span>
          </div>
          <ChevronsUpDown className="ml-auto h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0">
        <Command>
          <CommandInput placeholder="Search projects..." />
          <CommandList>
            <CommandEmpty>No projects found.</CommandEmpty>
            <CommandGroup heading="Projects">
              {projects.map((project) => (
                <CommandItem
                  key={project.id}
                  value={project.slug}
                  onSelect={() => handleSelect(project)}
                >
                  <FolderOpen className="mr-2 h-4 w-4" />
                  <div className="flex flex-col">
                    <span className="font-medium">{project.name}</span>
                    {project.description && (
                      <span className="text-xs text-muted-foreground">
                        {project.description.substring(0, 50)}
                        {project.description.length > 50 ? '...' : ''}
                      </span>
                    )}
                  </div>
                  <Check
                    className={cn(
                      "ml-auto h-4 w-4",
                      currentProject?.slug === project.slug
                        ? "opacity-100"
                        : "opacity-0"
                    )}
                  />
                </CommandItem>
              ))}
            </CommandGroup>
            <CommandSeparator />
            <CommandGroup>
              <CommandItem
                onSelect={() => {
                  setOpen(false)
                  router.push('/projects')
                }}
              >
                <Plus className="mr-2 h-4 w-4" />
                Create Project
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}