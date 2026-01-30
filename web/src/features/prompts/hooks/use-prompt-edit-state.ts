'use client'

import { useCallback } from 'react'
import { useQueryStates, parseAsString, parseAsBoolean } from 'nuqs'

export interface UsePromptEditStateReturn {
  // State (read from URL)
  sourceVersionId: string | null
  isRestoreFlow: boolean

  // Setters (update URL)
  setSourceVersionId: (versionId: string | null) => void
  setIsRestoreFlow: (isRestore: boolean) => void
  resetState: () => void
}

/**
 * Centralized hook for managing prompt edit page state via URL params.
 * Uses nuqs for type-safe URL synchronization with shallow routing.
 *
 * URL Params:
 * - version: Source version ID to load content from
 * - restore: Boolean flag indicating restore flow (different UX hints)
 *
 * Example URLs:
 * - /prompts/abc123/edit (new version from latest)
 * - /prompts/abc123/edit?version=xyz789 (edit from specific version)
 * - /prompts/abc123/edit?version=xyz789&restore=true (restore flow)
 */
export function usePromptEditState(): UsePromptEditStateReturn {
  const [query, setQuery] = useQueryStates({
    version: parseAsString,
    restore: parseAsBoolean,
  })

  // Setters that update URL
  const setSourceVersionId = useCallback(
    (versionId: string | null) => {
      setQuery({ version: versionId || null })
    },
    [setQuery]
  )

  const setIsRestoreFlow = useCallback(
    (isRestore: boolean) => {
      setQuery({ restore: isRestore || null })
    },
    [setQuery]
  )

  const resetState = useCallback(() => {
    setQuery({
      version: null,
      restore: null,
    })
  }, [setQuery])

  return {
    // State (read from URL)
    sourceVersionId: query.version,
    isRestoreFlow: query.restore ?? false,

    // Setters (update URL)
    setSourceVersionId,
    setIsRestoreFlow,
    resetState,
  }
}
