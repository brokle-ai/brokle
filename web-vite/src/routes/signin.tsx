import { createFileRoute, Link, useSearch } from '@tanstack/react-router'
import { z } from 'zod'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { AuthLayout } from '@/components/layout/auth-layout'
import {
  SignInForm,
  SignInToastHandler,
} from '@/features/authentication'

const searchSchema = z.object({
  redirect: z.string().optional().catch(undefined),
  logout: z.string().optional().catch(undefined),
  session: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/signin')({
  validateSearch: searchSchema,
  component: SignInPage,
})

function SignInPage() {
  const { redirect, logout, session } = useSearch({ from: '/signin' })

  return (
    <AuthLayout>
      <SignInToastHandler logout={logout} session={session} />
      <Card className="gap-4">
        <CardHeader>
          <CardTitle className="text-lg tracking-tight">Login</CardTitle>
          <CardDescription>
            Enter your email and password below to <br />
            log into your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <SignInForm redirectTo={redirect} />
        </CardContent>
        <CardFooter className="flex flex-col space-y-4">
          <div className="text-center text-sm">
            <span className="text-muted-foreground">Don&apos;t have an account? </span>
            <Link
              to="/signup"
              className="hover:text-primary font-medium underline underline-offset-4"
            >
              Sign up
            </Link>
          </div>
          <p className="text-muted-foreground px-8 text-center text-sm">
            By clicking login, you agree to our{' '}
            <a
              href="/terms"
              className="hover:text-primary underline underline-offset-4"
            >
              Terms of Service
            </a>{' '}
            and{' '}
            <a
              href="/privacy"
              className="hover:text-primary underline underline-offset-4"
            >
              Privacy Policy
            </a>
            .
          </p>
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}
