'use client'

import { useState } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { ChevronDown, Loader2, Plus, Settings } from 'lucide-react'
import Link from 'next/link'
import { useWorkspace } from '@/context/workspace-context'
import { WorkspaceError, classifyAPIError } from '@/context/workspace-errors'
import { useIsMobile } from '@/hooks/use-mobile'
import { buildProjectUrl, getProjectSlug } from '@/lib/utils/slug-utils'
import type { ProjectSummary } from '@/features/authentication'
import { getSmartRedirectUrl } from '@/lib/utils/smart-redirect'
import { toast } from 'sonner'
import { CreateProjectDialog } from './create-project-dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

const getStatusBadgeColor = (status: string) => {
  switch (status) {
    case 'archived':
      return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    case 'active':
    default:
      return ''
  }
}

interface ProjectSelectorProps {
  className?: string
  showStatusBadge?: boolean
}

export function ProjectSelector({ className, showStatusBadge = true }: ProjectSelectorProps) {
  const {
    currentOrganization,
    currentProject,
    isInitialized,
    loadingState,
    canInteract,
    switchProject,
  } = useWorkspace()

  // Projects come from current organization
  const projects = currentOrganization?.projects || []

  const router = useRouter()
  const pathname = usePathname()
  const isMobile = useIsMobile()
  const [dialogOpen, setDialogOpen] = useState(false)

  const handleProjectSwitch = async (project: ProjectSummary) => {
    if (project.id === currentProject?.id) return

    try {
      // Use context method (handles loading state internally)
      await switchProject(getProjectSlug(project))

      // Calculate redirect URL
      const redirectUrl = getSmartRedirectUrl({
        currentPath: pathname,
        targetProjectSlug: getProjectSlug(project),
        targetProjectId: project.id,
        targetProjectName: project.name
      })

      router.push(redirectUrl)
    } catch (error) {
      // Preserve already-classified WorkspaceError, only classify if needed
      const workspaceError = error instanceof WorkspaceError
        ? error  // Use as-is to preserve specific error codes/messages
        : classifyAPIError(error)  // Only classify unknown errors

      toast.error(workspaceError.userMessage)

      if (process.env.NODE_ENV === 'development') {
        console.error('[ProjectSelector] Switch failed:', {
          code: workspaceError.code,
          message: workspaceError.message,
          context: workspaceError.context,
          originalError: workspaceError.originalError
        })
      }
    }
  }

  // Loading state - only show when initialized and has current project
  if (!isInitialized || !currentOrganization || !currentProject) {
    return null // This component only shows when there's a current project
  }

  return (
    <>
    <DropdownMenu>
      <DropdownMenuTrigger
        className={cn(
          "flex items-center gap-1 [&_svg]:pointer-events-none [&_svg]:shrink-0",
          "text-sm text-primary hover:text-primary/80 transition-colors",
          !canInteract && "opacity-50 cursor-not-allowed",
          className
        )}
        disabled={!canInteract}
      >
        <span className="font-normal">{currentProject.name}</span>
        {showStatusBadge && currentProject.status === 'archived' && (
          <Badge
            variant="secondary"
            className={cn(
              "ml-1 px-1 py-0 text-xs font-normal capitalize",
              getStatusBadgeColor(currentProject.status)
            )}
          >
            Archived
          </Badge>
        )}
        {loadingState.isSwitchingProject ? (
          <Loader2 className="size-4 animate-spin" />
        ) : (
          <ChevronDown className="size-4" />
        )}
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
            href="/"
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
                className="flex cursor-pointer justify-between items-center"
                onClick={(e) => {
                  if (project.id === currentProject.id) {
                    e.preventDefault()
                    return
                  }
                  // Loading state handled by context
                  handleProjectSwitch(project)
                }}
              >
                <div className="flex items-center gap-2 min-w-0">
                  <span
                    className="overflow-hidden overflow-ellipsis whitespace-nowrap"
                    title={project.name}
                  >
                    {project.name}
                  </span>
                  {showStatusBadge && project.status === 'archived' && (
                    <Badge
                      variant="secondary"
                      className={cn(
                        "px-1 py-0 text-xs font-normal capitalize shrink-0",
                        getStatusBadgeColor(project.status)
                      )}
                    >
                      Archived
                    </Badge>
                  )}
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 hover:bg-background -my-1 ml-4"
                  aria-label={`Open ${project.name} settings`}
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    router.push(buildProjectUrl(project.name, project.id, 'settings'))
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
        <DropdownMenuItem
          onSelect={() => {
            setDialogOpen(true)
          }}
        >
          <Plus className="mr-1.5 h-4 w-4" aria-hidden="true" />
          New Project
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>

    {/* Dialog rendered as sibling, not nested */}
    <CreateProjectDialog
      organizationId={currentOrganization.id}
      open={dialogOpen}
      onOpenChange={setDialogOpen}
    />
  </>
  )
}