import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

type PromptStatus = 'draft' | 'published' | 'archived'

interface PromptStatusBadgeProps {
  status: PromptStatus
  className?: string
}

export function PromptStatusBadge({ status, className }: PromptStatusBadgeProps) {
  const variants = {
    draft: {
      label: 'Draft',
      className: 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300',
    },
    published: {
      label: 'Published',
      className: 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300',
    },
    archived: {
      label: 'Archived',
      className: 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300',
    },
  }

  const variant = variants[status]

  return (
    <Badge variant="outline" className={cn(variant.className, className)}>
      {variant.label}
    </Badge>
  )
}
