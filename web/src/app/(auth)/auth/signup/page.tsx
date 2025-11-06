'use client'

import { Suspense, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ArrowLeft, Loader2 } from 'lucide-react'
import {
  AuthLayout,
  TwoStepSignUpForm,
  AuthFormWrapper
} from '@/features/authentication'
import Link from 'next/link'

type SignupStep = 'auth' | 'personalization'

function SignUpContent() {
  const searchParams = useSearchParams()
  const invitationToken = searchParams.get('token') || undefined
  const oauthSessionId = searchParams.get('session') || undefined
  const [currentStep, setCurrentStep] = useState<SignupStep>(oauthSessionId ? 'personalization' : 'auth')

  return (
    <AuthLayout>
      <Card className='gap-4'>
        <CardHeader>
          <CardTitle className='text-lg tracking-tight'>
            {invitationToken ? 'Join your team' : 'Create your account'}
          </CardTitle>
          <CardDescription>
            {invitationToken
              ? 'Complete your account to join your team'
              : 'Get started with Brokle in seconds'}{' '}
            <br />
            Already have an account?{' '}
            <Link
              href='/auth/signin'
              className='hover:text-primary underline underline-offset-4'
            >
              Sign In
            </Link>
          </CardDescription>
        </CardHeader>
        <CardContent>
          <AuthFormWrapper>
            <TwoStepSignUpForm
              invitationToken={invitationToken}
              oauthSessionId={oauthSessionId}
              onStepChange={setCurrentStep}
            />
          </AuthFormWrapper>
        </CardContent>
        <CardFooter className="flex flex-col items-center gap-4">
          <p className='text-muted-foreground px-8 text-center text-sm'>
            By creating an account, you agree to our{' '}
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
          {currentStep === 'personalization' && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                // Trigger step change in the form component
                window.dispatchEvent(new CustomEvent('signup-go-back'))
              }}
              className="text-muted-foreground"
            >
              <ArrowLeft className="mr-2 h-4 w-4" /> Back
            </Button>
          )}
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}

export default function SignUpPage() {
  return (
    <Suspense
      fallback={
        <AuthLayout>
          <div className="flex min-h-[400px] items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        </AuthLayout>
      }
    >
      <SignUpContent />
    </Suspense>
  )
}