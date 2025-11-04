import { useMemo, useCallback } from 'react'
import { useRouter, useSearchParams, usePathname } from 'next/navigation'
import type {
  ColumnFiltersState,
  SortingState,
  PaginationState,
  OnChangeFn,
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
  pagination?: {
    enabled?: boolean
    pageKey?: string
    pageSizeKey?: string
    defaultPageSize?: number
  }
  columnFilters?: Array<
    | {
        columnId: string
        searchKey: string
        type?: 'string'
        deserialize?: (value: unknown) => unknown
        serialize?: (value: unknown) => string
      }
    | {
        columnId: string
        searchKey: string
        type: 'array'
        deserialize?: (value: unknown) => unknown
        serialize?: (value: unknown) => string
      }
  >
}

type UseTableUrlStateReturn = {
  // Read values
  globalFilter?: string
  columnFilters: ColumnFiltersState
  sorting: SortingState
  pagination: PaginationState
  // Write handlers
  onGlobalFilterChange: OnChangeFn<string>
  onColumnFiltersChange: OnChangeFn<ColumnFiltersState>
  onSortingChange: OnChangeFn<SortingState>
  onPaginationChange: OnChangeFn<PaginationState>
  // Utility
  ensurePageInRange: (totalRows: number) => void
}

export function useTableUrlState(
  params: UseTableUrlStateParams = {}
): UseTableUrlStateReturn {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const {
    globalFilter: globalFilterCfg,
    sorting: sortingCfg,
    pagination: paginationCfg,
    columnFilters: columnFiltersCfg = [],
  } = params

  const globalFilterKey = globalFilterCfg?.key ?? 'filter'
  const globalFilterEnabled = globalFilterCfg?.enabled ?? true

  const sortByKey = sortingCfg?.sortByKey ?? 'sortBy'
  const sortOrderKey = sortingCfg?.sortOrderKey ?? 'sortOrder'
  const sortingEnabled = sortingCfg?.enabled ?? true

  const pageKey = paginationCfg?.pageKey ?? 'page'
  const pageSizeKey = paginationCfg?.pageSizeKey ?? 'pageSize'
  const paginationEnabled = paginationCfg?.enabled ?? true
  const defaultPageSize = paginationCfg?.defaultPageSize ?? 10

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
          } catch (error) {
            console.warn(
              `[useTableUrlState] Failed to parse array filter for "${cfg.searchKey}":`,
              raw,
              error instanceof Error ? error.message : 'Unknown parsing error'
            )
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

  // Pagination state derived from URL
  const pagination: PaginationState = useMemo(() => {
    if (!paginationEnabled) {
      return { pageIndex: 0, pageSize: defaultPageSize }
    }

    const pageParam = searchParams.get(pageKey)
    const pageSizeParam = searchParams.get(pageSizeKey)

    const pageIndex = pageParam ? Math.max(0, parseInt(pageParam, 10) - 1) : 0
    const pageSize = pageSizeParam
      ? Math.max(1, parseInt(pageSizeParam, 10))
      : defaultPageSize

    return { pageIndex, pageSize }
  }, [paginationEnabled, searchParams, pageKey, pageSizeKey, defaultPageSize])

  // Helper to update URL params
  const updateUrlParams = useCallback(
    (updates: Record<string, string | null>) => {
      const newParams = new URLSearchParams(searchParams.toString())

      Object.entries(updates).forEach(([key, value]) => {
        if (value === null || value === '' || value === undefined) {
          newParams.delete(key)
        } else {
          newParams.set(key, value)
        }
      })

      const newUrl = `${pathname}?${newParams.toString()}`
      router.push(newUrl, { scroll: false })
    },
    [searchParams, pathname, router]
  )

  // Global filter change handler
  const onGlobalFilterChange: OnChangeFn<string> = useCallback(
    (updaterOrValue) => {
      if (!globalFilterEnabled) return

      const newValue =
        typeof updaterOrValue === 'function'
          ? updaterOrValue(globalFilter ?? '')
          : updaterOrValue

      updateUrlParams({
        [globalFilterKey]: newValue || null,
        [pageKey]: null, // Reset to first page when filtering
      })
    },
    [globalFilterEnabled, globalFilter, globalFilterKey, pageKey, updateUrlParams]
  )

  // Column filters change handler
  const onColumnFiltersChange: OnChangeFn<ColumnFiltersState> = useCallback(
    (updaterOrValue) => {
      const newFilters =
        typeof updaterOrValue === 'function'
          ? updaterOrValue(columnFilters)
          : updaterOrValue

      const updates: Record<string, string | null> = {
        [pageKey]: null, // Reset to first page when filtering
      }

      // Clear all existing filter params
      columnFiltersCfg.forEach((cfg) => {
        updates[cfg.searchKey] = null
      })

      // Set new filter values
      newFilters.forEach((filter) => {
        const cfg = columnFiltersCfg.find((c) => c.columnId === filter.id)
        if (!cfg) return

        const serialize = cfg.serialize ?? ((v: unknown) => String(v))

        if (cfg.type === 'array') {
          updates[cfg.searchKey] = JSON.stringify(filter.value)
        } else {
          updates[cfg.searchKey] = serialize(filter.value)
        }
      })

      updateUrlParams(updates)
    },
    [columnFilters, columnFiltersCfg, pageKey, updateUrlParams]
  )

  // Sorting change handler
  const onSortingChange: OnChangeFn<SortingState> = useCallback(
    (updaterOrValue) => {
      if (!sortingEnabled) return

      const newSorting =
        typeof updaterOrValue === 'function'
          ? updaterOrValue(sorting)
          : updaterOrValue

      const updates: Record<string, string | null> = {}

      if (newSorting.length === 0) {
        updates[sortByKey] = null
        updates[sortOrderKey] = null
      } else {
        const sort = newSorting[0]
        updates[sortByKey] = sort.id
        updates[sortOrderKey] = sort.desc ? 'desc' : 'asc'
      }

      updateUrlParams(updates)
    },
    [sortingEnabled, sorting, sortByKey, sortOrderKey, updateUrlParams]
  )

  // Pagination change handler
  const onPaginationChange: OnChangeFn<PaginationState> = useCallback(
    (updaterOrValue) => {
      if (!paginationEnabled) return

      const newPagination =
        typeof updaterOrValue === 'function'
          ? updaterOrValue(pagination)
          : updaterOrValue

      const updates: Record<string, string | null> = {}

      // Page is 1-indexed in URL, 0-indexed internally
      updates[pageKey] = String(newPagination.pageIndex + 1)

      // Only set page size if it differs from default
      if (newPagination.pageSize !== defaultPageSize) {
        updates[pageSizeKey] = String(newPagination.pageSize)
      } else {
        updates[pageSizeKey] = null
      }

      updateUrlParams(updates)
    },
    [
      paginationEnabled,
      pagination,
      pageKey,
      pageSizeKey,
      defaultPageSize,
      updateUrlParams,
    ]
  )

  // Utility to ensure current page is within valid range
  const ensurePageInRange = useCallback(
    (totalRows: number) => {
      if (!paginationEnabled) return

      const maxPage = Math.max(0, Math.ceil(totalRows / pagination.pageSize) - 1)

      if (pagination.pageIndex > maxPage) {
        updateUrlParams({
          [pageKey]: String(maxPage + 1), // 1-indexed in URL
        })
      }
    },
    [paginationEnabled, pagination, pageKey, updateUrlParams]
  )

  return {
    globalFilter: globalFilterEnabled ? (globalFilter ?? '') : undefined,
    columnFilters,
    sorting,
    pagination,
    onGlobalFilterChange,
    onColumnFiltersChange,
    onSortingChange,
    onPaginationChange,
    ensurePageInRange,
  }
}