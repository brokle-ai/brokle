// Auth API - Clean implementation without over-engineered abstractions
import { BrokleAPIClient } from '../core/client'
import type { 
  AuthResponse,
  LoginCredentials,
  SignUpCredentials,
  User,
  Organization,
  LoginResponse,
  UserResponse
} from '@/types/auth'

// Simple API client
const client = new BrokleAPIClient('/api')

export const login = async (credentials: LoginCredentials): Promise<AuthResponse> => {
  // Get token data from backend
  const backendResponse = await client.post<LoginResponse>(
    '/v1/auth/login',
    credentials, 
    { skipAuth: true }
  )

  // Get user details using the new token
  const userResponse = await client.get<UserResponse>('/v1/users/me', undefined, {
    headers: {
      'Authorization': `Bearer ${backendResponse.access_token}`
    }
  })

  // Map backend user response to frontend format
  const user: User = {
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

  // Get organization from backend
  let organization: Organization
  try {
    const orgResponse = await client.get<Array<{
      id: string
      name: string
      slug: string
      billing_email: string
      subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
      created_at: string
      updated_at: string
    }>>('/v1/organizations', undefined, {
      headers: {
        'Authorization': `Bearer ${backendResponse.access_token}`
      }
    })
    
    const firstOrg = Array.isArray(orgResponse) && orgResponse.length > 0 ? orgResponse[0] : null
    
    if (!firstOrg) {
      throw new Error('No organizations found for user')
    }
    
    organization = {
      id: firstOrg.id,
      name: firstOrg.name,
      slug: firstOrg.slug,
      plan: firstOrg.subscription_plan,
      members: [{
        userId: user.id,
        user: user,
        role: 'owner',
        joinedAt: new Date().toISOString(),
      }], 
      apiKeys: [],
      usage: {
        requestsThisMonth: 0,
        costsThisMonth: 0,
        modelsUsed: 0,
      },
      createdAt: firstOrg.created_at,
      updatedAt: firstOrg.updated_at,
    }
  } catch (orgError) {
    console.error('[AuthAPI] Failed to fetch organization during login:', orgError)
    throw orgError
  }

  return {
    user,
    organization,
    accessToken: backendResponse.access_token,
    refreshToken: backendResponse.refresh_token,
    expiresIn: backendResponse.expires_in,
  }
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

  const backendResponse = await client.post<LoginResponse>(
    '/v1/auth/signup',
    backendPayload,
    { skipAuth: true }
  )

  // Get user details
  const userResponse = await client.get<UserResponse>('/v1/users/me', undefined, {
    headers: {
      'Authorization': `Bearer ${backendResponse.access_token}`
    }
  })

  const user: User = {
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

  // Get organization from backend
  let organization: Organization
  try {
    const orgResponse = await client.get<Array<{
      id: string
      name: string
      slug: string
      billing_email: string
      subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
      created_at: string
      updated_at: string
    }>>('/v1/organizations', undefined, {
      headers: {
        'Authorization': `Bearer ${backendResponse.access_token}`
      }
    })
    
    const firstOrg = Array.isArray(orgResponse) && orgResponse.length > 0 ? orgResponse[0] : null
    if (!firstOrg) {
      throw new Error('No organizations found for user')
    }
    
    organization = {
      id: firstOrg.id,
      name: firstOrg.name,
      slug: firstOrg.slug,
      plan: firstOrg.subscription_plan,
      members: [],
      apiKeys: [],
      usage: {
        requestsThisMonth: 0,
        costsThisMonth: 0,
        modelsUsed: 0,
      },
      createdAt: firstOrg.created_at,
      updatedAt: firstOrg.updated_at,
    }
  } catch (orgError) {
    console.error('[AuthAPI] Failed to fetch organization during signup:', orgError)
    throw orgError
  }

  return {
    user,
    organization,
    accessToken: backendResponse.access_token,
    refreshToken: backendResponse.refresh_token,
    expiresIn: backendResponse.expires_in,
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
    const orgResponse = await client.get<Array<{
      id: string
      name: string
      slug: string
      billing_email: string
      subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
      created_at: string
      updated_at: string
    }>>('/v1/organizations')
    
    const firstOrg = Array.isArray(orgResponse) && orgResponse.length > 0 ? orgResponse[0] : null
    if (!firstOrg) {
      throw new Error('No organizations found for user')
    }
    
    return {
      id: firstOrg.id,
      name: firstOrg.name,
      slug: firstOrg.slug,
      plan: firstOrg.subscription_plan,
      members: [],
      apiKeys: [],
      usage: {
        requestsThisMonth: 0,
        costsThisMonth: 0,
        modelsUsed: 0,
      },
      createdAt: firstOrg.created_at,
      updatedAt: firstOrg.updated_at,
    }
  } catch (error) {
    console.error('[AuthAPI] Failed to fetch current organization:', error)
    throw error
  }
}

export const setDefaultOrganization = async (organizationId: string): Promise<void> => {
  await client.patch('/v1/users/me', {
    default_organization_id: organizationId
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
  return await client.post<{
    access_token: string
    refresh_token: string
    token_type: string
    expires_in: number
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

  const backendResponse = await client.post<LoginResponse>(
    '/v1/auth/complete-oauth-signup',
    backendPayload,
    { skipAuth: true }
  )

  // Get user details
  const userResponse = await client.get<UserResponse>('/v1/users/me', undefined, {
    headers: {
      'Authorization': `Bearer ${backendResponse.access_token}`
    }
  })

  const user: User = {
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

  // Get organization
  const orgResponse = await client.get<Array<{
    id: string
    name: string
    slug: string
    billing_email: string
    subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
    created_at: string
    updated_at: string
  }>>('/v1/organizations', undefined, {
    headers: {
      'Authorization': `Bearer ${backendResponse.access_token}`
    }
  })

  const firstOrg = Array.isArray(orgResponse) && orgResponse.length > 0 ? orgResponse[0] : null
  if (!firstOrg) {
    throw new Error('No organizations found for user')
  }

  const organization: Organization = {
    id: firstOrg.id,
    name: firstOrg.name,
    slug: firstOrg.slug,
    plan: firstOrg.subscription_plan,
    members: [],
    apiKeys: [],
    usage: {
      requestsThisMonth: 0,
      costsThisMonth: 0,
      modelsUsed: 0,
    },
    createdAt: firstOrg.created_at,
    updatedAt: firstOrg.updated_at,
  }

  return {
    user,
    organization,
    accessToken: backendResponse.access_token,
    refreshToken: backendResponse.refresh_token,
    expiresIn: backendResponse.expires_in,
  }
}
