export default function UsersLoading() {
  return (
    <div className="space-y-6">
      {/* Header skeleton */}
      <div className="mb-2 flex flex-wrap items-center justify-between space-y-2">
        <div className="space-y-2">
          <div className="h-8 w-48 bg-muted rounded animate-pulse" />
          <div className="h-4 w-96 bg-muted rounded animate-pulse" />
        </div>
        <div className="h-10 w-32 bg-muted rounded animate-pulse" />
      </div>

      {/* Table skeleton */}
      <div className="space-y-4">
        {/* Toolbar skeleton */}
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="h-9 w-64 bg-muted rounded animate-pulse" />
            <div className="h-9 w-24 bg-muted rounded animate-pulse" />
          </div>
          <div className="h-9 w-32 bg-muted rounded animate-pulse" />
        </div>

        {/* Table skeleton */}
        <div className="rounded-md border">
          <div className="h-12 border-b bg-muted/50" />
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-16 border-b bg-background" />
          ))}
        </div>

        {/* Pagination skeleton */}
        <div className="flex items-center justify-between">
          <div className="h-4 w-32 bg-muted rounded animate-pulse" />
          <div className="flex items-center space-x-2">
            <div className="h-8 w-8 bg-muted rounded animate-pulse" />
            <div className="h-8 w-8 bg-muted rounded animate-pulse" />
            <div className="h-4 w-16 bg-muted rounded animate-pulse" />
            <div className="h-8 w-8 bg-muted rounded animate-pulse" />
            <div className="h-8 w-8 bg-muted rounded animate-pulse" />
          </div>
        </div>
      </div>
    </div>
  )
}