import { useMemo, useState, type ReactNode } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  Check,
  ChevronsUpDown,
  Loader2,
  Plus,
  Trash2,
  UserPlus,
  Users,
} from 'lucide-react'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'
import { memberListQueryOptions } from '@/features/members/api/queries'
import {
  assignQueueUser,
  queueAssignmentsQueryOptions,
  queuesKeys,
  unassignQueueUser,
} from '../api/queries'
import type { AssignmentRole } from '../api/types'

interface AssignmentDialogProps {
  orgId: string
  projectId: string
  queueId: string
  queueName: string
  trigger?: ReactNode
}

function roleBadgeVariant(
  role: AssignmentRole,
): 'default' | 'secondary' | 'outline' {
  switch (role) {
    case 'admin':
      return 'default'
    case 'reviewer':
      return 'secondary'
    case 'annotator':
      return 'outline'
  }
}

function shortId(id: string): string {
  return id.length > 12 ? `${id.slice(0, 8)}…${id.slice(-4)}` : id
}

/**
 * Manage queue assignments dialog. Displays current assignees and
 * exposes an add-assignment row backed by the org members list.
 *
 * Note on display: the org members endpoint currently only returns
 * `user_id` + role/status/joined_at — no name/email. We render a
 * deterministic short-id label until the backend hydrates user
 * profile fields onto the membership row.
 */
