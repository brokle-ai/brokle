import { createFileRoute, useNavigate } from '@tanstack/react-router'
import {
  useMutation,
  useQueryClient,
  useSuspenseQuery,
} from '@tanstack/react-query'
import { useState } from 'react'
import { Copy, Loader2, Save, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { BrokleError } from '@/lib/api/errors'
import {
  deleteProject,
  projectKeys,
  projectMembershipQueryOptions,
  updateProject,
  type UpdateProjectRequest,
} from '@/features/projects/queries'

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/settings/',
)({
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      projectMembershipQueryOptions(params.projectId),
    ),
  component: ProjectSettingsIndex,
})

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function errorMessage(err: unknown, fallback: string): string {
  if (err instanceof BrokleError) return err.message
  if (err instanceof Error) return err.message
  return fallback
}

function ProjectSettingsIndex() {
  const { orgId, projectId } = Route.useParams()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const { data: project } = useSuspenseQuery(
    projectMembershipQueryOptions(projectId),
  )

  // Local edit state. Only flip `dirty` once the user types so the
  // Save button stays disabled on an unchanged form.
  const [name, setName] = useState(project.name)
  const [description, setDescription] = useState(project.description ?? '')
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState('')

  const dirty =
    name.trim() !== project.name ||
    (description ?? '').trim() !== (project.description ?? '').trim()

  const updateMutation = useMutation({
    mutationFn: async (data: UpdateProjectRequest) =>
      updateProject(projectId, data),
    onSuccess: async (updated) => {
      await queryClient.invalidateQueries({
        queryKey: projectKeys.detail(projectId),
      })
      setName(updated.name)
      setDescription(updated.description ?? '')
      toast.success('Project updated')
    },
    onError: (err) => {
      toast.error('Failed to update project', {
        description: errorMessage(err, 'Could not update project.'),
      })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteProject(projectId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: projectKeys.lists() })
      await queryClient.invalidateQueries({
        queryKey: projectKeys.detail(projectId),
      })
      toast.success('Project deleted')
      navigate({ to: '/o/$orgId', params: { orgId } })
    },
    onError: (err) => {
      toast.error('Failed to delete project', {
        description: errorMessage(err, 'Could not delete project.'),
      })
    },
  })

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault()
    const trimmedName = name.trim()
    if (trimmedName.length < 2 || trimmedName.length > 100) {
      toast.error('Project name must be 2–100 characters')
      return
    }
    if (description.length > 500) {
      toast.error('Description must be at most 500 characters')
      return
    }
    const payload: UpdateProjectRequest = {}
    if (trimmedName !== project.name) payload.name = trimmedName
    if (description.trim() !== (project.description ?? '').trim()) {
      payload.description = description.trim()
    }
    if (Object.keys(payload).length === 0) return
    updateMutation.mutate(payload)
  }

  const copyProjectId = () => {
    navigator.clipboard.writeText(project.id).then(
      () => toast.success('Project ID copied'),
      () => toast.error('Could not copy'),
    )
  }

  const canConfirmDelete = deleteConfirm === project.slug

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>General</CardTitle>
          <CardDescription>
            Manage basic project information and configuration.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSave} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="project-name">Name *</Label>
              <Input
                id="project-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Project name"
                minLength={2}
                maxLength={100}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="project-description">Description</Label>
              <Textarea
                id="project-description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Describe what this project is for…"
                rows={3}
                maxLength={500}
              />
              <p className="text-xs text-muted-foreground">
                {description.length}/500
              </p>
            </div>

            <dl className="grid grid-cols-1 gap-4 rounded-md border p-4 sm:grid-cols-2">
              <Field label="Slug" value={project.slug} mono />
              <Field
                label="Organization ID"
                value={project.organization_id}
                mono
              />
              <Field
                label="Project ID"
                value={project.id}
                mono
                action={
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={copyProjectId}
                  >
                    <Copy className="h-3 w-3" />
                  </Button>
                }
              />
              <Field
                label="Created"
                value={formatTimestamp(project.created_at)}
              />
              <Field
                label="Updated"
                value={formatTimestamp(project.updated_at)}
              />
            </dl>

            <Button type="submit" disabled={!dirty || updateMutation.isPending}>
              {updateMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving…
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  Save changes
                </>
              )}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card className="border-destructive/40">
        <CardHeader>
          <CardTitle className="text-destructive">Danger zone</CardTitle>
          <CardDescription>
            Irreversible actions. Deleting the project removes it from the
            organization — traces, evaluators, prompts, and API keys tied to
            it will be archived.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            variant="destructive"
            onClick={() => {
              setDeleteConfirm('')
              setDeleteOpen(true)
            }}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Delete project
          </Button>
        </CardContent>
      </Card>

      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-destructive">
              Delete project
            </DialogTitle>
            <DialogDescription>
              This cannot be undone. The project and its telemetry will be
              archived immediately.
            </DialogDescription>
          </DialogHeader>

          <Alert variant="destructive">
            <AlertDescription>
              Type <code className="font-mono text-xs">{project.slug}</code>{' '}
              below to confirm.
            </AlertDescription>
          </Alert>

          <Input
            value={deleteConfirm}
            onChange={(e) => setDeleteConfirm(e.target.value)}
            placeholder={project.slug}
            autoComplete="off"
          />

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteOpen(false)}
              disabled={deleteMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => deleteMutation.mutate()}
              disabled={!canConfirmDelete || deleteMutation.isPending}
            >
              {deleteMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting…
                </>
              ) : (
                'Delete permanently'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function Field({
  label,
  value,
  mono = false,
  action,
}: {
  label: string
  value: string
  mono?: boolean
  action?: React.ReactNode
}) {
  return (
    <div>
      <dt className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {label}
      </dt>
      <dd
        className={
          mono
            ? 'mt-1 flex items-center gap-1 font-mono text-xs text-foreground'
            : 'mt-1 text-sm text-foreground'
        }
      >
        <span>{value}</span>
        {action}
      </dd>
    </div>
  )
}
