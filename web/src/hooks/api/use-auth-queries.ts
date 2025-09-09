'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { 
  getCurrentUser,
  getCurrentOrganization,
  requestPasswordReset,
  confirmPasswordReset
} from '@/lib/api'
import { useAuth } from '@/hooks/auth/use-auth'
import type { 
  User, 
  LoginCredentials, 
  SignUpCredentials,
  AuthResponse 
} from '@/types/auth'
import { toast } from 'sonner'

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
  const { signup } = useAuth()

  return useMutation({
    mutationFn: async (credentials: SignUpCredentials) => {
      return signup(credentials)
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

// Logout mutation
export function useLogoutMutation() {
  const queryClient = useQueryClient()
  const { logout } = useAuth()

  return useMutation({
    mutationFn: async () => {
      await logout()
    },
    onSuccess: () => {
      // Clear all cached data
      queryClient.clear()
      
      toast.success('Logged out successfully', {
        description: 'You have been securely logged out.',
      })
    },
    onError: () => {
      // Still clear cache even if API call fails
      queryClient.clear()
      
      toast.warning('Logged out locally', {
        description: 'Session cleared locally. You may need to refresh other tabs.',
      })
    },
  })
}

// Update profile mutation
export function useUpdateProfileMutation() {
  const queryClient = useQueryClient()
  const { updateUser } = useAuth()

  return useMutation({
    mutationFn: async (data: Partial<User>) => {
      return updateUser(data)
    },
    onSuccess: (updatedUser: User) => {
      // Update cache with new user data
      queryClient.setQueryData(authQueryKeys.user(), updatedUser)
      
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
  const { changePassword } = useAuth()

  return useMutation({
    mutationFn: async (data: { currentPassword: string; newPassword: string }) => {
      await changePassword(data.currentPassword, data.newPassword)
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

