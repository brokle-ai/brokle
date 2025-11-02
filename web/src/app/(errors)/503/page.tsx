import type { Metadata } from 'next'
import { Construction } from 'lucide-react'
import { ErrorPage } from '@/components/error-page'

export const metadata: Metadata = {
  title: '503 - Service Unavailable | Brokle',
  description: 'Service is temporarily unavailable for maintenance',
}

/**
 * 503 Service Unavailable / Maintenance Page
 * Shown when the service is under maintenance
 * Next.js App Router: app/(errors)/503/page.tsx
 */
export default function ServiceUnavailable() {
  return (
    <ErrorPage
      statusCode={503}
      title="Under Maintenance"
      description="We're currently performing scheduled maintenance to improve your experience. We'll be back shortly. Thank you for your patience."
      icon={Construction}
      showBackButton={false}
      showHomeButton={false}
    />
  )
}
