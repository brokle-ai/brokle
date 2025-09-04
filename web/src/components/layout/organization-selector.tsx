'use client'

import * as React from 'react'
import { useState } from 'react'
import { ChevronDown, Building2, Settings, Users, Plus } from 'lucide-react'
import { useOrganization } from '@/context/organization-context'
import { useIsMobile } from '@/hooks/use-mobile'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface OrganizationSelectorProps {
  className?: string
}

export function OrganizationSelector({ className }: OrganizationSelectorProps) {
  const { 
    organizations, 
    currentOrganization, 
    switchOrganization,
    isLoading,
  } = useOrganization()
  
  const isMobile = useIsMobile()
  const [isOrgLoading, setIsOrgLoading] = useState(false)

  const handleOrgSwitch = async (orgSlug: string) => {
    if (isOrgLoading || orgSlug === currentOrganization?.slug) return
    
    try {
      setIsOrgLoading(true)
      await switchOrganization(orgSlug)
    } catch (error) {
      console.error('Failed to switch organization:', error)
    } finally {
      setIsOrgLoading(false)
    }
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

  // Loading state
  if (isLoading || !currentOrganization) {
    return (
      <div className={cn("animate-pulse bg-muted rounded h-8 w-32", className)}></div>
    )
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          className={cn(
            "gap-2 justify-start text-left font-normal",
            isMobile ? "max-w-[140px]" : "max-w-[180px]",
            "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
          )}
          disabled={isOrgLoading}
        >
          <Avatar className="h-4 w-4">
            <AvatarImage 
              src={`/api/organizations/${currentOrganization.slug}/avatar`} 
              alt={currentOrganization.name} 
            />
            <AvatarFallback className="bg-primary text-primary-foreground text-xs font-semibold">
              {getInitials(currentOrganization.name)}
            </AvatarFallback>
          </Avatar>
          <span className="truncate text-sm">
            {currentOrganization.name}
          </span>
          <ChevronDown className="ml-auto h-4 w-4 shrink-0" />
        </Button>
      </DropdownMenuTrigger>
      
      <DropdownMenuContent
        className={cn(
          "max-h-96 overflow-y-auto",
          isMobile ? "w-screen max-w-sm" : "w-72"
        )}
        align="start"
        side="bottom"
        sideOffset={4}
      >
        {/* Current Organization */}
        <DropdownMenuLabel className="text-xs text-muted-foreground">
          Current Organization
        </DropdownMenuLabel>
        <div className="px-2 py-2 border-b mb-1">
          <div className="flex items-center gap-2">
            <Avatar className="h-6 w-6">
              <AvatarImage 
                src={`/api/organizations/${currentOrganization.slug}/avatar`} 
                alt={currentOrganization.name} 
              />
              <AvatarFallback className="text-xs">
                {getInitials(currentOrganization.name)}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="font-medium text-sm">{currentOrganization.name}</span>
                <Badge 
                  variant="secondary" 
                  className={cn("text-xs px-1 py-0 h-4", getPlanBadgeColor(currentOrganization.plan))}
                >
                  {currentOrganization.plan}
                </Badge>
              </div>
            </div>
          </div>
        </div>

        {/* Switch Organization */}
        {organizations.filter(org => org.id !== currentOrganization.id).length > 0 && (
          <>
            <DropdownMenuLabel className="text-xs text-muted-foreground">
              Switch Organization
            </DropdownMenuLabel>
            {organizations.filter(org => org.id !== currentOrganization.id).map((org) => (
              <DropdownMenuItem
                key={org.id}
                onClick={() => handleOrgSwitch(org.slug)}
                className="gap-2 p-2 cursor-pointer"
                disabled={isOrgLoading}
              >
                <Avatar className="h-5 w-5">
                  <AvatarImage 
                    src={`/api/organizations/${org.slug}/avatar`} 
                    alt={org.name} 
                  />
                  <AvatarFallback className="text-xs">
                    {getInitials(org.name)}
                  </AvatarFallback>
                </Avatar>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-sm">{org.name}</span>
                    <Badge 
                      variant="secondary" 
                      className={cn("text-xs px-1 py-0 h-4", getPlanBadgeColor(org.plan))}
                    >
                      {org.plan}
                    </Badge>
                  </div>
                </div>
                <Building2 className="h-4 w-4 text-muted-foreground" />
              </DropdownMenuItem>
            ))}
            <DropdownMenuSeparator />
          </>
        )}
        
        {/* Organization Actions */}
        <DropdownMenuItem className="gap-2 p-2 cursor-pointer">
          <div className="bg-muted flex size-5 items-center justify-center rounded-sm">
            <Settings className="size-3" />
          </div>
          <span className="text-sm">Organization Settings</span>
        </DropdownMenuItem>
        
        <DropdownMenuItem className="gap-2 p-2 cursor-pointer">
          <div className="bg-muted flex size-5 items-center justify-center rounded-sm">
            <Users className="size-3" />
          </div>
          <span className="text-sm">Manage Members</span>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem className="gap-2 p-2 cursor-pointer">
          <div className="bg-primary text-primary-foreground flex size-5 items-center justify-center rounded-sm">
            <Plus className="size-3" />
          </div>
          <span className="text-sm font-medium">Create Organization</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}