export function AssignmentDialog({
  orgId,
  projectId,
  queueId,
  queueName,
  trigger,
}: AssignmentDialogProps) {
  const [open, setOpen] = useState(false)
  const [pickerOpen, setPickerOpen] = useState(false)
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null)
  const [role, setRole] = useState<AssignmentRole>('annotator')
  const queryClient = useQueryClient()

  const membersQuery = useQuery({
    ...memberListQueryOptions(orgId, { page: 1, limit: 200 }),
    enabled: open,
  })
  const assignmentsQuery = useQuery({
    ...queueAssignmentsQueryOptions(projectId, queueId),
    enabled: open,
  })

  const assignments = assignmentsQuery.data ?? []
  const members = membersQuery.data?.data ?? []

  const availableMembers = useMemo(() => {
    const taken = new Set(assignments.map((a) => a.user_id))
    return members.filter((m) => !taken.has(m.user_id))
  }, [members, assignments])

  const selectedMember = useMemo(
    () => members.find((m) => m.user_id === selectedUserId),
    [members, selectedUserId],
  )

  const invalidateAssignments = () =>
    queryClient.invalidateQueries({
      queryKey: queuesKeys.assignments(queueId),
    })

  const assignMutation = useMutation({
    mutationFn: async () => {
      if (!selectedUserId) throw new Error('No user selected')
      return assignQueueUser(projectId, queueId, {
        user_id: selectedUserId,
        role,
      })
    },
    onSuccess: () => {
      invalidateAssignments()
      setSelectedUserId(null)
      setRole('annotator')
      toast.success('Assignment added')
    },
    onError: (err) => {
      toast.error('Failed to assign user', {
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  const unassignMutation = useMutation({
    mutationFn: (userId: string) => unassignQueueUser(projectId, queueId, userId),
    onSuccess: () => {
      invalidateAssignments()
      toast.success('Assignment removed')
    },
    onError: (err) => {
      toast.error('Failed to remove assignment', {
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    },
  })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {trigger ?? (
          <Button variant="outline" size="sm">
            <Users className="mr-2 h-4 w-4" />
            Assignments
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px] max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Manage Assignments</DialogTitle>
          <DialogDescription>
            Assign users to &ldquo;{queueName}&rdquo;. Assigned users can
            annotate items based on their role.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          <div className="space-y-4 rounded-lg border p-4">
            <div className="flex items-center gap-2 text-sm font-medium">
              <UserPlus className="h-4 w-4" />
              Add Assignment
            </div>
            <div className="grid gap-4 sm:grid-cols-[1fr_150px_auto]">
              <div className="space-y-2">
                <Label>User</Label>
                {membersQuery.isLoading ? (
                  <div className="flex h-10 items-center gap-2 rounded-md border px-3 py-2 text-sm text-muted-foreground">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Loading members...
                  </div>
                ) : availableMembers.length === 0 ? (
                  <div className="flex h-10 items-center rounded-md border px-3 py-2 text-sm text-muted-foreground">
                    {members.length === 0
                      ? 'No members found'
                      : 'All members assigned'}
                  </div>
                ) : (
                  <Popover open={pickerOpen} onOpenChange={setPickerOpen}>
                    <PopoverTrigger asChild>
                      <Button
                        variant="outline"
                        role="combobox"
                        aria-expanded={pickerOpen}
                        className="w-full justify-between"
                      >
                        {selectedMember ? (
                          <div className="flex items-center gap-2">
                            <Avatar className="h-5 w-5">
                              <AvatarFallback className="text-xs">
                                {selectedMember.user_id
                                  .slice(0, 2)
                                  .toUpperCase()}
                              </AvatarFallback>
                            </Avatar>
                            <span className="font-mono text-xs">
                              {shortId(selectedMember.user_id)}
                            </span>
                          </div>
                        ) : (
                          <span className="text-muted-foreground">
                            Select a user...
                          </span>
                        )}
                        <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-[300px] p-0" align="start">
                      <Command>
                        <CommandInput placeholder="Search by user ID..." />
                        <CommandList>
                          <CommandEmpty>No users found.</CommandEmpty>
                          <CommandGroup>
                            {availableMembers.map((m) => (
                              <CommandItem
                                key={m.user_id}
                                value={m.user_id}
                                onSelect={() => {
                                  setSelectedUserId(m.user_id)
                                  setPickerOpen(false)
                                }}
                              >
                                <Check
                                  className={cn(
                                    'mr-2 h-4 w-4',
                                    selectedUserId === m.user_id
                                      ? 'opacity-100'
                                      : 'opacity-0',
                                  )}
                                />
                                <Avatar className="h-6 w-6 mr-2">
                                  <AvatarFallback className="text-xs">
                                    {m.user_id.slice(0, 2).toUpperCase()}
                                  </AvatarFallback>
                                </Avatar>
                                <div className="flex flex-col">
                                  <span className="font-mono text-xs">
                                    {shortId(m.user_id)}
                                  </span>
                                  <span className="text-[11px] text-muted-foreground">
                                    {m.status}
                                  </span>
                                </div>
                              </CommandItem>
                            ))}
                          </CommandGroup>
                        </CommandList>
                      </Command>
                    </PopoverContent>
                  </Popover>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="role">Role</Label>
                <Select
                  value={role}
                  onValueChange={(v) => setRole(v as AssignmentRole)}
                >
                  <SelectTrigger id="role">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="annotator">Annotator</SelectItem>
                    <SelectItem value="reviewer">Reviewer</SelectItem>
                    <SelectItem value="admin">Admin</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-end">
                <Button
                  onClick={() => assignMutation.mutate()}
                  disabled={!selectedUserId || assignMutation.isPending}
                >
                  {assignMutation.isPending ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Plus className="mr-2 h-4 w-4" />
                  )}
                  Assign
                </Button>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h4 className="text-sm font-medium">Current Assignments</h4>
              <span className="text-sm text-muted-foreground">
                {assignments.length} assigned
              </span>
            </div>

            {assignmentsQuery.isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : assignments.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <Users className="h-10 w-10 text-muted-foreground/50 mb-2" />
                <p className="text-sm text-muted-foreground">
                  No users assigned to this queue yet.
                </p>
              </div>
            ) : (
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>User</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead className="w-[100px]">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {assignments.map((a) => (
                      <TableRow key={a.id}>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Avatar className="h-7 w-7">
                              <AvatarFallback className="text-xs">
                                {a.user_id.slice(0, 2).toUpperCase()}
                              </AvatarFallback>
                            </Avatar>
                            <div>
                              <div className="font-mono text-xs">
                                {shortId(a.user_id)}
                              </div>
                              <div className="text-[11px] text-muted-foreground">
                                Assigned{' '}
                                {new Date(a.assigned_at).toLocaleDateString()}
                              </div>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant={roleBadgeVariant(a.role)}
                            className="capitalize"
                          >
                            {a.role}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() =>
                              unassignMutation.mutate(a.user_id)
                            }
                            disabled={unassignMutation.isPending}
                          >
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
