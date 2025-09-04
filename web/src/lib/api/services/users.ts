import { BrokleAPIClient } from '../core/client'
import type { 
  PaginatedResponse,
  QueryParams
} from '../core/types'
import type { User } from '@/features/users/data/schema'

export class UsersAPIClient extends BrokleAPIClient {
  constructor() {
    super('/users') // All user endpoints will be prefixed with /users
  }

  /**
   * Get paginated list of users
   * @param params Query parameters for filtering, sorting, and pagination
   * @returns Promise<PaginatedResponse<User>>
   */
  async getUsers(params?: QueryParams): Promise<PaginatedResponse<User>> {
    return this.getPaginated<User>('/v1/users', params, {
      includeOrgContext: true, // Include organization context headers
      includeProjectContext: false, // Users are org-level, not project-level
    })
  }

  /**
   * Get a specific user by ID
   * @param userId User ID
   * @returns Promise<User>
   */
  async getUser(userId: string): Promise<User> {
    return this.get<User>(`/v1/users/${userId}`, {}, {
      includeOrgContext: true,
    })
  }

  /**
   * Create a new user
   * @param userData User data
   * @returns Promise<User>
   */
  async createUser(userData: {
    firstName: string
    lastName: string
    email: string
    role: string
  }): Promise<User> {
    return this.post<User>('/v1/users', userData, {
      includeOrgContext: true,
    })
  }

  /**
   * Update an existing user
   * @param userId User ID
   * @param userData Updated user data
   * @returns Promise<User>
   */
  async updateUser(userId: string, userData: Partial<User>): Promise<User> {
    return this.put<User>(`/v1/users/${userId}`, userData, {
      includeOrgContext: true,
    })
  }

  /**
   * Delete a user
   * @param userId User ID
   * @returns Promise<void>
   */
  async deleteUser(userId: string): Promise<void> {
    return this.delete<void>(`/v1/users/${userId}`, {
      includeOrgContext: true,
    })
  }

  /**
   * Invite a new user to the organization
   * @param inviteData Invitation data
   * @returns Promise<User>
   */
  async inviteUser(inviteData: {
    email: string
    role: string
    firstName?: string
    lastName?: string
  }): Promise<User> {
    return this.post<User>('/v1/users/invite', inviteData, {
      includeOrgContext: true,
    })
  }

  /**
   * Resend invitation to a user
   * @param userId User ID
   * @returns Promise<void>
   */
  async resendInvitation(userId: string): Promise<void> {
    return this.post<void>(`/v1/users/${userId}/resend-invitation`, {}, {
      includeOrgContext: true,
    })
  }

  /**
   * Change user status (active, inactive, suspended)
   * @param userId User ID
   * @param status New status
   * @returns Promise<User>
   */
  async changeUserStatus(userId: string, status: 'active' | 'inactive' | 'suspended'): Promise<User> {
    return this.patch<User>(`/v1/users/${userId}/status`, { status }, {
      includeOrgContext: true,
    })
  }
}