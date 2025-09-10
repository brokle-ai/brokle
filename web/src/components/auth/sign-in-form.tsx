'use client'

import { HTMLAttributes } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { IconFacebook, IconGithub } from '@/assets/brand-icons'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
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
import { Alert, AlertDescription } from '@/components/ui/alert'
import { useLoginMutation } from '@/hooks/api/use-auth-queries'
import { Loader2, AlertTriangle } from 'lucide-react'

type SignInFormProps = HTMLAttributes<HTMLFormElement>

const formSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
  password: z
    .string()
    .min(1, 'Please enter your password')
    .min(7, 'Password must be at least 7 characters long'),
  rememberMe: z.boolean().default(false),
})

export function SignInForm({ className, ...props }: SignInFormProps) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const loginMutation = useLoginMutation()
  const [isRedirecting, setIsRedirecting] = useState(false)
  const [authError, setAuthError] = useState<string | null>(null)

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      email: '',
      password: '',
      rememberMe: false,
    },
  })

  async function onSubmit(data: z.infer<typeof formSchema>) {
    try {
      setAuthError(null)
      setIsRedirecting(false)

      await loginMutation.mutateAsync({
        email: data.email,
        password: data.password,
        rememberMe: data.rememberMe,
      })

      setIsRedirecting(true)

      // Get redirect URL from search params or default to root
      const redirectUrl = searchParams.get('redirect') || '/'
      
      // Use full page refresh to ensure middleware sees the new cookies
      // This is the most reliable way to handle post-login navigation
      window.location.href = redirectUrl
    } catch (error) {
      setIsRedirecting(false)
      
      // Enhanced error handling with user-friendly messages
      if (error instanceof Error) {
        if (error.message?.includes('Network')) {
          setAuthError('Unable to connect. Please check your internet connection and try again.')
        } else if (error.message?.includes('credentials')) {
          setAuthError('Invalid email or password. Please check your credentials and try again.')
        } else if (error.message?.includes('timeout')) {
          setAuthError('The request is taking too long. Please try again.')
        } else {
          setAuthError(error.message || 'Login failed. Please try again.')
        }
      } else {
        setAuthError('An unexpected error occurred. Please try again.')
      }
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-3', className)}
        {...props}
      >
        <FormField
          control={form.control}
          name='email'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl>
                <Input placeholder='name@example.com' {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='password'
          render={({ field }) => (
            <FormItem className='relative'>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <PasswordInput placeholder='********' {...field} />
              </FormControl>
              <FormMessage />
              <Link
                href='/auth/forgot-password'
                className='text-muted-foreground absolute end-0 -top-0.5 text-sm font-medium hover:opacity-75'
              >
                Forgot password?
              </Link>
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='rememberMe'
          render={({ field }) => (
            <FormItem className='flex flex-row items-start space-x-3 space-y-0'>
              <FormControl>
                <Checkbox
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              </FormControl>
              <div className='space-y-1 leading-none'>
                <FormLabel className='text-sm font-normal'>
                  Remember me for 30 days
                </FormLabel>
              </div>
            </FormItem>
          )}
        />
        
        <Button 
          className='mt-2' 
          disabled={loginMutation.isPending || isRedirecting} 
          type='submit'
        >
          {loginMutation.isPending ? (
            <>
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              Signing in...
            </>
          ) : isRedirecting ? (
            <>
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              Redirecting...
            </>
          ) : (
            'Sign In'
          )}
        </Button>

        <div className='relative my-2'>
          <div className='absolute inset-0 flex items-center'>
            <span className='w-full border-t' />
          </div>
          <div className='relative flex justify-center text-xs uppercase'>
            <span className='bg-background text-muted-foreground px-2'>
              Or continue with
            </span>
          </div>
        </div>

        <div className='grid grid-cols-2 gap-2'>
          <Button variant='outline' type='button' disabled={loginMutation.isPending || isRedirecting}>
            <IconGithub className='h-4 w-4' /> GitHub
          </Button>
          <Button variant='outline' type='button' disabled={loginMutation.isPending || isRedirecting}>
            <IconFacebook className='h-4 w-4' /> Facebook
          </Button>
        </div>
        
        {/* Enhanced error display */}
        {(authError || loginMutation.error) && (
          <Alert variant="destructive" className="mt-4">
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              {authError || loginMutation.error?.message || 'Login failed. Please try again.'}
            </AlertDescription>
          </Alert>
        )}
        
        {/* Success state when redirecting */}
        {isRedirecting && (
          <Alert className="mt-4">
            <Loader2 className="h-4 w-4 animate-spin" />
            <AlertDescription>
              Welcome back! Taking you to your dashboard...
            </AlertDescription>
          </Alert>
        )}
      </form>
    </Form>
  )
}