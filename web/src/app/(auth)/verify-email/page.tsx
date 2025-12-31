import { Metadata } from 'next'
import { Suspense } from 'react'
import Link from 'next/link'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { AuthLayout, OTPForm } from '@/features/authentication'
import { ROUTES } from '@/lib/routes'

export const metadata: Metadata = {
  title: 'Verify Email',
  description: 'Verify your email address',
}

function OtpFormWrapper() {
  // In a real app, you'd get the email from search params or session
  const email = 'user@example.com'

  return <OTPForm email={email} />
}

export default function VerifyEmailPage() {
  return (
    <AuthLayout>
      <Card className='gap-4'>
        <CardHeader>
          <CardTitle className='text-base tracking-tight'>
            Two-factor Authentication
          </CardTitle>
          <CardDescription>
            Please enter the authentication code. <br /> We have sent the
            authentication code to your email.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Suspense fallback={<div>Loading...</div>}>
            <OtpFormWrapper />
          </Suspense>
        </CardContent>
        <CardFooter>
          <p className='text-muted-foreground px-8 text-center text-sm'>
            Haven't received it?{' '}
            <Link
              href={ROUTES.SIGNIN}
              className='hover:text-primary underline underline-offset-4'
            >
              Resend a new code.
            </Link>
            .
          </p>
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}