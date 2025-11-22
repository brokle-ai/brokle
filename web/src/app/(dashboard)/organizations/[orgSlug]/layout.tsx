interface OrganizationLayoutProps {
  children: React.ReactNode
  params: Promise<{ orgSlug: string }>
}

/**
 * Organization Layout
 *
 * Simple pass-through layout for organization routes.
 * Security validation handled by Go backend JWT middleware on every API request.
 * Client WorkspaceContext provides UI state and error handling.
 */
export default function OrganizationLayout({
  children,
}: OrganizationLayoutProps) {
  // No server-side validation needed - Go backend validates on every API call
  // Client WorkspaceContext handles UX (error display, loading states)
  return <>{children}</>
}
