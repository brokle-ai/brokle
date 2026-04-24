'use client'

import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  useSafeMutation,
  MutationReentryError,
} from '@/lib/react-query/use-safe-mutation'
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

// isReentryBlocked filters the sentinel error thrown by useSafeMutation
// when a duplicate call is blocked while a prior call is still in flight.
// A blocked duplicate is not a user-facing failure — no toast, no alert.
function isReentryBlocked(error: unknown): boolean {
  return error instanceof MutationReentryError
}

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

  return useSafeMutation({
    mutationFn: async (credentials: LoginCredentials) => {
      return login(credentials)
    },
    onSuccess: (data: AuthResponse) => {
      queryClient.setQueryData(authQueryKeys.user(), data.user)
      queryClient.setQueryData(authQueryKeys.organization(), data.organization)

      toast.success('Welcome back!', {
        description: `Signed in as ${data.user?.email || 'Unknown User'}`,
      })
    },
    onError: (error: any) => {
      if (isReentryBlocked(error)) return
      toast.error('Login Failed', {
        description: error?.message || 'Invalid credentials',
      })
    },
  })
}

// Signup mutation
//
// No success toast: the full-page navigation to `/` is the success
// surface, and the signup forms render an inline "Redirecting…" alert
// while waiting. Toasting on success created a confusing
// success-then-error double-message whenever post-success client code
// threw (or a duplicate click fired the mutation twice).
export function useSignupMutation() {
  const queryClient = useQueryClient()

  return useSafeMutation({
    mutationFn: async (credentials: SignUpCredentials) => {
      return authSignup(credentials)
    },
    onSuccess: (data: AuthResponse) => {
      queryClient.setQueryData(authQueryKeys.user(), data.user)
      queryClient.setQueryData(authQueryKeys.organization(), data.organization)
    },
    onError: (error: any) => {
      if (isReentryBlocked(error)) return
      toast.error('Signup Failed', {
        description: error?.message || 'Failed to create account',
      })
    },
  })
}

// Complete OAuth Signup mutation
export function useCompleteOAuthSignupMutation() {
  const queryClient = useQueryClient()

  return useSafeMutation({
    mutationFn: async (data: {
      sessionId: string
      role: string
      organizationName?: string
      referralSource?: string
    }) => {
      return authCompleteOAuthSignup(data)
    },
    onSuccess: (data: AuthResponse) => {
      queryClient.setQueryData(authQueryKeys.user(), data.user)
      queryClient.setQueryData(authQueryKeys.organization(), data.organization)
    },
    onError: (error: any) => {
      if (isReentryBlocked(error)) return
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

  return useSafeMutation({
    mutationFn: async () => {
      // Show overlay
      if (typeof window !== 'undefined') {
        window.dispatchEvent(new CustomEvent('auth:logout-start'))
      }

      await logout()
    },
    onSuccess: () => {
      queryClient.clear()
      if (typeof window !== 'undefined') {
        window.location.href = signinWithStatus('logout_success')
      }
    },
    onError: (error: unknown) => {
      if (isReentryBlocked(error)) return
      try {
        queryClient.clear()
        if (typeof window !== 'undefined') {
          window.location.href = signinWithStatus('logout_error')
        }
      } catch (err) {
        console.error('[Logout] Error during logout error handling:', err)
      } finally {
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

  return useSafeMutation({
    mutationFn: async (data: Partial<User>) => {
      return authUpdateProfile(data)
    },
    onSuccess: (updatedUser: User) => {
      const currentUser = useAuthStore.getState().user
      const mergedUser = currentUser ? { ...currentUser, ...updatedUser } : updatedUser
      setUser(mergedUser)
      queryClient.setQueryData(authQueryKeys.user(), mergedUser)
      queryClient.invalidateQueries({ queryKey: ['workspace'] })

      toast.success('Profile Updated', {
        description: 'Your profile has been updated successfully.',
      })
    },
    onError: (error: any) => {
      if (isReentryBlocked(error)) return
      toast.error('Update Failed', {
        description: error?.message || 'Failed to update profile',
      })
    },
  })
}

// Change password mutation
export function useChangePasswordMutation() {
  return useSafeMutation({
    mutationFn: async (data: { currentPassword: string; newPassword: string }) => {
      await authChangePassword(data.currentPassword, data.newPassword)
    },
    onSuccess: () => {
      toast.success('Password Changed', {
        description: 'Your password has been updated successfully.',
      })
    },
    onError: (error: any) => {
      if (isReentryBlocked(error)) return
      toast.error('Password Change Failed', {
        description: error?.message || 'Failed to change password',
      })
    },
  })
}

// Request password reset mutation
export function useRequestPasswordResetMutation() {
  return useSafeMutation({
    mutationFn: async (email: string) => {
      await requestPasswordReset(email)
    },
    onSuccess: () => {
      toast.success('Reset Email Sent', {
        description: 'Check your email for password reset instructions.',
      })
    },
    onError: (error: any) => {
      if (isReentryBlocked(error)) return
      toast.error('Reset Failed', {
        description: error?.message || 'Failed to send reset email',
      })
    },
  })
}

// Confirm password reset mutation
export function useConfirmPasswordResetMutation() {
  return useSafeMutation({
    mutationFn: async (data: { token: string; password: string }) => {
      await confirmPasswordReset(data.token, data.password)
    },
    onSuccess: () => {
      toast.success('Password Reset', {
        description: 'Your password has been reset successfully.',
      })
    },
    onError: (error: any) => {
      if (isReentryBlocked(error)) return
      toast.error('Reset Failed', {
        description: error?.message || 'Failed to reset password',
      })
    },
  })
}

