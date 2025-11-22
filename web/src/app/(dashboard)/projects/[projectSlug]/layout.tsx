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
 */
export default function ProjectLayout({
  children,
}: ProjectLayoutProps) {
  // No server-side validation needed - Go backend validates on every API call
  // Client WorkspaceContext handles UX (error display, loading states)
  return <>{children}</>
}
