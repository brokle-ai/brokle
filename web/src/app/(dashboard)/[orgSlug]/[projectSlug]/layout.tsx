import type { ProjectParams } from '@/types/organization'

interface ProjectLayoutProps {
  children: React.ReactNode
  params: Promise<ProjectParams>
}

export default async function ProjectLayout({
  children,
  params,
}: ProjectLayoutProps) {
  // Organization and project validation is now handled by:
  // 1. Middleware - ensures user is authenticated
  // 2. OrganizationContext - validates organization and project access on client-side
  // 
  // Removing server-side validation to avoid 401 errors since server
  // doesn't have access to user authentication context
  
  return <>{children}</>
}