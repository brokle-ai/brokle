'use client'

import { ReactNode } from 'react'
import { useAuth } from '@/hooks/auth/use-auth'
import type { OrganizationRole } from '@/types/auth'

interface RoleGuardProps {
  children: ReactNode
  allowedRoles: OrganizationRole[]
  fallback?: ReactNode
  requireAll?: boolean
}

export function RoleGuard({
  children,
  allowedRoles,
  fallback = null,
  requireAll = false,
}: RoleGuardProps) {
  const { user, organization } = useAuth()

  if (!user || !organization) {
    return <>{fallback}</>
  }

  // Find user's role in the organization
  const userMembership = organization.members.find(member => member.userId === user.id)
  if (!userMembership) {
    return <>{fallback}</>
  }

  const userRole = userMembership.role
  
  // Check permissions
  const hasPermission = requireAll 
    ? allowedRoles.every(role => hasRolePermission(userRole, role))
    : allowedRoles.some(role => hasRolePermission(userRole, role))

  if (!hasPermission) {
    return <>{fallback}</>
  }

  return <>{children}</>
}

// Helper function to check role hierarchy
function hasRolePermission(userRole: OrganizationRole, requiredRole: OrganizationRole): boolean {
  const roleHierarchy: Record<OrganizationRole, number> = {
    viewer: 1,
    developer: 2,
    admin: 3,
    owner: 4,
  }

  return roleHierarchy[userRole] >= roleHierarchy[requiredRole]
}

// Convenience components for specific roles
export function OwnerOnly({ children, fallback }: { children: ReactNode, fallback?: ReactNode }) {
  return (
    <RoleGuard allowedRoles={['owner']} fallback={fallback}>
      {children}
    </RoleGuard>
  )
}

export function AdminOnly({ children, fallback }: { children: ReactNode, fallback?: ReactNode }) {
  return (
    <RoleGuard allowedRoles={['owner', 'admin']} fallback={fallback}>
      {children}
    </RoleGuard>
  )
}

export function DeveloperOnly({ children, fallback }: { children: ReactNode, fallback?: ReactNode }) {
  return (
    <RoleGuard allowedRoles={['owner', 'admin', 'developer']} fallback={fallback}>
      {children}
    </RoleGuard>
  )
}

export function ViewerOnly({ children, fallback }: { children: ReactNode, fallback?: ReactNode }) {
  return (
    <RoleGuard allowedRoles={['owner', 'admin', 'developer', 'viewer']} fallback={fallback}>
      {children}
    </RoleGuard>
  )
}