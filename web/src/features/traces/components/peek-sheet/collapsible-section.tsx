'use client'

import * as React from 'react'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { Badge } from '@/components/ui/badge'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'

// ============================================================================
// Types
// ============================================================================

interface CollapsibleSectionProps {
  /** Section title */
  title: string
  /** Optional icon to display before the title */
  icon?: React.ReactNode
  /** Optional count badge (e.g., number of items) */
  count?: number
  /** Optional type badge (e.g., "ChatML", "JSON") */
  typeBadge?: string
  /** Whether section is expanded by default */
  defaultExpanded?: boolean
  /** Children content to display when expanded */
  children: React.ReactNode
  /** Additional CSS classes */
  className?: string
  /** Whether to show empty state when children is null/undefined */
  emptyMessage?: string
  /** Disable collapsing (always show content) */
  alwaysOpen?: boolean
  /** Callback when expansion state changes */
  onOpenChange?: (open: boolean) => void
}

// ============================================================================
// CollapsibleSection Component
// ============================================================================

/**
 * CollapsibleSection - Expandable/collapsible section component
 *
 * Features:
 * - Chevron indicator (ChevronRight when collapsed, ChevronDown when expanded)
 * - Title with optional icon prefix
 * - Optional count badge (shows item count)
 * - Optional type badge (shows format type like "ChatML")
 * - Consistent styling for all OTEL data sections
 * - Hover state and smooth transitions
 */
export function CollapsibleSection({
  title,
  icon,
  count,
  typeBadge,
  defaultExpanded = false,
  children,
  className,
  emptyMessage,
  alwaysOpen = false,
  onOpenChange,
}: CollapsibleSectionProps) {
  const [isOpen, setIsOpen] = React.useState(defaultExpanded)

  const handleOpenChange = (open: boolean) => {
    if (!alwaysOpen) {
      setIsOpen(open)
      onOpenChange?.(open)
    }
  }

  const showEmpty = !children && emptyMessage
  const effectiveOpen = alwaysOpen ? true : isOpen

  return (
    <Collapsible
      open={effectiveOpen}
      onOpenChange={handleOpenChange}
      className={cn('border-b border-border/50 last:border-b-0', className)}
    >
      <CollapsibleTrigger
        asChild
        disabled={alwaysOpen}
      >
        <div
          className={cn(
            'flex items-center gap-2 py-2.5 px-1 transition-colors',
            !alwaysOpen && 'cursor-pointer hover:bg-muted/50 rounded-md -mx-1'
          )}
        >
          {/* Chevron indicator */}
          {!alwaysOpen && (
            <div className='flex-shrink-0 w-4'>
              {effectiveOpen ? (
                <ChevronDown className='h-4 w-4 text-muted-foreground' />
              ) : (
                <ChevronRight className='h-4 w-4 text-muted-foreground' />
              )}
            </div>
          )}

          {/* Optional icon */}
          {icon && (
            <div className='flex-shrink-0'>
              {icon}
            </div>
          )}

          {/* Title */}
          <span className='text-sm font-medium text-foreground'>
            {title}
          </span>

          {/* Type badge (e.g., ChatML, JSON) */}
          {typeBadge && (
            <Badge variant='outline' className='text-[10px] px-1.5 py-0 h-4 font-normal'>
              {typeBadge}
            </Badge>
          )}

          {/* Count badge */}
          {typeof count === 'number' && (
            <Badge
              variant='secondary'
              className={cn(
                'text-[10px] px-1.5 py-0 h-4 font-normal ml-auto',
                count === 0 && 'text-muted-foreground'
              )}
            >
              {count === 0 ? 'empty' : count}
            </Badge>
          )}
        </div>
      </CollapsibleTrigger>

      <CollapsibleContent>
        <div className='pb-3 pt-1'>
          {showEmpty ? (
            <div className='py-3 text-center'>
              <p className='text-sm text-muted-foreground italic'>{emptyMessage}</p>
            </div>
          ) : (
            children
          )}
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}

// ============================================================================
// Empty State Component - For consistent empty state display
// ============================================================================

interface EmptyStateProps {
  message: string
  className?: string
}

export function EmptyState({ message, className }: EmptyStateProps) {
  return (
    <div className={cn('py-4 text-center', className)}>
      <p className='text-sm text-muted-foreground italic'>{message}</p>
    </div>
  )
}
