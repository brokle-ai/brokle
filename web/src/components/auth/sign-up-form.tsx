'use client'

import { HTMLAttributes } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useSearchParams } from 'next/navigation'
import { IconFacebook, IconGithub } from '@/assets/brand-icons'
import { cn } from '@/lib/utils'
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
import { Alert, AlertDescription } from '@/components/ui/alert'
import { useSignupMutation } from '@/hooks/api/use-auth-queries'
import { Loader2, AlertTriangle } from 'lucide-react'

type SignUpFormProps = HTMLAttributes<HTMLFormElement>

const formSchema = z
  .object({
    firstName: z.string().min(1, 'Please enter your first name'),
    lastName: z.string().min(1, 'Please enter your last name'),
    email: z.string().email('Please enter a valid email address'),
    password: z
      .string()
      .min(1, 'Please enter your password')
      .min(7, 'Password must be at least 7 characters long'),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match.",
    path: ['confirmPassword'],
  })

export function SignUpForm({ className, ...props }: SignUpFormProps) {
  const searchParams = useSearchParams()
  const signupMutation = useSignupMutation()
  const [isRedirecting, setIsRedirecting] = useState(false)
  const [authError, setAuthError] = useState<string | null>(null)

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      firstName: '',
      lastName: '',
      email: '',
      password: '',
      confirmPassword: '',
    },
  })

  async function onSubmit(data: z.infer<typeof formSchema>) {
    try {
      setAuthError(null)
      setIsRedirecting(false)

      await signupMutation.mutateAsync({
        firstName: data.firstName,
        lastName: data.lastName,
        email: data.email,
        password: data.password,
      })

      setIsRedirecting(true)

      // Get redirect URL from search params or default to dashboard
      const redirectUrl = searchParams.get('redirect') || '/dashboard'
      
      // Use full page refresh to ensure middleware sees the new cookies
      // This is the most reliable way to handle post-signup navigation
      window.location.href = redirectUrl
    } catch (error) {
      setIsRedirecting(false)
      
      // Enhanced error handling with user-friendly messages
      if (error instanceof Error) {
        if (error.message?.includes('Network')) {
          setAuthError('Unable to connect. Please check your internet connection and try again.')
        } else if (error.message?.includes('email')) {
          setAuthError('This email is already registered. Please try signing in instead.')
        } else if (error.message?.includes('timeout')) {
          setAuthError('The request is taking too long. Please try again.')
        } else {
          setAuthError(error.message || 'Registration failed. Please try again.')
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
        <div className='grid grid-cols-2 gap-3'>
          <FormField
            control={form.control}
            name='firstName'
            render={({ field }) => (
              <FormItem>
                <FormLabel>First Name</FormLabel>
                <FormControl>
                  <Input placeholder='John' {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='lastName'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Last Name</FormLabel>
                <FormControl>
                  <Input placeholder='Doe' {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
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
            <FormItem>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <PasswordInput placeholder='********' {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='confirmPassword'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Confirm Password</FormLabel>
              <FormControl>
                <PasswordInput placeholder='********' {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button 
          className='mt-2' 
          disabled={signupMutation.isPending || isRedirecting} 
          type='submit'
        >
          {signupMutation.isPending ? (
            <>
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              Creating Account...
            </>
          ) : isRedirecting ? (
            <>
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              Redirecting...
            </>
          ) : (
            'Create Account'
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
          <Button variant='outline' type='button' disabled={signupMutation.isPending || isRedirecting}>
            <IconGithub className='h-4 w-4' /> GitHub
          </Button>
          <Button variant='outline' type='button' disabled={signupMutation.isPending || isRedirecting}>
            <IconFacebook className='h-4 w-4' /> Facebook
          </Button>
        </div>
        
        {/* Enhanced error display */}
        {(authError || signupMutation.error) && (
          <Alert variant="destructive" className="mt-4">
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              {authError || signupMutation.error?.message || 'Registration failed. Please try again.'}
            </AlertDescription>
          </Alert>
        )}
        
        {/* Success state when redirecting */}
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
  )
}