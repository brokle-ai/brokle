'use client'

import * as React from 'react'
import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { ChevronsUpDown, Plus, Building2, FolderOpen, Settings, Users } from 'lucide-react'
import { useOrganization } from '@/context/org-context'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface OrganizationSwitcherProps {
  showProjects?: boolean
  className?: string
}

export function OrganizationSwitcher({
  showProjects = true,
  className
}: OrganizationSwitcherProps) {
  const router = useRouter()
  const { isMobile } = useSidebar()
  const {
    organizations,
    currentOrganization,
    currentProject,
    projects,
    switchOrganization,
    switchProject,
    isLoading,
    error
  } = useOrganization()

  const [isOrgLoading, setIsOrgLoading] = useState(false)
  const [isProjectLoading, setIsProjectLoading] = useState(false)

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

  const handleProjectSwitch = async (projectSlug: string) => {
    if (isProjectLoading || projectSlug === currentProject?.slug) return
    
    try {
      setIsProjectLoading(true)
      await switchProject(projectSlug)
    } catch (error) {
      console.error('Failed to switch project:', error)
    } finally {
      setIsProjectLoading(false)
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

  if (isLoading) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" className="animate-pulse">
            <div className="bg-muted flex aspect-square size-8 items-center justify-center rounded-lg">
              <Building2 className="size-4" />
            </div>
            <div className="grid flex-1 text-left text-sm leading-tight">
              <div className="h-4 bg-muted rounded w-24 mb-1"></div>
              <div className="h-3 bg-muted rounded w-16"></div>
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    )
  }

  if (error || !currentOrganization) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" className="text-destructive">
            <div className="bg-destructive/10 flex aspect-square size-8 items-center justify-center rounded-lg">
              <Building2 className="size-4" />
            </div>
            <div className="grid flex-1 text-left text-sm leading-tight">
              <span className="truncate font-semibold">Error</span>
              <span className="truncate text-xs">No organization</span>
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    )
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className={cn(
                "data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground",
                className
              )}
              disabled={isOrgLoading}
            >
              <Avatar className="h-8 w-8">
                <AvatarImage 
                  src={`/api/organizations/${currentOrganization.slug}/avatar`} 
                  alt={currentOrganization.name} 
                />
                <AvatarFallback className="bg-primary text-primary-foreground text-xs font-semibold">
                  {getInitials(currentOrganization.name)}
                </AvatarFallback>
              </Avatar>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">
                  {currentOrganization.name}
                </span>
                <div className="flex items-center gap-1">
                  <Badge 
                    variant="secondary" 
                    className={cn("text-xs px-1 py-0 h-4", getPlanBadgeColor(currentOrganization.plan))}
                  >
                    {currentOrganization.plan}
                  </Badge>
                  {showProjects && currentProject && (
                    <>
                      <span className="text-xs text-muted-foreground">•</span>
                      <span className="truncate text-xs text-muted-foreground">
                        {currentProject.name}
                      </span>
                    </>
                  )}
                </div>
              </div>
              <ChevronsUpDown className="ml-auto size-4" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          
          <DropdownMenuContent
            className="w-[--radix-dropdown-menu-trigger-width] min-w-64 rounded-lg"
            align="start"
            side={isMobile ? 'bottom' : 'right'}
            sideOffset={4}
          >
            {/* Organizations Section */}
            <DropdownMenuLabel className="text-muted-foreground text-xs">
              Organizations
            </DropdownMenuLabel>
            
            {organizations.map((org, index) => (
              <DropdownMenuItem
                key={org.id}
                onClick={() => handleOrgSwitch(org.slug)}
                className="gap-2 p-2 cursor-pointer"
                disabled={isOrgLoading}
              >
                <Avatar className="h-6 w-6">
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
                    <span className="truncate font-medium">{org.name}</span>
                    <Badge 
                      variant="secondary" 
                      className={cn("text-xs px-1 py-0 h-4", getPlanBadgeColor(org.plan))}
                    >
                      {org.plan}
                    </Badge>
                  </div>
                  {org.id === currentOrganization.id && (
                    <span className="text-xs text-muted-foreground">Current</span>
                  )}
                </div>
                <DropdownMenuShortcut>⌘{index + 1}</DropdownMenuShortcut>
              </DropdownMenuItem>
            ))}

            <DropdownMenuSeparator />

            {/* Projects Section */}
            {showProjects && projects.length > 0 && (
              <>
                <DropdownMenuLabel className="text-muted-foreground text-xs">
                  Projects in {currentOrganization.name}
                </DropdownMenuLabel>
                
                {projects.slice(0, 5).map((project) => (
                  <DropdownMenuItem
                    key={project.id}
                    onClick={() => handleProjectSwitch(project.slug)}
                    className="gap-2 p-2 cursor-pointer"
                    disabled={isProjectLoading}
                  >
                    <div className="bg-muted flex size-6 items-center justify-center rounded-sm">
                      <FolderOpen className="size-3" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <span className="truncate font-medium">{project.name}</span>
                      {project.id === currentProject?.id && (
                        <span className="text-xs text-muted-foreground block">Current</span>
                      )}
                    </div>
                    {project.status === 'active' && (
                      <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                    )}
                  </DropdownMenuItem>
                ))}

                {projects.length > 5 && (
                  <DropdownMenuItem className="gap-2 p-2 text-muted-foreground">
                    <FolderOpen className="size-4" />
                    <span>+{projects.length - 5} more projects</span>
                  </DropdownMenuItem>
                )}

                <DropdownMenuSeparator />
              </>
            )}

            {/* Management Actions */}
            <DropdownMenuItem className="gap-2 p-2 cursor-pointer">
              <div className="bg-muted flex size-6 items-center justify-center rounded-sm">
                <Settings className="size-3" />
              </div>
              <span>Organization Settings</span>
            </DropdownMenuItem>
            
            <DropdownMenuItem className="gap-2 p-2 cursor-pointer">
              <div className="bg-muted flex size-6 items-center justify-center rounded-sm">
                <Users className="size-3" />
              </div>
              <span>Manage Members</span>
            </DropdownMenuItem>

            <DropdownMenuSeparator />

            <DropdownMenuItem
              className="gap-2 p-2 cursor-pointer"
              onClick={() => router.push('/organizations/create')}
            >
              <div className="bg-primary text-primary-foreground flex size-6 items-center justify-center rounded-sm">
                <Plus className="size-3" />
              </div>
              <span className="font-medium">Create Organization</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}