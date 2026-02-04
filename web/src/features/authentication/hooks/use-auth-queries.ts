'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getCurrentUser,
  getCurrentOrganization,
  requestPasswordReset,
  confirmPasswordReset
} from '../api/auth-api'
import {
  signup as authSignup,
  completeOAuthSignup as authCompleteOAuthSignup,
  updateProfile as authUpdateProfile,
  changePassword as authChangePassword
} from '../api/auth-api'
import { useAuth } from './use-auth'
import { useAuthStore } from '../stores/auth-store'
import type {
  User,
  LoginCredentials,
  SignUpCredentials,
  AuthResponse
} from '../types'
import { toast } from 'sonner'
import { signinWithStatus } from '@/lib/routes'

// Query keys for consistent caching
export const authQueryKeys = {
  all: ['auth'] as const,
  user: () => [...authQueryKeys.all, 'user'] as const,
  organization: () => [...authQueryKeys.all, 'organization'] as const,
} as const

// Current user query
export function useCurrentUser() {
  const { isAuthenticated } = useAuth()
  
  return useQuery({
    queryKey: authQueryKeys.user(),
    queryFn: () => getCurrentUser(),
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (failureCount, error: any) => {
      // Don't retry on auth errors
      if (error?.statusCode === 401) return false
      return failureCount < 3
    },
  })
}

// Current organization query
export function useCurrentOrganization() {
  const { isAuthenticated } = useAuth()
  
  return useQuery({
    queryKey: authQueryKeys.organization(),
    queryFn: () => getCurrentOrganization(),
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (failureCount, error: any) => {
      if (error?.statusCode === 401) return false
      return failureCount < 3
    },
  })
}

// Login mutation
export function useLoginMutation() {
  const queryClient = useQueryClient()
  const { login } = useAuth()

  return useMutation({
    mutationFn: async (credentials: LoginCredentials) => {
      return login(credentials)
    },
    onSuccess: (data: AuthResponse) => {
      // Update query cache with new user data
      queryClient.setQueryData(authQueryKeys.user(), data.user)
      queryClient.setQueryData(authQueryKeys.organization(), data.organization)
      

      toast.success('Welcome back!', {
        description: `Signed in as ${data.user?.email || 'Unknown User'}`,
      })
    },
    onError: (error: any) => {
      toast.error('Login Failed', {
        description: error?.message || 'Invalid credentials',
      })
    },
  })
}

// Signup mutation
export function useSignupMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (credentials: SignUpCredentials) => {
      return authSignup(credentials)
    },
    onSuccess: (data: AuthResponse) => {
      // Update query cache with new user data
      queryClient.setQueryData(authQueryKeys.user(), data.user)
      queryClient.setQueryData(authQueryKeys.organization(), data.organization)

      toast.success('Account Created!', {
        description: `Welcome to Brokle, ${data.user.firstName || data.user.email}!`,
      })
    },
    onError: (error: any) => {
      toast.error('Signup Failed', {
        description: error?.message || 'Failed to create account',
      })
    },
  })
}

// Complete OAuth Signup mutation
export function useCompleteOAuthSignupMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: {
      sessionId: string
      role: string
      organizationName?: string
      referralSource?: string
    }) => {
      return authCompleteOAuthSignup(data)
    },
    onSuccess: (data: AuthResponse) => {
      // Update query cache
      queryClient.setQueryData(authQueryKeys.user(), data.user)
      queryClient.setQueryData(authQueryKeys.organization(), data.organization)

      toast.success('Account Created!', {
        description: `Welcome to Brokle, ${data.user.firstName || data.user.email}!`,
      })
    },
    onError: (error: any) => {
      toast.error('OAuth Signup Failed', {
        description: error?.message || 'Failed to complete OAuth signup',
      })
    },
  })
}

// Logout mutation
export function useLogoutMutation() {
  const queryClient = useQueryClient()
  const logout = useAuthStore(state => state.logout)

  return useMutation({
    mutationFn: async () => {
      // Show overlay
      if (typeof window !== 'undefined') {
        window.dispatchEvent(new CustomEvent('auth:logout-start'))
      }

      await logout()
    },
    onSuccess: () => {
      // Clear all cached data
      queryClient.clear()

      // Hard redirect (toast shows on signin page)
      if (typeof window !== 'undefined') {
        window.location.href = signinWithStatus('logout_success')
      }
    },
    onError: () => {
      try {
        // Still clear cache even if API call fails
        queryClient.clear()

        // Hard redirect with error param
        if (typeof window !== 'undefined') {
          window.location.href = signinWithStatus('logout_error')
        }
      } catch (error) {
        console.error('[Logout] Error during logout error handling:', error)
      } finally {
        // Ensure overlay clears even if redirect fails
        if (typeof window !== 'undefined') {
          window.dispatchEvent(new CustomEvent('auth:logout-end'))
        }
      }
    },
  })
}

// Update profile mutation
export function useUpdateProfileMutation() {
  const queryClient = useQueryClient()
  const setUser = useAuthStore((state) => state.setUser)

  return useMutation({
    mutationFn: async (data: Partial<User>) => {
      return authUpdateProfile(data)
    },
    onSuccess: (updatedUser: User) => {
      // Merge with existing user to preserve org data not returned by update endpoint
      const currentUser = useAuthStore.getState().user
      const mergedUser = currentUser ? { ...currentUser, ...updatedUser } : updatedUser
      // Update both caches with merged user for consistency
      setUser(mergedUser)
      queryClient.setQueryData(authQueryKeys.user(), mergedUser)
      // Invalidate workspace cache so sidebar reflects the updated name
      queryClient.invalidateQueries({ queryKey: ['workspace'] })

      toast.success('Profile Updated', {
        description: 'Your profile has been updated successfully.',
      })
    },
    onError: (error: any) => {
      toast.error('Update Failed', {
        description: error?.message || 'Failed to update profile',
      })
    },
  })
}

// Change password mutation
export function useChangePasswordMutation() {
  return useMutation({
    mutationFn: async (data: { currentPassword: string; newPassword: string }) => {
      await authChangePassword(data.currentPassword, data.newPassword)
    },
    onSuccess: () => {
      toast.success('Password Changed', {
        description: 'Your password has been updated successfully.',
      })
    },
    onError: (error: any) => {
      toast.error('Password Change Failed', {
        description: error?.message || 'Failed to change password',
      })
    },
  })
}

// Request password reset mutation
export function useRequestPasswordResetMutation() {
  return useMutation({
    mutationFn: async (email: string) => {
      await requestPasswordReset(email)
    },
    onSuccess: () => {
      toast.success('Reset Email Sent', {
        description: 'Check your email for password reset instructions.',
      })
    },
    onError: (error: any) => {
      toast.error('Reset Failed', {
        description: error?.message || 'Failed to send reset email',
      })
    },
  })
}

// Confirm password reset mutation
export function useConfirmPasswordResetMutation() {
  return useMutation({
    mutationFn: async (data: { token: string; password: string }) => {
      await confirmPasswordReset(data.token, data.password)
    },
    onSuccess: () => {
      toast.success('Password Reset', {
        description: 'Your password has been reset successfully.',
      })
    },
    onError: (error: any) => {
      toast.error('Reset Failed', {
        description: error?.message || 'Failed to reset password',
      })
    },
  })
}

