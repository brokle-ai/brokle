'use client'

import { useState, useEffect, useCallback } from 'react'
import { Mail, RefreshCw, Trash2, Clock, Loader2, AlertCircle } from 'lucide-react'
import { formatDistanceToNow, format } from 'date-fns'
import { useWorkspace } from '@/context/workspace-context'
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
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'
import {
  getPendingInvitations,
  resendInvitation,
  revokeInvitation,
  type Invitation
} from '../api/invitations-api'

interface PendingInvitationsProps {
  onInvitationsChange?: () => void
}

export function PendingInvitations({ onInvitationsChange }: PendingInvitationsProps) {
  const { currentOrganization } = useWorkspace()

  const [invitations, setInvitations] = useState<Invitation[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null)

  const fetchInvitations = useCallback(async () => {
    if (!currentOrganization) return

    try {
      setError(null)
      const invitationsData = await getPendingInvitations(currentOrganization.id)
      setInvitations(invitationsData)
    } catch (err) {
      console.error('Failed to fetch invitations:', err)
      setError('Failed to load pending invitations')
    } finally {
      setIsLoading(false)
    }
  }, [currentOrganization])

  useEffect(() => {
    fetchInvitations()
  }, [fetchInvitations])

  const handleResend = async (invitation: Invitation) => {
    if (!currentOrganization) return

    setActionLoadingId(invitation.id)
    try {
      await resendInvitation(currentOrganization.id, invitation.id)
      toast.success(`Invitation resent to ${invitation.email}`)
      await fetchInvitations()
      onInvitationsChange?.()
    } catch (err: any) {
      console.error('Failed to resend invitation:', err)
      const errorMessage = err?.message || 'Failed to resend invitation'
      toast.error(errorMessage)
    } finally {
      setActionLoadingId(null)
    }
  }

  const handleRevoke = async (invitation: Invitation) => {
    if (!currentOrganization) return

    setActionLoadingId(invitation.id)
    try {
      await revokeInvitation(currentOrganization.id, invitation.id)
      toast.success(`Invitation to ${invitation.email} revoked`)
      await fetchInvitations()
      onInvitationsChange?.()
    } catch (err: any) {
      console.error('Failed to revoke invitation:', err)
      const errorMessage = err?.message || 'Failed to revoke invitation'
      toast.error(errorMessage)
    } finally {
      setActionLoadingId(null)
    }
  }

  const getExpirationStatus = (expiresAt: Date) => {
    const now = new Date()
    const diffMs = expiresAt.getTime() - now.getTime()
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffMs < 0) {
      return { variant: 'destructive' as const, text: 'Expired', className: '' }
    }
    if (diffDays <= 1) {
      return { variant: 'outline' as const, text: 'Expires soon', className: 'border-orange-500 text-orange-600' }
    }
    return { variant: 'secondary' as const, text: `Expires in ${diffDays} days`, className: '' }
  }

  if (!currentOrganization) {
    return null
  }

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Mail className="h-5 w-5" />
            Pending Invitations
          </CardTitle>
          <CardDescription>
            Invitations waiting to be accepted
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex items-center gap-4">
                <Skeleton className="h-10 w-10 rounded-full" />
                <div className="space-y-2 flex-1">
                  <Skeleton className="h-4 w-[200px]" />
                  <Skeleton className="h-3 w-[150px]" />
                </div>
                <Skeleton className="h-8 w-20" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Mail className="h-5 w-5" />
            Pending Invitations
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8 text-muted-foreground">
            <AlertCircle className="mr-2 h-4 w-4" />
            {error}
            <Button variant="link" onClick={fetchInvitations} className="ml-2">
              Retry
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (invitations.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Mail className="h-5 w-5" />
            Pending Invitations
          </CardTitle>
          <CardDescription>
            Invitations waiting to be accepted
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center py-8 text-center text-muted-foreground">
            <Mail className="h-12 w-12 mb-4 opacity-50" />
            <p>No pending invitations</p>
            <p className="text-sm">Invited members will appear here until they accept.</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Mail className="h-5 w-5" />
          Pending Invitations
          <Badge variant="secondary" className="ml-2">
            {invitations.length}
          </Badge>
        </CardTitle>
        <CardDescription>
          Invitations waiting to be accepted
        </CardDescription>
      </CardHeader>
      <CardContent>
        <TooltipProvider>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Invited By</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {invitations.map((invitation) => {
                const expirationStatus = getExpirationStatus(invitation.expiresAt)
                const isActionLoading = actionLoadingId === invitation.id

                return (
                  <TableRow key={invitation.id}>
                    <TableCell>
                      <div className="flex flex-col">
                        <span className="font-medium">{invitation.email}</span>
                        {invitation.message && (
                          <span className="text-xs text-muted-foreground truncate max-w-[200px]">
                            &quot;{invitation.message}&quot;
                          </span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {invitation.roleName}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col">
                        <span className="text-sm">{invitation.invitedByName}</span>
                        <span className="text-xs text-muted-foreground">
                          {formatDistanceToNow(invitation.createdAt, { addSuffix: true })}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Tooltip>
                        <TooltipTrigger>
                          <Badge variant={expirationStatus.variant} className={expirationStatus.className}>
                            <Clock className="mr-1 h-3 w-3" />
                            {expirationStatus.text}
                          </Badge>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Expires: {format(invitation.expiresAt, 'PPpp')}</p>
                          {invitation.resentCount > 0 && (
                            <p className="text-xs">Resent {invitation.resentCount} time{invitation.resentCount !== 1 ? 's' : ''}</p>
                          )}
                        </TooltipContent>
                      </Tooltip>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-2">
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleResend(invitation)}
                              disabled={isActionLoading}
                            >
                              {isActionLoading ? (
                                <Loader2 className="h-4 w-4 animate-spin" />
                              ) : (
                                <RefreshCw className="h-4 w-4" />
                              )}
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p>Resend invitation</p>
                          </TooltipContent>
                        </Tooltip>

                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="text-destructive hover:text-destructive"
                              disabled={isActionLoading}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Revoke Invitation</AlertDialogTitle>
                              <AlertDialogDescription>
                                Are you sure you want to revoke the invitation to{' '}
                                <strong>{invitation.email}</strong>? They will no longer be able to
                                join this organization using the existing invitation link.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction
                                onClick={() => handleRevoke(invitation)}
                                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                              >
                                Revoke Invitation
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </TooltipProvider>
      </CardContent>
    </Card>
  )
}
