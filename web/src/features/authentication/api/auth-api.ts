// Auth API - Clean implementation without over-engineered abstractions
import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  AuthResponse,
  LoginCredentials,
  SignUpCredentials,
  User,
  Organization,
  LoginResponse,
  UserResponse
} from '../types'

// Simple API client
const client = new BrokleAPIClient('/api')

export const login = async (credentials: LoginCredentials): Promise<AuthResponse> => {
  if (process.env.NODE_ENV === 'development') {
    console.debug('[AuthAPI] Login called with credentials:', { email: credentials.email })
  }

  // Backend now returns: { user, expires_at, expires_in } (tokens in httpOnly cookies)
  const backendResponse = await client.post<{
    user: UserResponse
    expires_at: number  // Milliseconds
    expires_in: number  // Milliseconds
  }>(
    '/v1/auth/login',
    credentials,
    { skipAuth: true }
  )

  if (process.env.NODE_ENV === 'development') {
    console.debug('[AuthAPI] Login response received:', {
      hasUser: !!backendResponse.user,
      hasExpiresAt: !!backendResponse.expires_at,
      backendResponse
    })
  }

  // Defensive check
  if (!backendResponse || !backendResponse.user) {
    console.error('[AuthAPI] Invalid login response - missing user:', backendResponse)
    throw new Error('Login response missing user data')
  }

  // Map backend user response to frontend format
  const user: User = {
    id: backendResponse.user.id,
    email: backendResponse.user.email,
    firstName: backendResponse.user.first_name,
    lastName: backendResponse.user.last_name,
    name: `${backendResponse.user.first_name} ${backendResponse.user.last_name}`.trim(),
    role: 'user',
    organizationId: '',
    defaultOrganizationId: backendResponse.user.default_organization_id,
    projects: [],
    createdAt: backendResponse.user.created_at,
    updatedAt: backendResponse.user.created_at,
    isEmailVerified: backendResponse.user.is_email_verified,
    onboardingCompletedAt: backendResponse.user.onboarding_completed_at,
  }

  if (process.env.NODE_ENV === 'development') {
    console.debug('[AuthAPI] User mapped:', { userId: user.id, email: user.email })
  }

  // Get organization from backend (cookies sent automatically, no manual auth header)
  let organization: Organization
  try {
    if (process.env.NODE_ENV === 'development') {
      console.debug('[AuthAPI] Fetching organizations...')
    }

    const orgResponse = await client.get<Array<{
      id: string
      name: string
      billing_email: string
      subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
      created_at: string
      updated_at: string
    }>>('/v1/organizations')

    if (process.env.NODE_ENV === 'development') {
      console.debug('[AuthAPI] Organization response:', {
        isArray: Array.isArray(orgResponse),
        length: Array.isArray(orgResponse) ? orgResponse.length : 0,
        orgResponse
      })
    }

    // Select organization based on user's default_organization_id preference
    let selectedOrg = Array.isArray(orgResponse) && orgResponse.length > 0 ? orgResponse[0] : null

    // If user has a default organization preference, find it in the list
    if (user.defaultOrganizationId && Array.isArray(orgResponse)) {
      const defaultOrg = orgResponse.find(org => org.id === user.defaultOrganizationId)
      if (defaultOrg) {
        selectedOrg = defaultOrg
        console.log('[AuthAPI] Using default organization:', { id: defaultOrg.id, name: defaultOrg.name })
      } else {
        console.log('[AuthAPI] Default org not found in user orgs, using first org')
      }
    }

    if (!selectedOrg) {
      console.error('[AuthAPI] No organizations found in response')
      throw new Error('No organizations found for user')
    }

    if (process.env.NODE_ENV === 'development') {
      console.debug('[AuthAPI] First organization:', { id: selectedOrg.id, name: selectedOrg.name })
    }

    organization = {
      id: selectedOrg.id,
      name: selectedOrg.name,
      plan: selectedOrg.subscription_plan,
      members: [{
        userId: user.id,
        user: user,
        role: 'owner',
        joinedAt: new Date().toISOString(),
      }],
      apiKeys: [],
      usage: {
        requests_this_month: 0,
        cost_this_month: 0,
        models_used: 0,
      },
      createdAt: selectedOrg.created_at,
      updatedAt: selectedOrg.updated_at,
    }
  } catch (orgError) {
    console.error('[AuthAPI] Failed to fetch organization during login:', orgError)
    throw orgError
  }

  const authResponse = {
    user,
    organization,
    expiresAt: backendResponse.expires_at,  // Milliseconds
    expiresIn: backendResponse.expires_in,  // Milliseconds
  }

  if (process.env.NODE_ENV === 'development') {
    console.debug('[AuthAPI] Login complete, returning:', {
      hasUser: !!authResponse.user,
      hasOrg: !!authResponse.organization,
      userId: authResponse.user?.id,
      orgId: authResponse.organization?.id
    })
  }

  return authResponse
}

