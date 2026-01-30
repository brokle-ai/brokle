import { useMemo } from 'react'
import type { ReadonlyURLSearchParams } from 'next/navigation'

export type TableSearchParams = {
  page: number
  pageSize: number
  filter: string
  status: string[]
  priority: string[]
  type: string[]
  sortBy: string | null
  sortOrder: 'asc' | 'desc' | null
  peek: string | null
  tab: string | null
}

type SearchParamsInput = ReadonlyURLSearchParams | Record<string, string | string[] | undefined>

/**
 * Parse table search params from URL
 * Works with both client (ReadonlyURLSearchParams) and server (plain object) sources
 */
export function parseTableSearchParams(searchParams: SearchParamsInput): TableSearchParams {
  const get = (key: string): string | null => {
    if (searchParams instanceof URLSearchParams) {
      return searchParams.get(key)
    }
    const value = searchParams[key]
    if (Array.isArray(value)) {
      return value[0] ?? null
    }
    return value ?? null
  }

  // Parse pagination
  const page = Math.max(1, parseInt(get('page') ?? '1', 10))
  const pageSize = Math.max(1, parseInt(get('pageSize') ?? '50', 10))

  // Parse global filter
  const filter = get('filter') ?? ''

  // Parse status filter (JSON array)
  let status: string[] = []
  const statusParam = get('status')
  if (statusParam) {
    try {
      const parsed = JSON.parse(statusParam)
      status = Array.isArray(parsed) ? parsed : []
    } catch {
      status = []
    }
  }

  // Parse priority filter (JSON array)
  let priority: string[] = []
  const priorityParam = get('priority')
  if (priorityParam) {
    try {
      const parsed = JSON.parse(priorityParam)
      priority = Array.isArray(parsed) ? parsed : []
    } catch {
      priority = []
    }
  }

  // Parse type filter (JSON array) - for prompts
  let type: string[] = []
  const typeParam = get('type')
  if (typeParam) {
    try {
      const parsed = JSON.parse(typeParam)
      type = Array.isArray(parsed) ? parsed : []
    } catch {
      type = []
    }
  }

  // Parse sorting
  const sortBy = get('sortBy')
  const sortOrderParam = get('sortOrder')
  const sortOrder = sortOrderParam === 'asc' || sortOrderParam === 'desc' ? sortOrderParam : null

  // Parse peek and tab parameters
  const peek = get('peek')
  const tab = get('tab')

  return {
    page,
    pageSize,
    filter,
    status,
    priority,
    type,
    sortBy,
    sortOrder,
    peek,
    tab,
  }
}

/**
 * Hook version for client components
 * Accepts ReadonlyURLSearchParams from useSearchParams()
 * Memoized to prevent infinite loops from array reference changes
 */
export function useTableSearchParams(searchParams: ReadonlyURLSearchParams): TableSearchParams {
  return useMemo(() => parseTableSearchParams(searchParams), [searchParams])
}
