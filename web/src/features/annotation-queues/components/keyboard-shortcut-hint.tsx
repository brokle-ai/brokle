'use client'

import { useEffect, useState } from 'react'
import { cn } from '@/lib/utils'
import { Keyboard } from 'lucide-react'

interface KeyboardShortcutHintProps {
  /** Additional CSS classes */
  className?: string
  /** Whether to show in compact mode */
  compact?: boolean
  /** Whether to show the keyboard icon */
  showIcon?: boolean
}

/**
 * Displays available keyboard shortcuts for the annotation panel
 *
 * Follows Opik pattern:
 * - Shows platform-aware modifier keys (⌘ on Mac, Ctrl on Windows/Linux)
 * - Compact display with kbd styling
 * - Shows most important shortcuts
 */
export function KeyboardShortcutHint({
  className,
  compact = false,
  showIcon = true,
}: KeyboardShortcutHintProps) {
  const [isMac, setIsMac] = useState(true)

  // Detect platform on client side
  useEffect(() => {
    setIsMac(navigator.platform.toLowerCase().includes('mac'))
  }, [])

  const submitKey = isMac ? '⌘⏎' : 'Ctrl+⏎'

  const shortcuts = [
    { key: submitKey, label: 'Submit' },
    { key: 'S', label: 'Skip' },
    { key: 'Esc', label: 'Release' },
  ]

  if (compact) {
    return (
      <div className={cn('flex items-center gap-2 text-[10px] text-muted-foreground', className)}>
        {showIcon && <Keyboard className="h-3 w-3" />}
        {shortcuts.map((s, i) => (
          <span key={s.key}>
            <Kbd>{s.key}</Kbd>
            {i < shortcuts.length - 1 && <span className="mx-1">·</span>}
          </span>
        ))}
      </div>
    )
  }

  return (
    <div
      className={cn(
        'flex items-center gap-4 text-xs text-muted-foreground border-t pt-3 mt-4',
        className
      )}
    >
      {showIcon && <Keyboard className="h-3.5 w-3.5 flex-shrink-0" />}
      <span className="font-medium text-foreground/70">Shortcuts:</span>
      <div className="flex items-center gap-3 flex-wrap">
        {shortcuts.map((shortcut) => (
          <ShortcutItem key={shortcut.key} shortcut={shortcut.key} label={shortcut.label} />
        ))}
      </div>
    </div>
  )
}

/**
 * Individual shortcut display
 */
function ShortcutItem({ shortcut, label }: { shortcut: string; label: string }) {
  return (
    <span className="flex items-center gap-1">
      <Kbd>{shortcut}</Kbd>
      <span className="opacity-75">{label}</span>
    </span>
  )
}

/**
 * Styled keyboard key display
 */
function Kbd({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <kbd
      className={cn(
        'inline-flex items-center justify-center px-1.5 py-0.5',
        'bg-muted border border-border rounded text-[10px] font-mono',
        'min-w-[1.5rem] h-5',
        className
      )}
    >
      {children}
    </kbd>
  )
}

/**
 * Standalone Kbd component for use elsewhere
 */
export { Kbd }
