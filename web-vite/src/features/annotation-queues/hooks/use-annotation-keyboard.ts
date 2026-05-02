import { useCallback, useEffect } from 'react'

interface UseAnnotationKeyboardOptions {
  /** Cmd/Ctrl + Enter — submit current scores. Works inside inputs. */
  onSubmit: () => void
  /** S — skip the current item. Suppressed inside inputs. */
  onSkip: () => void
  /** Escape — release the lock. First Esc inside an input blurs it. */
  onRelease: () => void
  /** P — previous item (optional). */
  onPrevious?: () => void
  /** N — next item (optional). */
  onNext?: () => void
  /** C — focus the comment field (optional). */
  onFocusComment?: () => void
  /** Wholesale enable/disable (e.g. when no item is claimed). */
  enabled?: boolean
}

/**
 * Document-level keyboard handler for the annotation review pane.
 * Modelled on Opik's hotkey set:
 *   ⌘/Ctrl+Enter — submit (works in inputs so reviewers can fire from
 *                  the comment field without losing focus)
 *   Esc          — blur input first; release the item if pressed again
 *                  outside an input
 *   S / P / N / C — single-letter shortcuts; suppressed inside inputs
 *                   so they don't interfere with typing
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

      const target = e.target as HTMLElement | null
      const isInputField =
        !!target &&
        (target.tagName === 'INPUT' ||
          target.tagName === 'TEXTAREA' ||
          target.tagName === 'SELECT' ||
          target.isContentEditable)

      if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        onSubmit()
        return
      }

      if (e.key === 'Escape') {
        if (isInputField && target) {
          target.blur()
        } else {
          e.preventDefault()
          onRelease()
        }
        return
      }

      if (isInputField) return
      if (e.metaKey || e.ctrlKey || e.altKey) return

      if (e.key === 'p' && onPrevious) {
        e.preventDefault()
        onPrevious()
        return
      }
      if (e.key === 'n' && onNext) {
        e.preventDefault()
        onNext()
        return
      }
      if (e.key === 's') {
        e.preventDefault()
        onSkip()
        return
      }
      if (e.key === 'c' && onFocusComment) {
        e.preventDefault()
        onFocusComment()
        return
      }
    },
    [enabled, onSubmit, onSkip, onRelease, onPrevious, onNext, onFocusComment],
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])
}
