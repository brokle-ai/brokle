'use client'

import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface VariableBadgeProps {
  name: string
  className?: string
}

export function VariableBadge({ name, className }: VariableBadgeProps) {
  return (
    <Badge
      variant="secondary"
      className={cn(
        'font-mono text-xs bg-amber-100 text-amber-800 border-amber-200',
        'dark:bg-amber-900/30 dark:text-amber-400 dark:border-amber-800',
        className
      )}
    >
      {`{{${name}}}`}
    </Badge>
  )
}

interface VariableListProps {
  variables: string[]
  className?: string
  maxVisible?: number
}

export function VariableList({ variables, className, maxVisible = 5 }: VariableListProps) {
  if (variables.length === 0) {
    return (
      <span className="text-sm text-muted-foreground italic">
        No variables
      </span>
    )
  }

  const visibleVariables = variables.slice(0, maxVisible)
  const hiddenCount = variables.length - maxVisible

  return (
    <div className={cn('flex flex-wrap gap-1', className)}>
      {visibleVariables.map((variable) => (
        <VariableBadge key={variable} name={variable} />
      ))}
      {hiddenCount > 0 && (
        <Badge variant="outline" className="text-xs">
          +{hiddenCount} more
        </Badge>
      )}
    </div>
  )
}
