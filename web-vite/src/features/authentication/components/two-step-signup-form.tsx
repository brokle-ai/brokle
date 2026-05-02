import { type HTMLAttributes, useEffect, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { z } from 'zod'
import { AlertTriangle, Loader2, Lock, Mail } from 'lucide-react'
import { IconFacebook, IconGithub } from '@/assets/brand-icons'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { PasswordInput } from '@/components/ui/password-input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  completeOAuthSignup,
  signup,
} from '@/features/authentication/api/auth-api'
import type { InvitationDetails } from '@/features/authentication/types'
import { buildOAuthUrl } from '@/features/authentication/utils/oauth'
import { BrokleError } from '@/lib/api/errors'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

type SignupStep = 'auth' | 'personalization'

interface TwoStepSignUpFormProps extends HTMLAttributes<HTMLDivElement> {
  invitationToken?: string
  invitationDetails?: InvitationDetails | null
  oauthSessionId?: string
  redirectTo?: string
  onStepChange?: (step: SignupStep) => void
}

function deriveSignupError(error: unknown): string {
  if (error instanceof BrokleError) {
    return error.body.error.message
  }
  if (!(error instanceof Error)) {
    return 'An unexpected error occurred. Please try again.'
  }
  if (error.message.includes('email')) {
    return 'This email is already registered. Please try signing in instead.'
  }
  if (error.message.includes('organization')) {
    return 'Failed to create organization. Please try again.'
  }
  if (error.message.includes('Network')) {
    return 'Unable to connect. Please check your internet connection and try again.'
  }
  return error.message || 'Registration failed. Please try again.'
}

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

type AuthStepValues = z.infer<typeof authStepSchema>

export function TwoStepSignUpForm({
  className,
  invitationToken,
  invitationDetails,
  oauthSessionId,
  redirectTo,
  onStepChange,
  ...props
}: TwoStepSignUpFormProps) {
  const setUser = useAuthStore((s) => s.setUser)

  const signupMutation = useMutation({ mutationFn: signup })
  const oauthSignupMutation = useMutation({ mutationFn: completeOAuthSignup })

  const [step, setStep] = useState<SignupStep>(
    oauthSessionId ? 'personalization' : 'auth',
  )

  useEffect(() => {
    onStepChange?.(step)
  }, [step, onStepChange])

  useEffect(() => {
    const handleGoBack = () => setStep('auth')
    window.addEventListener('signup-go-back', handleGoBack)
    return () => window.removeEventListener('signup-go-back', handleGoBack)
  }, [])

  const [authData, setAuthData] = useState<{ email: string; password: string } | null>(
    null,
  )
  const [isRedirecting, setIsRedirecting] = useState(false)
  const [authError, setAuthError] = useState<string | null>(null)
  const hasSucceededRef = useRef(false)

  const isInvitationSignup = Boolean(invitationToken || invitationDetails)

  const personalizationStepSchema = z.object({
    firstName: z.string().min(1, 'Please enter your first name'),
    lastName: z.string().min(1, 'Please enter your last name'),
    organizationName: isInvitationSignup
      ? z.string().optional()
      : z.string().min(1, 'Please enter your organization name'),
    role: z.string().min(1, 'Please select your role'),
    referralSource: z.string().optional(),
  })
  type PersonalValues = z.infer<typeof personalizationStepSchema>

  const authForm = useForm<AuthStepValues>({
    resolver: zodResolver(authStepSchema),
    defaultValues: {
      email: invitationDetails?.email ?? '',
      password: '',
      confirmPassword: '',
    },
  })

  useEffect(() => {
    if (invitationDetails?.email) {
      authForm.setValue('email', invitationDetails.email)
    }
  }, [invitationDetails, authForm])

  const personalForm = useForm<PersonalValues>({
    resolver: zodResolver(personalizationStepSchema),
    defaultValues: {
      firstName: '',
      lastName: '',
      organizationName: '',
      role: '',
      referralSource: '',
    },
  })

  const handleOAuth = (provider: 'google' | 'github') => () => {
    window.location.href = buildOAuthUrl(provider, invitationToken)
  }

  const handleAuthSubmit = (data: AuthStepValues) => {
    setAuthData({ email: data.email, password: data.password })
    setStep('personalization')
  }

  const handlePersonalizationSubmit = async (data: PersonalValues) => {
    if (hasSucceededRef.current) return
    setAuthError(null)

    try {
      if (oauthSessionId) {
        const resp = await oauthSignupMutation.mutateAsync({
          session_id: oauthSessionId,
          role: data.role,
          organization_name: data.organizationName,
          referral_source: data.referralSource,
        })
        setUser(resp.user)
      } else {
        if (!authData) throw new Error('Auth data not found')
        const resp = await signupMutation.mutateAsync({
          email: authData.email,
          password: authData.password,
          first_name: data.firstName,
          last_name: data.lastName,
          role: data.role,
          organization_name: data.organizationName,
          referral_source: data.referralSource,
          invitation_token: invitationToken,
        })
        setUser(resp.user)
      }

      hasSucceededRef.current = true
      setIsRedirecting(true)
      window.location.href = redirectTo ?? '/'
    } catch (error) {
      setIsRedirecting(false)
      setAuthError(deriveSignupError(error))
    }
  }

  if (step === 'auth') {
    return (
      <div className={cn('space-y-4', className)} {...props}>
        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-background text-muted-foreground px-2">
              Sign up with
            </span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <Button variant="outline" type="button" onClick={handleOAuth('github')}>
            <IconGithub className="mr-2 h-4 w-4" /> GitHub
          </Button>
          <Button variant="outline" type="button" onClick={handleOAuth('google')}>
            <IconFacebook className="mr-2 h-4 w-4" /> Google
          </Button>
        </div>

        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-background text-muted-foreground px-2">
              Or continue with email
            </span>
          </div>
        </div>

        <Form {...authForm}>
          <form
            onSubmit={authForm.handleSubmit(handleAuthSubmit)}
            className="space-y-4"
          >
            {invitationDetails ? (
              <div className="space-y-2">
                <FormLabel>Email</FormLabel>
                <div className="bg-muted/50 flex items-center gap-2 rounded-md border p-3">
                  <Mail className="text-muted-foreground h-4 w-4" />
                  <span className="flex-1 text-sm font-medium">
                    {invitationDetails.email}
                  </span>
                  <Badge variant="secondary" className="gap-1">
                    <Lock className="h-3 w-3" />
                    Locked
                  </Badge>
                </div>
                <input
                  type="hidden"
                  {...authForm.register('email')}
                  value={invitationDetails.email}
                />
              </div>
            ) : (
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
                        autoComplete="email"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            <FormField
              control={authForm.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Password</FormLabel>
                  <FormControl>
                    <PasswordInput
                      placeholder="********"
                      autoComplete="new-password"
                      {...field}
                    />
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
                    <PasswordInput
                      placeholder="********"
                      autoComplete="new-password"
                      {...field}
                    />
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

  const showOrgField = !invitationToken && !invitationDetails
  const isPending = signupMutation.isPending || oauthSignupMutation.isPending

  return (
    <div className={cn('space-y-6', className)} {...props}>
      <div className="space-y-2">
        <h2 className="text-2xl font-semibold tracking-tight">
          Tell us about yourself
        </h2>
      </div>

      <Form {...personalForm}>
        <form
          onSubmit={personalForm.handleSubmit(handlePersonalizationSubmit)}
          className="space-y-4"
        >
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
            disabled={isPending || isRedirecting}
          >
            {isPending ? (
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
