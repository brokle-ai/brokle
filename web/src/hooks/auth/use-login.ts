'use client'

import { useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { useAuth } from './use-auth'
import type { LoginCredentials, AuthResponse } from '@/types/auth'
import { toast } from 'sonner'

interface UseLoginReturn {
  login: (credentials: LoginCredentials) => Promise<AuthResponse | null>
  isLoading: boolean
  error: string | null
  clearError: () => void
}

export function useLogin(): UseLoginReturn {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  
  const { login: authLogin } = useAuth()
  const router = useRouter()
  const searchParams = useSearchParams()

  const login = async (credentials: LoginCredentials): Promise<AuthResponse | null> => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await authLogin(credentials)
      
      // Show success message
      toast.success('Welcome back!', {
        description: `Signed in as ${response.user.email}`,
      })

      // Get redirect URL from search params or default to dashboard
      const redirectUrl = searchParams.get('redirect') || '/dashboard'
      
      // Small delay to show success message
      setTimeout(() => {
        router.push(redirectUrl)
      }, 500)

      return response
    } catch (err: any) {
      const errorMessage = err?.message || 'Login failed. Please try again.'
      setError(errorMessage)
      
      toast.error('Login Failed', {
        description: errorMessage,
      })
      
      return null
    } finally {
      setIsLoading(false)
    }
  }

  const clearError = () => {
    setError(null)
  }

  return {
    login,
    isLoading,
    error,
    clearError,
  }
}