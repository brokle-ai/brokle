import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useMutation, useQueryClient, useSuspenseQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  AlertTriangle,
  Copy,
  Loader2,
  Save,
  Shield,
  Trash2,
  Users,
} from 'lucide-react'
import { toast } from 'sonner'
import { BrokleError } from '@/lib/api/errors'
import {
  deleteOrganization,
  organizationKeys,
  organizationMembershipQueryOptions,
  updateOrganization,
  type UpdateOrganizationRequest,
} from '@/features/organizations/queries'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Label } from '@/components/ui/label'

// Org-level settings landing — currently four sections: General,
// Members (link), Security/SSO (placeholder), Danger Zone. Each
// project also has its own settings route under
// `/o/$orgId/p/$projectId/settings`; the org-level page is what the
// settings cog in the org switcher now points at.
export const Route = createFileRoute('/_authenticated/o/$orgId/settings')({
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      organizationMembershipQueryOptions(params.orgId),
    ),
  errorComponent: ({ error }) => {
    if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
      return (
        <main className="mx-auto max-w-4xl p-6">
          <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
            <p className="text-sm font-medium text-destructive">
              Unable to load settings
            </p>
            <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
          </div>
        </main>
      )
    }
    throw error
  },
  component: OrgSettingsPage,
})

function OrgSettingsPage() {
  const { orgId } = Route.useParams()
  const { data: org } = useSuspenseQuery(organizationMembershipQueryOptions(orgId))

  return (
    <main className="mx-auto w-full max-w-4xl space-y-6 p-6">
      <header>
        <h1 className="text-2xl font-semibold">Organization Settings</h1>
        <p className="text-sm text-muted-foreground">
          Manage details, members, and security for {org.name}.
        </p>
      </header>

      <GeneralSection orgId={orgId} />
      <MembersSection orgId={orgId} />
      <SecuritySection />
      <DangerZoneSection orgId={orgId} orgName={org.name} orgSlug={org.slug} />
    </main>
  )
}

const generalSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters').max(100),
  slug: z
    .string()
    .min(2, 'Slug must be at least 2 characters')
    .max(63)
    .regex(/^[a-z0-9-]+$/, 'Lowercase letters, digits, hyphens only'),
  billing_email: z.string().email('Must be a valid email').or(z.literal('')),
})

type GeneralFormData = z.infer<typeof generalSchema>

function GeneralSection({ orgId }: { orgId: string }) {
  const qc = useQueryClient()
  const { data: org } = useSuspenseQuery(organizationMembershipQueryOptions(orgId))

  const form = useForm<GeneralFormData>({
    resolver: zodResolver(generalSchema),
    defaultValues: {
      name: org.name,
      slug: org.slug,
      billing_email: org.billing_email,
    },
  })

  // Reset the form when the underlying org changes (e.g. after another
  // tab updates it). RHF's `reset` is stable across renders.
  useEffect(() => {
    form.reset({
      name: org.name,
      slug: org.slug,
      billing_email: org.billing_email,
    })
  }, [form, org.billing_email, org.name, org.slug])

  const mutation = useMutation({
    mutationFn: (data: UpdateOrganizationRequest) => updateOrganization(orgId, data),
    onSuccess: (updated) => {
      toast.success('Organization updated')
      qc.setQueryData(organizationKeys.detail(orgId), updated)
      void qc.invalidateQueries({ queryKey: organizationKeys.lists() })
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to update organization')
    },
  })

  const onSubmit = (values: GeneralFormData) => {
    // Send only changed fields to keep the payload minimal and avoid
    // tripping uniqueness validators on unchanged slugs.
    const patch: UpdateOrganizationRequest = {}
    if (values.name !== org.name) patch.name = values.name
    if (values.slug !== org.slug) patch.slug = values.slug
    if (values.billing_email !== org.billing_email)
      patch.billing_email = values.billing_email
    if (Object.keys(patch).length === 0) {
      toast.info('No changes to save')
      return
    }
    mutation.mutate(patch)
  }

  const copyId = () => {
    void navigator.clipboard.writeText(org.id)
    toast.success('Organization ID copied')
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>General</CardTitle>
        <CardDescription>Update your organization profile.</CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="space-y-6"
          >
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Organization Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Acme Corp" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="slug"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Slug</FormLabel>
                  <FormControl>
                    <Input placeholder="acme-corp" {...field} />
                  </FormControl>
                  <FormDescription>
                    Used in URLs and API references.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="billing_email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Billing Email</FormLabel>
                  <FormControl>
                    <Input
                      type="email"
                      placeholder="billing@acme.com"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    Invoices and billing alerts go to this address.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="rounded-lg border p-4">
              <Label className="text-sm font-medium text-muted-foreground">
                Organization ID
              </Label>
              <div className="mt-2 flex items-center gap-2">
                <code className="rounded bg-muted px-2 py-1 font-mono text-xs">
                  {org.id}
                </code>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={copyId}
                >
                  <Copy className="h-3 w-3" />
                </Button>
              </div>
              <p className="mt-2 text-xs text-muted-foreground">
                Use this ID for API integration and support requests.
              </p>
            </div>

            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  Save Changes
                </>
              )}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  )
}

