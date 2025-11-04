'use client'

import { useEffect } from 'react'
import { AuthenticatedLayout } from "@/components/layout/authenticated-layout"
import { WorkspaceProvider } from '@/context/workspace-context'
import { useAuthStore } from '@/stores/auth-store'

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const initializeAuth = useAuthStore(state => state.initializeAuth)
  const isLoading = useAuthStore(state => state.isLoading)

  // Initialize auth on mount
  useEffect(() => {
    initializeAuth()
  }, [initializeAuth])

  // Show loading state while initializing auth
  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto mb-4" />
          <p className="text-sm text-muted-foreground">Loading...</p>
        </div>
      </div>
    )
  }

  return (
    <WorkspaceProvider>
      <AuthenticatedLayout>
        {children}
      </AuthenticatedLayout>
    </WorkspaceProvider>
  )
}
