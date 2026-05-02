import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { AuthLayout } from '@/components/layout/auth-layout'
import { ForgotPasswordForm } from '@/features/authentication'

export const Route = createFileRoute('/forgot-password')({
  component: ForgotPasswordPage,
})

function ForgotPasswordPage() {
  const [submittedEmail, setSubmittedEmail] = useState<string | null>(null)

  if (submittedEmail) {
    return (
      <AuthLayout>
        <Card className="gap-4">
          <CardHeader>
            <CardTitle className="text-lg tracking-tight">Check your inbox</CardTitle>
            <CardDescription>
              If an account exists for <strong>{submittedEmail}</strong>, you&apos;ll
              receive a reset link shortly.
            </CardDescription>
          </CardHeader>
          <CardFooter>
            <Link
              to="/signin"
              className="text-muted-foreground hover:text-primary mx-auto text-sm underline underline-offset-4"
            >
              Back to sign in
            </Link>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout>
      <Card className="gap-4">
        <CardHeader>
          <CardTitle className="text-lg tracking-tight">Forgot Password</CardTitle>
          <CardDescription>
            Enter your registered email and <br /> we will send you a link to
            reset your password.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ForgotPasswordForm onSuccess={setSubmittedEmail} />
        </CardContent>
        <CardFooter>
          <p className="text-muted-foreground mx-auto px-8 text-center text-sm text-balance">
            Don&apos;t have an account?{' '}
            <Link
              to="/signup"
              className="hover:text-primary underline underline-offset-4"
            >
              Sign up
            </Link>
            .
          </p>
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}
