'use client'

import * as React from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  Hash,
  List,
  ToggleLeft,
  Trash2,
  Loader2,
  ChevronDown,
  ChevronUp,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { formatDistanceToNow } from 'date-fns'
import type { Annotation } from '../api/scores-api'

interface AnnotationItemProps {
  annotation: Annotation
  currentUserId?: string
  onDelete: (scoreId: string) => void
  isDeleting?: boolean
  canDelete?: boolean
}

/**
 * AnnotationItem - Display a single annotation score
 *
 * Features:
 * - Data type icon (numeric/categorical/boolean)
 * - Value badge with type-based coloring
 * - Expandable reason/explanation
 * - Source badge for automated scores
 * - Relative timestamp
 * - Delete button (owner only, hover reveal)
 */
export function AnnotationItem({
  annotation,
  currentUserId,
  onDelete,
  isDeleting = false,
  canDelete = true,
}: AnnotationItemProps) {
  const [isExpanded, setIsExpanded] = React.useState(false)
  const [isHovered, setIsHovered] = React.useState(false)

  const isOwner = annotation.created_by === currentUserId
  const showDelete = canDelete && isOwner && annotation.source === 'annotation'
  const hasLongReason = (annotation.reason?.length ?? 0) > 100

  // Format the value based on data type
  const formattedValue = React.useMemo(() => {
    switch (annotation.type) {
      case 'NUMERIC':
        return annotation.value?.toString() ?? '-'
      case 'CATEGORICAL':
        return annotation.string_value ?? '-'
      case 'BOOLEAN':
        return annotation.value === 1 ? 'Yes' : annotation.value === 0 ? 'No' : '-'
      default:
        return annotation.value?.toString() ?? annotation.string_value ?? '-'
    }
  }, [annotation])

  // Get data type icon
  const DataTypeIcon = React.useMemo(() => {
    switch (annotation.type) {
      case 'NUMERIC':
        return Hash
      case 'CATEGORICAL':
        return List
      case 'BOOLEAN':
        return ToggleLeft
      default:
        return Hash
    }
  }, [annotation.type])

  // Get value badge variant based on data type and value
  const valueBadgeVariant = React.useMemo(() => {
    if (annotation.type === 'BOOLEAN') {
      return annotation.value === 1 ? 'default' : 'destructive'
    }
    return 'secondary'
  }, [annotation.type, annotation.value])

  // Format source label
  const sourceLabel = React.useMemo(() => {
    switch (annotation.source) {
      case 'annotation':
        return 'Human'
      case 'api':
        return 'SDK'
      case 'eval':
        return 'Evaluation'
      default:
        return annotation.source
    }
  }, [annotation.source])

  const handleDelete = () => {
    if (!isDeleting) {
      onDelete(annotation.id)
    }
  }

  return (
    <div
      className={cn(
        'group relative rounded-md border p-3 transition-colors',
        isHovered && showDelete && 'bg-muted/50'
      )}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Main content row */}
      <div className='flex items-start justify-between gap-2'>
        <div className='flex-1 min-w-0'>
          {/* Name and type */}
          <div className='flex items-center gap-2'>
            <DataTypeIcon className='h-3.5 w-3.5 text-muted-foreground shrink-0' />
            <span className='font-medium text-sm truncate'>{annotation.name}</span>
            {annotation.source !== 'annotation' && (
              <Badge variant='outline' className='text-[10px] px-1.5 py-0'>
                {sourceLabel}
              </Badge>
            )}
          </div>

          {/* Timestamp */}
          <p className='text-xs text-muted-foreground mt-0.5'>
            {formatDistanceToNow(new Date(annotation.timestamp), { addSuffix: true })}
          </p>
        </div>

        {/* Value badge */}
        <div className='flex items-center gap-2'>
          <Badge
            variant={valueBadgeVariant}
            className={cn(
              'font-mono text-xs',
              annotation.type === 'NUMERIC' && 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
              annotation.type === 'CATEGORICAL' && 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200'
            )}
          >
            {formattedValue}
          </Badge>

          {/* Delete button - shown on hover for owners */}
          {showDelete && (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant='ghost'
                    size='icon'
                    className={cn(
                      'h-6 w-6 opacity-0 transition-opacity',
                      isHovered && 'opacity-100'
                    )}
                    onClick={handleDelete}
                    disabled={isDeleting}
                  >
                    {isDeleting ? (
                      <Loader2 className='h-3.5 w-3.5 animate-spin' />
                    ) : (
                      <Trash2 className='h-3.5 w-3.5 text-muted-foreground hover:text-destructive' />
                    )}
                    <span className='sr-only'>Delete annotation</span>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Delete annotation</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )}
        </div>
      </div>

      {/* Reason/explanation */}
      {annotation.reason && (
        <div className='mt-2'>
          <p
            className={cn(
              'text-xs text-muted-foreground',
              !isExpanded && hasLongReason && 'line-clamp-2'
            )}
          >
            {annotation.reason}
          </p>
          {hasLongReason && (
            <Button
              variant='ghost'
              size='sm'
              className='h-6 px-1 mt-1 text-xs'
              onClick={() => setIsExpanded(!isExpanded)}
            >
              {isExpanded ? (
                <>
                  <ChevronUp className='h-3 w-3 mr-1' />
                  Show less
                </>
              ) : (
                <>
                  <ChevronDown className='h-3 w-3 mr-1' />
                  Show more
                </>
              )}
            </Button>
          )}
        </div>
      )}
    </div>
  )
}
