'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from './use-auth'
import type { SignUpCredentials, AuthResponse } from '@/types/auth'
import { toast } from 'sonner'

interface UseSignUpReturn {
  signup: (credentials: SignUpCredentials) => Promise<AuthResponse | null>
  isLoading: boolean
  error: string | null
  clearError: () => void
}

export function useSignUp(): UseSignUpReturn {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  
  const { signup: authSignup } = useAuth()
  const router = useRouter()

  const signup = async (credentials: SignUpCredentials): Promise<AuthResponse | null> => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await authSignup(credentials)
      
      // Show success message
      toast.success('Account Created!', {
        description: `Welcome to Brokle, ${response.user.firstName || response.user.email}!`,
      })

      // Redirect to onboarding or dashboard
      setTimeout(() => {
        router.push('/dashboard')
      }, 500)

      return response
    } catch (err: any) {
      const errorMessage = err?.message || 'Signup failed. Please try again.'
      setError(errorMessage)
      
      toast.error('Signup Failed', {
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
    signup,
    isLoading,
    error,
    clearError,
  }
}