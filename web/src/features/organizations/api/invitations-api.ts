// Invitations API - Invitation management endpoints
// Full invitation lifecycle: create, list, resend, revoke, accept

import { BrokleAPIClient } from '@/lib/api/core/client'

// API response types matching backend
export interface InvitationAPIResponse {
  id: string
  email: string
  status: 'pending' | 'accepted' | 'expired' | 'revoked'
  token_preview: string
  role_id: string
  role_name: string
  message?: string
  invited_by_id: string
  invited_by_email: string
  invited_by_name: string
  expires_at: string
  accepted_at?: string
  resent_count: number
  created_at: string
  updated_at: string
}

export interface UserInvitationAPIResponse {
  id: string
  email: string
  status: 'pending' | 'accepted' | 'expired' | 'revoked'
  role_name: string
  message?: string
  organization_id: string
  organization_name: string
  invited_by_name: string
  expires_at: string
  created_at: string
}

export interface AcceptInvitationResponse {
  organization_id: string
  organization_name: string
  role_name: string
  message: string
}

// Backend InvitationDetailsResponse structure
export interface ValidateTokenResponse {
  valid: boolean // Added for frontend convenience - derived from !is_expired
  organization_id: string
  organization_name: string
  email: string
  role_name: string // Backend sends 'role', we alias it
  expires_at: string
  invited_by_name: string // Backend sends 'inviter_name', we alias it
  is_expired?: boolean
}

// Frontend types
export interface Invitation {
  id: string
  email: string
  status: 'pending' | 'accepted' | 'expired' | 'revoked'
  tokenPreview: string
  roleId: string
  roleName: string
  message?: string
  invitedById: string
  invitedByEmail: string
  invitedByName: string
  expiresAt: Date
  acceptedAt?: Date
  resentCount: number
  createdAt: Date
  updatedAt: Date
}

export interface UserInvitation {
  id: string
  email: string
  status: 'pending' | 'accepted' | 'expired' | 'revoked'
  roleName: string
  message?: string
  organizationId: string
  organizationName: string
  invitedByName: string
  expiresAt: Date
  createdAt: Date
}

export interface RoleAPIResponse {
  id: string
  name: string
  display_name: string
  description: string
  scope_type: string
  scopes: string[]
  is_system_role: boolean
  created_at: string
  updated_at: string
}

export interface Role {
  id: string
  name: string
  displayName: string
  description: string
  scopeType: string
  scopes: string[]
  isSystemRole: boolean
}

// Client instance
const client = new BrokleAPIClient('/api')

/**
 * Create an invitation to join an organization
 * @param organizationId - Organization ID
 * @param data.email - Email address to invite
 * @param data.role_id - Role ID to assign
 * @param data.message - Optional personal message
 */
export const createInvitation = async (
  organizationId: string,
  data: {
    email: string
    role_id: string
    message?: string
  }
): Promise<Invitation> => {
  const response = await client.post<InvitationAPIResponse>(
    `/v1/organizations/${organizationId}/invitations`,
    data,
    {
      includeOrgContext: true,
      customOrgId: organizationId
    }
  )
  return mapInvitationFromAPI(response)
}

/**
 * Get pending invitations for an organization
 * @param organizationId - Organization ID
 */
export const getPendingInvitations = async (
  organizationId: string
): Promise<Invitation[]> => {
  const response = await client.get<{ invitations: InvitationAPIResponse[]; total: number }>(
    `/v1/organizations/${organizationId}/invitations`,
    { status: 'pending' },
    {
      includeOrgContext: true,
      customOrgId: organizationId
    }
  )

  return response.invitations.map(mapInvitationFromAPI)
}

/**
 * Resend an invitation email
 * @param organizationId - Organization ID
 * @param invitationId - Invitation ID
 */
export const resendInvitation = async (
  organizationId: string,
  invitationId: string
): Promise<Invitation> => {
  const response = await client.post<InvitationAPIResponse>(
    `/v1/organizations/${organizationId}/invitations/${invitationId}/resend`,
    {},
    {
      includeOrgContext: true,
      customOrgId: organizationId
    }
  )
  return mapInvitationFromAPI(response)
}

