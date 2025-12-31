'use client'

import { ShieldAlert } from 'lucide-react'
import { ErrorPage } from '@/components/error-page'
import { ROUTES } from '@/lib/routes'

/**
 * 401 Unauthorized Error Page
 * Shown when user needs to authenticate to access a resource
 * Next.js App Router: app/(errors)/401/page.tsx
 */
export default function Unauthorized() {
  return (
    <ErrorPage
      statusCode={401}
      title="Authentication Required"
      description="You need to sign in to access this page. Please log in with your credentials to continue."
      icon={ShieldAlert}
      showBackButton={false}
      customAction={{
        label: 'Sign In',
        href: ROUTES.SIGNIN,
      }}
    />
  )
}
