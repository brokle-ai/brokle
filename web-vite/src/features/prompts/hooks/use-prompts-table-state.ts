import { useMemo, useCallback } from 'react'
import { useQueryStates, parseAsInteger, parseAsString } from 'nuqs'
import type { PromptType } from '../types'

const promptTypes = ['text', 'chat'] as const

export interface UsePromptsTableStateReturn {
  page: number
  pageSize: number
  search: string | null
  types: PromptType[]
  sortBy: string | null
  sortOrder: 'asc' | 'desc' | null
  setSearch: (search: string) => void
  setTypes: (types: PromptType[]) => void
  setPagination: (page: number, pageSize?: number) => void
  setSorting: (sortBy: string | null, sortOrder: 'asc' | 'desc' | null) => void
  resetAll: () => void
  hasActiveFilters: boolean
}

export function usePromptsTableState(): UsePromptsTableStateReturn {
  const [query, setQuery] = useQueryStates({
    page: parseAsInteger.withDefault(1),
    pageSize: parseAsInteger.withDefault(10),
    search: parseAsString,
    types: parseAsString,
    sortBy: parseAsString,
    sortOrder: parseAsString,
  })

  const types = useMemo((): PromptType[] => {
    if (!query.types) return []
    try {
      const parsed = JSON.parse(query.types)
      return Array.isArray(parsed)
        ? parsed.filter((t): t is PromptType =>
            promptTypes.includes(t as PromptType),
          )
        : []
    } catch {
      return []
    }
  }, [query.types])

  const setSearch = useCallback(
    (search: string) => {
      setQuery({ search: search || null, page: 1 })
    },
    [setQuery],
  )

  const setTypes = useCallback(
    (newTypes: PromptType[]) => {
      setQuery({
        types: newTypes.length > 0 ? JSON.stringify(newTypes) : null,
        page: 1,
      })
    },
    [setQuery],
  )

  const setPagination = useCallback(
    (page: number, pageSize?: number) => {
      setQuery({
        page: Math.max(1, page),
        ...(pageSize !== undefined && { pageSize: Math.max(1, pageSize) }),
      })
    },
    [setQuery],
  )

  const setSorting = useCallback(
    (sortBy: string | null, sortOrder: 'asc' | 'desc' | null) => {
      setQuery({ sortBy: sortBy || null, sortOrder: sortOrder || null })
    },
    [setQuery],
  )

  const resetAll = useCallback(() => {
    setQuery({
      page: 1,
      pageSize: null,
      search: null,
      types: null,
      sortBy: null,
      sortOrder: null,
    })
  }, [setQuery])

  return {
    page: Math.max(1, query.page),
    pageSize: Math.max(1, query.pageSize),
    search: query.search,
    types,
    sortBy: query.sortBy,
    sortOrder: query.sortOrder as 'asc' | 'desc' | null,
    setSearch,
    setTypes,
    setPagination,
    setSorting,
    resetAll,
    hasActiveFilters: !!query.search || types.length > 0,
  }
}
