'use client'

import { useState } from 'react'
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
  Filter
} from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { useAuth } from '@/hooks/auth/use-auth'
import { useHasAccess } from '@/hooks/rbac/use-has-access'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
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
import type { OrganizationRole, OrganizationMember } from '@/types/organization'

interface MemberManagementProps {
  className?: string
}

export function MemberManagement({ className }: MemberManagementProps) {
  const { user } = useAuth()
  const { currentOrganization } = useWorkspace()

  const [searchTerm, setSearchTerm] = useState('')
  const [roleFilter, setRoleFilter] = useState<OrganizationRole | 'all'>('all')

  // Scope-based permission checks
  const canInviteMembers = useHasAccess({ scope: "members:invite" })
  const canUpdateMembers = useHasAccess({ scope: "members:update" })
  const canRemoveMembers = useHasAccess({ scope: "members:remove" })

  if (!currentOrganization || !user) {
    return null
  }

  // TODO: Fetch members from API when members page loads
  // WorkspaceProvider only returns org + projects, not members
  const members = [] // Placeholder until members API is integrated

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
    // TODO: Implement role change API call
    toast.success(`Member role updated to ${newRole}`)
  }

  const handleRemoveMember = async (memberId: string, memberName: string) => {
    // TODO: Implement remove member API call
    toast.success(`${memberName} has been removed from the organization`)
  }

  const canEditMember = (member: OrganizationMember) => {
    // Need either update or remove permissions to show actions menu
    if (!canUpdateMembers && !canRemoveMembers) return false
    if (member.email === user.email) return false // Can't edit yourself
    if (member.role === 'owner') return false // Can't edit owner
    return true
  }

  return (
    <div className={className}>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5" />
                Team Members
              </CardTitle>
              <CardDescription>
                Manage who has access to {currentOrganization.name} and their permissions
              </CardDescription>
            </div>
            
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
        </CardHeader>
        
        <CardContent className="space-y-4">
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
                      <div className="flex items-center gap-3">
                        <Avatar className="h-8 w-8">
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

          {/* Role Descriptions */}
          <div className="border-t pt-4">
            <h4 className="font-medium mb-3">Role Permissions</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Crown className="h-4 w-4 text-yellow-500" />
                  <span className="font-medium">Owner</span>
                </div>
                <p className="text-muted-foreground text-xs pl-6">
                  Full control over the organization, billing, and all projects
                </p>
                
                <div className="flex items-center gap-2">
                  <Shield className="h-4 w-4 text-blue-500" />
                  <span className="font-medium">Admin</span>
                </div>
                <p className="text-muted-foreground text-xs pl-6">
                  Manage members, projects, and organization settings
                </p>
              </div>
              
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-green-500" />
                  <span className="font-medium">Developer</span>
                </div>
                <p className="text-muted-foreground text-xs pl-6">
                  Create and manage projects, view analytics and costs
                </p>
                
                <div className="flex items-center gap-2">
                  <Eye className="h-4 w-4 text-gray-500" />
                  <span className="font-medium">Viewer</span>
                </div>
                <p className="text-muted-foreground text-xs pl-6">
                  Read-only access to projects and basic analytics
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}