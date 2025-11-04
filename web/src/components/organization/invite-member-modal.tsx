'use client'

import { useState } from 'react'
import { UserPlus, Mail, Loader2, Copy, Check } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { getOrgSlug } from '@/lib/utils/slug-utils'
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
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import type { OrganizationRole } from '@/types/organization'

interface InviteMemberModalProps {
  trigger?: React.ReactNode
  onSuccess?: () => void
}

export function InviteMemberModal({ trigger, onSuccess }: InviteMemberModalProps) {
  const { currentOrganization } = useWorkspace()
  
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [inviteMethod, setInviteMethod] = useState<'email' | 'link'>('email')
  const [emails, setEmails] = useState('')
  const [role, setRole] = useState<OrganizationRole>('developer')
  const [message, setMessage] = useState('')
  const [inviteLink, setInviteLink] = useState('')
  const [linkCopied, setLinkCopied] = useState(false)

  if (!currentOrganization) {
    return null
  }

  const validateEmails = (emailString: string) => {
    const emailList = emailString
      .split(/[,\n]/)
      .map(email => email.trim())
      .filter(email => email.length > 0)

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    const validEmails = emailList.filter(email => emailRegex.test(email))
    const invalidEmails = emailList.filter(email => !emailRegex.test(email))

    return { validEmails, invalidEmails, total: emailList.length }
  }

  const handleSendInvites = async () => {
    if (inviteMethod === 'email') {
      const { validEmails, invalidEmails } = validateEmails(emails)
      
      if (validEmails.length === 0) {
        toast.error('Please enter at least one valid email address')
        return
      }
      
      if (invalidEmails.length > 0) {
        toast.error(`Invalid email addresses: ${invalidEmails.join(', ')}`)
        return
      }

      setIsLoading(true)
      
      try {
        // TODO: Implement API call to send invitations
        await new Promise(resolve => setTimeout(resolve, 1000)) // Simulate API call
        
        toast.success(`Invitations sent to ${validEmails.length} email${validEmails.length !== 1 ? 's' : ''}`)
        
        // Reset form
        setEmails('')
        setMessage('')
        setIsOpen(false)
        onSuccess?.()
      } catch (error) {
        console.error('Failed to send invitations:', error)
        toast.error('Failed to send invitations. Please try again.')
      } finally {
        setIsLoading(false)
      }
    }
  }

  const generateInviteLink = async () => {
    setIsLoading(true)
    
    try {
      // TODO: Implement API call to generate invite link
      await new Promise(resolve => setTimeout(resolve, 500)) // Simulate API call

      const mockLink = `https://app.brokle.com/invite/${getOrgSlug(currentOrganization)}?token=abc123def456&role=${role}`
      setInviteLink(mockLink)
      toast.success('Invite link generated successfully')
    } catch (error) {
      console.error('Failed to generate invite link:', error)
      toast.error('Failed to generate invite link. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const copyInviteLink = async () => {
    try {
      await navigator.clipboard.writeText(inviteLink)
      setLinkCopied(true)
      toast.success('Invite link copied to clipboard')
      setTimeout(() => setLinkCopied(false), 2000)
    } catch (error) {
      toast.error('Failed to copy link')
    }
  }

  const getRoleDescription = (role: OrganizationRole) => {
    switch (role) {
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

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        {trigger || defaultTrigger}
      </DialogTrigger>
      
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <UserPlus className="h-5 w-5" />
            Invite Members to {currentOrganization.name}
          </DialogTitle>
          <DialogDescription>
            Invite team members to collaborate on AI projects in your organization.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Invite Method Selection */}
          <div className="flex gap-2 p-1 bg-muted rounded-lg">
            <Button
              variant={inviteMethod === 'email' ? 'default' : 'ghost'}
              size="sm"
              className="flex-1"
              onClick={() => setInviteMethod('email')}
            >
              <Mail className="mr-2 h-4 w-4" />
              Email Invites
            </Button>
            <Button
              variant={inviteMethod === 'link' ? 'default' : 'ghost'}
              size="sm"
              className="flex-1"
              onClick={() => setInviteMethod('link')}
            >
              <Copy className="mr-2 h-4 w-4" />
              Share Link
            </Button>
          </div>

          {/* Role Selection */}
          <div className="space-y-2">
            <Label htmlFor="role">Role *</Label>
            <Select value={role} onValueChange={(value: OrganizationRole) => setRole(value)}>
              <SelectTrigger>
                <SelectValue placeholder="Select a role" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="admin">Admin</SelectItem>
                <SelectItem value="developer">Developer</SelectItem>
                <SelectItem value="viewer">Viewer</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              {getRoleDescription(role)}
            </p>
          </div>

          {inviteMethod === 'email' ? (
            // Email Invites
            <>
              <div className="space-y-2">
                <Label htmlFor="emails">Email Addresses *</Label>
                <Textarea
                  id="emails"
                  placeholder="Enter email addresses separated by commas or new lines&#10;example@company.com, another@company.com"
                  value={emails}
                  onChange={(e) => setEmails(e.target.value)}
                  rows={4}
                  className="resize-none"
                />
                {emails && (
                  <div className="text-xs text-muted-foreground">
                    {(() => {
                      const { validEmails, invalidEmails, total } = validateEmails(emails)
                      return (
                        <div className="flex flex-wrap gap-1">
                          <span>{validEmails.length} valid</span>
                          {invalidEmails.length > 0 && (
                            <span className="text-destructive">, {invalidEmails.length} invalid</span>
                          )}
                          <span> of {total} total</span>
                        </div>
                      )
                    })()}
                  </div>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="message">Personal Message (Optional)</Label>
                <Textarea
                  id="message"
                  placeholder="Add a personal message to the invitation..."
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  rows={3}
                  className="resize-none"
                />
              </div>
            </>
          ) : (
            // Invite Link
            <div className="space-y-4">
              {!inviteLink ? (
                <Card>
                  <CardHeader className="pb-3">
                    <CardTitle className="text-lg">Generate Invite Link</CardTitle>
                    <CardDescription>
                      Create a shareable link that anyone can use to join your organization
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button onClick={generateInviteLink} disabled={isLoading} className="w-full">
                      {isLoading ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Generating...
                        </>
                      ) : (
                        'Generate Invite Link'
                      )}
                    </Button>
                  </CardContent>
                </Card>
              ) : (
                <Card>
                  <CardHeader className="pb-3">
                    <CardTitle className="text-lg">Invite Link Generated</CardTitle>
                    <CardDescription>
                      Share this link with people you want to invite
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="flex gap-2">
                      <Input
                        value={inviteLink}
                        readOnly
                        className="font-mono text-sm"
                      />
                      <Button onClick={copyInviteLink} variant="outline">
                        {linkCopied ? (
                          <Check className="h-4 w-4" />
                        ) : (
                          <Copy className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                    
                    <div className="text-xs text-muted-foreground">
                      • This link will expire in 7 days<br />
                      • New members will join as <Badge variant="outline" className="text-xs">{role}</Badge><br />
                      • You can revoke this link at any time
                    </div>
                  </CardContent>
                </Card>
              )}
            </div>
          )}
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
          
          {inviteMethod === 'email' && (
            <Button onClick={handleSendInvites} disabled={isLoading || !emails.trim()}>
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Sending...
                </>
              ) : (
                <>
                  <Mail className="mr-2 h-4 w-4" />
                  Send Invitations
                </>
              )}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}