# Frontend Feature Patterns

## Complete Feature Example: Authentication

This shows the ACTUAL structure of the authentication feature as implemented in `web/src/features/authentication/`.

### Actual Directory Structure

```
features/authentication/
├── components/
│   ├── sign-in-form.tsx
│   ├── sign-up-form.tsx
│   ├── two-step-signup-form.tsx
│   ├── forgot-password-form.tsx
│   ├── otp-form.tsx
│   ├── auth-guard.tsx
│   ├── auth-layout.tsx
│   ├── auth-status.tsx
│   ├── logout-button.tsx
│   ├── auth-form-wrapper.tsx
│   ├── signin-toast-handler.tsx
│   └── unauthorized-fallback.tsx
├── hooks/
│   ├── use-auth.ts
│   ├── use-auth-guard.ts
│   └── use-auth-queries.ts
├── stores/
│   └── auth-store.ts
├── api/
│   └── auth-api.ts
├── types/
│   └── index.ts
├── __tests__/
│   └── (test files)
└── index.ts
```

### Public API (`web/src/features/authentication/index.ts:3-69`)

**ACTUAL Exports**:

```typescript
// Hooks
export { useAuth }
export { useAuthGuard }
export {
  useCurrentUser,
  useCurrentOrganization,
  useLoginMutation,
  useSignupMutation,
  useCompleteOAuthSignupMutation,
  useUpdateProfileMutation,
  useChangePasswordMutation,
  useRequestPasswordResetMutation,
  useConfirmPasswordResetMutation,
  useLogoutMutation,
  authQueryKeys,
}

// Components
export { SignInForm }
export { SignUpForm }
export { TwoStepSignUpForm }
export { ForgotPasswordForm }
export { OTPForm }
export { AuthGuard }
export { UnauthorizedFallback }
export { LogoutButton }
export { AuthStatus }
export { AuthLayout }
export { AuthFormWrapper }
export { SignInToastHandler }

// Store
export { useAuthStore }

// API Functions
export { exchangeLoginSession }

// Types
export type {
  User, UserRole, AuthState, LoginCredentials,
  SignUpCredentials, AuthResponse, Organization,
  // ... (see index.ts:41-69 for full list)
}
```

### Component Example

```typescript
// features/authentication/components/sign-in-form.tsx
'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { useSignIn } from '../hooks/use-sign-in'

const signInSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
})

type SignInFormData = z.infer<typeof signInSchema>

interface SignInFormProps {
  onSuccess?: () => void
  redirectTo?: string
}

export function SignInForm({ onSuccess, redirectTo = '/dashboard' }: SignInFormProps) {
  const { signIn, isLoading, error } = useSignIn()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<SignInFormData>({
    resolver: zodResolver(signInSchema),
  })

  const onSubmit = async (data: SignInFormData) => {
    const success = await signIn(data, redirectTo)
    if (success && onSuccess) {
      onSuccess()
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="space-y-2">
        <Label htmlFor="email">Email</Label>
        <Input
          id="email"
          type="email"
          placeholder="you@example.com"
          {...register('email')}
          disabled={isLoading}
        />
        {errors.email && (
          <p className="text-sm text-destructive">{errors.email.message}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="password">Password</Label>
        <Input
          id="password"
          type="password"
          {...register('password')}
          disabled={isLoading}
        />
        {errors.password && (
          <p className="text-sm text-destructive">{errors.password.message}</p>
        )}
      </div>

      <Button type="submit" className="w-full" disabled={isLoading}>
        {isLoading ? 'Signing in...' : 'Sign In'}
      </Button>
    </form>
  )
}
```

### Hook Example

```typescript
// features/authentication/hooks/use-auth.ts
'use client'

import { useAuthStore } from '../stores/auth-store'
import { useRouter } from 'next/navigation'
import { signOut as apiSignOut } from '../api/auth-api'

export function useAuth() {
  const router = useRouter()
  const { user, isAuthenticated, setUser, clearUser } = useAuthStore()

  const signOut = async () => {
    try {
      await apiSignOut()
      clearUser()
      router.push('/sign-in')
    } catch (error) {
      console.error('Sign out failed:', error)
    }
  }

  const checkAuth = () => {
    // Check if user is authenticated
    return isAuthenticated
  }

  const requireAuth = (redirectTo: string = '/sign-in') => {
    if (!isAuthenticated) {
      router.push(redirectTo)
      return false
    }
    return true
  }

  return {
    user,
    isAuthenticated,
    signOut,
    checkAuth,
    requireAuth,
  }
}
```

