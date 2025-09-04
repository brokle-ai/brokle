'use client'

import { HTMLAttributes } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
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
import { useRequestPasswordResetMutation } from '@/hooks/api/use-auth-queries'
import { Loader2 } from 'lucide-react'

type ForgotPasswordFormProps = HTMLAttributes<HTMLFormElement>

const formSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
})

export function ForgotPasswordForm({ className, ...props }: ForgotPasswordFormProps) {
  const requestResetMutation = useRequestPasswordResetMutation()

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: { email: '' },
  })

  async function onSubmit(data: z.infer<typeof formSchema>) {
    try {
      await requestResetMutation.mutateAsync(data.email)
    } catch (error) {
      // Error is already handled by the mutation
      console.error('Password reset request failed:', error)
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-2', className)}
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
        <Button className='mt-2' disabled={requestResetMutation.isPending} type='submit'>
          {requestResetMutation.isPending ? (
            <>
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              Sending...
            </>
          ) : (
            'Continue'
          )}
        </Button>
        
        {requestResetMutation.error && (
          <div className='mt-2 text-sm text-red-600 text-center'>
            {requestResetMutation.error.message || 'Failed to send reset email. Please try again.'}
          </div>
        )}
      </form>
    </Form>
  )
}