import { useMemo } from 'react'
import { useSearchParams } from 'next/navigation'
import type {
  ColumnFiltersState,
  SortingState,
} from '@tanstack/react-table'

type UseTableUrlStateParams = {
  globalFilter?: {
    enabled?: boolean
    key?: string
  }
  sorting?: {
    enabled?: boolean
    sortByKey?: string
    sortOrderKey?: string
  }
  columnFilters?: Array<
    | {
        columnId: string
        searchKey: string
        type?: 'string'
        deserialize?: (value: unknown) => unknown
      }
    | {
        columnId: string
        searchKey: string
        type: 'array'
        deserialize?: (value: unknown) => unknown
      }
  >
}

type UseTableUrlStateReturn = {
  globalFilter?: string
  columnFilters: ColumnFiltersState
  sorting: SortingState
}

export function useTableUrlState(
  params: UseTableUrlStateParams = {}
): UseTableUrlStateReturn {
  const searchParams = useSearchParams()
  
  const {
    globalFilter: globalFilterCfg,
    sorting: sortingCfg,
    columnFilters: columnFiltersCfg = [],
  } = params

  const globalFilterKey = globalFilterCfg?.key ?? 'filter'
  const globalFilterEnabled = globalFilterCfg?.enabled ?? true

  const sortByKey = sortingCfg?.sortByKey ?? 'sortBy'
  const sortOrderKey = sortingCfg?.sortOrderKey ?? 'sortOrder'
  const sortingEnabled = sortingCfg?.enabled ?? true

  // Column filters state derived from URL
  const columnFilters: ColumnFiltersState = useMemo(() => {
    const collected: ColumnFiltersState = []
    for (const cfg of columnFiltersCfg) {
      const raw = searchParams.get(cfg.searchKey)
      const deserialize = cfg.deserialize ?? ((v: unknown) => v)
      if (cfg.type === 'string') {
        const value = raw ? (deserialize(raw) as string) : ''
        if (typeof value === 'string' && value.trim() !== '') {
          collected.push({ id: cfg.columnId, value })
        }
      } else {
        // default to array type
        let value: unknown[] = []
        if (raw) {
          try {
            value = JSON.parse(raw)
          } catch {
            value = []
          }
        }
        value = deserialize(value) as unknown[]
        if (Array.isArray(value) && value.length > 0) {
          collected.push({ id: cfg.columnId, value })
        }
      }
    }
    return collected
  }, [columnFiltersCfg, searchParams])

  // Sorting state derived from URL
  const sorting: SortingState = useMemo(() => {
    if (!sortingEnabled) return []
    
    const sortBy = searchParams.get(sortByKey)
    const sortOrder = searchParams.get(sortOrderKey)
    
    if (sortBy && (sortOrder === 'asc' || sortOrder === 'desc')) {
      return [{
        id: sortBy,
        desc: sortOrder === 'desc'
      }]
    }
    
    return []
  }, [sortingEnabled, searchParams, sortByKey, sortOrderKey])

  // Global filter state derived from URL
  const globalFilter = useMemo(() => {
    if (!globalFilterEnabled) return undefined
    return searchParams.get(globalFilterKey) || ''
  }, [globalFilterEnabled, searchParams, globalFilterKey])

  return {
    globalFilter: globalFilterEnabled ? (globalFilter ?? '') : undefined,
    columnFilters,
    sorting,
  }
}