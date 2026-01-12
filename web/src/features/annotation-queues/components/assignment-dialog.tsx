'use client'

import { useState } from 'react'
import { Users, Plus, Trash2, Loader2, UserPlus } from 'lucide-react'
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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  useQueueAssignmentsQuery,
  useAssignUserMutation,
  useUnassignUserMutation,
} from '../hooks/use-annotation-queues'
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
  const [newUserId, setNewUserId] = useState('')
  const [newRole, setNewRole] = useState<AssignmentRole>('annotator')

  const { data: assignments, isLoading } = useQueueAssignmentsQuery(projectId, queueId)
  const assignMutation = useAssignUserMutation(projectId, queueId)
  const unassignMutation = useUnassignUserMutation(projectId, queueId)

  const handleAssign = async () => {
    if (!newUserId.trim()) return

    await assignMutation.mutateAsync({
      user_id: newUserId.trim(),
      role: newRole,
    })
    setNewUserId('')
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
                <Label htmlFor="user-id">User ID</Label>
                <Input
                  id="user-id"
                  placeholder="Enter user ID"
                  value={newUserId}
                  onChange={(e) => setNewUserId(e.target.value)}
                />
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
                  disabled={!newUserId.trim() || assignMutation.isPending}
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
                      <TableHead>User ID</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead className="w-[100px]">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {assignments.map((assignment: QueueAssignment) => (
                      <TableRow key={assignment.id}>
                        <TableCell className="font-mono text-sm">
                          {assignment.user_id.substring(0, 16)}...
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
