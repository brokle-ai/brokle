'use client'

import { useEffect, useState, useCallback, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { useWorkspace } from '@/context/workspace-context'
import { CreateOrganizationDialog } from '@/features/organizations'
import { CreateProjectDialog } from '@/features/projects'
import { PageLoader } from '@/components/shared/loading'
import { buildProjectUrl } from '@/lib/utils/slug-utils'
import { getLastProjectSlug, clearLastProjectSlug } from '@/lib/utils/project-persistence'
import { extractIdFromCompositeSlug, isValidCompositeSlug } from '@/lib/utils/slug-utils'
import { Button } from '@/components/ui/button'
import { Plus, Building2, FolderPlus } from 'lucide-react'
import { ROUTES } from '@/lib/routes'

export default function RootPage() {
  const router = useRouter()
  const {
    user,
    organizations,
    currentOrganization,
    isLoading,
    isInitialized,
    error,
  } = useWorkspace()

  const [isRedirecting, setIsRedirecting] = useState(false)
  const [orgDialogOpen, setOrgDialogOpen] = useState(false)
  const [projectDialogOpen, setProjectDialogOpen] = useState(false)

  // Use ref to prevent race conditions
  const redirectInitiatedRef = useRef(false)
  // Track previous error state for recovery detection
  const prevErrorRef = useRef<unknown>(null)

  // Check if user has any organization
  const hasOrganizations = organizations.length > 0

  // Check if CURRENT organization has projects (not all orgs - respect org switching!)
  const hasProjects = (currentOrganization?.projects.length ?? 0) > 0

  // Find first available project from CURRENT organization only
  const getFirstProject = useCallback(() => {
    if (!currentOrganization?.projects.length) return null

    return {
      project: currentOrganization.projects[0],
      organization: currentOrganization,
    }
  }, [currentOrganization])

  const redirectToProject = useCallback(() => {
    if (redirectInitiatedRef.current) return
    if (!currentOrganization) return // Wait for organization to be set

    redirectInitiatedRef.current = true
    setIsRedirecting(true)

    // Step 1: Check for last visited project in localStorage (only within current org)
    const lastProjectSlug = getLastProjectSlug()
    if (lastProjectSlug && isValidCompositeSlug(lastProjectSlug)) {
      try {
        const projectId = extractIdFromCompositeSlug(lastProjectSlug)

        // Only check current organization's projects
        const project = currentOrganization.projects.find(p => p.id === projectId)
        if (project) {
          // Found in current org - redirect to last project
          const url = buildProjectUrl(project.name, project.id)
          router.push(url)
          return
        }

        // Project not in current org - don't clear, might be valid for other org
      } catch {
        // Invalid slug format - clear it
        clearLastProjectSlug()
      }
    }

    // Step 2: Redirect to first available project in current org
    const firstProject = getFirstProject()
    if (firstProject) {
      const url = buildProjectUrl(
        firstProject.project.name,
        firstProject.project.id
      )
      router.push(url)
      return
    }

    // Step 3: No projects in current org - reset redirect flag to show empty state
    redirectInitiatedRef.current = false
    setIsRedirecting(false)
  }, [currentOrganization, getFirstProject, router])

  useEffect(() => {
    if (isLoading || !isInitialized) return

    if (!user) {
      router.push(ROUTES.SIGNIN)
      return
    }

    // Handle workspace fetch error - skip redirect logic (will show error state)
    // Note: Do NOT set redirectInitiatedRef here - it blocks redirects after error recovery
    if (error) {
      return
    }

    // Trigger redirect once data is loaded
    if (!isRedirecting && hasProjects) {
      // Use queueMicrotask to avoid synchronous setState in effect
      queueMicrotask(redirectToProject)
    }
  }, [
    isLoading,
    isInitialized,
    error,
    user,
    hasProjects,
    router,
    redirectToProject,
    isRedirecting,
  ])

  // Reset redirect flag when error clears (recovery scenario)
  useEffect(() => {
    if (prevErrorRef.current && !error) {
      // Error just cleared - reset redirect flag to allow redirect to proceed
      redirectInitiatedRef.current = false
      // Use queueMicrotask to avoid synchronous setState in effect
      queueMicrotask(() => setIsRedirecting(false))
    }
    prevErrorRef.current = error
  }, [error])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      redirectInitiatedRef.current = false
    }
  }, [])

  if (isLoading || !isInitialized) {
    return <PageLoader message="Loading your workspace..." />
  }

  if (!user) {
    return null // Will redirect to signin
  }

  // Redirecting to project
  if (hasProjects && isRedirecting) {
    return <PageLoader message="Loading your workspace..." />
  }

  // No organization or error - show create organization empty state
  if (!hasOrganizations || error) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="flex flex-col items-center justify-center text-center max-w-md">
          <div className="mb-6 rounded-full bg-muted p-6">
            <Building2 className="h-16 w-16 text-muted-foreground" />
          </div>
          <h1 className="text-2xl font-bold mb-3">Welcome to Brokle</h1>
          <p className="text-muted-foreground mb-8">
            Create your first organization to start managing AI projects and
            team members.
          </p>

          <Button size="lg" onClick={() => setOrgDialogOpen(true)}>
            <Plus className="mr-2 h-5 w-5" />
            Create Your First Organization
          </Button>

          <CreateOrganizationDialog
            open={orgDialogOpen}
            onOpenChange={setOrgDialogOpen}
          />
        </div>
      </div>
    )
  }

  // Has organization but no projects - show create project empty state
  if (!hasProjects) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="flex flex-col items-center justify-center text-center max-w-md">
          <div className="mb-6 rounded-full bg-muted p-6">
            <FolderPlus className="h-16 w-16 text-muted-foreground" />
          </div>
          <h1 className="text-2xl font-bold mb-3">Create Your First Project</h1>
          <p className="text-muted-foreground mb-8">
            Projects help you organize your AI applications, track usage, and
            manage API keys.
          </p>

          <Button size="lg" onClick={() => setProjectDialogOpen(true)}>
            <Plus className="mr-2 h-5 w-5" />
            Create Your First Project
          </Button>

          {currentOrganization && (
            <CreateProjectDialog
              organizationId={currentOrganization.id}
              open={projectDialogOpen}
              onOpenChange={setProjectDialogOpen}
            />
          )}
        </div>
      </div>
    )
  }

  return <PageLoader message="Loading your workspace..." />
}
