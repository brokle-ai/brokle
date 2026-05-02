import { type HTMLAttributes } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { z } from 'zod'
import { Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { PasswordInput } from '@/components/ui/password-input'
import { resetPassword } from '@/features/authentication/api/auth-api'
import { BrokleError } from '@/lib/api/errors'
import { cn } from '@/lib/utils'

type ResetPasswordFormProps = HTMLAttributes<HTMLFormElement> & {
  token: string
  onSuccess?: () => void
}

const formSchema = z
  .object({
    newPassword: z
      .string()
      .min(1, 'Please enter a new password')
      .min(8, 'Password must be at least 8 characters long'),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: "Passwords don't match.",
    path: ['confirmPassword'],
  })

type FormValues = z.infer<typeof formSchema>

export function ResetPasswordForm({
  className,
  token,
  onSuccess,
  ...props
}: ResetPasswordFormProps) {
  const mutation = useMutation({
    mutationFn: (newPassword: string) => resetPassword(token, newPassword),
  })

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: { newPassword: '', confirmPassword: '' },
  })

  async function onSubmit(data: FormValues) {
    try {
      await mutation.mutateAsync(data.newPassword)
      onSuccess?.()
    } catch {
      // surfaced via mutation.error below
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-3', className)}
        {...props}
      >
        <FormField
          control={form.control}
          name="newPassword"
          render={({ field }) => (
            <FormItem>
              <FormLabel>New Password</FormLabel>
              <FormControl>
                <PasswordInput
                  placeholder="********"
                  autoComplete="new-password"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="confirmPassword"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Confirm Password</FormLabel>
              <FormControl>
                <PasswordInput
                  placeholder="********"
                  autoComplete="new-password"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button className="mt-2" type="submit" disabled={mutation.isPending}>
          {mutation.isPending ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Resetting...
            </>
          ) : (
            'Reset Password'
          )}
        </Button>

        {mutation.error instanceof BrokleError && (
          <div className="mt-2 text-center text-sm text-red-600">
            {mutation.error.body.error.message ||
              'Failed to reset password. The link may have expired.'}
          </div>
        )}
      </form>
    </Form>
  )
}