export const signup = async (credentials: SignUpCredentials): Promise<AuthResponse> => {
  // Map frontend format to backend format
  const backendPayload = {
    first_name: credentials.firstName,
    last_name: credentials.lastName,
    email: credentials.email,
    password: credentials.password,
    role: credentials.role,
    organization_name: credentials.organizationName,
    referral_source: credentials.referralSource,
    invitation_token: credentials.invitationToken,
  }

  // Backend now returns: { user, expires_at, expires_in } (tokens in httpOnly cookies)
  const backendResponse = await client.post<{
    user: UserResponse
    expires_at: number  // Milliseconds
    expires_in: number  // Milliseconds
  }>(
    '/v1/auth/signup',
    backendPayload,
    { skipAuth: true }
  )

  // Map backend user response to frontend format
  const user: User = {
    id: backendResponse.user.id,
    email: backendResponse.user.email,
    firstName: backendResponse.user.first_name,
    lastName: backendResponse.user.last_name,
    name: `${backendResponse.user.first_name} ${backendResponse.user.last_name}`.trim(),
    role: 'user',
    organizationId: '',
    defaultOrganizationId: backendResponse.user.default_organization_id,
    projects: [],
    createdAt: backendResponse.user.created_at,
    updatedAt: backendResponse.user.created_at,
    isEmailVerified: backendResponse.user.is_email_verified,
    onboardingCompletedAt: backendResponse.user.onboarding_completed_at,
  }

  // Get organization from backend (cookies sent automatically, no manual auth header)
  let organization: Organization
  try {
    const orgResponse = await client.get<Array<{
      id: string
      name: string
      billing_email: string
      subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
      created_at: string
      updated_at: string
    }>>('/v1/organizations')

    // Select organization based on user's default_organization_id preference
    let selectedOrg = Array.isArray(orgResponse) && orgResponse.length > 0 ? orgResponse[0] : null

    // If user has a default organization preference, find it in the list
    if (user.defaultOrganizationId && Array.isArray(orgResponse)) {
      const defaultOrg = orgResponse.find(org => org.id === user.defaultOrganizationId)
      if (defaultOrg) {
        selectedOrg = defaultOrg
      }
    }

    if (!selectedOrg) {
      throw new Error('No organizations found for user')
    }

    organization = {
      id: selectedOrg.id,
      name: selectedOrg.name,
      plan: selectedOrg.subscription_plan,
      members: [],
      apiKeys: [],
      usage: {
        requests_this_month: 0,
        cost_this_month: 0,
        models_used: 0,
      },
      createdAt: selectedOrg.created_at,
      updatedAt: selectedOrg.updated_at,
    }
  } catch (orgError) {
    console.error('[AuthAPI] Failed to fetch organization during signup:', orgError)
    throw orgError
  }

  return {
    user,
    organization,
    expiresAt: backendResponse.expires_at,  // Milliseconds
    expiresIn: backendResponse.expires_in,  // Milliseconds
  }
}

export const logout = async (): Promise<void> => {
  try {
    await client.post('/v1/auth/logout', {})
  } catch (error) {
    console.warn('Logout request failed:', error)
  }
}

export const getCurrentUser = async (): Promise<User> => {
  const userResponse = await client.get<UserResponse>('/v1/users/me')

  return {
    id: userResponse.id,
    email: userResponse.email,
    firstName: userResponse.first_name,
    lastName: userResponse.last_name,
    name: `${userResponse.first_name} ${userResponse.last_name}`.trim(),
    role: 'user',
    organizationId: '',
    defaultOrganizationId: userResponse.default_organization_id,
    projects: [],
    createdAt: userResponse.created_at,
    updatedAt: userResponse.created_at,
    isEmailVerified: userResponse.is_email_verified,
    onboardingCompletedAt: userResponse.onboarding_completed_at,
    organizations: userResponse.organizations || [],
  }
}

export const updateProfile = async (data: Partial<User>): Promise<User> => {
  const backendData = {
    first_name: data.firstName,
    last_name: data.lastName,
  }

  const userResponse = await client.patch<UserResponse>('/v1/users/me', backendData)
  
  return {
    id: userResponse.id,
    email: userResponse.email,
    firstName: userResponse.first_name,
    lastName: userResponse.last_name,
    name: `${userResponse.first_name} ${userResponse.last_name}`.trim(),
    role: 'user',
    organizationId: '',
    defaultOrganizationId: userResponse.default_organization_id,
    projects: [],
    createdAt: userResponse.created_at,
    updatedAt: userResponse.created_at,
    isEmailVerified: userResponse.is_email_verified,
    onboardingCompletedAt: userResponse.onboarding_completed_at,
  }
}

export const changePassword = async (currentPassword: string, newPassword: string): Promise<void> => {
  await client.patch('/v1/auth/change-password', {
    current_password: currentPassword,
    new_password: newPassword,
  })
}

export const requestPasswordReset = async (email: string): Promise<void> => {
  await client.post('/v1/auth/forgot-password', { email }, { skipAuth: true })
}

