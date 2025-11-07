import { Skeleton } from '@/components/ui/skeleton'

export default function OrganizationLoading() {
  return (
    <div className="space-y-6 p-6">
      <Skeleton className="h-10 w-48" />

      <div className="grid gap-4 md:grid-cols-3">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>

      <Skeleton className="h-64 w-full" />
    </div>
  )
}
