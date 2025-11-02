'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { NavigationProgress } from "@/components/navigation-progress";
import { Toaster } from "@/components/ui/sonner";
import { toast } from "sonner";
import { ThemeProvider } from "@/context/theme-context";
import { SearchProvider } from "@/context/search-context";
import { DirectionProvider } from "@/context/direction-context";
import { ErrorBoundary } from './error-boundary'

interface ClientProvidersProps {
  children: React.ReactNode
}

export function ClientProviders({ children }: ClientProvidersProps) {
  const router = useRouter()
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
      toast.error('Session expired. Please log in again.')
      router.push('/auth/signin?session=expired')
    }

    window.addEventListener('auth:session-expired', handleSessionExpired)

    return () => {
      window.removeEventListener('auth:session-expired', handleSessionExpired)
    }
  }, [router])

  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <DirectionProvider>
          <ThemeProvider>
            <SearchProvider>
              <NavigationProgress />
              <Toaster duration={5000} />
              {children}
              <ReactQueryDevtools initialIsOpen={false} />
            </SearchProvider>
          </ThemeProvider>
        </DirectionProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}