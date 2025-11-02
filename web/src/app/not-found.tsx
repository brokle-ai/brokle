import type { Metadata } from 'next'
import { FileQuestion } from 'lucide-react'
import { ErrorPage } from '@/components/error-page'

export const metadata: Metadata = {
  title: '404 - Page Not Found | Brokle',
  description: 'The page you are looking for does not exist',
}

/**
 * Global 404 Not Found Page
 * Automatically used by Next.js for any unmatched routes
 * Next.js App Router: app/not-found.tsx
 */
export default function NotFound() {
  return (
    <ErrorPage
      statusCode={404}
      title="Page Not Found"
      description="The page you're looking for doesn't exist or might have been moved. Please check the URL or return to the dashboard."
      icon={FileQuestion}
    />
  )
}