```typescript
// features/authentication/hooks/use-sign-in.ts
'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { signIn as apiSignIn } from '../api/auth-api'
import { useAuthStore } from '../stores/auth-store'
import type { SignInRequest } from '../types'

export function useSignIn() {
  const router = useRouter()
  const { setUser } = useAuthStore()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const signIn = async (credentials: SignInRequest, redirectTo: string = '/dashboard') => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await apiSignIn(credentials)
      setUser(response.user)
      router.push(redirectTo)
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Sign in failed'
      setError(message)
      return false
    } finally {
      setIsLoading(false)
    }
  }

  return {
    signIn,
    isLoading,
    error,
  }
}
```

### API Client Example

```typescript
// features/authentication/api/auth-api.ts
import { apiClient } from '@/lib/api/core/client'
import type {
  SignInRequest,
  SignUpRequest,
  AuthResponse,
  User,
} from '../types'

export async function signIn(credentials: SignInRequest): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>('/auth/sign-in', credentials)
  return response.data
}

export async function signUp(data: SignUpRequest): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>('/auth/sign-up', data)
  return response.data
}

export async function signOut(): Promise<void> {
  await apiClient.post('/auth/sign-out')
}

export async function getCurrentUser(): Promise<User> {
  const response = await apiClient.get<User>('/auth/me')
  return response.data
}

export async function refreshToken(): Promise<{ token: string }> {
  const response = await apiClient.post<{ token: string }>('/auth/refresh')
  return response.data
}
```

### Store Example

```typescript
// features/authentication/stores/auth-store.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { AuthUser } from '../types'

interface AuthState {
  user: AuthUser | null
  isAuthenticated: boolean
  setUser: (user: AuthUser | null) => void
  clearUser: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      setUser: (user) =>
        set({
          user,
          isAuthenticated: !!user,
        }),
      clearUser: () =>
        set({
          user: null,
          isAuthenticated: false,
        }),
    }),
    {
      name: 'auth-storage',
    }
  )
)
```

### Types Example

```typescript
// features/authentication/types/index.ts

export interface AuthUser {
  id: string
  email: string
  name: string
  role: 'admin' | 'user' | 'viewer'
  avatar?: string
  createdAt: string
}

export interface SignInRequest {
  email: string
  password: string
}

export interface SignUpRequest {
  email: string
  password: string
  name: string
}

export interface AuthResponse {
  user: AuthUser
  token: string
  refreshToken: string
  expiresIn: number
}

export interface Session {
  token: string
  refreshToken: string
  expiresAt: number
}
```

## Route Integration Example

```typescript
// app/(auth)/sign-in/page.tsx
import { SignInForm, OAuthButtons } from '@/features/authentication'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import Link from 'next/link'

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Sign In</CardTitle>
          <CardDescription>Enter your credentials to access your account</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <SignInForm />

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-background px-2 text-muted-foreground">
                Or continue with
              </span>
            </div>
          </div>

          <OAuthButtons />

          <div className="text-center text-sm">
            Don't have an account?{' '}
            <Link href="/sign-up" className="text-primary hover:underline">
              Sign up
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
```

## Testing Example

```typescript
// features/authentication/__tests__/sign-in-form.test.tsx
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { SignInForm } from '../components/sign-in-form'

// Mock the hook
vi.mock('../hooks/use-sign-in', () => ({
  useSignIn: () => ({
    signIn: vi.fn().mockResolvedValue(true),
    isLoading: false,
    error: null,
  }),
}))

describe('SignInForm', () => {
  it('renders form fields', () => {
    render(<SignInForm />)

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('validates email format', async () => {
    render(<SignInForm />)

    const emailInput = screen.getByLabelText(/email/i)
    fireEvent.change(emailInput, { target: { value: 'invalid-email' } })
    fireEvent.blur(emailInput)

    await waitFor(() => {
      expect(screen.getByText(/invalid email/i)).toBeInTheDocument()
    })
  })

  it('validates password length', async () => {
    render(<SignInForm />)

    const passwordInput = screen.getByLabelText(/password/i)
    fireEvent.change(passwordInput, { target: { value: '123' } })
    fireEvent.blur(passwordInput)

    await waitFor(() => {
      expect(screen.getByText(/at least 8 characters/i)).toBeInTheDocument()
    })
  })
})
```

## Key Patterns Summary

1. **Public API**: Only export what external code needs via `index.ts`
2. **Encapsulation**: Hide implementation details (stores, API functions)
3. **Hooks as Interface**: Expose functionality through hooks, not direct store access
4. **Type Safety**: Define comprehensive TypeScript types
5. **Error Handling**: Handle errors at hook level, display at component level
6. **Loading States**: Manage loading state in hooks
7. **Form Validation**: Use Zod schemas with React Hook Form
8. **Routing**: Use Next.js router for navigation
9. **Testing**: Test components and hooks independently
10. **Naming**: Clear, descriptive names (use-sign-in, not use-login)
