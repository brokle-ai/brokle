// Auth API - Latest endpoints for dashboard application
// Direct functions using optimal backend endpoints

import { BrokleAPIClient } from '../core/client'
import { getTokenManager } from '@/lib/auth/token-manager'
import { SecureStorage } from '@/lib/auth/storage'
import type { 
  AuthResponse,
  AuthTokens,
  LoginCredentials,
  SignUpCredentials,
  User,
  Organization,
  LoginResponse,
  UserResponse
} from '@/types/auth'

// Flexible base client - versions specified per endpoint
const client = new BrokleAPIClient('/api')
const tokenManager = getTokenManager()

// Connect token refresh callback to avoid circular dependency
tokenManager.setRefreshCallback(() => refreshTokens())

// Direct auth functions - latest & optimal endpoints
export const login = async (credentials: LoginCredentials): Promise<AuthResponse> => {
    // Get token data from backend
    const backendResponse = await client.post<LoginResponse>(
      '/v1/auth/login',
      credentials, 
      { skipAuth: true }
    )

    // Store tokens immediately
    await tokenManager.setTokens({
      accessToken: backendResponse.access_token,
      refreshToken: backendResponse.refresh_token,
      expiresIn: backendResponse.expires_in,
      tokenType: 'Bearer',
    })

    // Get user details using the new token
    const userResponse = await client.get<UserResponse>('/v1/users/me')

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
      onboardingCompleted: userResponse.onboarding_completed,
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
      }>>('/v1/organizations')
      
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
    }
    
    const backendResponse = await client.post<LoginResponse>(
      '/v1/auth/signup', 
      backendPayload, 
      { skipAuth: true }
    )

    // Store tokens immediately
    await tokenManager.setTokens({
      accessToken: backendResponse.access_token,
      refreshToken: backendResponse.refresh_token,
      expiresIn: backendResponse.expires_in,
      tokenType: 'Bearer',
    })

    // Get user details
    const userResponse = await client.get<UserResponse>('/v1/users/me')

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
      onboardingCompleted: userResponse.onboarding_completed,
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
      }>>('/v1/organizations')
      
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
    } finally {
      tokenManager.clearTokens()
    }
  }

export const refreshTokens = async (refreshToken?: string): Promise<AuthTokens> => {
    const tokenToUse = refreshToken || getStoredRefreshToken()
    
    if (!tokenToUse) {
      throw new Error('No refresh token available')
    }

    const backendResponse = await client.post<LoginResponse>(
      '/v1/auth/refresh',
      { refresh_token: tokenToUse },
      { skipAuth: true }
    )

    return {
      accessToken: backendResponse.access_token,
      refreshToken: backendResponse.refresh_token,
      expiresIn: backendResponse.expires_in,
      tokenType: 'Bearer',
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
      onboardingCompleted: userResponse.onboarding_completed,
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
      onboardingCompleted: userResponse.onboarding_completed,
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

export const completeOnboarding = async (): Promise<void> => {
    await client.patch('/v1/users/me', {
      onboarding_completed: true
    })
  }

// Helper function to get stored refresh token
const getStoredRefreshToken = (): string | null => {
  if (typeof window === 'undefined') return null
  return SecureStorage.getRefreshToken()
}