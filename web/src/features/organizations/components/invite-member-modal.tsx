'use client'

import { useState, useEffect } from 'react'
import { UserPlus, Mail, Loader2 } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { toast } from 'sonner'
import {
  createInvitation,
  getAvailableRolesForInvitation,
  type Role
} from '../api/invitations-api'

interface InviteMemberModalProps {
  trigger?: React.ReactNode
  onSuccess?: () => void
}

export function InviteMemberModal({ trigger, onSuccess }: InviteMemberModalProps) {
  const { currentOrganization } = useWorkspace()

  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isLoadingRoles, setIsLoadingRoles] = useState(false)
  const [email, setEmail] = useState('')
  const [roleId, setRoleId] = useState('')
  const [message, setMessage] = useState('')
  const [availableRoles, setAvailableRoles] = useState<Role[]>([])

  // Fetch available roles when modal opens
  useEffect(() => {
    if (isOpen && availableRoles.length === 0) {
      setIsLoadingRoles(true)
      getAvailableRolesForInvitation()
        .then((roles) => {
          setAvailableRoles(roles)
          // Set default role to developer if available
          const developerRole = roles.find(r => r.name === 'developer')
          if (developerRole) {
            setRoleId(developerRole.id)
          } else if (roles.length > 0) {
            setRoleId(roles[0].id)
          }
        })
        .catch((error) => {
          console.error('Failed to fetch roles:', error)
          toast.error('Failed to load available roles')
        })
        .finally(() => setIsLoadingRoles(false))
    }
  }, [isOpen, availableRoles.length])

  if (!currentOrganization) {
    return null
  }

  const selectedRole = availableRoles.find(r => r.id === roleId)

  const isValidEmail = (emailStr: string) => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return emailRegex.test(emailStr.trim())
  }

  const handleSendInvite = async () => {
    const trimmedEmail = email.trim()

    if (!trimmedEmail) {
      toast.error('Please enter an email address')
      return
    }

    if (!isValidEmail(trimmedEmail)) {
      toast.error('Please enter a valid email address')
      return
    }

    if (!roleId) {
      toast.error('Please select a role')
      return
    }

    setIsLoading(true)

    try {
      await createInvitation(currentOrganization.id, {
        email: trimmedEmail,
        role_id: roleId,
        message: message.trim() || undefined
      })

      toast.success(`Invitation sent to ${trimmedEmail}`)

      // Reset form and close
      setEmail('')
      setMessage('')
      setIsOpen(false)
      onSuccess?.()
    } catch (error: unknown) {
      console.error('Failed to send invitation:', error)

      // Handle specific error messages
      if (error && typeof error === 'object' && 'message' in error) {
        const errorMessage = String(error.message).toLowerCase()
        if (errorMessage.includes('already a member')) {
          toast.error('This user is already a member of the organization')
        } else if (errorMessage.includes('pending')) {
          toast.error('An invitation is already pending for this email')
        } else {
          toast.error(String(error.message))
        }
      } else {
        toast.error('Failed to send invitation. Please try again.')
      }
    } finally {
      setIsLoading(false)
    }
  }

  const getRoleDescription = (role: Role | undefined) => {
    if (!role) return ''
    if (role.description) return role.description
    // Fallback descriptions for system roles
    switch (role.name) {
      case 'admin':
        return 'Can manage members, projects, and organization settings'
      case 'developer':
        return 'Can create and manage projects, view analytics and costs'
      case 'viewer':
        return 'Read-only access to projects and basic analytics'
      default:
        return ''
    }
  }

  const defaultTrigger = (
    <Button>
      <UserPlus className="mr-2 h-4 w-4" />
      Invite Member
    </Button>
  )

  const canSubmit = email.trim() && isValidEmail(email.trim()) && roleId

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        {trigger || defaultTrigger}
      </DialogTrigger>

      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <UserPlus className="h-5 w-5" />
            Invite Member
          </DialogTitle>
          <DialogDescription>
            Send an invitation to join {currentOrganization.name}.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Email */}
          <div className="space-y-2">
            <Label htmlFor="email">Email Address</Label>
            <Input
              id="email"
              type="email"
              placeholder="colleague@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isLoading}
            />
          </div>

          {/* Role Selection */}
          <div className="space-y-2">
            <Label htmlFor="role">Role</Label>
            <Select
              value={roleId}
              onValueChange={setRoleId}
              disabled={isLoadingRoles || isLoading}
            >
              <SelectTrigger>
                <SelectValue placeholder={isLoadingRoles ? 'Loading roles...' : 'Select a role'} />
              </SelectTrigger>
              <SelectContent>
                {availableRoles.map((role) => (
                  <SelectItem key={role.id} value={role.id}>
                    {role.displayName || role.name.charAt(0).toUpperCase() + role.name.slice(1)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {selectedRole && (
              <p className="text-xs text-muted-foreground">
                {getRoleDescription(selectedRole)}
              </p>
            )}
          </div>

          {/* Personal Message */}
          <div className="space-y-2">
            <Label htmlFor="message">
              Personal Message <span className="text-muted-foreground">(optional)</span>
            </Label>
            <Textarea
              id="message"
              placeholder="Add a personal note to the invitation..."
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              rows={2}
              className="resize-none"
              disabled={isLoading}
            />
          </div>
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => setIsOpen(false)}
            disabled={isLoading}
          >
            Cancel
          </Button>

          <Button onClick={handleSendInvite} disabled={isLoading || !canSubmit}>
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Sending...
              </>
            ) : (
              <>
                <Mail className="mr-2 h-4 w-4" />
                Send Invitation
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
