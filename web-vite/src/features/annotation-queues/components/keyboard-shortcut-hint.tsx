import { useEffect, useState, type ReactNode } from 'react'
import { Keyboard } from 'lucide-react'
import { cn } from '@/lib/utils'

interface KeyboardShortcutHintProps {
  className?: string
  compact?: boolean
  showIcon?: boolean
}

/**
 * Renders the platform-aware keyboard hints used by the review pane.
 * Mac shows ⌘⏎; everything else shows Ctrl+⏎. The display defaults to
 * Mac on first paint to avoid SSR/hydration mismatch and corrects on
 * the first effect tick — visible for one frame at most on non-Mac.
 */
export function KeyboardShortcutHint({
  className,
  compact = false,
  showIcon = true,
}: KeyboardShortcutHintProps) {
  const [isMac, setIsMac] = useState(true)

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
      <div
        className={cn(
          'flex items-center gap-2 text-[10px] text-muted-foreground',
          className,
        )}
      >
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
        className,
      )}
    >
      {showIcon && <Keyboard className="h-3.5 w-3.5 flex-shrink-0" />}
      <span className="font-medium text-foreground/70">Shortcuts:</span>
      <div className="flex flex-wrap items-center gap-3">
        {shortcuts.map((s) => (
          <span key={s.key} className="flex items-center gap-1">
            <Kbd>{s.key}</Kbd>
            <span className="opacity-75">{s.label}</span>
          </span>
        ))}
      </div>
    </div>
  )
}

export function Kbd({
  children,
  className,
}: {
  children: ReactNode
  className?: string
}) {
  return (
    <kbd
      className={cn(
        'inline-flex items-center justify-center px-1.5 py-0.5',
        'bg-muted border border-border rounded text-[10px] font-mono',
        'min-w-[1.5rem] h-5',
        className,
      )}
    >
      {children}
    </kbd>
  )
}
