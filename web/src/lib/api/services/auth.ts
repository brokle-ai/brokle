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
  ApiKey,
  LoginResponse,
  UserResponse
} from '@/types/auth'

export class AuthAPIClient extends BrokleAPIClient {
  private tokenManager = getTokenManager()

  constructor() {
    super('/auth') // All auth endpoints will be prefixed with /auth
    
    // Connect token refresh callback to fix circular dependency
    this.tokenManager.setRefreshCallback(() => this.refreshTokens())
  }

  // Authentication endpoints (these use skipAuth for login/signup)
  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    // Get token data from backend
    const backendResponse = await this.post<LoginResponse>(
      '/v1/auth/login', 
      credentials, 
      { skipAuth: true }
    )

    // Store tokens immediately
    await this.tokenManager.setTokens({
      accessToken: backendResponse.access_token,
      refreshToken: backendResponse.refresh_token,
      expiresIn: backendResponse.expires_in,
      tokenType: 'Bearer',
    })

    // Get user details using the new token (no context headers needed)
    const userResponse = await this.get<UserResponse>('/v1/auth/me')

    // Map backend user response to frontend format
    const user: User = {
      id: userResponse.id,
      email: userResponse.email,
      firstName: userResponse.first_name,
      lastName: userResponse.last_name,
      name: `${userResponse.first_name} ${userResponse.last_name}`.trim(),
      role: 'user', // Default role
      organizationId: '', // Will be populated from org endpoint
      defaultOrganizationId: userResponse.default_organization_id,
      projects: [], // Will be populated from separate endpoint
      createdAt: userResponse.created_at,
      updatedAt: userResponse.created_at,
      isEmailVerified: userResponse.is_email_verified,
      onboardingCompleted: userResponse.onboarding_completed,
    }

