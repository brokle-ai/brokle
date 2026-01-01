'use client'

import { useState, useCallback, useMemo, useEffect } from 'react'
import { useSearchParams, useRouter, usePathname } from 'next/navigation'
import { nanoid } from 'nanoid'
import type { FilterCondition, FilterOperator } from '../api/traces-api'
import { operatorRequiresValue } from '../config/filter-columns'

export interface UseFilterStateOptions {
  /** Sync filters with URL query params */
  syncWithUrl?: boolean
  /** Initial filters to apply */
  initialFilters?: FilterCondition[]
  /** Initial search query */
  initialSearch?: string
  /** Initial search types */
  initialSearchTypes?: ('id' | 'content' | 'all')[]
  /** Maximum number of filters */
  maxFilters?: number
}

export interface UseFilterStateReturn {
  // Filter state
  filters: FilterCondition[]
  searchQuery: string
  searchTypes: ('id' | 'content' | 'all')[]

  // Filter operations
  addFilter: (column?: string, operator?: FilterOperator, value?: any) => void
  updateFilter: (id: string, updates: Partial<FilterCondition>) => void
  removeFilter: (id: string) => void
  clearFilters: () => void
  setFilters: (filters: FilterCondition[]) => void

  // Search operations
  setSearchQuery: (query: string) => void
  setSearchTypes: (types: ('id' | 'content' | 'all')[]) => void

  // State checks
  hasFilters: boolean
  hasSearch: boolean
  isFilterValid: (filter: FilterCondition) => boolean
  getInvalidFilters: () => FilterCondition[]

  // URL sync
  getUrlParams: () => URLSearchParams
  loadFromUrl: () => void

  // For presets
  toPresetFormat: () => {
    filters: FilterCondition[]
    search_query?: string
    search_types?: string[]
  }
  loadFromPreset: (preset: {
    filters?: FilterCondition[]
    search_query?: string
    search_types?: string[]
  }) => void
}

/**
 * Serialize a filter condition to URL-safe string
 */
function serializeFilter(filter: FilterCondition): string {
  const value = Array.isArray(filter.value)
    ? filter.value.join('|')
    : String(filter.value ?? '')
  return `${filter.column}:${filter.operator}:${encodeURIComponent(value)}`
}

/**
 * Deserialize a URL string to filter condition
 */
function deserializeFilter(str: string): FilterCondition | null {
  const parts = str.split(':')
  if (parts.length < 2) return null

  const column = parts[0]
  const operator = parts[1] as FilterOperator
  const rawValue = parts.slice(2).join(':')
  const value = decodeURIComponent(rawValue)

  // Handle multi-value operators
  const parsedValue =
    operator === 'IN' || operator === 'NOT IN' ? value.split('|') : value

  return {
    id: nanoid(8),
    column,
    operator,
    value: parsedValue || null,
  }
}

/**
 * Hook for managing filter builder state
 */
