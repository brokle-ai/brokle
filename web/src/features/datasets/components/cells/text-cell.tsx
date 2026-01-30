'use client'

import { cn } from '@/lib/utils'
import type { RowHeight } from './types'

interface TextCellProps {
  value: string | undefined | null
  rowHeight?: RowHeight
  className?: string
}

const ROW_HEIGHT_CONFIG: Record<RowHeight, { maxLines: number; className: string }> = {
  small: { maxLines: 1, className: 'line-clamp-1' },
  medium: { maxLines: 2, className: 'line-clamp-2' },
  large: { maxLines: 6, className: 'line-clamp-6' },
}

export function TextCell({ value, rowHeight = 'medium', className }: TextCellProps) {
  if (!value) {
    return <span className="text-muted-foreground">-</span>
  }

  const config = ROW_HEIGHT_CONFIG[rowHeight]

  return (
    <span
      className={cn(
        'text-sm break-words',
        config.className,
        className
      )}
      title={value}
    >
      {value}
    </span>
  )
}
