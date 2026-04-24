import { useEffect, useCallback } from 'react'

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
  onExecute?: () => void
  onExecuteAll?: () => void
  onSave?: () => void
  onNewWindow?: () => void
  onStop?: () => void
  enabled?: boolean
}

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

      if (e.key === 'Enter' && cmdKey && e.shiftKey && onExecuteAll) {
        e.preventDefault()
        onExecuteAll()
        return
      }

      if (e.key === 'Enter' && cmdKey && !e.shiftKey && onExecute) {
        e.preventDefault()
        onExecute()
        return
      }

      if (e.key === 's' && cmdKey && !e.shiftKey && onSave) {
        e.preventDefault()
        onSave()
        return
      }

      if (e.key === 'n' && cmdKey && !e.shiftKey && onNewWindow) {
        e.preventDefault()
        onNewWindow()
        return
      }

      if (e.key === 'Escape') {
        if (isInputField) {
          target.blur()
        } else if (onStop) {
          e.preventDefault()
          onStop()
        }
        return
      }
    },
    [enabled, onExecute, onExecuteAll, onSave, onNewWindow, onStop],
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])
}

export function getShortcutDisplay(
  hotkey: (typeof PLAYGROUND_HOTKEYS)[keyof typeof PLAYGROUND_HOTKEYS],
): string {
  if (typeof navigator === 'undefined') return hotkey.display
  const isMac = navigator.platform.toLowerCase().includes('mac')
  return isMac
    ? hotkey.display
    : (hotkey as { displayWin?: string }).displayWin || hotkey.display
}

export function isMacPlatform(): boolean {
  if (typeof navigator === 'undefined') return false
  return navigator.platform.toLowerCase().includes('mac')
}

export function formatShortcut(shortcut: string): string {
  if (typeof navigator === 'undefined') return shortcut
  const isMac = navigator.platform.toLowerCase().includes('mac')

  if (!isMac) {
    return shortcut
      .replace('⌘', 'Ctrl+')
      .replace('⇧', 'Shift+')
      .replace('⌥', 'Alt+')
      .replace('⏎', 'Enter')
  }

  return shortcut
}
