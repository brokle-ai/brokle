import type { ReactNode } from 'react'

type DataTableEmptyStateProps = {
  icon: ReactNode
  title: string
  description: string
}

/**
 * Empty state component for data tables
 * Used when a project has no data yet (not for filtered-to-zero results)
 *
 * Displays:
 * - Icon (muted color)
 * - Title (semibold)
 * - Description (guides user to header action)
 *
 * No CTA button - action is in PageHeader for consistency
 */
export function DataTableEmptyState({
  icon,
  title,
  description,
}: DataTableEmptyStateProps) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center py-16">
      <div className="flex flex-col items-center text-center max-w-md space-y-4">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-muted">
          <div className="h-8 w-8 text-muted-foreground">
            {icon}
          </div>
        </div>
        <div className="space-y-2">
          <h3 className="text-lg font-semibold">{title}</h3>
          <p className="text-sm text-muted-foreground">{description}</p>
        </div>
      </div>
    </div>
  )
}
