import { Link } from '@tanstack/react-router'
import { type HTMLAttributes, useRef, useState } from 'react'
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
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

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
  const setUser = useAuthStore((s) => s.setUser)
  const [isRedirecting, setIsRedirecting] = useState(false)
  const [authError, setAuthError] = useState<string | null>(null)
  // Fire-once guard, same invariant as web/'s SignInForm — blocks any
  // post-success duplicate submit delivered while the browser is still
  // tearing down for navigation.
  const hasSucceededRef = useRef(false)

  const loginMutation = useMutation({
    mutationFn: login,
  })

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: { email: '', password: '' },
  })

  async function onSubmit(data: FormValues) {
    if (hasSucceededRef.current) return
    setAuthError(null)
    setIsRedirecting(false)

    try {
      const resp = await loginMutation.mutateAsync(data)
      setUser(resp.user)
      hasSucceededRef.current = true
      setIsRedirecting(true)
      window.location.href = redirectTo ?? '/'
    } catch (error) {
      setIsRedirecting(false)
      setAuthError(deriveLoginError(error))
    }
  }

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

        {isRedirecting && (
          <Alert>
            <Loader2 className="h-4 w-4 animate-spin" />
            <AlertDescription>
              Welcome back! Taking you to your dashboard...
            </AlertDescription>
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
        <Button
          className="mt-2"
          type="submit"
          disabled={loginMutation.isPending || isRedirecting}
        >
          {loginMutation.isPending ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Signing in...
            </>
          ) : isRedirecting ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Redirecting...
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
            disabled={loginMutation.isPending || isRedirecting}
            onClick={handleOAuth('github')}
          >
            <IconGithub className="h-4 w-4" /> GitHub
          </Button>
          <Button
            variant="outline"
            type="button"
            disabled={loginMutation.isPending || isRedirecting}
            onClick={handleOAuth('google')}
          >
            <IconFacebook className="h-4 w-4" /> Google
          </Button>
        </div>
      </form>
    </Form>
  )
}
