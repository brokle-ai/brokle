'use client'

import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface LabelBadgeProps {
  label: string
  version?: number
  isProtected?: boolean
  className?: string
}

const labelColors: Record<string, string> = {
  production: 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900/30 dark:text-green-400 dark:border-green-800',
  staging: 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900/30 dark:text-yellow-400 dark:border-yellow-800',
  development: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900/30 dark:text-blue-400 dark:border-blue-800',
  latest: 'bg-purple-100 text-purple-800 border-purple-200 dark:bg-purple-900/30 dark:text-purple-400 dark:border-purple-800',
}

export function LabelBadge({ label, version, isProtected, className }: LabelBadgeProps) {
  const colorClass = labelColors[label.toLowerCase()] || 'bg-gray-100 text-gray-800 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700'

  return (
    <Badge
      variant="outline"
      className={cn(
        'font-medium',
        colorClass,
        isProtected && 'ring-1 ring-offset-1 ring-amber-400/50',
        className
      )}
    >
      {label}
      {version !== undefined && (
        <span className="ml-1 opacity-70">v{version}</span>
      )}
      {isProtected && (
        <span className="ml-1" title="Protected label">
          🔒
        </span>
      )}
    </Badge>
  )
}

interface LabelListProps {
  labels: Array<{ name: string; version?: number }>
  protectedLabels?: string[]
  className?: string
  maxVisible?: number
}

export function LabelList({ labels, protectedLabels = [], className, maxVisible = 3 }: LabelListProps) {
  const visibleLabels = labels.slice(0, maxVisible)
  const hiddenCount = labels.length - maxVisible

  return (
    <div className={cn('flex flex-wrap gap-1', className)}>
      {visibleLabels.map((label) => (
        <LabelBadge
          key={label.name}
          label={label.name}
          version={label.version}
          isProtected={protectedLabels.includes(label.name)}
        />
      ))}
      {hiddenCount > 0 && (
        <Badge variant="secondary" className="text-xs">
          +{hiddenCount} more
        </Badge>
      )}
    </div>
  )
}
