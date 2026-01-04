/**
 * Auto-Save Status Indicator
 *
 * Displays the current auto-save status with appropriate icons and colors.
 */

'use client'

import { Check, Cloud, CloudOff, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { AutoSaveStatus } from '../hooks/use-auto-save'

interface AutoSaveIndicatorProps {
  status: AutoSaveStatus
  error?: string | null
  className?: string
}

export function AutoSaveIndicator({ status, error, className }: AutoSaveIndicatorProps) {
  if (status === 'idle') {
    return null
  }

  return (
    <div
      className={cn(
        'flex items-center gap-1.5 text-xs transition-opacity',
        status === 'error' && 'text-destructive',
        status === 'saved' && 'text-muted-foreground',
        status === 'saving' && 'text-muted-foreground',
        status === 'pending' && 'text-yellow-600 dark:text-yellow-500',
        className
      )}
    >
      {status === 'pending' && (
        <>
          <Cloud className="h-3.5 w-3.5" />
          <span>Unsaved changes</span>
        </>
      )}
      {status === 'saving' && (
        <>
          <Loader2 className="h-3.5 w-3.5 animate-spin" />
          <span>Saving...</span>
        </>
      )}
      {status === 'saved' && (
        <>
          <Check className="h-3.5 w-3.5" />
          <span>Saved</span>
        </>
      )}
      {status === 'error' && (
        <>
          <CloudOff className="h-3.5 w-3.5" />
          <span title={error ?? 'Save failed'}>Save failed</span>
        </>
      )}
    </div>
  )
}
