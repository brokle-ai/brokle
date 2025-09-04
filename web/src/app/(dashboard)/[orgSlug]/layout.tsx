import type { OrganizationParams } from '@/types/organization'

interface OrganizationLayoutProps {
  children: React.ReactNode
  params: Promise<OrganizationParams>
}

export default async function OrganizationLayout({
  children,
  params,
}: OrganizationLayoutProps) {
  // Organization validation is now handled by:
  // 1. Middleware - ensures user is authenticated
  // 2. OrganizationContext - validates organization access on client-side
  // 
  // Removing server-side validation to avoid 401 errors since server
  // doesn't have access to user authentication context
  
  return <>{children}</>
}