'use client'

import { useCallback } from 'react'
import { useQueryStates, parseAsString } from 'nuqs'

export type PromptDetailTab = 'prompt' | 'traces' | 'sdk'

const validTabs: PromptDetailTab[] = ['prompt', 'traces', 'sdk']

export interface UsePromptDetailStateReturn {
  // State (read from URL)
  selectedVersionId: string | null
  activeTab: PromptDetailTab

  // Setters (update URL)
  setSelectedVersionId: (versionId: string | null) => void
  setActiveTab: (tab: PromptDetailTab) => void
  setVersionAndTab: (versionId: string | null, tab: PromptDetailTab) => void
  resetState: () => void
}

/**
 * Centralized hook for managing prompt detail page state via URL params.
 * Uses nuqs for type-safe URL synchronization with shallow routing.
 *
 * URL Params:
 * - version: Selected version ID to display
 * - tab: Active tab ('prompt' | 'traces' | 'sdk')
 *
 * Example URLs:
 * - /prompts/abc123 (default: latest version, prompt tab)
 * - /prompts/abc123?version=xyz789 (specific version)
 * - /prompts/abc123?tab=sdk (SDK tab)
 * - /prompts/abc123?version=xyz789&tab=sdk (specific version, SDK tab)
 */
export function usePromptDetailState(): UsePromptDetailStateReturn {
  const [query, setQuery] = useQueryStates({
    version: parseAsString,
    tab: parseAsString,
  })

  // Validate and normalize tab value
  const activeTab: PromptDetailTab = validTabs.includes(query.tab as PromptDetailTab)
    ? (query.tab as PromptDetailTab)
    : 'prompt'

  // Setters that update URL
  const setSelectedVersionId = useCallback(
    (versionId: string | null) => {
      setQuery({ version: versionId || null })
    },
    [setQuery]
  )

  const setActiveTab = useCallback(
    (tab: PromptDetailTab) => {
      // Only set if different from default to keep URL clean
      setQuery({ tab: tab === 'prompt' ? null : tab })
    },
    [setQuery]
  )

  const setVersionAndTab = useCallback(
    (versionId: string | null, tab: PromptDetailTab) => {
      setQuery({
        version: versionId || null,
        tab: tab === 'prompt' ? null : tab,
      })
    },
    [setQuery]
  )

  const resetState = useCallback(() => {
    setQuery({
      version: null,
      tab: null,
    })
  }, [setQuery])

  return {
    // State (read from URL)
    selectedVersionId: query.version,
    activeTab,

    // Setters (update URL)
    setSelectedVersionId,
    setActiveTab,
    setVersionAndTab,
    resetState,
  }
}