    // Get organization from backend or create default
    let organization: Organization
    try {
      // Try to get user's organizations from backend
      const orgResponse = await this.get<Array<{
        id: string
        name: string
        slug: string
        billing_email: string
        subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
        created_at: string
        updated_at: string
      }>>('/v1/organizations/user')
      
      // Debug logging to understand the response structure
      console.debug('[AuthAPIClient] Login - Organization response:', orgResponse)
      
      // Use the first organization from the response (orgResponse should be the array directly after extractData)
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
        apiKeys: [], // Will be populated separately if needed
        usage: {
          requestsThisMonth: 0,
          costsThisMonth: 0,
          modelsUsed: 0,
        },
        createdAt: firstOrg.created_at,
        updatedAt: firstOrg.updated_at,
      }
    } catch (orgError) {
      // Re-throw with context - let UI handle the error properly
      console.error('[AuthAPIClient] Failed to fetch organization during login:', orgError)
      throw orgError // Let BrokleAPIError propagate with proper error details
    }

    return {
      user,
      organization,
      accessToken: backendResponse.access_token,
      refreshToken: backendResponse.refresh_token,
      expiresIn: backendResponse.expires_in,
    }
  }

  async signup(credentials: SignUpCredentials): Promise<AuthResponse> {
    // Map frontend format to backend format
    const backendPayload = {
      first_name: credentials.firstName,
      last_name: credentials.lastName,
      email: credentials.email,
      password: credentials.password,
    }
    
    const backendResponse = await this.post<LoginResponse>(
      '/v1/auth/register', 
      backendPayload, 
      { skipAuth: true }
    )

    // Store tokens immediately
    await this.tokenManager.setTokens({
      accessToken: backendResponse.access_token,
      refreshToken: backendResponse.refresh_token,
      expiresIn: backendResponse.expires_in,
      tokenType: 'Bearer',
    })

    // Get user details (no context headers needed)
    const userResponse = await this.get<UserResponse>('/v1/auth/me')

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

    // Get organization from backend or create default
    let organization: Organization
    try {
      // Try to get user's organizations from backend
      const orgResponse = await this.get<Array<{
        id: string
        name: string
        slug: string
        billing_email: string
        subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
        created_at: string
        updated_at: string
      }>>('/v1/organizations/user')
      
      // Debug logging to understand the response structure
      console.debug('[AuthAPIClient] Signup - Organization response:', orgResponse)
      
      // Use the first organization from the response (orgResponse should be the array directly after extractData)
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
      // Re-throw with context - let UI handle the error properly
      console.error('[AuthAPIClient] Failed to fetch organization during signup:', orgError)
      throw orgError // Let BrokleAPIError propagate with proper error details
    }

    return {
      user,
      organization,
      accessToken: backendResponse.access_token,
      refreshToken: backendResponse.refresh_token,
      expiresIn: backendResponse.expires_in,
    }
  }

  async logout(): Promise<void> {
    try {
      // Call logout endpoint (authenticated - token will be added automatically)
      await this.post('/v1/auth/logout', {})
    } catch (error) {
      // Log but don't throw - we want to clear local tokens regardless
      console.warn('Logout request failed:', error)
    } finally {
      // Always clear local tokens
      this.tokenManager.clearTokens()
    }
  }

  async refreshTokens(refreshToken?: string): Promise<AuthTokens> {
    const tokenToUse = refreshToken || this.getStoredRefreshToken()
    
    if (!tokenToUse) {
      throw new Error('No refresh token available')
    }

    const backendResponse = await this.post<LoginResponse>(
      '/v1/sessions/refresh',
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

  // User management (all authenticated - tokens added automatically)
  async getCurrentUser(): Promise<User> {
    const userResponse = await this.get<UserResponse>('/v1/auth/me')
    
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

  async updateProfile(data: Partial<User>): Promise<User> {
    // Map frontend format to backend format
    const backendData = {
      first_name: data.firstName,
      last_name: data.lastName,
      // Add other fields as needed
    }

    const userResponse = await this.patch<UserResponse>('/v1/auth/profile', backendData)
    
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

  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    await this.patch('/v1/auth/password', {
      current_password: currentPassword,
      new_password: newPassword,
    })
  }

  // Password reset (public endpoints)
  async requestPasswordReset(email: string): Promise<void> {
    await this.post('/v1/auth/password-reset', { email }, { skipAuth: true })
  }

  async confirmPasswordReset(token: string, password: string): Promise<void> {
    await this.post(
      '/v1/auth/password-reset/confirm',
      { token, password },
      { skipAuth: true }
    )
  }

  // Organization management (authenticated)
  async getCurrentOrganization(): Promise<Organization> {
    try {
      // Get user's organizations from backend
      const orgResponse = await this.get<Array<{
        id: string
        name: string
        slug: string
        billing_email: string
        subscription_plan: 'free' | 'pro' | 'business' | 'enterprise'
        created_at: string
        updated_at: string
      }>>('/v1/organizations/user')
      
      // Debug logging to understand the response structure
      console.debug('[AuthAPIClient] getCurrentOrganization - Organization response:', orgResponse)
      
      // Use the first organization from the response (orgResponse should be the array directly after extractData)
      const firstOrg = Array.isArray(orgResponse) && orgResponse.length > 0 ? orgResponse[0] : null
      if (!firstOrg) {
        throw new Error('No organizations found for user')
      }
      
      return {
        id: firstOrg.id,
        name: firstOrg.name,
        slug: firstOrg.slug,
        plan: firstOrg.subscription_plan,
        members: [], // Will be populated separately if needed
        apiKeys: [], // Will be populated separately if needed
        usage: {
          requestsThisMonth: 0,
          costsThisMonth: 0,
          modelsUsed: 0,
        },
        createdAt: firstOrg.created_at,
        updatedAt: firstOrg.updated_at,
      }
    } catch (error) {
      // Re-throw with context - let UI handle the error properly
      console.error('[AuthAPIClient] Failed to fetch current organization:', error)
      throw error // Let BrokleAPIError propagate with proper error details
    }
  }

  async updateOrganization(data: Partial<Organization>): Promise<Organization> {
    // Map frontend format to backend format
    const backendData = {
      name: data.name,
      slug: data.slug,
      // Add other organization fields as needed
    }

    try {
      const orgResponse = await this.patch<{
        id: string
        name: string
        slug: string
        plan: 'free' | 'pro' | 'business' | 'enterprise'
        created_at: string
        updated_at: string
      }>('/v1/auth/organization', backendData)
      
      return {
        id: orgResponse.id,
        name: orgResponse.name,
        slug: orgResponse.slug,
        plan: orgResponse.plan,
        members: data.members || [],
        apiKeys: data.apiKeys || [],
        usage: data.usage || {
          requestsThisMonth: 0,
          costsThisMonth: 0,
          modelsUsed: 0,
        },
        createdAt: orgResponse.created_at,
        updatedAt: orgResponse.updated_at,
      }
    } catch (error) {
      console.error('[AuthAPIClient] Failed to update organization:', error)
      throw new Error('Organization update failed')
    }
  }

  async inviteUser(email: string, role: string): Promise<void> {
    await this.post('/v1/auth/organization/invite', { email, role })
  }

  async removeUser(userId: string): Promise<void> {
    await this.delete(`/v1/auth/organization/members/${userId}`)
  }

  // API Key management (authenticated)
  async getApiKeys(): Promise<ApiKey[]> {
    // Placeholder - implement when API keys endpoint is ready
    return []
  }

  async createApiKey(data: {
    name: string
    permissions: string[]
    expiresAt?: string
  }): Promise<ApiKey> {
    return this.post<ApiKey>('/v1/auth/api-keys', data)
  }

  async revokeApiKey(keyId: string): Promise<void> {
    await this.delete(`/v1/auth/api-keys/${keyId}`)
  }

  // Two-factor authentication (authenticated)
  async enableTwoFactor(): Promise<{
    qrCode: string
    secret: string
    backupCodes: string[]
  }> {
    return this.post('/v1/auth/2fa/enable')
  }

  async confirmTwoFactor(token: string, secret: string): Promise<void> {
    await this.post('/v1/auth/2fa/confirm', { token, secret })
  }

  async disableTwoFactor(token: string): Promise<void> {
    await this.post('/v1/auth/2fa/disable', { token })
  }

  async verifyTwoFactor(token: string): Promise<void> {
    await this.post('/v1/auth/2fa/verify', { token })
  }

  async setDefaultOrganization(organizationId: string): Promise<void> {
    await this.patch('/v1/organizations/default', { 
      organization_id: organizationId 
    })
  }

  async completeOnboarding(): Promise<void> {
    await this.patch('/v1/auth/profile', {
      onboarding_completed: true
    })
  }

  // Helper method to get stored refresh token using centralized storage
  private getStoredRefreshToken(): string | null {
    if (typeof window === 'undefined') return null
    
    // Use the centralized storage method which tries localStorage first, then cookie
    return SecureStorage.getRefreshToken()
  }
}