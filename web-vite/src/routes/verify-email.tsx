import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { ArrowLeft, CheckCircle2 } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { AuthLayout } from '@/components/layout/auth-layout'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from '@/components/ui/input-otp'

const searchSchema = z.object({
  email: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/verify-email')({
  validateSearch: searchSchema,
  component: VerifyEmailPage,
})

const formSchema = z.object({
  otp: z.string().min(6, 'Please enter the complete verification code'),
})

type FormValues = z.infer<typeof formSchema>

function VerifyEmailPage() {
  const { email } = Route.useSearch()
  const [isLoading, setIsLoading] = useState(false)
  const [isVerified, setIsVerified] = useState(false)
  const [canResend, setCanResend] = useState(true)

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: { otp: '' },
  })

  // NOTE: web/ ships this page as a stubbed OTP flow
  // (`/verify-email/page.tsx` has a TODO call). Kept at parity — the
  // backend `/verify-email` wire-up is a future task.
  function onSubmit(_data: FormValues) {
    setIsLoading(true)
    setTimeout(() => {
      setIsLoading(false)
      setIsVerified(true)
    }, 1500)
  }

  function onResend() {
    setCanResend(false)
    setTimeout(() => setCanResend(true), 60_000)
  }

  if (isVerified) {
    return (
      <AuthLayout>
        <Card className="gap-4">
          <CardContent className="grid gap-4 pt-6 text-center">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-green-100">
              <CheckCircle2 className="h-6 w-6 text-green-600" />
            </div>
            <div className="grid gap-2">
              <h3 className="text-lg font-semibold">Email verified!</h3>
              <p className="text-muted-foreground text-sm">
                Your email has been successfully verified.
              </p>
            </div>
            <Button asChild>
              <Link to="/">Continue to Dashboard</Link>
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
          <CardTitle className="text-base tracking-tight">
            Two-factor Authentication
          </CardTitle>
          <CardDescription>
            Please enter the authentication code. <br /> We have sent the
            authentication code to {email ?? 'your email'}.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSubmit)}
              className="grid gap-4"
            >
              <FormField
                control={form.control}
                name="otp"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Verification Code</FormLabel>
                    <FormControl>
                      <InputOTP maxLength={6} {...field}>
                        <InputOTPGroup className="mx-auto">
                          <InputOTPSlot index={0} />
                          <InputOTPSlot index={1} />
                          <InputOTPSlot index={2} />
                          <InputOTPSlot index={3} />
                          <InputOTPSlot index={4} />
                          <InputOTPSlot index={5} />
                        </InputOTPGroup>
                      </InputOTP>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button disabled={isLoading} className="w-full">
                {isLoading ? 'Verifying...' : 'Verify Email'}
              </Button>
            </form>
          </Form>
          <div className="mt-4 text-center text-sm">
            <span className="text-muted-foreground">
              Didn&apos;t receive the code?{' '}
            </span>
            <Button
              variant="link"
              size="sm"
              disabled={!canResend}
              onClick={onResend}
              className="px-0"
            >
              {canResend ? 'Resend code' : 'Resend in 60s'}
            </Button>
          </div>
        </CardContent>
        <CardFooter>
          <Link
            to="/signin"
            className="text-muted-foreground hover:text-foreground mx-auto inline-flex items-center justify-center text-sm"
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to login
          </Link>
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}
