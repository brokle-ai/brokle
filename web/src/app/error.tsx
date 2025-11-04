'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  const router = useRouter()

  useEffect(() => {
    // Log error to monitoring service (e.g., Sentry)
    console.error('Application error occurred:', error)
    
    // In production, you would send this to your error tracking service
    // Example: Sentry.captureException(error)
  }, [error])

  return (
    <div className='h-svh'>
      <div className='m-auto flex h-full w-full flex-col items-center justify-center gap-2'>
        <h1 className='text-[7rem] leading-tight font-bold'>500</h1>
        <span className='font-medium'>Oops! Something went wrong {`:')`}</span>
        <p className='text-muted-foreground text-center'>
          We apologize for the inconvenience. <br /> Please try again later.
        </p>
        <div className='mt-6 flex gap-4'>
          <Button variant='outline' onClick={reset}>
            Try Again
          </Button>
          <Button onClick={() => router.push('/')}>Back to Dashboard</Button>
        </div>
      </div>
    </div>
  )
}