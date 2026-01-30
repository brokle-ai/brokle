'use client'

import { useEffect, useCallback } from 'react'

/**
 * Keyboard shortcut configuration for the Playground
 */
export const PLAYGROUND_HOTKEYS = {
  EXECUTE: {
    key: 'Enter',
    modifier: 'cmd',
    display: '⌘⏎',
    displayWin: 'Ctrl+⏎',
    description: 'Execute active window',
  },
  EXECUTE_ALL: {
    key: 'Enter',
    modifier: 'cmd+shift',
    display: '⌘⇧⏎',
    displayWin: 'Ctrl+Shift+⏎',
    description: 'Execute all windows',
  },
  SAVE: {
    key: 's',
    modifier: 'cmd',
    display: '⌘S',
    displayWin: 'Ctrl+S',
    description: 'Save session',
  },
  NEW_WINDOW: {
    key: 'n',
    modifier: 'cmd',
    display: '⌘N',
    displayWin: 'Ctrl+N',
    description: 'Add new window',
  },
  STOP: {
    key: 'Escape',
    display: 'Esc',
    description: 'Stop execution / Blur input',
  },
} as const

interface UsePlaygroundKeyboardOptions {
  /** Called when user presses Cmd/Ctrl+Enter */
  onExecute?: () => void
  /** Called when user presses Cmd/Ctrl+Shift+Enter */
  onExecuteAll?: () => void
  /** Called when user presses Cmd/Ctrl+S */
  onSave?: () => void
  /** Called when user presses Cmd/Ctrl+N */
  onNewWindow?: () => void
  /** Called when user presses Escape (when not in an input field) */
  onStop?: () => void
  /** Whether keyboard shortcuts are enabled */
  enabled?: boolean
}

/**
 * Hook to handle keyboard shortcuts for the Playground
 *
 * Shortcuts:
 * - Cmd/Ctrl+Enter = Execute active window
 * - Cmd/Ctrl+Shift+Enter = Execute all windows
 * - Cmd/Ctrl+S = Save session (when available)
 * - Cmd/Ctrl+N = Add new window
 * - Escape = Stop execution (blurs input first if focused)
 *
 * Important: All shortcuts use modifier keys to avoid interfering with typing.
 * Escape blurs the input first, then stops if pressed again.
 */
export function usePlaygroundKeyboard({
  onExecute,
  onExecuteAll,
  onSave,
  onNewWindow,
  onStop,
  enabled = true,
}: UsePlaygroundKeyboardOptions) {
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (!enabled) return

      const target = e.target as HTMLElement
      const isInputField =
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.tagName === 'SELECT' ||
        target.isContentEditable

      const isMac = navigator.platform.toLowerCase().includes('mac')
      const cmdKey = isMac ? e.metaKey : e.ctrlKey

      // Cmd/Ctrl + Shift + Enter = Execute all (check first due to more specific modifier)
      if (e.key === 'Enter' && cmdKey && e.shiftKey && onExecuteAll) {
        e.preventDefault()
        onExecuteAll()
        return
      }

      // Cmd/Ctrl + Enter = Execute active window (works even in input fields for quick execution)
      if (e.key === 'Enter' && cmdKey && !e.shiftKey && onExecute) {
        e.preventDefault()
        onExecute()
        return
      }

      // Cmd/Ctrl + S = Save session
      if (e.key === 's' && cmdKey && !e.shiftKey && onSave) {
        e.preventDefault()
        onSave()
        return
      }

      // Cmd/Ctrl + N = New window
      if (e.key === 'n' && cmdKey && !e.shiftKey && onNewWindow) {
        e.preventDefault()
        onNewWindow()
        return
      }

      // Escape = Blur input OR stop execution
      if (e.key === 'Escape') {
        if (isInputField) {
          // First escape: just blur the input
          target.blur()
        } else if (onStop) {
          // Second escape (or when not in input): stop execution
          e.preventDefault()
          onStop()
        }
        return
      }
    },
    [enabled, onExecute, onExecuteAll, onSave, onNewWindow, onStop]
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])
}

/**
 * Get the platform-specific display string for a shortcut
 */
export function getShortcutDisplay(
  hotkey: (typeof PLAYGROUND_HOTKEYS)[keyof typeof PLAYGROUND_HOTKEYS]
): string {
  if (typeof navigator === 'undefined') return hotkey.display
  const isMac = navigator.platform.toLowerCase().includes('mac')
  return isMac ? hotkey.display : (hotkey as { displayWin?: string }).displayWin || hotkey.display
}

/**
 * Check if the current platform is Mac
 */
export function isMacPlatform(): boolean {
  if (typeof navigator === 'undefined') return false
  return navigator.platform.toLowerCase().includes('mac')
}

/**
 * Helper to render a keyboard shortcut badge
 */
export function formatShortcut(shortcut: string): string {
  if (typeof navigator === 'undefined') return shortcut
  const isMac = navigator.platform.toLowerCase().includes('mac')

  // Convert Mac-style shortcuts to Windows-style if needed
  if (!isMac) {
    return shortcut
      .replace('⌘', 'Ctrl+')
      .replace('⇧', 'Shift+')
      .replace('⌥', 'Alt+')
      .replace('⏎', 'Enter')
  }

  return shortcut
}
