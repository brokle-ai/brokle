'use client'

import * as React from 'react'
import { useState } from 'react'
import { ChevronDown, FolderOpen, Plus, Settings } from 'lucide-react'
import { useOrganization } from '@/context/organization-context'
import { useIsMobile } from '@/hooks/use-mobile'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface ProjectSelectorProps {
  className?: string
}

export function ProjectSelector({ className }: ProjectSelectorProps) {
  const { 
    currentOrganization,
    currentProject,
    projects,
    switchProject,
    switchOrganization,
    isLoading,
  } = useOrganization()
  
  const isMobile = useIsMobile()
  const [isProjectLoading, setIsProjectLoading] = useState(false)

  const handleProjectSwitch = async (projectSlug: string) => {
    if (isProjectLoading || projectSlug === currentProject?.slug) return
    
    try {
      setIsProjectLoading(true)
      await switchProject(projectSlug)
    } catch (error) {
      console.error('Failed to switch project:', error)
    } finally {
      setIsProjectLoading(false)
    }
  }

  const handleGoToOrganization = async () => {
    if (!currentOrganization || isProjectLoading) return
    
    try {
      setIsProjectLoading(true)
      await switchOrganization(currentOrganization.slug)
    } catch (error) {
      console.error('Failed to switch to organization:', error)
    } finally {
      setIsProjectLoading(false)
    }
  }

  // Loading state
  if (isLoading || !currentOrganization || !currentProject) {
    return null // This component only shows when there's a current project
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          className={cn(
            "gap-2 justify-start text-left font-normal",
            isMobile ? "max-w-[140px]" : "max-w-[180px]",
            "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
          )}
          disabled={isProjectLoading}
        >
          <div className="bg-muted flex size-4 items-center justify-center rounded-sm">
            <FolderOpen className="size-3" />
          </div>
          <span className="truncate text-sm">
            {currentProject.name}
          </span>
          <ChevronDown className="ml-auto h-4 w-4 shrink-0" />
        </Button>
      </DropdownMenuTrigger>
      
      <DropdownMenuContent
        className={cn(
          "max-h-96 overflow-y-auto",
          isMobile ? "w-screen max-w-sm" : "w-64"
        )}
        align="start"
        side="bottom"
        sideOffset={4}
      >
        {/* Current Project */}
        <DropdownMenuLabel className="text-xs text-muted-foreground">
          Current Project
        </DropdownMenuLabel>
        <div className="px-2 py-2 border-b mb-1">
          <div className="flex items-center gap-2">
            <div className="bg-muted flex size-6 items-center justify-center rounded-sm">
              <FolderOpen className="size-3" />
            </div>
            <div className="flex-1 min-w-0">
              <span className="font-medium text-sm">{currentProject.name}</span>
              {currentProject.status === 'active' && (
                <div className="flex items-center gap-1 text-xs text-muted-foreground">
                  <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                  <span>Active</span>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Go to Organization Overview */}
        <DropdownMenuItem
          onClick={handleGoToOrganization}
          className="gap-2 p-2 cursor-pointer"
          disabled={isProjectLoading}
        >
          <div className="bg-muted flex size-5 items-center justify-center rounded-sm">
            <FolderOpen className="size-3" />
          </div>
          <span className="text-sm">All Projects</span>
        </DropdownMenuItem>

        {/* Switch Project */}
        {projects.filter(project => project.id !== currentProject.id).length > 0 && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuLabel className="text-xs text-muted-foreground">
              Switch Project
            </DropdownMenuLabel>
            
            {projects.filter(project => project.id !== currentProject.id).map((project) => (
              <DropdownMenuItem
                key={project.id}
                onClick={() => handleProjectSwitch(project.slug)}
                className="gap-2 p-2 cursor-pointer"
                disabled={isProjectLoading}
              >
                <div className="bg-muted flex size-5 items-center justify-center rounded-sm">
                  <FolderOpen className="size-3" />
                </div>
                <span className="text-sm">{project.name}</span>
                {project.status === 'active' && (
                  <div className="w-2 h-2 bg-green-500 rounded-full ml-auto"></div>
                )}
              </DropdownMenuItem>
            ))}
          </>
        )}

        <DropdownMenuSeparator />
        
        {/* Project Actions */}
        <DropdownMenuItem className="gap-2 p-2 cursor-pointer">
          <div className="bg-muted flex size-5 items-center justify-center rounded-sm">
            <Settings className="size-3" />
          </div>
          <span className="text-sm">Project Settings</span>
        </DropdownMenuItem>

        <DropdownMenuItem className="gap-2 p-2 cursor-pointer">
          <div className="bg-primary text-primary-foreground flex size-5 items-center justify-center rounded-sm">
            <Plus className="size-3" />
          </div>
          <span className="text-sm font-medium">New Project</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}