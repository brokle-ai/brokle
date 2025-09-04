'use client'

import { useEffect } from 'react'
import { Button } from '@/components/ui/button'

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    // Log error to monitoring service (e.g., Sentry)
    console.error('Global error occurred:', error)
  }, [error])

  return (
    <html>
      <body>
        <div className='h-screen flex items-center justify-center bg-background'>
          <div className='flex flex-col items-center justify-center gap-4 text-center'>
            <h1 className='text-[7rem] leading-tight font-bold text-destructive'>500</h1>
            <div className='space-y-2'>
              <h2 className='text-2xl font-semibold'>Something went wrong</h2>
              <p className='text-muted-foreground max-w-md'>
                We apologize for the inconvenience. Our team has been notified and is working to fix this issue.
              </p>
            </div>
            <div className='flex gap-4 mt-6'>
              <Button variant='outline' onClick={reset}>
                Try Again
              </Button>
              <Button onClick={() => window.location.href = '/'}>
                Go to Dashboard
              </Button>
            </div>
          </div>
        </div>
      </body>
    </html>
  )
}