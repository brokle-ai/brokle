import { Metadata } from 'next'
import { Suspense } from 'react'
import Link from 'next/link'
import {
  AuthLayout,
  SignInForm,
  SignInToastHandler,
  AuthFormWrapper
} from '@/features/authentication'
import { ROUTES } from '@/lib/routes'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export const metadata: Metadata = {
  title: 'Sign In',
  description: 'Sign in to your Brokle account',
}

export default function SignInPage() {
  return (
    <AuthLayout>
      <Suspense fallback={null}>
        <SignInToastHandler />
      </Suspense>
      <Card className='gap-4'>
        <CardHeader>
          <CardTitle className='text-lg tracking-tight'>Login</CardTitle>
          <CardDescription>
            Enter your email and password below to <br />
            log into your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <AuthFormWrapper>
            <SignInForm />
          </AuthFormWrapper>
        </CardContent>
        <CardFooter className="flex flex-col space-y-4">
          <div className="text-center text-sm">
            <span className="text-muted-foreground">Don't have an account? </span>
            <Link
              href={ROUTES.SIGNUP}
              className="font-medium underline underline-offset-4 hover:text-primary"
            >
              Sign up
            </Link>
          </div>
          <p className='text-muted-foreground px-8 text-center text-sm'>
            By clicking login, you agree to our{' '}
            <Link
              href='/terms'
              className='hover:text-primary underline underline-offset-4'
            >
              Terms of Service
            </Link>{' '}
            and{' '}
            <Link
              href='/privacy'
              className='hover:text-primary underline underline-offset-4'
            >
              Privacy Policy
            </Link>
            .
          </p>
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}