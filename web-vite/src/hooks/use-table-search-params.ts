// Port of web/'s useTableSearchParams (which reads Next.js
// ReadonlyURLSearchParams) adapted to TanStack Router. The consumer
// passes any object-shaped search params and we normalize into the
// TableSearchParams contract used by ported features.

import { useMemo } from 'react'

export interface TableSearchParams {
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

type SearchValue = string | string[] | number | boolean | undefined | null
type SearchInput = URLSearchParams | Record<string, SearchValue>

function readString(src: SearchInput, key: string): string | null {
  if (src instanceof URLSearchParams) return src.get(key)
  const v = src[key]
  if (v === undefined || v === null) return null
  if (Array.isArray(v)) return v[0] ?? null
  return String(v)
}

function parseJsonArray(raw: string | null): string[] {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed.map(String) : []
  } catch {
    return []
  }
}

export function parseTableSearchParams(src: SearchInput): TableSearchParams {
  const page = Math.max(1, parseInt(readString(src, 'page') ?? '1', 10))
  const pageSize = Math.max(1, parseInt(readString(src, 'pageSize') ?? '50', 10))
  const filter = readString(src, 'filter') ?? ''
  const status = parseJsonArray(readString(src, 'status'))
  const priority = parseJsonArray(readString(src, 'priority'))
  const type = parseJsonArray(readString(src, 'type'))
  const sortBy = readString(src, 'sortBy')
  const sortOrderRaw = readString(src, 'sortOrder')
  const sortOrder =
    sortOrderRaw === 'asc' || sortOrderRaw === 'desc' ? sortOrderRaw : null
  const peek = readString(src, 'peek')
  const tab = readString(src, 'tab')

  return { page, pageSize, filter, status, priority, type, sortBy, sortOrder, peek, tab }
}

export function useTableSearchParams(src: SearchInput): TableSearchParams {
  return useMemo(() => parseTableSearchParams(src), [src])
}
