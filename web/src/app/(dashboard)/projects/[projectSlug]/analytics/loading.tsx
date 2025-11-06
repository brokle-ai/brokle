import { Skeleton } from '@/components/ui/skeleton'

export default function AnalyticsLoading() {
  return (
    <div className="space-y-6 p-6">
      <Skeleton className="h-8 w-48" />

      <div className="grid gap-4 md:grid-cols-4">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>

      <Skeleton className="h-96 w-full" />
      <Skeleton className="h-64 w-full" />
    </div>
  )
}
