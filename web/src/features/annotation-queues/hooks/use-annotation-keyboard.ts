'use client'

import { useEffect, useCallback } from 'react'

/**
 * Keyboard shortcut configuration (Opik-style)
 * Simple letter keys for common actions, modifier+key for important actions
 */
export const ANNOTATION_HOTKEYS = {
  PREVIOUS: { key: 'p', display: 'P', description: 'Previous item' },
  NEXT: { key: 'n', display: 'N', description: 'Next item' },
  SUBMIT: {
    key: 'Enter',
    modifier: true,
    display: '⌘⏎',
    displayWin: 'Ctrl+⏎',
    description: 'Submit scores',
  },
  SKIP: { key: 's', display: 'S', description: 'Skip item' },
  RELEASE: { key: 'Escape', display: 'Esc', description: 'Release item' },
  FOCUS_COMMENT: { key: 'c', display: 'C', description: 'Focus comment' },
} as const

interface UseAnnotationKeyboardOptions {
  /** Called when user presses Cmd/Ctrl+Enter */
  onSubmit: () => void
  /** Called when user presses S key */
  onSkip: () => void
  /** Called when user presses Escape key */
  onRelease: () => void
  /** Called when user presses P key (optional - for multi-item navigation) */
  onPrevious?: () => void
  /** Called when user presses N key (optional - for multi-item navigation) */
  onNext?: () => void
  /** Called when user presses C key - should focus a comment field */
  onFocusComment?: () => void
  /** Whether keyboard shortcuts are enabled (disable when no item is claimed) */
  enabled?: boolean
}

/**
 * Hook to handle keyboard shortcuts for the annotation panel
 *
 * Follows Opik pattern:
 * - P = Previous item
 * - N = Next item
 * - Cmd/Ctrl+Enter = Submit scores
 * - S = Skip item
 * - Escape = Release item / blur input
 * - C = Focus comment field
 *
 * Important: Simple letter keys are ignored when user is typing in an input field
 * Cmd/Ctrl+Enter works even when focused on an input (allows quick submit)
 * Escape blurs the input first, then releases if pressed again
 */
export function useAnnotationKeyboard({
  onSubmit,
  onSkip,
  onRelease,
  onPrevious,
  onNext,
  onFocusComment,
  enabled = true,
}: UseAnnotationKeyboardOptions) {
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (!enabled) return

      const target = e.target as HTMLElement
      const isInputField =
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.tagName === 'SELECT' ||
        target.isContentEditable

      // Cmd/Ctrl + Enter = Submit (works even in input fields - Opik pattern)
      if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        onSubmit()
        return
      }

      // Escape = Blur input OR release item
      if (e.key === 'Escape') {
        if (isInputField) {
          // First escape: just blur the input
          target.blur()
        } else {
          // Second escape (or when not in input): release the item
          e.preventDefault()
          onRelease()
        }
        return
      }

      // Skip remaining shortcuts if user is typing in an input field
      // This prevents accidental actions while entering comments/values
      if (isInputField) return

      // Skip single-letter shortcuts if modifier keys are held
      // This allows browser shortcuts (Ctrl+S, Ctrl+P, etc.) to work normally
      if (e.metaKey || e.ctrlKey || e.altKey) return

      // P = Previous item
      if (e.key === 'p' && onPrevious) {
        e.preventDefault()
        onPrevious()
        return
      }

      // N = Next item
      if (e.key === 'n' && onNext) {
        e.preventDefault()
        onNext()
        return
      }

      // S = Skip item
      if (e.key === 's') {
        e.preventDefault()
        onSkip()
        return
      }

      // C = Focus comment field (Opik pattern)
      if (e.key === 'c' && onFocusComment) {
        e.preventDefault()
        onFocusComment()
        return
      }
    },
    [enabled, onSubmit, onSkip, onRelease, onPrevious, onNext, onFocusComment]
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])
}

/**
 * Get the platform-specific display string for the submit shortcut
 */
export function getSubmitShortcutDisplay(): string {
  if (typeof navigator === 'undefined') return '⌘⏎'
  const isMac = navigator.platform.toLowerCase().includes('mac')
  return isMac ? '⌘⏎' : 'Ctrl+⏎'
}

/**
 * Check if the current platform is Mac
 */
export function isMacPlatform(): boolean {
  if (typeof navigator === 'undefined') return false
  return navigator.platform.toLowerCase().includes('mac')
}
