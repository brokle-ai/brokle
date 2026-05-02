import { rawFetch } from '@/lib/api/client'
import type { SessionUser } from '@/stores/auth-store'

// Conceptual port of web/src/features/authentication/api/auth-api.ts,
// rewritten against web-vite's `rawFetch` transport (shared error
// envelope, CSRF, 401 retry). The SDK/JS-style per-manager HTTP helper
// is intentionally NOT reintroduced — `rawFetch` is the one transport.

async function postJSON<T>(path: string, body: unknown): Promise<T> {
  const resp = await rawFetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  return (await resp.json()) as T
}

async function getJSON<T>(path: string): Promise<T> {
  const resp = await rawFetch(path, { method: 'GET' })
  return (await resp.json()) as T
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  user: SessionUser
  expires_at?: number
  expires_in?: number
}

export function login(data: LoginRequest): Promise<LoginResponse> {
  return postJSON<LoginResponse>('/api/v1/auth/login', data)
}

export interface SignupRequest {
  email: string
  password: string
  first_name: string
  last_name: string
  role: string
  organization_name?: string
  referral_source?: string
  invitation_token?: string
}

export function signup(data: SignupRequest): Promise<LoginResponse> {
  return postJSON<LoginResponse>('/api/v1/auth/signup', data)
}

export interface CompleteOAuthSignupRequest {
  session_id: string
  role: string
  organization_name?: string
  referral_source?: string
}

export function completeOAuthSignup(
  data: CompleteOAuthSignupRequest,
): Promise<LoginResponse> {
  return postJSON<LoginResponse>('/api/v1/auth/complete-oauth-signup', data)
}

export function exchangeLoginSession(sessionId: string): Promise<LoginResponse> {
  return postJSON<LoginResponse>(
    `/api/v1/auth/exchange-session/${encodeURIComponent(sessionId)}`,
    {},
  )
}

export function forgotPassword(email: string): Promise<void> {
  return postJSON<void>('/api/v1/auth/forgot-password', { email })
}

export function resetPassword(token: string, newPassword: string): Promise<void> {
  return postJSON<void>('/api/v1/auth/reset-password', {
    token,
    new_password: newPassword,
  })
}

export interface ValidateInvitationResponse {
  organization_id: string
  organization_name: string
  email: string
  role: string
  expires_at: string
  inviter_name: string
  is_expired: boolean
}

export function validateInvitation(
  token: string,
): Promise<ValidateInvitationResponse> {
  return getJSON<ValidateInvitationResponse>(
    `/api/v1/invitations/validate/${encodeURIComponent(token)}`,
  )
}

export interface AcceptInvitationResponse {
  organization_id: string
  organization_name: string
}

export function acceptInvitation(token: string): Promise<AcceptInvitationResponse> {
  return postJSON<AcceptInvitationResponse>('/api/v1/invitations/accept', {
    token,
  })
}

export function declineInvitation(token: string): Promise<void> {
  return postJSON<void>('/api/v1/invitations/decline', { token })
}