export const confirmPasswordReset = async (token: string, password: string): Promise<void> => {
  await client.post(
    '/v1/auth/reset-password',
    { token, password },
    { skipAuth: true }
  )
}

export const getCurrentOrganization = async (): Promise<Organization> => {
  try {
    // First, get user profile to check for default_organization_id
    const userResponse = await client.get<UserResponse>('/v1/users/me')
    const defaultOrgId = userResponse.default_organization_id

    // Fetch all organizations
    const orgResponse = await client.get<Array<{
      id: string
      name: string
      billing_email: string
      subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
      created_at: string
      updated_at: string
    }>>('/v1/organizations')

    if (!Array.isArray(orgResponse) || orgResponse.length === 0) {
      throw new Error('No organizations found for user')
    }

    // Select organization based on user's default_organization_id preference
    let selectedOrg = orgResponse[0] // Default to first

    // If user has a default organization preference, find it in the list
    if (defaultOrgId) {
      const defaultOrg = orgResponse.find(org => org.id === defaultOrgId)
      if (defaultOrg) {
        selectedOrg = defaultOrg
        console.log('[AuthAPI] Using default organization for current org:', defaultOrg.name)
      } else {
        console.log('[AuthAPI] Default org not found in user orgs, using first org')
      }
    }

    return {
      id: selectedOrg.id,
      name: selectedOrg.name,
      plan: selectedOrg.subscription_plan,
      members: [],
      apiKeys: [],
      usage: {
        requests_this_month: 0,
        cost_this_month: 0,
        models_used: 0,
      },
      createdAt: selectedOrg.created_at,
      updatedAt: selectedOrg.updated_at,
    }
  } catch (error) {
    console.error('[AuthAPI] Failed to fetch current organization:', error)
    throw error
  }
}

export const setDefaultOrganization = async (organizationId: string): Promise<void> => {
  await client.put('/v1/users/me/default-organization', {
    organization_id: organizationId
  })
}

// Validate invitation token (public endpoint)
export const validateInvitation = async (token: string) => {
  return await client.get<{
    organization_name: string
    organization_id: string
    inviter_name: string
    role: string
    email: string
    expires_at: string
    is_expired: boolean
  }>(`/v1/invitations/validate/${token}`, undefined, { skipAuth: true })
}

// Exchange OAuth login session for tokens (existing user OAuth login)
export const exchangeLoginSession = async (sessionId: string) => {
  // Backend now returns: { user, expires_at, expires_in } (tokens in httpOnly cookies)
  return await client.post<{
    user: UserResponse
    expires_at: number  // Milliseconds
    expires_in: number  // Milliseconds
  }>(`/v1/auth/exchange-session/${sessionId}`, {}, { skipAuth: true })
}

// Complete OAuth signup (Step 2)
export const completeOAuthSignup = async (data: {
  sessionId: string
  role: string
  organizationName?: string
  referralSource?: string
}): Promise<AuthResponse> => {
  const backendPayload = {
    session_id: data.sessionId,
    role: data.role,
    organization_name: data.organizationName,
    referral_source: data.referralSource,
  }

  // Backend now returns: { user, organization, expires_at, expires_in } (tokens in httpOnly cookies)
  const backendResponse = await client.post<{
    user: UserResponse
    organization: {
      id: string
      name: string
      billing_email: string
      subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
      created_at: string
      updated_at: string
    }
    expires_at: number  // Milliseconds
    expires_in: number  // Milliseconds
  }>(
    '/v1/auth/complete-oauth-signup',
    backendPayload,
    { skipAuth: true }
  )

  // Map user from response (no /me call needed)
  const user: User = {
    id: backendResponse.user.id,
    email: backendResponse.user.email,
    firstName: backendResponse.user.first_name,
    lastName: backendResponse.user.last_name,
    name: `${backendResponse.user.first_name} ${backendResponse.user.last_name}`.trim(),
    role: 'user',
    organizationId: '',
    defaultOrganizationId: backendResponse.user.default_organization_id,
    projects: [],
    createdAt: backendResponse.user.created_at,
    updatedAt: backendResponse.user.created_at,
    isEmailVerified: backendResponse.user.is_email_verified,
    onboardingCompletedAt: backendResponse.user.onboarding_completed_at,
  }

  // Map organization from response (no /organizations call needed)
  const organization: Organization = {
    id: backendResponse.organization.id,
    name: backendResponse.organization.name,
    plan: backendResponse.organization.subscription_plan,
    members: [],
    apiKeys: [],
    usage: {
      requests_this_month: 0,
      cost_this_month: 0,
      models_used: 0,
    },
    createdAt: backendResponse.organization.created_at,
    updatedAt: backendResponse.organization.updated_at,
  }

  return {
    user,
    organization,
    expiresAt: backendResponse.expires_at,  // Milliseconds
    expiresIn: backendResponse.expires_in,  // Milliseconds
  }
}
