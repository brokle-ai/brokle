'use client'

import { HTMLAttributes, useState, useEffect } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useSearchParams } from 'next/navigation'
import { IconFacebook, IconGithub } from '@/assets/brand-icons'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { PasswordInput } from '@/components/ui/password-input'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Loader2, AlertTriangle } from 'lucide-react'
import { useSignupMutation, useCompleteOAuthSignupMutation } from '../hooks/use-auth-queries'
import { validateInvitation } from '../api/auth-api'
import type { InvitationDetails } from '../types'

type SignupStep = 'auth' | 'personalization'

interface TwoStepSignUpFormProps extends HTMLAttributes<HTMLDivElement> {
  invitationToken?: string
  oauthSessionId?: string
  onStepChange?: (step: SignupStep) => void
}

// Step 1 Schema: Authentication (Email + Password)
const authStepSchema = z
  .object({
    email: z.string().email('Please enter a valid email address'),
    password: z
      .string()
      .min(1, 'Please enter your password')
      .min(8, 'Password must be at least 8 characters long'),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match.",
    path: ['confirmPassword'],
  })

export function TwoStepSignUpForm({
  className,
  invitationToken,
  oauthSessionId,
  onStepChange,
  ...props
}: TwoStepSignUpFormProps) {
  const searchParams = useSearchParams()
  const signupMutation = useSignupMutation()
  const oauthSignupMutation = useCompleteOAuthSignupMutation()

  // State
  const [step, setStep] = useState<SignupStep>(oauthSessionId ? 'personalization' : 'auth')

  // Notify parent when step changes
  useEffect(() => {
    onStepChange?.(step)
  }, [step, onStepChange])

  // Listen for back button click from parent
  useEffect(() => {
    const handleGoBack = () => setStep('auth')
    window.addEventListener('signup-go-back', handleGoBack)
    return () => window.removeEventListener('signup-go-back', handleGoBack)
  }, [])

  const [authData, setAuthData] = useState<{ email: string; password: string } | null>(null)
  const [invitationDetails, setInvitationDetails] = useState<InvitationDetails | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isRedirecting, setIsRedirecting] = useState(false)
  const [authError, setAuthError] = useState<string | null>(null)
  const [isLoadingInvitation, setIsLoadingInvitation] = useState(false)

  // Load invitation details if token provided
  useEffect(() => {
    if (invitationToken) {
      setIsLoadingInvitation(true)
      validateInvitation(invitationToken)
        .then((details) => {
          // Map snake_case API response to camelCase TypeScript type
          setInvitationDetails({
            organizationName: details.organization_name,
            organizationId: details.organization_id,
            inviterName: details.inviter_name,
            role: details.role,
            email: details.email,
            expiresAt: details.expires_at,
            isExpired: details.is_expired,
          })
          setIsLoadingInvitation(false)
        })
        .catch((error: any) => {
          console.error('Failed to validate invitation:', error)
          setAuthError('Invalid or expired invitation link')
          setIsLoadingInvitation(false)
        })
    }
  }, [invitationToken])

  // Determine if this is invitation-based signup
  const isInvitationSignup = !!(invitationToken || invitationDetails)

  // Step 2 Schema: Personalization (conditional validation based on signup type)
  const personalizationStepSchema = z.object({
    firstName: z.string().min(1, 'Please enter your first name'),
    lastName: z.string().min(1, 'Please enter your last name'),
    organizationName: isInvitationSignup
      ? z.string().optional() // Optional for invitation signup
      : z.string().min(1, 'Please enter your organization name'), // Required for fresh signup
    role: z.string().min(1, 'Please select your role'),
    referralSource: z.string().optional(),
  })

  // Step 1: Authentication Form
  const authForm = useForm<z.infer<typeof authStepSchema>>({
    resolver: zodResolver(authStepSchema),
    defaultValues: {
      email: invitationDetails?.email || '',
      password: '',
      confirmPassword: '',
    },
  })

  // Step 2: Personalization Form
  const personalForm = useForm<z.infer<typeof personalizationStepSchema>>({
    resolver: zodResolver(personalizationStepSchema),
    defaultValues: {
      firstName: '',
      lastName: '',
      organizationName: '',
      role: '',
      referralSource: '',
    },
  })

  // OAuth button handlers
  const handleGoogleSignup = () => {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    const params = new URLSearchParams()
    if (invitationToken) {
      params.set('invitation_token', invitationToken)
    }
    window.location.href = `${apiUrl}/api/v1/auth/google?${params.toString()}`
  }

  const handleGitHubSignup = () => {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    const params = new URLSearchParams()
    if (invitationToken) {
      params.set('invitation_token', invitationToken)
    }
    window.location.href = `${apiUrl}/api/v1/auth/github?${params.toString()}`
  }

  // Step 1 Submit: Move to Step 2
  const handleAuthSubmit = (data: z.infer<typeof authStepSchema>) => {
    setAuthData({ email: data.email, password: data.password })
    setStep('personalization')
  }

  // Step 2 Submit: Complete Registration
  const handlePersonalizationSubmit = async (data: z.infer<typeof personalizationStepSchema>) => {
    try {
      setAuthError(null)
      setIsSubmitting(true)

      if (oauthSessionId) {
        // OAuth completion flow - Use mutation hook (stores tokens automatically)
        await oauthSignupMutation.mutateAsync({
          sessionId: oauthSessionId,
          role: data.role,
          organizationName: data.organizationName,
          referralSource: data.referralSource,
        })
      } else {
        // Email/password signup - Use mutation hook (stores tokens automatically)
        if (!authData) {
          throw new Error('Auth data not found')
        }

        await signupMutation.mutateAsync({
          email: authData.email,
          password: authData.password,
          firstName: data.firstName,
          lastName: data.lastName,
          role: data.role,
          organizationName: data.organizationName,
          referralSource: data.referralSource,
          invitationToken: invitationToken,
        })
      }

      setIsRedirecting(true)

      // Redirect to dashboard
      const redirectUrl = searchParams.get('redirect') || '/'
      window.location.href = redirectUrl
    } catch (error) {
      setIsSubmitting(false)
      setIsRedirecting(false)

      if (error instanceof Error) {
        if (error.message?.includes('email')) {
          setAuthError('This email is already registered. Please try signing in instead.')
        } else if (error.message?.includes('organization')) {
          setAuthError('Failed to create organization. Please try again.')
        } else {
          setAuthError(error.message || 'Registration failed. Please try again.')
        }
      } else {
        setAuthError('An unexpected error occurred. Please try again.')
      }
    }
  }

  // Loading state for invitation validation
  if (invitationToken && isLoadingInvitation) {
    return (
      <div className="flex flex-col items-center justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        <p className="mt-4 text-sm text-muted-foreground">Validating invitation...</p>
      </div>
    )
  }

  // STEP 1: Authentication
  if (step === 'auth') {
    return (
      <div className={cn('space-y-4', className)} {...props}>
        {/* OAuth Buttons */}
        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-background px-2 text-muted-foreground">Sing up With</span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <Button variant="outline" type="button" disabled={isSubmitting} onClick={handleGitHubSignup}>
            <IconGithub className="mr-2 h-4 w-4" /> GitHub
          </Button>
          <Button variant="outline" type="button" disabled={isSubmitting} onClick={handleGoogleSignup}>
            <IconFacebook className="mr-2 h-4 w-4" /> Google
          </Button>
        </div>

        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-background px-2 text-muted-foreground">Or continue with email</span>
          </div>
        </div>

        {/* Email/Password Form */}
        <Form {...authForm}>
          <form onSubmit={authForm.handleSubmit(handleAuthSubmit)} className="space-y-4">
            <FormField
              control={authForm.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <Input
                      type="email"
                      placeholder="name@example.com"
                      disabled={!!invitationDetails}
                      {...field}
                      value={invitationDetails?.email || field.value}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={authForm.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Password</FormLabel>
                  <FormControl>
                    <PasswordInput placeholder="********" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={authForm.control}
              name="confirmPassword"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Confirm Password</FormLabel>
                  <FormControl>
                    <PasswordInput placeholder="********" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {authError && (
              <Alert variant="destructive">
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription>{authError}</AlertDescription>
              </Alert>
            )}

            <Button type="submit" className="w-full">
              Continue
            </Button>
          </form>
        </Form>
      </div>
    )
  }

  // STEP 2: Personalization
  const showOrgField = !invitationToken && !invitationDetails

  return (
    <div className={cn('space-y-6', className)} {...props}>
      <div className="space-y-2">
        <h2 className="text-2xl font-semibold tracking-tight">Tell us about yourself</h2>
        {invitationDetails && (
          <Alert className="mt-4">
            <AlertDescription>
              You're joining <strong>{invitationDetails.organizationName}</strong> as{' '}
              <strong>{invitationDetails.role}</strong>
            </AlertDescription>
          </Alert>
        )}
      </div>

      <Form {...personalForm}>
        <form onSubmit={personalForm.handleSubmit(handlePersonalizationSubmit)} className="space-y-4">
            <FormField
              control={personalForm.control}
              name="firstName"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>First Name</FormLabel>
                  <FormControl>
                    <Input placeholder="John" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={personalForm.control}
              name="lastName"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Last Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Doe" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

          {showOrgField && (
            <FormField
              control={personalForm.control}
              name="organizationName"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Organization Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Acme Corp" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          )}

          <FormField
            control={personalForm.control}
            name="role"
            render={({ field }) => (
              <FormItem>
                <FormLabel>What is your role?</FormLabel>
                <Select onValueChange={field.onChange} defaultValue={field.value}>
                  <FormControl>
                    <SelectTrigger className="!w-full">
                      <SelectValue placeholder="Select your role" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="engineer">Engineer</SelectItem>
                    <SelectItem value="product">Product Manager</SelectItem>
                    <SelectItem value="designer">Designer</SelectItem>
                    <SelectItem value="executive">Executive</SelectItem>
                    <SelectItem value="other">Other</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={personalForm.control}
            name="referralSource"
            render={({ field }) => (
              <FormItem>
                <FormLabel>How did you hear about us? (Optional)</FormLabel>
                <Select onValueChange={field.onChange} defaultValue={field.value}>
                  <FormControl>
                    <SelectTrigger className="!w-full">
                      <SelectValue placeholder="Select an option" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="search">Search Engine</SelectItem>
                    <SelectItem value="social">Social Media</SelectItem>
                    <SelectItem value="friend">Friend/Colleague</SelectItem>
                    <SelectItem value="blog">Blog/Article</SelectItem>
                    <SelectItem value="other">Other</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />

          {authError && (
            <Alert variant="destructive">
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>{authError}</AlertDescription>
            </Alert>
          )}

          <Button
            type="submit"
            className="w-full"
            disabled={signupMutation.isPending || oauthSignupMutation.isPending || isSubmitting || isRedirecting}
          >
            {(signupMutation.isPending || oauthSignupMutation.isPending || isSubmitting) ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Creating Account...
              </>
            ) : isRedirecting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Redirecting...
              </>
            ) : (
              'Create Account'
            )}
          </Button>

          {isRedirecting && (
            <Alert className="mt-4">
              <Loader2 className="h-4 w-4 animate-spin" />
              <AlertDescription>
                Welcome to Brokle! Taking you to your dashboard...
              </AlertDescription>
            </Alert>
          )}
        </form>
      </Form>
    </div>
  )
}
