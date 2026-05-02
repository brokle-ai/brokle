import { useCallback } from 'react'
import { useQueryStates, parseAsString } from 'nuqs'

export type PromptDetailTab = 'prompt' | 'traces' | 'sdk'

const validTabs: PromptDetailTab[] = ['prompt', 'traces', 'sdk']

export interface UsePromptDetailStateReturn {
  selectedVersionId: string | null
  activeTab: PromptDetailTab
  setSelectedVersionId: (versionId: string | null) => void
  setActiveTab: (tab: PromptDetailTab) => void
  setVersionAndTab: (versionId: string | null, tab: PromptDetailTab) => void
  resetState: () => void
}

export function usePromptDetailState(): UsePromptDetailStateReturn {
  const [query, setQuery] = useQueryStates({
    version: parseAsString,
    tab: parseAsString,
  })

  const activeTab: PromptDetailTab = validTabs.includes(query.tab as PromptDetailTab)
    ? (query.tab as PromptDetailTab)
    : 'prompt'

  const setSelectedVersionId = useCallback(
    (versionId: string | null) => {
      setQuery({ version: versionId || null })
    },
    [setQuery],
  )

  const setActiveTab = useCallback(
    (tab: PromptDetailTab) => {
      setQuery({ tab: tab === 'prompt' ? null : tab })
    },
    [setQuery],
  )

  const setVersionAndTab = useCallback(
    (versionId: string | null, tab: PromptDetailTab) => {
      setQuery({
        version: versionId || null,
        tab: tab === 'prompt' ? null : tab,
      })
    },
    [setQuery],
  )

  const resetState = useCallback(() => {
    setQuery({ version: null, tab: null })
  }, [setQuery])

  return {
    selectedVersionId: query.version,
    activeTab,
    setSelectedVersionId,
    setActiveTab,
    setVersionAndTab,
    resetState,
  }
}
