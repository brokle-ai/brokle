import type { Metadata } from 'next'
import { ServerCrash } from 'lucide-react'
import { ErrorPage } from '@/components/error-page'

export const metadata: Metadata = {
  title: '500 - Internal Server Error | Brokle',
  description: 'An internal server error occurred',
}

/**
 * 500 Internal Server Error Page
 * Shown when an unexpected server error occurs
 * Next.js App Router: app/(errors)/500/page.tsx
 */
export default function InternalServerError() {
  return (
    <ErrorPage
      statusCode={500}
      title="Internal Server Error"
      description="Something went wrong on our end. Our team has been notified and is working to fix the issue. Please try again later."
      icon={ServerCrash}
      showBackButton={false}
    />
  )
}
