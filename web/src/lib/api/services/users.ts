// Users API - Latest endpoints for dashboard application
// Direct functions using stable user management endpoints

import { BrokleAPIClient } from '../core/client'
import type { 
  PaginatedResponse,
  QueryParams
} from '../core/types'
import type { User } from '@/features/users/data/schema'

// User management types
export interface CreateUserData {
  firstName: string
  lastName: string
  email: string
  role: string
}

// Flexible base client - versions specified per endpoint
const client = new BrokleAPIClient('/api')

// Direct user management functions
export const getUsers = async (params?: QueryParams): Promise<PaginatedResponse<User>> => {
    return client.getPaginated<User>('/v1/users', params, { // v1: Stable user listing
      includeOrgContext: true,
      includeProjectContext: false,
    })
  }

export const getUser = async (userId: string): Promise<User> => {
    return client.get<User>(`/v1/users/${userId}`, {}, { // v1: Stable user profile
      includeOrgContext: true,
    })
  }

export const createUser = async (userData: CreateUserData): Promise<User> => {
    return client.post<User>('/v1/users', userData, {
      includeOrgContext: true,
    })
  }

export const updateUser = async (userId: string, userData: Partial<User>): Promise<User> => {
    return client.put<User>(`/users/${userId}`, userData, {
      includeOrgContext: true,
    })
  }

export const deleteUser = async (userId: string): Promise<void> => {
    return client.delete<void>(`/users/${userId}`, {
      includeOrgContext: true,
    })
  }

export const resendInvitation = async (userId: string): Promise<void> => {
    return client.post<void>(`/users/${userId}/resend-invitation`, {}, {
      includeOrgContext: true,
    })
  }

export const changeUserStatus = async (
    userId: string, 
    status: 'active' | 'inactive' | 'suspended'
  ): Promise<User> => {
    return client.patch<User>(`/users/${userId}/status`, { status }, {
      includeOrgContext: true,
    })
  }

