'use client'

import { useState, useCallback } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { ChevronDown, Settings, Plus, Loader2 } from 'lucide-react'
import Link from 'next/link'
import { toast } from 'sonner'
import { useWorkspace } from '@/context/workspace-context'
import { WorkspaceError, classifyAPIError } from '@/context/workspace-errors'
import { useIsMobile } from '@/hooks/use-mobile'
import { getSmartRedirectUrl } from '@/lib/utils/smart-redirect'
import { generateCompositeSlug, extractIdFromCompositeSlug, buildProjectUrl } from '@/lib/utils/slug-utils'
import { CreateOrganizationDialog } from '@/features/organizations'
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

interface OrganizationSelectorProps {
  className?: string
  showPlanBadge?: boolean
}

// Pure function for plan badge colors - moved to module scope for performance
const getPlanBadgeColor = (plan: string) => {
  switch (plan) {
    case 'enterprise':
      return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300'
    case 'business':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
    case 'pro':
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
    case 'free':
      return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
  }
}

export function OrganizationSelector({ className, showPlanBadge = false }: OrganizationSelectorProps) {
  const {
    organizations,
    currentOrganization,
    isInitialized,
    loadingState,
    canInteract,
    switchOrganization,
  } = useWorkspace()

  const pathname = usePathname()
  const router = useRouter()
  const isMobile = useIsMobile()
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)

  const handleOrgSwitch = useCallback(async (compositeSlug: string) => {
    // Generate current composite slug for comparison
    const currentCompositeSlug = currentOrganization
      ? generateCompositeSlug(currentOrganization.name, currentOrganization.id)
      : null

    if (compositeSlug === currentCompositeSlug) return

    // Extract ID to find target org
    const targetOrgId = extractIdFromCompositeSlug(compositeSlug)
    const targetOrg = organizations.find(org => org.id === targetOrgId)
    if (!targetOrg) {
      if (process.env.NODE_ENV === 'development') {
        console.error('[OrganizationSelector] Organization not found:', compositeSlug)
      }
      return
    }

    try {
      // Use context method (handles loading state internally)
      // Always complete the switch first (PostHog pattern)
      await switchOrganization(compositeSlug)

      // Determine redirect URL based on whether target org has projects
      const targetProject = targetOrg.projects[0]

      let redirectUrl: string
      if (targetProject) {
        // Target org has projects - use smart redirect to preserve page context
        redirectUrl = getSmartRedirectUrl({
          currentPath: pathname,
          targetProjectSlug: generateCompositeSlug(targetProject.name, targetProject.id),
          targetProjectId: targetProject.id,
          targetProjectName: targetProject.name
        })
      } else {
        // Target org has NO projects - redirect to root (PostHog pattern)
        // Root page will show "Create Your First Project" empty state
        redirectUrl = '/'
      }

      // Navigate
      router.push(redirectUrl)
    } catch (error) {
      // Preserve already-classified WorkspaceError, only classify if needed
      const workspaceError = error instanceof WorkspaceError
        ? error  // Use as-is to preserve specific error codes/messages
        : classifyAPIError(error)  // Only classify unknown errors

      toast.error(workspaceError.userMessage)

      if (process.env.NODE_ENV === 'development') {
        console.error('[OrganizationSelector] Switch failed:', {
          code: workspaceError.code,
          message: workspaceError.message,
          context: workspaceError.context,
          originalError: workspaceError.originalError
        })
      }
    }
  }, [currentOrganization, organizations, pathname, router, switchOrganization])

  // Loading state - show shimmer only if not initialized yet
  if (!isInitialized) {
    return (
      <div className={cn("animate-pulse bg-muted rounded h-6 w-32", className)}></div>
    )
  }

  // Show nothing if no current organization after initialization
  if (!currentOrganization) {
    return null
  }

  return (
    <>
    <DropdownMenu open={dropdownOpen} onOpenChange={setDropdownOpen}>
      <DropdownMenuTrigger
        className={cn(
          "flex items-center gap-1 [&_svg]:pointer-events-none [&_svg]:shrink-0",
          "text-sm text-primary hover:text-primary/80 transition-colors",
          !canInteract && "opacity-50 cursor-not-allowed",
          className
        )}
        disabled={!canInteract}
      >
        <span className="font-normal">{currentOrganization.name}</span>
        {showPlanBadge && (
          <Badge
            variant="secondary"
            className={cn(
              "ml-1 px-1 py-0 text-xs font-normal capitalize",
              getPlanBadgeColor(currentOrganization.plan)
            )}
          >
            {currentOrganization.plan}
          </Badge>
        )}
        {loadingState.isSwitchingOrg ? (
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
        {/* Organizations overview link */}
        <DropdownMenuItem className="font-semibold" asChild>
          <Link href="/" className="cursor-pointer">
            Organizations
          </Link>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        {/* All organizations list */}
        <div className="max-h-36 overflow-y-auto">
          {organizations && organizations.map((org) => {
            // Since org pages no longer exist, link to first project in org
            const firstProject = org.projects[0]
            const targetUrl = org.id === currentOrganization.id
              ? pathname
              : firstProject
                ? getSmartRedirectUrl({
                    currentPath: pathname,
                    targetProjectSlug: generateCompositeSlug(firstProject.name, firstProject.id),
                    targetProjectId: firstProject.id,
                    targetProjectName: firstProject.name
                  })
                : '/' // Fallback to home if no projects

            return (
            <DropdownMenuItem
              key={org.id}
              asChild
            >
              <Link
                href={targetUrl}
                className="flex cursor-pointer justify-between"
                onClick={(e) => {
                  if (org.id === currentOrganization.id) {
                    e.preventDefault()
                    return
                  }
                  // Allow "open in new tab" behavior for modified clicks
                  if (e.metaKey || e.ctrlKey || e.button === 1) {
                    return
                  }
                  // Prevent default link navigation and dropdown auto-close
                  e.preventDefault()
                  e.stopPropagation()
                  const compositeSlug = generateCompositeSlug(org.name, org.id)
                  handleOrgSwitch(compositeSlug)
                }}
              >
                <span
                  className="max-w-36 overflow-hidden overflow-ellipsis whitespace-nowrap"
                  title={org.name}
                >
                  {org.name}
                </span>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 hover:bg-background -my-1 ml-4"
                  aria-label={`Open ${org.name} settings`}
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    // Navigate to TARGET org's settings via its first project
                    const targetProject = org.projects[0]
                    if (targetProject) {
                      router.push(buildProjectUrl(targetProject.name, targetProject.id, 'settings/organization'))
                    } else {
                      // Target org has no projects - switch to it first
                      // Root page will show "Create Project" empty state
                      const compositeSlug = generateCompositeSlug(org.name, org.id)
                      handleOrgSwitch(compositeSlug)
                    }
                  }}
                >
                  <Settings className="h-3 w-3" />
                  <span className="sr-only">Open {org.name} settings</span>
                </Button>
              </Link>
            </DropdownMenuItem>
            )
          })}
        </div>

        <DropdownMenuSeparator />

        {/* Create new organization */}
        <DropdownMenuItem
          onSelect={() => {
            setDialogOpen(true)
          }}
        >
          <Plus className="mr-1.5 h-4 w-4" aria-hidden="true" />
          New Organization
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>

    {/* Dialog rendered as sibling, not nested */}
    <CreateOrganizationDialog
      open={dialogOpen}
      onOpenChange={setDialogOpen}
    />
  </>
  )
}