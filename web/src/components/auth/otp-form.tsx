'use client'

import { HTMLAttributes, useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Link from 'next/link'
import { ArrowLeft, CheckCircle2 } from 'lucide-react'
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
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from '@/components/ui/input-otp'

type OtpFormProps = HTMLAttributes<HTMLFormElement> & {
  email?: string
}

const formSchema = z.object({
  otp: z.string().min(6, 'Please enter the complete verification code'),
})

export function OtpForm({ className, email = 'your email', ...props }: OtpFormProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [isVerified, setIsVerified] = useState(false)
  const [canResend, setCanResend] = useState(true)

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      otp: '',
    },
  })

  function onSubmit(data: z.infer<typeof formSchema>) {
    setIsLoading(true)
    // TODO: Implement actual OTP verification logic
    console.log(data)

    setTimeout(() => {
      setIsLoading(false)
      setIsVerified(true)
    }, 2000)
  }

  function onResend() {
    setCanResend(false)
    // TODO: Implement resend OTP logic
    console.log('Resending OTP...')
    
    // Re-enable resend after 60 seconds
    setTimeout(() => {
      setCanResend(true)
    }, 60000)
  }

  if (isVerified) {
    return (
      <div className='grid gap-4 text-center'>
        <div className='mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-green-100'>
          <CheckCircle2 className='h-6 w-6 text-green-600' />
        </div>
        <div className='grid gap-2'>
          <h3 className='text-lg font-semibold'>Email verified!</h3>
          <p className='text-sm text-muted-foreground'>
            Your email has been successfully verified.
          </p>
        </div>
        <Button asChild>
          <Link href='/'>Continue to Dashboard</Link>
        </Button>
      </div>
    )
  }

  return (
    <div className='grid gap-4'>
      <div className='grid gap-2 text-center'>
        <h1 className='text-2xl font-semibold tracking-tight'>
          Verify your email
        </h1>
        <p className='text-sm text-muted-foreground'>
          Enter the 6-digit verification code sent to {email}
        </p>
      </div>
      <Form {...form}>
        <form
          onSubmit={form.handleSubmit(onSubmit)}
          className={cn('grid gap-4', className)}
          {...props}
        >
          <FormField
            control={form.control}
            name='otp'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Verification Code</FormLabel>
                <FormControl>
                  <InputOTP maxLength={6} {...field}>
                    <InputOTPGroup className='mx-auto'>
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
          <Button disabled={isLoading} className='w-full'>
            {isLoading ? 'Verifying...' : 'Verify Email'}
          </Button>
        </form>
      </Form>
      <div className='text-center text-sm'>
        <span className='text-muted-foreground'>Didn't receive the code? </span>
        <Button
          variant='link'
          size='sm'
          disabled={!canResend}
          onClick={onResend}
          className='px-0'
        >
          {canResend ? 'Resend code' : 'Resend in 60s'}
        </Button>
      </div>
      <Link
        href='/auth/signin'
        className='inline-flex items-center justify-center text-sm text-muted-foreground hover:text-foreground'
      >
        <ArrowLeft className='mr-2 h-4 w-4' />
        Back to login
      </Link>
    </div>
  )
}