'use client'

import { useEffect } from 'react'
import { useParams } from 'next/navigation'
import { setLastProjectSlug } from '@/lib/utils/project-persistence'
import { isValidCompositeSlug } from '@/lib/utils/slug-utils'

interface ProjectLayoutProps {
  children: React.ReactNode
  params: Promise<{ projectSlug: string }>
}

/**
 * Project Layout
 *
 * Simple pass-through layout for project routes.
 * Security validation handled by Go backend JWT middleware on every API request.
 * Client WorkspaceContext provides UI state and error handling.
 *
 * Also tracks last visited project for smart redirect on next login.
 */
export default function ProjectLayout({
  children,
}: ProjectLayoutProps) {
  const params = useParams()
  const projectSlug = params?.projectSlug as string

  // Track last visited project for smart redirect
  useEffect(() => {
    if (projectSlug && isValidCompositeSlug(projectSlug)) {
      setLastProjectSlug(projectSlug)
    }
  }, [projectSlug])

  // No server-side validation needed - Go backend validates on every API call
  // Client WorkspaceContext handles UX (error display, loading states)
  return <>{children}</>
}
