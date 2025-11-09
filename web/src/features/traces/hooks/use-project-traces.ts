'use client'

import { useEffect, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useTraces } from '../context/traces-context'
import { traces as allTraces } from '../data/traces'
import type { Trace } from '../data/schema'

export function useProjectTraces() {
  const searchParams = useSearchParams()
  const { projectSlug } = useTraces()
  const [data, setData] = useState<Trace[]>([])
  const [totalCount, setTotalCount] = useState(0)
  const [isLoading, setIsLoading] = useState(false)

  const { page, pageSize, filter, status, sortBy, sortOrder } =
    useTableSearchParams(searchParams)

  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true)
      try {
        // TODO: Replace with real API call when ready
        // const result = await getProjectTraces({
        //   projectSlug,
        //   page,
        //   pageSize,
        //   filter,
        //   status,
        //   sortBy,
        //   sortOrder,
        // })

        // MOCK: Simulate server-side filtering, sorting, and pagination
        let filtered = [...allTraces]

        // Apply global filter (search by ID or name)
        if (filter) {
          const searchLower = filter.toLowerCase()
          filtered = filtered.filter(
            (trace) =>
              trace.id.toLowerCase().includes(searchLower) ||
              trace.name.toLowerCase().includes(searchLower)
          )
        }

        // Apply status filter
        if (status.length > 0) {
          filtered = filtered.filter((trace) => status.includes(trace.status))
        }

        // Apply sorting
        if (sortBy) {
          filtered.sort((a, b) => {
            const aVal = a[sortBy as keyof Trace]
            const bVal = b[sortBy as keyof Trace]

            if (aVal === undefined || bVal === undefined) return 0

            let comparison = 0
            if (aVal instanceof Date && bVal instanceof Date) {
              comparison = aVal.getTime() - bVal.getTime()
            } else if (typeof aVal === 'string' && typeof bVal === 'string') {
              comparison = aVal.localeCompare(bVal)
            } else if (typeof aVal === 'number' && typeof bVal === 'number') {
              comparison = aVal - bVal
            } else if (aVal < bVal) {
              comparison = -1
            } else if (aVal > bVal) {
              comparison = 1
            }

            return sortOrder === 'desc' ? -comparison : comparison
          })
        }

        // Store total count after filtering
        const total = filtered.length
        setTotalCount(total)

        // Apply pagination
        const start = (page - 1) * pageSize
        const end = start + pageSize
        const paginated = filtered.slice(start, end)

        setData(paginated)
      } finally {
        setIsLoading(false)
      }
    }

    fetchData()
  }, [page, pageSize, filter, status, sortBy, sortOrder, projectSlug])

  return { data, totalCount, isLoading }
}
