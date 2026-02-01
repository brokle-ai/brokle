'use client'

import {
  Crown,
  Shield,
  User,
  Eye,
  MoreVertical,
  Mail,
  Calendar,
} from 'lucide-react'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
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
import type { OrganizationRole } from '../types'

export interface MemberRow {
  id: string
  name: string
  email: string
  role: OrganizationRole
  joined_at: string
  avatar?: string
}

interface CreateMembersColumnsOptions {
  currentUserEmail: string
  organizationName: string
  canUpdateMembers: boolean
  canRemoveMembers: boolean
  onRoleChange: (memberId: string, newRole: OrganizationRole) => void
  onRemoveMember: (memberId: string, memberName: string) => void
}

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

export function createMembersColumns({
  currentUserEmail,
  organizationName,
  canUpdateMembers,
  canRemoveMembers,
  onRoleChange,
  onRemoveMember,
}: CreateMembersColumnsOptions): ColumnDef<MemberRow>[] {
  const canEditMember = (member: MemberRow) => {
    if (!canUpdateMembers && !canRemoveMembers) return false
    if (member.email === currentUserEmail) return false
    if (member.role === 'owner') return false
    return true
  }

  return [
    {
      accessorKey: 'name',
      header: 'Member',
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <Avatar className="h-7 w-7">
            <AvatarImage src={row.original.avatar} alt={row.original.name} />
            <AvatarFallback className="text-xs">
              {getInitials(row.original.name)}
            </AvatarFallback>
          </Avatar>
          <div>
            <div className="font-medium">
              {row.original.name}
              {row.original.email === currentUserEmail && (
                <Badge variant="outline" className="ml-2 text-xs">
                  You
                </Badge>
              )}
            </div>
            <div className="text-sm text-muted-foreground">
              {row.original.email}
            </div>
          </div>
        </div>
      ),
    },
    {
      accessorKey: 'role',
      header: 'Role',
      cell: ({ row }) => (
        <Badge className={`flex items-center gap-1 w-fit ${getRoleBadgeColor(row.original.role)}`}>
          {getRoleIcon(row.original.role)}
          {row.original.role.charAt(0).toUpperCase() + row.original.role.slice(1)}
        </Badge>
      ),
    },
    {
      accessorKey: 'joined_at',
      header: 'Joined',
      cell: ({ row }) => (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Calendar className="h-4 w-4" />
          {new Date(row.original.joined_at).toLocaleDateString()}
        </div>
      ),
    },
    {
      id: 'actions',
      header: () => <span className="sr-only">Actions</span>,
      cell: ({ row }) => {
        const member = row.original

        if (!canEditMember(member)) {
          return null
        }

        return (
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
                    onClick={() => onRoleChange(member.id, role)}
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
                        Are you sure you want to remove <strong>{member.name}</strong> from {organizationName}?
                        They will lose access to all projects and data in this organization.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Cancel</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={() => onRemoveMember(member.id, member.name)}
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
        )
      },
    },
  ]
}
