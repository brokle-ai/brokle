'use client'

import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Users, Plus, Trash2, Loader2, UserPlus, Check, ChevronsUpDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Label } from '@/components/ui/label'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useWorkspace } from '@/context/workspace-context'
import { getOrganizationMembers } from '@/features/organizations/api/members-api'
import {
  useQueueAssignmentsQuery,
  useAssignUserMutation,
  useUnassignUserMutation,
} from '../hooks/use-annotation-queues'
import { cn } from '@/lib/utils'
import type { AssignmentRole, QueueAssignment } from '../types'

interface AssignmentDialogProps {
  projectId: string
  queueId: string
  queueName: string
  trigger?: React.ReactNode
}

function getRoleBadgeVariant(role: AssignmentRole): 'default' | 'secondary' | 'outline' {
  switch (role) {
    case 'admin':
      return 'default'
    case 'reviewer':
      return 'secondary'
    case 'annotator':
      return 'outline'
  }
}

export function AssignmentDialog({
  projectId,
  queueId,
  queueName,
  trigger,
}: AssignmentDialogProps) {
  const [open, setOpen] = useState(false)
  const [userSelectOpen, setUserSelectOpen] = useState(false)
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null)
  const [newRole, setNewRole] = useState<AssignmentRole>('annotator')

  const { currentOrganization } = useWorkspace()

  // Fetch organization members
  const { data: membersResponse, isLoading: isMembersLoading } = useQuery({
    queryKey: ['organization-members', currentOrganization?.id],
    queryFn: async () => {
      if (!currentOrganization?.id) throw new Error('No organization selected')
      return getOrganizationMembers(currentOrganization.id)
    },
    enabled: !!currentOrganization?.id && open,
    staleTime: 2 * 60 * 1000,
  })

  const members = membersResponse?.data ?? []

  const { data: assignments, isLoading } = useQueueAssignmentsQuery(projectId, queueId)
  const assignMutation = useAssignUserMutation(projectId, queueId)
  const unassignMutation = useUnassignUserMutation(projectId, queueId)

  // Filter out already assigned members
  const availableMembers = useMemo(() => {
    const assignedUserIds = new Set(assignments?.map((a) => a.user_id) ?? [])
    return members.filter((m) => !assignedUserIds.has(m.userId))
  }, [members, assignments])

  const selectedMember = members.find((m) => m.userId === selectedUserId)

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((word) => word[0])
      .join('')
      .substring(0, 2)
      .toUpperCase()
  }

  const handleAssign = async () => {
    if (!selectedUserId) return

    await assignMutation.mutateAsync({
      user_id: selectedUserId,
      role: newRole,
    })
    setSelectedUserId(null)
    setNewRole('annotator')
  }

  const handleUnassign = async (userId: string) => {
    await unassignMutation.mutateAsync(userId)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {trigger || (
          <Button variant="outline" size="sm">
            <Users className="mr-2 h-4 w-4" />
            Assignments
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Manage Assignments</DialogTitle>
          <DialogDescription>
            Assign users to &quot;{queueName}&quot; queue. Assigned users can annotate items based on their role.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Add New Assignment */}
          <div className="space-y-4 rounded-lg border p-4">
            <div className="flex items-center gap-2 text-sm font-medium">
              <UserPlus className="h-4 w-4" />
              Add Assignment
            </div>
            <div className="grid gap-4 sm:grid-cols-[1fr_150px_auto]">
              <div className="space-y-2">
                <Label>User</Label>
                {isMembersLoading ? (
                  <div className="flex items-center gap-2 h-10 px-3 py-2 text-sm text-muted-foreground border rounded-md">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Loading members...
                  </div>
                ) : availableMembers.length === 0 ? (
                  <div className="h-10 px-3 py-2 text-sm text-muted-foreground border rounded-md flex items-center">
                    {members.length === 0 ? 'No members found' : 'All members assigned'}
                  </div>
                ) : (
                  <Popover open={userSelectOpen} onOpenChange={setUserSelectOpen}>
                    <PopoverTrigger asChild>
                      <Button
                        variant="outline"
                        role="combobox"
                        aria-expanded={userSelectOpen}
                        className="w-full justify-between"
                      >
                        {selectedMember ? (
                          <div className="flex items-center gap-2">
                            <Avatar className="h-5 w-5">
                              <AvatarFallback className="text-xs">
                                {getInitials(selectedMember.name)}
                              </AvatarFallback>
                            </Avatar>
                            <span className="truncate">{selectedMember.name}</span>
                          </div>
                        ) : (
                          <span className="text-muted-foreground">Select a user...</span>
                        )}
                        <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-[300px] p-0" align="start">
                      <Command>
                        <CommandInput placeholder="Search users..." />
                        <CommandList>
                          <CommandEmpty>No users found.</CommandEmpty>
                          <CommandGroup>
                            {availableMembers.map((member) => (
                              <CommandItem
                                key={member.userId}
                                value={member.name}
                                onSelect={() => {
                                  setSelectedUserId(member.userId)
                                  setUserSelectOpen(false)
                                }}
                              >
                                <Check
                                  className={cn(
                                    'mr-2 h-4 w-4',
                                    selectedUserId === member.userId ? 'opacity-100' : 'opacity-0'
                                  )}
                                />
                                <Avatar className="h-6 w-6 mr-2">
                                  <AvatarFallback className="text-xs">
                                    {getInitials(member.name)}
                                  </AvatarFallback>
                                </Avatar>
                                <div className="flex flex-col">
                                  <span>{member.name}</span>
                                  <span className="text-xs text-muted-foreground">
                                    {member.email}
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
                  value={newRole}
                  onValueChange={(value: AssignmentRole) => setNewRole(value)}
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
                  onClick={handleAssign}
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

          {/* Current Assignments */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h4 className="text-sm font-medium">Current Assignments</h4>
              <span className="text-sm text-muted-foreground">
                {assignments?.length ?? 0} assigned
              </span>
            </div>

            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : !assignments || assignments.length === 0 ? (
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
                    {assignments.map((assignment: QueueAssignment) => {
                      const member = members.find((m) => m.userId === assignment.user_id)
                      return (
                        <TableRow key={assignment.id}>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <Avatar className="h-7 w-7">
                                <AvatarFallback className="text-xs">
                                  {member ? getInitials(member.name) : '?'}
                                </AvatarFallback>
                              </Avatar>
                              <div>
                                <div className="font-medium">
                                  {member?.name ?? 'Unknown User'}
                                </div>
                                <div className="text-xs text-muted-foreground">
                                  {member?.email ?? assignment.user_id.substring(0, 16) + '...'}
                                </div>
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <Badge
                              variant={getRoleBadgeVariant(assignment.role)}
                              className="capitalize"
                            >
                              {assignment.role}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleUnassign(assignment.user_id)}
                              disabled={unassignMutation.isPending}
                            >
                              <Trash2 className="h-4 w-4 text-destructive" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      )
                    })}
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
