// Wire types for the organization members endpoint. Shape comes from
// internal/transport/http/handlers/organization/types.go `memberResponse`
// and `listMembersBody`. The endpoint is NOT paginated — the server
// returns the full member list per call along with the count. If that
// changes, swap `MemberListResponse` for the standard paginated shape
// used in organizations/queries.ts.

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
