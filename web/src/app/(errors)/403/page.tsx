'use client'

import { ShieldOff } from 'lucide-react'
import { ErrorPage } from '@/components/error-page'

/**
 * 403 Forbidden Error Page
 * Shown when user is authenticated but lacks permissions
 * Next.js App Router: app/(errors)/403/page.tsx
 */
export default function Forbidden() {
  return (
    <ErrorPage
      statusCode={403}
      title="Access Denied"
      description="You don't have permission to access this resource. Contact your administrator if you believe this is an error."
      icon={ShieldOff}
      customAction={{
        label: 'Contact Support',
        href: '/support',
      }}
    />
  )
}
