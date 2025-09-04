'use client'

import { Shield, ArrowLeft, Mail, Building2, FolderOpen } from 'lucide-react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface AccessDeniedProps {
  type?: 'organization' | 'project' | 'page'
  title?: string
  description?: string
  orgName?: string
  projectName?: string
  requiredRole?: string
  currentRole?: string
  backUrl?: string
  className?: string
}

export function AccessDenied({
  type = 'page',
  title,
  description,
  orgName,
  projectName,
  requiredRole,
  currentRole,
  backUrl,
  className
}: AccessDeniedProps) {
  const router = useRouter()

  const getDefaultContent = () => {
    switch (type) {
      case 'organization':
        return {
          title: title || `Access Denied to ${orgName || 'Organization'}`,
          description: description || "You don't have permission to access this organization.",
          icon: Building2,
          reasons: [
            'You are not a member of this organization',
            'Your account may not have the required permissions',
            'The organization may be private or restricted',
            'You may need to be invited by an organization owner or admin'
          ]
        }
      case 'project':
        return {
          title: title || `Access Denied to ${projectName || 'Project'}`,
          description: description || "You don't have permission to access this project.",
          icon: FolderOpen,
          reasons: [
            'You may not have access to this specific project',
            'The project may be restricted to certain team members',
            'You may need higher permissions to view this project',
            'The project may be archived or disabled'
          ]
        }
      default:
        return {
          title: title || 'Access Denied',
          description: description || "You don't have permission to access this page.",
          icon: Shield,
          reasons: [
            'You may not have the required role or permissions',
            'This page may be restricted to certain users',
            'Your session may have expired',
            'Contact your organization admin for access'
          ]
        }
    }
  }

  const content = getDefaultContent()
  const Icon = content.icon

  const handleGoBack = () => {
    if (backUrl) {
      router.push(backUrl)
    } else {
      router.back()
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-6">
      <Card className="w-full max-w-lg text-center">
        <CardHeader className="pb-6">
          <div className="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
            <Icon className="h-8 w-8 text-destructive" />
          </div>
          <CardTitle className="text-xl">{content.title}</CardTitle>
          <CardDescription className="text-base">
            {content.description}
          </CardDescription>
        </CardHeader>
        
        <CardContent className="space-y-6">
          {/* Role Information */}
          {requiredRole && (
            <Alert>
              <Shield className="h-4 w-4" />
              <AlertDescription>
                This action requires <strong>{requiredRole}</strong> role.
                {currentRole && (
                  <span> Your current role is <strong>{currentRole}</strong>.</span>
                )}
              </AlertDescription>
            </Alert>
          )}

          {/* Possible Reasons */}
          <div className="text-left">
            <h4 className="text-sm font-medium mb-3">This could happen if:</h4>
            <ul className="text-sm text-muted-foreground space-y-1">
              {content.reasons.map((reason, index) => (
                <li key={index} className="flex items-start gap-2">
                  <span className="text-muted-foreground mt-1">•</span>
                  <span>{reason}</span>
                </li>
              ))}
            </ul>
          </div>

          {/* Actions */}
          <div className="flex flex-col gap-2 pt-4">
            <Button onClick={handleGoBack} className="w-full">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Go Back
            </Button>
            
            <div className="grid grid-cols-2 gap-2">
              <Button variant="outline" asChild>
                <Link href="/">
                  <Building2 className="mr-2 h-4 w-4" />
                  Organizations
                </Link>
              </Button>
              
              <Button variant="outline" asChild>
                <Link href="/help-center">
                  <Mail className="mr-2 h-4 w-4" />
                  Get Help
                </Link>
              </Button>
            </div>
          </div>

          {/* Contact Info */}
          <div className="text-xs text-muted-foreground border-t pt-4">
            If you believe this is an error, please contact your organization administrator
            or <Link href="/support" className="text-primary hover:underline">contact support</Link>.
          </div>
        </CardContent>
      </Card>
    </div>
  )
}