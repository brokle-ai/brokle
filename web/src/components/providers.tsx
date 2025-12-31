'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { useState, useEffect } from 'react'
import { NavigationProgress } from "@/components/navigation-progress";
import { Toaster } from "@/components/ui/sonner";
import { ThemeProvider } from "@/context/theme-context";
import { SearchProvider } from "@/context/search-context";
import { DirectionProvider } from "@/context/direction-context";
import { ErrorBoundary } from './error-boundary'
import { useAuthStore } from '@/features/authentication'
import { Loader2 } from 'lucide-react'
import { signinWithStatus } from '@/lib/routes'

interface ClientProvidersProps {
  children: React.ReactNode
}

export function ClientProviders({ children }: ClientProvidersProps) {
  const [isLoggingOut, setIsLoggingOut] = useState(false)
  const [queryClient] = useState(() => new QueryClient({
    defaultOptions: {
      queries: {
        // With SSR, we usually want to set some default stale time
        // above 0 to avoid refetching immediately on the client
        staleTime: 60 * 1000, // 1 minute
        retry: (failureCount, error: any) => {
          // Don't retry on auth errors
          if (error?.statusCode === 401 || error?.statusCode === 403) {
            return false
          }
          return failureCount < 3
        },
      },
      mutations: {
        retry: (failureCount, error: any) => {
          // Don't retry on client errors
          if (error?.statusCode >= 400 && error?.statusCode < 500) {
            return false
          }
          return failureCount < 2
        },
      },
    },
  }))

  // Session expiry event listener
  useEffect(() => {
    const handleSessionExpired = () => {
      // Show overlay (prevents flash during navigation)
      setIsLoggingOut(true)

      // Hard redirect (toast shows on signin page)
      window.location.href = signinWithStatus('expired')
    }

    window.addEventListener('auth:session-expired', handleSessionExpired)

    return () => {
      window.removeEventListener('auth:session-expired', handleSessionExpired)
    }
  }, [])

  // Cross-tab logout sync listener
  useEffect(() => {
    const handleLogoutSignal = (e: StorageEvent) => {
      if (e.key === 'logout_signal' && e.newValue) {
        console.debug('[Auth] Cross-tab logout detected')

        // Show overlay immediately (prevents flash)
        setIsLoggingOut(true)

        // Clear auth state
        useAuthStore.getState().clearAuth()

        // Clear React Query cache
        queryClient.clear()

        // Hard redirect
        window.location.href = signinWithStatus('ended')
      }
    }

    window.addEventListener('storage', handleLogoutSignal)
    return () => window.removeEventListener('storage', handleLogoutSignal)
  }, [queryClient])

  // Logout overlay event listeners
  useEffect(() => {
    const handleLogoutStart = () => setIsLoggingOut(true)
    const handleLogoutEnd = () => setIsLoggingOut(false)  // Clear overlay on error/failure

    window.addEventListener('auth:logout-start', handleLogoutStart)
    window.addEventListener('auth:logout-end', handleLogoutEnd)

    return () => {
      window.removeEventListener('auth:logout-start', handleLogoutStart)
      window.removeEventListener('auth:logout-end', handleLogoutEnd)
    }
  }, [])

  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <DirectionProvider>
          <ThemeProvider>
            <SearchProvider>
              <NavigationProgress />
              <Toaster duration={5000} />

              {/* Logout/Session loading overlay */}
              {isLoggingOut && (
                <div className="fixed inset-0 z-[100] bg-background/80 backdrop-blur-sm flex items-center justify-center">
                  <div className="text-center">
                    <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4 text-primary" />
                    <p className="text-sm text-muted-foreground">Logging out...</p>
                  </div>
                </div>
              )}

              {children}
              <ReactQueryDevtools initialIsOpen={false} />
            </SearchProvider>
          </ThemeProvider>
        </DirectionProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}