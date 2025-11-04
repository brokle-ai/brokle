'use client'

import * as React from 'react'
import { useState } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { ChevronDown, FolderOpen, Plus, Settings } from 'lucide-react'
import Link from 'next/link'
import { useWorkspace } from '@/context/workspace-context'
import { useProjectOnly } from '@/hooks/use-project-only'
import { useIsMobile } from '@/hooks/use-mobile'
import { buildProjectUrl, getProjectSlug } from '@/lib/utils/slug-utils'
import { getSmartRedirectUrl } from '@/lib/utils/smart-redirect'
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
    isInitialized,
  } = useWorkspace()

  // Projects come from current organization
  const projects = currentOrganization?.projects || []
  
  const router = useRouter()
  const pathname = usePathname()
  const isMobile = useIsMobile()
  const [isSwitchLoading, setIsSwitchLoading] = useState(false)

  const handleProjectSwitch = async (project: any) => {
    if (isSwitchLoading || project.id === currentProject?.id) return
    
    try {
      setIsSwitchLoading(true)
      
      // Use smart redirect to determine the appropriate URL
      const redirectUrl = getSmartRedirectUrl({
        currentPath: pathname,
        targetProjectSlug: getProjectSlug(project),
        targetProjectId: project.id,
        targetProjectName: project.name
      })
      
      router.push(redirectUrl)
    } catch (error) {
      console.error('Failed to switch project:', error)
    } finally {
      setIsSwitchLoading(false)
    }
  }

  const handleGoToOrganization = async () => {
    if (!currentOrganization || isSwitchLoading) return
    
    try {
      setIsSwitchLoading(true)
      
      // Use smart redirect to determine the appropriate URL (cross-context switch)
      const redirectUrl = getSmartRedirectUrl({
        currentPath: pathname,
        targetOrgSlug: currentOrganization.slug,
        targetOrgId: currentOrganization.id,
        targetOrgName: currentOrganization.name
      })
      
      router.push(redirectUrl)
    } catch (error) {
      console.error('Failed to navigate to organization:', error)
    } finally {
      setIsSwitchLoading(false)
    }
  }

  // Loading state - only show when initialized and has current project
  if (!isInitialized || !currentOrganization || !currentProject) {
    return null // This component only shows when there's a current project
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        className={cn(
          "flex items-center gap-1 [&_svg]:pointer-events-none [&_svg]:shrink-0",
          "text-sm text-primary hover:text-primary/80 transition-colors",
          isSwitchLoading && "opacity-50 cursor-not-allowed",
          className
        )}
        disabled={isSwitchLoading}
      >
        <span className="font-normal">{currentProject.name}</span>
        <ChevronDown className="size-4" />
      </DropdownMenuTrigger>
      
      <DropdownMenuContent
        className={cn(
          "max-h-96 overflow-y-auto",
          isMobile ? "w-screen max-w-sm" : "w-64"
        )}
        align="start"
      >
        {/* Projects overview link */}
        <DropdownMenuItem className="font-semibold" asChild>
          <Link
            href={`/organization/${currentOrganization.id}`}
            className="cursor-pointer"
          >
            Projects
          </Link>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        {/* All projects list */}
        <div className="max-h-36 overflow-y-auto">
          {projects.map((project) => (
            <DropdownMenuItem
              key={project.id}
              asChild
            >
              <Link
                href={
                  project.id === currentProject.id
                    ? pathname
                    : getSmartRedirectUrl({
                        currentPath: pathname,
                        targetProjectSlug: getProjectSlug(project),
                        targetProjectId: project.id,
                        targetProjectName: project.name
                      })
                }
                className="flex cursor-pointer justify-between"
                onClick={(e) => {
                  if (project.id === currentProject.id) {
                    e.preventDefault()
                    return
                  }
                  setIsSwitchLoading(true)
                }}
              >
                <span
                  className="max-w-36 overflow-hidden overflow-ellipsis whitespace-nowrap"
                  title={project.name}
                >
                  {project.name}
                </span>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 hover:bg-background -my-1 ml-4"
                  aria-label={`Open ${project.name} settings`}
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    router.push(`/project/${project.id}/settings`)
                  }}
                >
                  <Settings className="h-3 w-3" />
                  <span className="sr-only">Open {project.name} settings</span>
                </Button>
              </Link>
            </DropdownMenuItem>
          ))}
        </div>

        <DropdownMenuSeparator />

        {/* Create new project */}
        <DropdownMenuItem asChild>
          <Button
            variant="ghost"
            size="sm"
            className="h-8 w-full text-sm font-normal justify-start"
            asChild
          >
            <Link href="/projects/create">
              <Plus className="mr-1.5 h-4 w-4" aria-hidden="true" />
              New Project
            </Link>
          </Button>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}