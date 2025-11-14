'use client'

import { useEffect, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import { useTableSearchParams } from '@/hooks/use-table-search-params'
import { useTasks } from '../context/tasks-context'
import { tasks as allTasks } from '../data/tasks'
import type { Task } from '../data/schema'

export function useProjectTasks() {
  const searchParams = useSearchParams()
  const { projectSlug } = useTasks()
  const [data, setData] = useState<Task[]>([])
  const [totalCount, setTotalCount] = useState(0)
  const [isLoading, setIsLoading] = useState(false)

  const { page, pageSize, filter, status, priority, sortBy, sortOrder } =
    useTableSearchParams(searchParams)

  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true)
      try {
        // TODO: Replace with real API call when backend is ready
        // const result = await getProjectTasks({
        //   projectSlug,
        //   page,
        //   pageSize,
        //   filter,
        //   status,
        //   priority,
        //   sortBy,
        //   sortOrder,
        // })
        // setData(result.tasks)
        // setTotalCount(result.totalCount)

        // MOCK: Simulate server-side filtering, sorting, and pagination
        let filtered = [...allTasks]

        // Apply global filter
        if (filter) {
          const searchLower = filter.toLowerCase()
          filtered = filtered.filter(
            (task) =>
              task.id.toLowerCase().includes(searchLower) ||
              task.title.toLowerCase().includes(searchLower)
          )
        }

        // Apply status filter
        if (status.length > 0) {
          filtered = filtered.filter((task) => status.includes(task.status))
        }

        // Apply priority filter
        if (priority.length > 0) {
          filtered = filtered.filter((task) => priority.includes(task.priority))
        }

        // Apply sorting
        if (sortBy) {
          filtered.sort((a, b) => {
            const aVal = a[sortBy as keyof Task]
            const bVal = b[sortBy as keyof Task]

            if (aVal === undefined || bVal === undefined) return 0

            let comparison = 0
            if (typeof aVal === 'string' && typeof bVal === 'string') {
              comparison = aVal.localeCompare(bVal)
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

        // Apply pagination (slice to current page)
        const start = (page - 1) * pageSize
        const end = start + pageSize
        const paginated = filtered.slice(start, end)

        setData(paginated)
      } finally {
        setIsLoading(false)
      }
    }

    fetchData()
  }, [page, pageSize, filter, status, priority, sortBy, sortOrder, projectSlug])

  return { data, totalCount, isLoading }
}
