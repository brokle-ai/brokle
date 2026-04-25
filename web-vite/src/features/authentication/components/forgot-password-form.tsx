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
import { Input } from '@/components/ui/input'
import { forgotPassword } from '@/features/authentication/api/auth-api'
import { BrokleError } from '@/lib/api/errors'
import { cn } from '@/lib/utils'

type ForgotPasswordFormProps = HTMLAttributes<HTMLFormElement> & {
  onSuccess?: (email: string) => void
}

const formSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
})

type FormValues = z.infer<typeof formSchema>

export function ForgotPasswordForm({
  className,
  onSuccess,
  ...props
}: ForgotPasswordFormProps) {
  const mutation = useMutation({
    mutationFn: (email: string) => forgotPassword(email),
  })

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: { email: '' },
  })

  async function onSubmit(data: FormValues) {
    try {
      await mutation.mutateAsync(data.email)
    } catch {
      // Anti-enumeration: fall through to the success UI regardless so
      // attackers can't learn whether an email exists. The actual
      // BrokleError still surfaces below for genuine network failures.
    }
    onSuccess?.(data.email)
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
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl>
                <Input placeholder="name@example.com" autoComplete="email" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button className="mt-2" type="submit" disabled={mutation.isPending}>
          {mutation.isPending ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Sending...
            </>
          ) : (
            'Continue'
          )}
        </Button>

        {mutation.error instanceof BrokleError && (
          <div className="mt-2 text-center text-sm text-red-600">
            {mutation.error.body.error.message || 'Failed to send reset email. Please try again.'}
          </div>
        )}
      </form>
    </Form>
  )
}
