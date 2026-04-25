import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { z } from 'zod'
import { CheckCircle2 } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { AuthLayout } from '@/components/layout/auth-layout'
import { ResetPasswordForm } from '@/features/authentication/components/reset-password-form'

const searchSchema = z.object({
  token: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/reset-password')({
  validateSearch: searchSchema,
  component: ResetPasswordPage,
})

function ResetPasswordPage() {
  const { token } = Route.useSearch()
  const [done, setDone] = useState(false)

  if (!token) {
    return (
      <AuthLayout>
        <Card className="gap-4">
          <CardHeader>
            <CardTitle className="text-lg tracking-tight">
              Invalid reset link
            </CardTitle>
            <CardDescription>
              This password reset link is missing its token. Request a new
              link to continue.
            </CardDescription>
          </CardHeader>
          <CardFooter>
            <Button asChild className="mx-auto">
              <Link to="/forgot-password">Request a new link</Link>
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  if (done) {
    return (
      <AuthLayout>
        <Card className="gap-4">
          <CardContent className="grid gap-4 pt-6 text-center">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-green-100">
              <CheckCircle2 className="h-6 w-6 text-green-600" />
            </div>
            <div className="grid gap-2">
              <h3 className="text-lg font-semibold">
                Password reset successfully
              </h3>
              <p className="text-muted-foreground text-sm">
                You can now sign in with your new password.
              </p>
            </div>
            <Button asChild>
              <Link to="/signin">Continue to sign in</Link>
            </Button>
          </CardContent>
        </Card>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout>
      <Card className="gap-4">
        <CardHeader>
          <CardTitle className="text-lg tracking-tight">
            Reset your password
          </CardTitle>
          <CardDescription>
            Choose a new password for your account.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ResetPasswordForm token={token} onSuccess={() => setDone(true)} />
        </CardContent>
        <CardFooter>
          <p className="text-muted-foreground mx-auto text-center text-sm">
            <Link
              to="/signin"
              className="hover:text-primary underline underline-offset-4"
            >
              Back to sign in
            </Link>
          </p>
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}