function MembersSection({ orgId }: { orgId: string }) {
  // Member management lives on each project's settings page (orgs
  // don't currently have an org-scoped members surface in the route
  // tree). Send the user there.
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Users className="h-5 w-5" />
          Members
        </CardTitle>
        <CardDescription>
          Invite teammates and manage roles for this organization.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Button asChild variant="outline">
          <Link to="/o/$orgId" params={{ orgId }}>
            Open project to manage members
          </Link>
        </Button>
      </CardContent>
    </Card>
  )
}

function SecuritySection() {
  // SSO/SAML/SCIM endpoints aren't surfaced on the dashboard plane
  // yet; render a placeholder so the section is discoverable but
  // honest about availability.
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          Security & SSO
        </CardTitle>
        <CardDescription>
          Configure SSO, SAML, and SCIM provisioning for your organization.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            Single sign-on configuration is coming soon. Contact support to
            enable enterprise SSO for your organization.
          </AlertDescription>
        </Alert>
      </CardContent>
    </Card>
  )
}

function DangerZoneSection({
  orgId,
  orgName,
  orgSlug,
}: {
  orgId: string
  orgName: string
  orgSlug: string
}) {
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [open, setOpen] = useState(false)
  const [confirmation, setConfirmation] = useState('')

  const mutation = useMutation({
    mutationFn: () => deleteOrganization(orgId),
    onSuccess: () => {
      toast.success('Organization deleted')
      qc.removeQueries({ queryKey: organizationKeys.detail(orgId) })
      void qc.invalidateQueries({ queryKey: organizationKeys.lists() })
      setOpen(false)
      void navigate({ to: '/o' })
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to delete organization')
    },
  })

  const matches = confirmation.trim() === orgSlug

  return (
    <Card className="border-destructive/40">
      <CardHeader>
        <CardTitle className="text-destructive">Danger Zone</CardTitle>
        <CardDescription>
          Permanently delete this organization. This action cannot be undone.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            All projects, API keys, traces, datasets, and members will be
            permanently removed.
          </AlertDescription>
        </Alert>

        <Dialog
          open={open}
          onOpenChange={(v) => {
            if (!v && mutation.isPending) return
            setOpen(v)
            if (!v) setConfirmation('')
          }}
        >
          <Button
            variant="destructive"
            onClick={() => setOpen(true)}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Delete {orgName}
          </Button>

          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle className="text-destructive">
                Delete Organization
              </DialogTitle>
              <DialogDescription>
                This permanently deletes <strong>{orgName}</strong> and all
                associated data.
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-3">
              <Label htmlFor="confirm-slug">
                Type the organization slug{' '}
                <code className="rounded bg-muted px-1 py-0.5 font-mono text-xs">
                  {orgSlug}
                </code>{' '}
                to confirm
              </Label>
              <Input
                id="confirm-slug"
                value={confirmation}
                onChange={(e) => setConfirmation(e.target.value)}
                placeholder={orgSlug}
                autoComplete="off"
              />
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setOpen(false)}
                disabled={mutation.isPending}
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={() => mutation.mutate()}
                disabled={!matches || mutation.isPending}
              >
                {mutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Deleting...
                  </>
                ) : (
                  'Delete Forever'
                )}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>
  )
}
