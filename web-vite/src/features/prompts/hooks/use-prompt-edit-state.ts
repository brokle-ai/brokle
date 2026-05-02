import { useCallback } from 'react'
import { useQueryStates, parseAsString, parseAsBoolean } from 'nuqs'

export interface UsePromptEditStateReturn {
  sourceVersionId: string | null
  isRestoreFlow: boolean
  setSourceVersionId: (versionId: string | null) => void
  setIsRestoreFlow: (isRestore: boolean) => void
  resetState: () => void
}

export function usePromptEditState(): UsePromptEditStateReturn {
  const [query, setQuery] = useQueryStates({
    version: parseAsString,
    restore: parseAsBoolean,
  })

  const setSourceVersionId = useCallback(
    (versionId: string | null) => {
      setQuery({ version: versionId || null })
    },
    [setQuery],
  )

  const setIsRestoreFlow = useCallback(
    (isRestore: boolean) => {
      setQuery({ restore: isRestore || null })
    },
    [setQuery],
  )

  const resetState = useCallback(() => {
    setQuery({ version: null, restore: null })
  }, [setQuery])

  return {
    sourceVersionId: query.version,
    isRestoreFlow: query.restore ?? false,
    setSourceVersionId,
    setIsRestoreFlow,
    resetState,
  }
}
