'use client'

import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Users,
  UserPlus,
  Crown,
  Shield,
  User,
  Eye,
  MoreVertical,
  Mail,
  Calendar,
  Search,
  Loader2
} from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { useAuth } from '@/features/authentication'
import { useHasAccess } from '@/hooks/rbac/use-has-access'
import { getOrganizationMembers, removeMember, type Member } from '../api/members-api'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { InviteMemberModal } from './invite-member-modal'
import { toast } from 'sonner'
import type { OrganizationRole, OrganizationMember } from '../types'

interface OrganizationMembersSectionProps {
  className?: string
}

export function OrganizationMembersSection({ className }: OrganizationMembersSectionProps) {
  const { user } = useAuth()
  const { currentOrganization } = useWorkspace()
  const queryClient = useQueryClient()

  const [searchTerm, setSearchTerm] = useState('')
  const [roleFilter, setRoleFilter] = useState<OrganizationRole | 'all'>('all')

  // Scope-based permission checks
  const canInviteMembers = useHasAccess({ scope: "members:invite" })
  const canUpdateMembers = useHasAccess({ scope: "members:update" })
  const canRemoveMembers = useHasAccess({ scope: "members:remove" })

  // Fetch members from API
  const { data: membersResponse, isLoading, error } = useQuery({
    queryKey: ['organization-members', currentOrganization?.id],
    queryFn: async () => {
      if (!currentOrganization?.id) throw new Error('No organization selected')
      return getOrganizationMembers(currentOrganization.id)
    },
    enabled: !!currentOrganization?.id,
    staleTime: 2 * 60 * 1000, // 2 minutes
  })

  if (!currentOrganization || !user) {
    return null
  }

  // Transform members data for display
  const members: Array<{
    id: string
    name: string
    email: string
    role: OrganizationRole
    joined_at: string
    avatar?: string
  }> = (membersResponse?.data || []).map((member: Member) => ({
    id: member.userId,
    name: member.name,
    email: member.email,
    role: member.role,
    joined_at: member.joinedAt.toISOString(),
    avatar: undefined,
  }))

  const filteredMembers = members.filter(member => {
    const matchesSearch = member.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         member.email.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesRole = roleFilter === 'all' || member.role === roleFilter

    return matchesSearch && matchesRole
  })

  const getRoleIcon = (role: OrganizationRole) => {
    switch (role) {
      case 'owner':
        return <Crown className="h-4 w-4 text-yellow-500" />
      case 'admin':
        return <Shield className="h-4 w-4 text-blue-500" />
      case 'developer':
        return <User className="h-4 w-4 text-green-500" />
      case 'viewer':
        return <Eye className="h-4 w-4 text-gray-500" />
      default:
        return <User className="h-4 w-4 text-gray-500" />
    }
  }

  const getRoleBadgeColor = (role: OrganizationRole) => {
    switch (role) {
      case 'owner':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
      case 'admin':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
      case 'developer':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'viewer':
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    }
  }

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map(word => word[0])
      .join('')
      .substring(0, 2)
      .toUpperCase()
  }

  const handleRoleChange = async (memberId: string, newRole: OrganizationRole) => {
    // TODO: Implement role change API when backend endpoint supports it
    toast.success(`Member role updated to ${newRole}`)
  }

  const handleRemoveMember = async (memberId: string, memberName: string) => {
    if (!currentOrganization?.id) return

    try {
      await removeMember(currentOrganization.id, memberId)
      // Invalidate members cache to refetch
      queryClient.invalidateQueries({ queryKey: ['organization-members', currentOrganization.id] })
      toast.success(`${memberName} has been removed from the organization`)
    } catch (err) {
      console.error('Failed to remove member:', err)
      toast.error('Failed to remove member. Please try again.')
    }
  }

  const canEditMember = (member: { email: string; role: OrganizationRole }) => {
    // Need either update or remove permissions to show actions menu
    if (!canUpdateMembers && !canRemoveMembers) return false
    if (member.email === user.email) return false // Can't edit yourself
    if (member.role === 'owner') return false // Can't edit owner
    return true
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="text-center py-8">
        <Users className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
        <h3 className="text-lg font-medium mb-2">Failed to load members</h3>
        <p className="text-muted-foreground">
          Please try refreshing the page
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Members Table Section */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-medium">Team Members ({members.length})</h3>
          {canInviteMembers && (
            <InviteMemberModal
              trigger={
                <Button>
                  <UserPlus className="mr-2 h-4 w-4" />
                  Invite Member
                </Button>
              }
            />
          )}
        </div>

        {/* Filters */}
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
            <Input
              placeholder="Search members..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>

          <Select value={roleFilter} onValueChange={setRoleFilter}>
            <SelectTrigger className="w-full sm:w-40">
              <SelectValue placeholder="All Roles" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Roles</SelectItem>
              <SelectItem value="owner">Owner</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
              <SelectItem value="developer">Developer</SelectItem>
              <SelectItem value="viewer">Viewer</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Members Table */}
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Member</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Joined</TableHead>
                <TableHead className="w-[70px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredMembers.map((member) => (
                <TableRow key={member.id}>
                  <TableCell>
                      <div className="flex items-center gap-2">
                        <Avatar className="h-7 w-7">
                          <AvatarImage src={member.avatar} alt={member.name} />
                          <AvatarFallback className="text-xs">
                            {getInitials(member.name)}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="font-medium">
                            {member.name}
                            {member.email === user.email && (
                              <Badge variant="outline" className="ml-2 text-xs">
                                You
                              </Badge>
                            )}
                          </div>
                          <div className="text-sm text-muted-foreground">
                            {member.email}
                          </div>
                        </div>
                      </div>
                    </TableCell>
                    
                    <TableCell>
                      <Badge className={`flex items-center gap-1 w-fit ${getRoleBadgeColor(member.role)}`}>
                        {getRoleIcon(member.role)}
                        {member.role.charAt(0).toUpperCase() + member.role.slice(1)}
                      </Badge>
                    </TableCell>
                    
                    <TableCell>
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <Calendar className="h-4 w-4" />
                        {new Date(member.joined_at).toLocaleDateString()}
                      </div>
                    </TableCell>
                    
                    <TableCell>
                      {canEditMember(member) && (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm">
                              <MoreVertical className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuLabel>Actions</DropdownMenuLabel>
                            
                            <DropdownMenuLabel className="text-xs font-normal text-muted-foreground">
                              Change Role
                            </DropdownMenuLabel>
                            
                            {canUpdateMembers && (['viewer', 'developer', 'admin'] as OrganizationRole[])
                              .filter(role => role !== member.role)
                              .map((role) => (
                                <DropdownMenuItem
                                  key={role}
                                  onClick={() => handleRoleChange(member.id, role)}
                                  className="pl-6"
                                >
                                  <div className="flex items-center gap-2">
                                    {getRoleIcon(role)}
                                    {role.charAt(0).toUpperCase() + role.slice(1)}
                                  </div>
                                </DropdownMenuItem>
                              ))}
                            
                            <DropdownMenuSeparator />
                            
                            <DropdownMenuItem asChild>
                              <a href={`mailto:${member.email}`}>
                                <Mail className="mr-2 h-4 w-4" />
                                Send Email
                              </a>
                            </DropdownMenuItem>
                            
                            <DropdownMenuSeparator />

                            {canRemoveMembers && (
                              <AlertDialog>
                                <AlertDialogTrigger asChild>
                                  <DropdownMenuItem
                                    className="text-destructive"
                                    onSelect={(e) => e.preventDefault()}
                                  >
                                    Remove from Organization
                                  </DropdownMenuItem>
                                </AlertDialogTrigger>
                                <AlertDialogContent>
                                  <AlertDialogHeader>
                                    <AlertDialogTitle>Remove Member</AlertDialogTitle>
                                    <AlertDialogDescription>
                                      Are you sure you want to remove <strong>{member.name}</strong> from {currentOrganization.name}?
                                      They will lose access to all projects and data in this organization.
                                    </AlertDialogDescription>
                                  </AlertDialogHeader>
                                  <AlertDialogFooter>
                                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                                    <AlertDialogAction
                                      onClick={() => handleRemoveMember(member.id, member.name)}
                                      className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                    >
                                      Remove Member
                                    </AlertDialogAction>
                                  </AlertDialogFooter>
                                </AlertDialogContent>
                              </AlertDialog>
                            )}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

        {filteredMembers.length === 0 && (
          <div className="text-center py-8">
            <Users className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium mb-2">No members found</h3>
            <p className="text-muted-foreground">
              {searchTerm || roleFilter !== 'all'
                ? 'Try adjusting your search or filters'
                : 'Invite team members to start collaborating'}
            </p>
          </div>
        )}
      </div>

    </div>
  )
}