import { Link, useNavigate, useRouter } from '@tanstack/react-router'
import { type HTMLAttributes, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { z } from 'zod'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { IconFacebook, IconGithub } from '@/assets/brand-icons'
import { Alert, AlertDescription } from '@/components/ui/alert'
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
import { login } from '@/features/authentication/api/auth-api'
import { buildOAuthUrl } from '@/features/authentication/utils/oauth'
import { BrokleError } from '@/lib/api/errors'
import { resetSession } from '@/lib/auth/session'
import { cn } from '@/lib/utils'

type SignInFormProps = HTMLAttributes<HTMLFormElement> & {
  redirectTo?: string
}

const formSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
  password: z
    .string()
    .min(1, 'Please enter your password')
    .min(7, 'Password must be at least 7 characters long'),
})

type FormValues = z.infer<typeof formSchema>

function deriveLoginError(error: unknown): string {
  if (error instanceof BrokleError) {
    return error.body.error.message
  }
  if (error instanceof Error) {
    if (error.message.includes('Network')) {
      return 'Unable to connect. Please check your internet connection and try again.'
    }
    if (error.message.includes('credentials')) {
      return 'Invalid email or password. Please check your credentials and try again.'
    }
    return error.message || 'Login failed. Please try again.'
  }
  return 'An unexpected error occurred. Please try again.'
}

export function SignInForm({ className, redirectTo, ...props }: SignInFormProps) {
  const navigate = useNavigate()
  const router = useRouter()
  const [authError, setAuthError] = useState<string | null>(null)

  const loginMutation = useMutation({
    mutationFn: login,
  })

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: { email: '', password: '' },
  })

  // The form's only responsibility is "trigger navigation after the
  // credential POST succeeds." Post-login validation lives in
  // `_authenticated.beforeLoad` (the /me probe). On failure modes:
  //   - cookie dropped → beforeLoad redirects to /signin?session=expired,
  //     this component re-mounts fresh with a toast (handled by
  //     SignInToastHandler).
  //   - /me 5xx → RootErrorFallback renders.
  // No local "isRedirecting" or fire-once latch — those leaked across
  // redirect-back-to-/signin because TanStack Router reuses the route
  // component on same-route redirects, and `navigate()` doesn't reject
  // for redirects/errors so a try/catch would never reset them.
  // `loginMutation.isPending || form.formState.isSubmitting` is the
  // correct duplicate-submit guard.
  async function onSubmit(data: FormValues) {
    setAuthError(null)
    try {
      await loginMutation.mutateAsync(data)
      resetSession(router.options.context.queryClient)
      await navigate({ to: redirectTo ?? '/', replace: true })
    } catch (error) {
      setAuthError(deriveLoginError(error))
    }
  }

  const isSubmitting = loginMutation.isPending || form.formState.isSubmitting

  const handleOAuth = (provider: 'google' | 'github') => () => {
    window.location.href = buildOAuthUrl(provider)
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-3', className)}
        {...props}
      >
        {authError && (
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>{authError}</AlertDescription>
          </Alert>
        )}

        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl>
                <Input placeholder="name@example.com" autoComplete="email" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem className="relative">
              <FormLabel>Password</FormLabel>
              <FormControl>
                <PasswordInput
                  placeholder="********"
                  autoComplete="current-password"
                  {...field}
                />
              </FormControl>
              <FormMessage />
              <Link
                to="/forgot-password"
                className="text-muted-foreground hover:text-primary absolute end-0 -top-0.5 text-sm font-medium hover:opacity-75"
              >
                Forgot password?
              </Link>
            </FormItem>
          )}
        />
        <Button className="mt-2" type="submit" disabled={isSubmitting}>
          {isSubmitting ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Signing in...
            </>
          ) : (
            'Sign In'
          )}
        </Button>

        <div className="relative my-2">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-background text-muted-foreground px-2">
              Or continue with
            </span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-2">
          <Button
            variant="outline"
            type="button"
            disabled={isSubmitting}
            onClick={handleOAuth('github')}
          >
            <IconGithub className="h-4 w-4" /> GitHub
          </Button>
          <Button
            variant="outline"
            type="button"
            disabled={isSubmitting}
            onClick={handleOAuth('google')}
          >
            <IconFacebook className="h-4 w-4" /> Google
          </Button>
        </div>
      </form>
    </Form>
  )
}