export function useFilterState(
  options: UseFilterStateOptions = {}
): UseFilterStateReturn {
  const {
    syncWithUrl = false,
    initialFilters = [],
    initialSearch = '',
    initialSearchTypes = ['id'],
    maxFilters = 20,
  } = options

  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  // State
  const [filters, setFiltersState] = useState<FilterCondition[]>(initialFilters)
  const [searchQuery, setSearchQueryState] = useState(initialSearch)
  const [searchTypes, setSearchTypesState] =
    useState<('id' | 'content' | 'all')[]>(initialSearchTypes)

  // Load from URL on mount if syncWithUrl is enabled
  useEffect(() => {
    if (syncWithUrl) {
      loadFromUrl()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Update URL when filters change
  useEffect(() => {
    if (syncWithUrl) {
      const params = getUrlParams()
      const newUrl = `${pathname}?${params.toString()}`
      router.replace(newUrl, { scroll: false })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filters, searchQuery, searchTypes, syncWithUrl])

  // Filter validation
  const isFilterValid = useCallback((filter: FilterCondition): boolean => {
    if (!filter.column || !filter.operator) return false
    if (operatorRequiresValue(filter.operator)) {
      if (filter.value === null || filter.value === undefined) return false
      if (typeof filter.value === 'string' && filter.value.trim() === '')
        return false
      if (Array.isArray(filter.value) && filter.value.length === 0) return false
    }
    return true
  }, [])

  // Add a new filter
  const addFilter = useCallback(
    (
      column: string = '',
      operator: FilterOperator = '=',
      value: any = ''
    ) => {
      if (filters.length >= maxFilters) return

      const newFilter: FilterCondition = {
        id: nanoid(8),
        column,
        operator,
        value,
      }
      setFiltersState((prev) => [...prev, newFilter])
    },
    [filters.length, maxFilters]
  )

  // Update an existing filter
  const updateFilter = useCallback(
    (id: string, updates: Partial<FilterCondition>) => {
      setFiltersState((prev) =>
        prev.map((f) => (f.id === id ? { ...f, ...updates } : f))
      )
    },
    []
  )

  // Remove a filter
  const removeFilter = useCallback((id: string) => {
    setFiltersState((prev) => prev.filter((f) => f.id !== id))
  }, [])

  // Clear all filters
  const clearFilters = useCallback(() => {
    setFiltersState([])
    setSearchQueryState('')
    setSearchTypesState(['id'])
  }, [])

  // Set filters directly
  const setFilters = useCallback((newFilters: FilterCondition[]) => {
    setFiltersState(newFilters)
  }, [])

  // Set search query
  const setSearchQuery = useCallback((query: string) => {
    setSearchQueryState(query)
  }, [])

  // Set search types
  const setSearchTypes = useCallback(
    (types: ('id' | 'content' | 'all')[]) => {
      setSearchTypesState(types)
    },
    []
  )

  // Get invalid filters
  const getInvalidFilters = useCallback(() => {
    return filters.filter((f) => !isFilterValid(f))
  }, [filters, isFilterValid])

  // Build URL params
  const getUrlParams = useCallback(() => {
    const params = new URLSearchParams()

    // Add filters
    const validFilters = filters.filter(isFilterValid)
    if (validFilters.length > 0) {
      params.set('filters', validFilters.map(serializeFilter).join(','))
    }

    // Add search
    if (searchQuery) {
      params.set('search', searchQuery)
    }
    // Serialize search_types unless it's the default ['id']
    const isDefaultSearchTypes = searchTypes.length === 1 && searchTypes[0] === 'id'
    if (!isDefaultSearchTypes) {
      params.set('search_types', searchTypes.join(','))
    }

    return params
  }, [filters, isFilterValid, searchQuery, searchTypes])

  // Load from URL
  const loadFromUrl = useCallback(() => {
    // Parse filters
    const filtersParam = searchParams.get('filters')
    if (filtersParam) {
      const parsedFilters = filtersParam
        .split(',')
        .map(deserializeFilter)
        .filter(Boolean) as FilterCondition[]
      setFiltersState(parsedFilters)
    }

    // Parse search
    const searchParam = searchParams.get('search')
    if (searchParam) {
      setSearchQueryState(searchParam)
    }

    // Parse search types
    const searchTypesParam = searchParams.get('search_types')
    if (searchTypesParam) {
      const types = searchTypesParam.split(',') as ('id' | 'content' | 'all')[]
      setSearchTypesState(types)
    }
  }, [searchParams])

  // Convert to preset format
  const toPresetFormat = useCallback(() => {
    const result: {
      filters: FilterCondition[]
      search_query?: string
      search_types?: string[]
    } = {
      filters: filters.filter(isFilterValid),
    }

    if (searchQuery) {
      result.search_query = searchQuery
    }
    if (searchTypes.length > 0) {
      result.search_types = searchTypes
    }

    return result
  }, [filters, isFilterValid, searchQuery, searchTypes])

  // Load from preset
  const loadFromPreset = useCallback(
    (preset: {
      filters?: FilterCondition[]
      search_query?: string
      search_types?: string[]
    }) => {
      if (preset.filters) {
        // Ensure each filter has an id
        setFiltersState(
          preset.filters.map((f) => ({
            ...f,
            id: f.id || nanoid(8),
          }))
        )
      }
      if (preset.search_query !== undefined) {
        setSearchQueryState(preset.search_query)
      }
      if (preset.search_types) {
        setSearchTypesState(
          preset.search_types as ('id' | 'content' | 'all')[]
        )
      }
    },
    []
  )

  // Memoized state checks
  const hasFilters = useMemo(() => filters.length > 0, [filters])
  const hasSearch = useMemo(() => searchQuery.length > 0, [searchQuery])

  return {
    filters,
    searchQuery,
    searchTypes,
    addFilter,
    updateFilter,
    removeFilter,
    clearFilters,
    setFilters,
    setSearchQuery,
    setSearchTypes,
    hasFilters,
    hasSearch,
    isFilterValid,
    getInvalidFilters,
    getUrlParams,
    loadFromUrl,
    toPresetFormat,
    loadFromPreset,
  }
}
