'use client'

import { useState, useEffect, useCallback } from 'react'
import { useSearchParams } from 'next/navigation'
import { User } from '@/features/users/data/schema'
import type { PaginatedResponse, Pagination } from '@/lib/api/core/types'
import { api } from '@/lib/api'
import { BrokleAPIError } from '@/lib/api/core/types'

// Internal pagination format for UI components
interface PaginationMeta {
  page: number
  pageSize: number
  total: number
  totalPages: number
  hasNextPage: boolean
  hasPreviousPage: boolean
}

// Convert API pagination to internal format
const fromApiPagination = (apiPagination: Pagination): PaginationMeta => ({
  page: apiPagination.page,
  pageSize: apiPagination.limit, // API uses 'limit', UI uses 'pageSize'
  total: apiPagination.total,
  totalPages: apiPagination.totalPages,
  hasNextPage: apiPagination.hasNext,
  hasPreviousPage: apiPagination.hasPrev,
})

type UsersResponse = PaginatedResponse<User>

interface UseUsersDataReturn {
  data: User[]
  pagination: PaginationMeta | null
  loading: boolean
  error: string | null
  refetch: () => void
}

export function useUsersData(): UseUsersDataReturn {
  const searchParams = useSearchParams()
  const [data, setData] = useState<User[]>([])
  const [pagination, setPagination] = useState<PaginationMeta | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchUsers = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      
      // Build query params object for API call
      const queryParams: Record<string, any> = {}
      
      // Add pagination params
      const page = searchParams.get('page') || '1'
      const pageSize = searchParams.get('pageSize') || '10'
      queryParams.page = parseInt(page, 10)
      queryParams.limit = parseInt(pageSize, 10) // API uses 'limit' not 'pageSize'
      
      // Add filter params
      const username = searchParams.get('username')
      if (username) queryParams.username = username
      
      const status = searchParams.get('status')
      if (status) queryParams.status = status
      
      const role = searchParams.get('role')
      if (role) queryParams.role = role
      
      // Add sorting params
      const sortBy = searchParams.get('sortBy')
      if (sortBy) queryParams.sortBy = sortBy
      
      const sortOrder = searchParams.get('sortOrder')
      if (sortOrder) queryParams.sortOrder = sortOrder
      
      // Use BrokleAPIClient for proper error handling and authentication
      const result: UsersResponse = await api.users.getUsers(queryParams)
      
      setData(result.data)
      setPagination(fromApiPagination(result.pagination))
    } catch (err) {
      let errorMessage = 'Failed to fetch users'
      
      if (err instanceof BrokleAPIError) {
        // Use user-friendly error message from API
        errorMessage = err.message
        
        // Log detailed error info in development
        if (process.env.NODE_ENV === 'development') {
          console.error('[useUsersData] API Error:', {
            code: err.code,
            message: err.message,
            details: err.details,
            requestId: err.requestId,
            statusCode: err.statusCode
          })
        }
      } else if (err instanceof Error) {
        errorMessage = err.message
      }
      
      setError(errorMessage)
      setData([])
      setPagination(null)
    } finally {
      setLoading(false)
    }
  }, [searchParams])

  // Fetch data when search params change
  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  const refetch = () => {
    fetchUsers()
  }

  return {
    data,
    pagination,
    loading,
    error,
    refetch,
  }
}