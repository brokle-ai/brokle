// Members API - Organization member management endpoints

import { BrokleAPIClient } from '@/lib/api/core/client'
import type { PaginatedResponse } from '@/lib/api/core/types'

// API response type matching backend OrganizationMember
export interface MemberAPIResponse {
  user_id: string
  email: string
  first_name: string
  last_name: string
  role: string
  status: string
  joined_at: string
  invited_by?: string
  created_at: string
  updated_at: string
}

// Frontend member type
export interface Member {
  id: string
  userId: string
  email: string
  firstName: string
  lastName: string
  name: string
  role: 'owner' | 'admin' | 'developer' | 'viewer'
  status: string
  joinedAt: Date
  invitedBy?: string
  createdAt: Date
  updatedAt: Date
}

// Client instance
const client = new BrokleAPIClient('/api')

/**
 * Get all members of an organization
 * @param organizationId - Organization ID
 * @param page - Page number
 * @param limit - Items per page
 */
export const getOrganizationMembers = async (
  organizationId: string,
  page = 1,
  limit = 50
): Promise<PaginatedResponse<Member>> => {
  const response = await client.getPaginated<MemberAPIResponse>(
    `/v1/organizations/${organizationId}/members`,
    { page, limit },
    {
      includeOrgContext: true,
      customOrgId: organizationId
    }
  )

  return {
    data: response.data.map(mapMemberFromAPI),
    pagination: response.pagination,
  }
}

/**
 * Remove a member from an organization
 * @param organizationId - Organization ID
 * @param userId - User ID to remove
 */
export const removeMember = async (
  organizationId: string,
  userId: string
): Promise<void> => {
  await client.delete(
    `/v1/organizations/${organizationId}/members/${userId}`,
    {
      includeOrgContext: true,
      customOrgId: organizationId
    }
  )
}

/**
 * Update a member's role
 * @param organizationId - Organization ID
 * @param userId - User ID
 * @param roleId - New role ID
 */
export const updateMemberRole = async (
  organizationId: string,
  userId: string,
  roleId: string
): Promise<void> => {
  await client.patch(
    `/v1/organizations/${organizationId}/members/${userId}`,
    { role_id: roleId },
    {
      includeOrgContext: true,
      customOrgId: organizationId
    }
  )
}

// Mapping function
const mapMemberFromAPI = (apiMember: MemberAPIResponse): Member => {
  const firstName = apiMember.first_name || ''
  const lastName = apiMember.last_name || ''
  const name = `${firstName} ${lastName}`.trim() || apiMember.email

  return {
    id: apiMember.user_id, // Use user_id as the member ID for actions
    userId: apiMember.user_id,
    email: apiMember.email,
    firstName,
    lastName,
    name,
    role: apiMember.role as 'owner' | 'admin' | 'developer' | 'viewer',
    status: apiMember.status,
    joinedAt: new Date(apiMember.joined_at),
    invitedBy: apiMember.invited_by,
    createdAt: new Date(apiMember.created_at),
    updatedAt: new Date(apiMember.updated_at),
  }
}
