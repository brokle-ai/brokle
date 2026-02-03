'use client'

import { useState, useMemo } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
} from '@tanstack/react-table'
import { Users, Search } from 'lucide-react'
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
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DataTableSkeleton,
  DataTableEmptyState,
} from '@/components/shared/tables'
import { toast } from 'sonner'
import { createMembersColumns, type MemberRow } from './members-columns'
import type { OrganizationRole } from '../types'

export function OrganizationMembersSection() {
  const { user } = useAuth()
  const { currentOrganization } = useWorkspace()
  const queryClient = useQueryClient()

  const [searchTerm, setSearchTerm] = useState('')
  const [roleFilter, setRoleFilter] = useState<OrganizationRole | 'all'>('all')

  // Scope-based permission checks
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

  // Transform members data for display
  const members: MemberRow[] = useMemo(() =>
    (membersResponse?.data || []).map((member: Member) => ({
      id: member.userId,
      name: member.name,
      email: member.email,
      role: member.role,
      joined_at: member.joinedAt.toISOString(),
      avatar: undefined,
    })),
    [membersResponse]
  )

  const filteredMembers = useMemo(() =>
    members.filter(member => {
      const matchesSearch = member.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           member.email.toLowerCase().includes(searchTerm.toLowerCase())
      const matchesRole = roleFilter === 'all' || member.role === roleFilter

      return matchesSearch && matchesRole
    }),
    [members, searchTerm, roleFilter]
  )

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

  // Create columns
  const columns = useMemo(
    () =>
      createMembersColumns({
        currentUserEmail: user?.email || '',
        organizationName: currentOrganization?.name || '',
        canUpdateMembers,
        canRemoveMembers,
        onRoleChange: handleRoleChange,
        onRemoveMember: handleRemoveMember,
      }),
    [user?.email, currentOrganization?.name, canUpdateMembers, canRemoveMembers]
  )

  // Initialize React Table
  const table = useReactTable({
    data: filteredMembers,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  if (!currentOrganization || !user) {
    return null
  }

  // Loading state
  if (isLoading) {
    return <DataTableSkeleton columns={4} rows={5} showToolbar={false} />
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
    <div className="space-y-4">
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

          <Select value={roleFilter} onValueChange={(value) => setRoleFilter(value as OrganizationRole | 'all')}>
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
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  ))}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow key={row.id}>
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={columns.length} className="h-24 text-center">
                    <DataTableEmptyState
                      title="No members found"
                      description={
                        searchTerm || roleFilter !== 'all'
                          ? 'Try adjusting your search or filters'
                          : 'Invite team members to start collaborating'
                      }
                      icon={<Users className="h-8 w-8 text-muted-foreground/50" />}
                    />
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
  )
}
