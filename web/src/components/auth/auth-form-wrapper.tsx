'use client'

import { Suspense } from 'react'
import { Loader2 } from 'lucide-react'

function AuthFormFallback() {
  return (
    <div className="flex items-center justify-center p-8">
      <Loader2 className="h-6 w-6 animate-spin" />
    </div>
  )
}

export function AuthFormWrapper({ children }: { children: React.ReactNode }) {
  return (
    <Suspense fallback={<AuthFormFallback />}>
      {children}
    </Suspense>
  )
}