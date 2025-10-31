import { Skeleton } from '@/components/ui/skeleton'
import { Loader2 } from 'lucide-react'

interface PageLoaderProps {
  message?: string
  type?: 'skeleton' | 'spinner'
}

export function PageLoader({ message = 'Loading...', type = 'skeleton' }: PageLoaderProps) {
  if (type === 'spinner') {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center space-y-4">
          <Loader2 className="h-8 w-8 animate-spin mx-auto text-primary" />
          <p className="text-muted-foreground">{message}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="space-y-4 text-center w-full max-w-6xl px-6">
        <Skeleton className="h-8 w-64 mx-auto" />
        <Skeleton className="h-5 w-96 mx-auto" />
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-48 w-full" />
          ))}
        </div>
      </div>
    </div>
  )
}
