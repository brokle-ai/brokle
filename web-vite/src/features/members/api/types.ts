// Wire types for the organization members and invitations endpoints.
// Shapes come from internal/transport/http/handlers/organization/types.go
// — `memberResponse`, `listMembersBody`, `invitationResponse`,
// `listPendingInvitationsBody`, `createInvitationBody`.
//
// The member-list endpoint is NOT paginated — it returns the full
// members array per call with a count.

export interface MemberListItem {
  user_id: string
  role_id: string
  status: string
  invited_by?: string
  // RFC 3339 timestamps on the wire; components format on render.
  joined_at: string
  created_at: string
  updated_at: string
}

export interface MemberListResponse {
  members: MemberListItem[]
  total: number
}

// Role hydration shape returned alongside invitations (and used to
// populate the invite-role select via GET /api/v1/rbac/roles).
export interface RoleRef {
  id: string
  name: string
  display_name?: string
}

export interface InviterRef {
  id: string
  email: string
  first_name?: string
  last_name?: string
}

export interface Invitation {
  id: string
  organization_id: string
  email: string
  status: 'pending' | 'accepted' | 'expired' | 'revoked'
  token_preview?: string
  role_id: string
  role?: RoleRef
  inviter?: InviterRef
  invited_by_id?: string
  message?: string
  resent_count: number
  resent_at?: string
  expires_at: string
  created_at: string
  updated_at: string
}

export interface InvitationListResponse {
  invitations: Invitation[]
  total: number
}

export interface InviteMemberRequest {
  email: string
  role_id: string
  message?: string
}

// Response of GET /api/v1/rbac/roles. `auth.Role` carries more fields
// than the invite UI uses — narrow here to the discriminators the
// select needs.
export interface RoleListItem {
  id: string
  name: string
  scope_type: string
  description?: string
}

export interface RoleListResponse {
  data: RoleListItem[]
}