/**
 * Revoke an invitation
 * @param organizationId - Organization ID
 * @param invitationId - Invitation ID
 */
export const revokeInvitation = async (
  organizationId: string,
  invitationId: string
): Promise<void> => {
  await client.delete(
    `/v1/organizations/${organizationId}/invitations/${invitationId}`,
    {
      includeOrgContext: true,
      customOrgId: organizationId
    }
  )
}

/**
 * Accept an invitation (requires user to be logged in)
 * @param token - Invitation token from email link
 */
export const acceptInvitation = async (
  token: string
): Promise<AcceptInvitationResponse> => {
  return await client.post<AcceptInvitationResponse>(
    '/v1/invitations/accept',
    { token }
  )
}

/**
 * Get invitations for the current user
 */
export const getUserInvitations = async (): Promise<UserInvitation[]> => {
  const response = await client.get<{ invitations: UserInvitationAPIResponse[]; total: number }>(
    '/v1/invitations'
  )

  return response.invitations.map(mapUserInvitationFromAPI)
}

// Backend response type for validation
interface BackendValidateResponse {
  organization_id: string
  organization_name: string
  email: string
  role: string
  expires_at: string
  inviter_name: string
  is_expired: boolean
}

/**
 * Validate an invitation token (public endpoint, no auth required)
 * @param token - Invitation token from email link
 */
export const validateInvitationToken = async (
  token: string
): Promise<ValidateTokenResponse> => {
  const response = await client.get<BackendValidateResponse>(
    `/v1/invitations/validate/${token}`
  )

  // Transform backend response to frontend format
  return {
    valid: !response.is_expired,
    organization_id: response.organization_id,
    organization_name: response.organization_name,
    email: response.email,
    role_name: response.role,
    expires_at: response.expires_at,
    invited_by_name: response.inviter_name,
    is_expired: response.is_expired,
  }
}

/**
 * Get available roles for member invitation
 * Returns organization-scoped roles that can be assigned to new members
 */
export const getAvailableRolesForInvitation = async (): Promise<Role[]> => {
  const response = await client.get<RoleAPIResponse[]>(
    '/v1/rbac/roles',
    { scope_type: 'organization' }
  )

  // Filter to roles that can be assigned (exclude owner - only one owner allowed)
  return response
    .map(mapRoleFromAPI)
    .filter(role => role.name !== 'owner')
}

// Mapping functions
const mapRoleFromAPI = (apiRole: RoleAPIResponse): Role => {
  return {
    id: apiRole.id,
    name: apiRole.name,
    displayName: apiRole.display_name,
    description: apiRole.description,
    scopeType: apiRole.scope_type,
    scopes: apiRole.scopes || [],
    isSystemRole: apiRole.is_system_role,
  }
}

const mapInvitationFromAPI = (apiInvitation: InvitationAPIResponse): Invitation => {
  return {
    id: apiInvitation.id,
    email: apiInvitation.email,
    status: apiInvitation.status,
    tokenPreview: apiInvitation.token_preview,
    roleId: apiInvitation.role_id,
    roleName: apiInvitation.role_name,
    message: apiInvitation.message,
    invitedById: apiInvitation.invited_by_id,
    invitedByEmail: apiInvitation.invited_by_email,
    invitedByName: apiInvitation.invited_by_name,
    expiresAt: new Date(apiInvitation.expires_at),
    acceptedAt: apiInvitation.accepted_at ? new Date(apiInvitation.accepted_at) : undefined,
    resentCount: apiInvitation.resent_count,
    createdAt: new Date(apiInvitation.created_at),
    updatedAt: new Date(apiInvitation.updated_at),
  }
}

const mapUserInvitationFromAPI = (apiInvitation: UserInvitationAPIResponse): UserInvitation => {
  return {
    id: apiInvitation.id,
    email: apiInvitation.email,
    status: apiInvitation.status,
    roleName: apiInvitation.role_name,
    message: apiInvitation.message,
    organizationId: apiInvitation.organization_id,
    organizationName: apiInvitation.organization_name,
    invitedByName: apiInvitation.invited_by_name,
    expiresAt: new Date(apiInvitation.expires_at),
    createdAt: new Date(apiInvitation.created_at),
  }
}
