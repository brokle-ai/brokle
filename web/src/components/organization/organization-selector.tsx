'use client'

import * as React from 'react'
import { useState } from 'react'
import { Building2, ArrowRight, Plus } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useEffect } from 'react'
import { getUserOrganizations } from '@/lib/api/services/organizations'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Skeleton } from '@/components/ui/skeleton'
import { buildOrgUrl } from '@/lib/utils/slug-utils'
import { cn } from '@/lib/utils'
import type { Organization } from '@/types/organization'

interface OrganizationSelectorProps {
  className?: string
}

export function OrganizationSelector({ className }: OrganizationSelectorProps) {
  const router = useRouter()
  const [organizations, setOrganizations] = useState<Organization[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedOrgId, setSelectedOrgId] = useState<string | null>(null)

  // Fetch organizations using API client
  useEffect(() => {
    const fetchOrganizations = async () => {
      try {
        setIsLoading(true)
        
        // Use the existing API client
        const response = await getUserOrganizations()
        setOrganizations(response.data)
        setError(null)
      } catch (err) {
        console.error('Failed to fetch organizations:', err)
        setError(err instanceof Error ? err.message : 'Failed to load organizations')
      } finally {
        setIsLoading(false)
      }
    }

    fetchOrganizations()
  }, [])

  const handleOrgSelect = (organization: Organization) => {
    setSelectedOrgId(organization.id)
    const orgUrl = buildOrgUrl(organization.name, organization.id)
    router.push(orgUrl)
  }

  const getPlanBadgeColor = (plan: string) => {
    switch (plan) {
      case 'enterprise':
        return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300'
      case 'business':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
      case 'pro':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'free':
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

  const formatUsage = (count: number) => {
    if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`
    if (count >= 1000) return `${(count / 1000).toFixed(1)}K`
    return count.toString()
  }

  if (isLoading) {
    return (
      <div className={cn("space-y-6", className)}>
        <div className="text-center">
          <Skeleton className="h-8 w-64 mx-auto mb-2" />
          <Skeleton className="h-5 w-96 mx-auto" />
        </div>
        
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader className="pb-3">
                <div className="flex items-center gap-3">
                  <Skeleton className="h-10 w-10 rounded-lg" />
                  <div className="space-y-2">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-3 w-16" />
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <Skeleton className="h-3 w-full" />
                  <Skeleton className="h-3 w-2/3" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    )
  }

  if (error || !organizations || organizations.length === 0) {
    return (
      <div className={cn("space-y-6", className)}>
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground mb-2">
            No Organizations Found
          </h1>
          <p className="text-muted-foreground">
            You don't have access to any organizations yet.
          </p>
        </div>
        
        <div className="flex justify-center">
          <Card className="w-full max-w-md">
            <CardHeader className="text-center pb-3">
              <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
                <Building2 className="h-8 w-8" />
              </div>
              <CardTitle>Create Organization</CardTitle>
              <CardDescription>
                Get started by creating your first organization
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button className="w-full" onClick={() => router.push('/organizations/create')}>
                <Plus className="mr-2 h-4 w-4" />
                Create Organization
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  return (
    <div className={cn("space-y-6", className)}>
      <div className="text-center">
        <h1 className="text-3xl font-bold text-foreground mb-2">
          Select Organization
        </h1>
        <p className="text-lg text-muted-foreground">
          Choose an organization to access your AI infrastructure dashboard
        </p>
      </div>
      
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {organizations.map((org) => (
          <Card 
            key={org.id}
            className={cn(
              "cursor-pointer transition-all duration-200 hover:shadow-md hover:scale-[1.02]",
              selectedOrgId === org.id && "ring-2 ring-primary ring-offset-2"
            )}
            onClick={() => handleOrgSelect(org)}
          >
            <CardHeader className="pb-3">
              <div className="flex items-center gap-3">
                <Avatar className="h-10 w-10">
                  <AvatarImage 
                    src={`/api/organizations/${org.slug}/avatar`} 
                    alt={org.name} 
                  />
                  <AvatarFallback className="bg-primary text-primary-foreground font-semibold">
                    {getInitials(org.name)}
                  </AvatarFallback>
                </Avatar>
                <div className="space-y-1">
                  <CardTitle className="text-lg leading-none">
                    {org.name}
                  </CardTitle>
                  <Badge 
                    variant="secondary" 
                    className={cn("text-xs", getPlanBadgeColor(org.plan))}
                  >
                    {org.plan.charAt(0).toUpperCase() + org.plan.slice(1)}
                  </Badge>
                </div>
                <ArrowRight className="ml-auto h-5 w-5 text-muted-foreground" />
              </div>
            </CardHeader>
            
            <CardContent className="space-y-3">
              <div className="text-sm text-muted-foreground">
                {org.members.length} member{org.members.length !== 1 ? 's' : ''}
              </div>
              
              {org.usage && (
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <div className="font-medium text-foreground">
                      {formatUsage(org.usage.requests_this_month || 0)}
                    </div>
                    <div className="text-muted-foreground text-xs">
                      Requests/month
                    </div>
                  </div>
                  <div>
                    <div className="font-medium text-foreground">
                      ${(org.usage.cost_this_month || 0).toFixed(2)}
                    </div>
                    <div className="text-muted-foreground text-xs">
                      Cost/month
                    </div>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        ))}
        
        {/* Create new organization card */}
        <Card
          className="cursor-pointer transition-all duration-200 hover:shadow-md hover:scale-[1.02] border-dashed"
          onClick={() => router.push('/organizations/create')}
        >
          <CardHeader className="pb-3">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
                <Plus className="h-5 w-5" />
              </div>
              <div>
                <CardTitle className="text-lg leading-none">
                  Create Organization
                </CardTitle>
              </div>
              <ArrowRight className="ml-auto h-5 w-5 text-muted-foreground" />
            </div>
          </CardHeader>
          
          <CardContent>
            <CardDescription>
              Set up a new organization for your team or project
            </CardDescription>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}