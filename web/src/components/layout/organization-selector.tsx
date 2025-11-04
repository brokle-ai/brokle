'use client'

import * as React from 'react'
import { useState } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { ChevronDown, Building2, Settings, Users, Plus } from 'lucide-react'
import Link from 'next/link'
import { useWorkspace } from '@/context/workspace-context'
import { useIsMobile } from '@/hooks/use-mobile'
import { getSmartRedirectUrl } from '@/lib/utils/smart-redirect'
import { generateCompositeSlug, extractIdFromCompositeSlug } from '@/lib/utils/slug-utils'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface OrganizationSelectorProps {
  className?: string
  showPlanBadge?: boolean
}

export function OrganizationSelector({ className, showPlanBadge = false }: OrganizationSelectorProps) {
  const {
    organizations,
    currentOrganization,
    isInitialized,
  } = useWorkspace()
  
  const pathname = usePathname()
  const router = useRouter()
  const isMobile = useIsMobile()
  const [isOrgLoading, setIsOrgLoading] = useState(false)

  const handleOrgSwitch = async (compositeSlug: string) => {
    // Generate current composite slug for comparison
    const currentCompositeSlug = currentOrganization
      ? generateCompositeSlug(currentOrganization.name, currentOrganization.id)
      : null

    if (isOrgLoading || compositeSlug === currentCompositeSlug) return

    // Extract ID from composite slug
    const targetOrgId = extractIdFromCompositeSlug(compositeSlug)

    // Find the organization object by ID
    const targetOrg = organizations.find(org => org.id === targetOrgId)
    if (!targetOrg) {
      console.error('Organization not found for composite slug:', compositeSlug)
      return
    }
    
    try {
      setIsOrgLoading(true)
      
      // Use smart redirect to determine the appropriate URL
      const redirectUrl = getSmartRedirectUrl({
        currentPath: pathname,
        targetOrgSlug: compositeSlug,
        targetOrgId: targetOrg.id,
        targetOrgName: targetOrg.name
      })

      // Navigate to smart redirect URL
      // Next.js handles context cleanup automatically during navigation
      router.push(redirectUrl)
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

  // Loading state - show shimmer only if not initialized yet
  if (!isInitialized) {
    return (
      <div className={cn("animate-pulse bg-muted rounded h-6 w-32", className)}></div>
    )
  }

  // Show nothing if no current organization after initialization
  if (!currentOrganization) {
    return null
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        className={cn(
          "flex items-center gap-1 [&_svg]:pointer-events-none [&_svg]:shrink-0",
          "text-sm text-primary hover:text-primary/80 transition-colors",
          isOrgLoading && "opacity-50 cursor-not-allowed",
          className
        )}
        disabled={isOrgLoading}
      >
        <span className="font-normal">{currentOrganization.name}</span>
        {showPlanBadge && currentOrganization.plan !== 'free' && (
          <Badge
            variant="secondary"
            className={cn(
              "ml-1 px-1 py-0 text-xs font-normal capitalize",
              getPlanBadgeColor(currentOrganization.plan)
            )}
          >
            {currentOrganization.plan}
          </Badge>
        )}
        <ChevronDown className="size-4" />
      </DropdownMenuTrigger>
      
      <DropdownMenuContent
        className={cn(
          "max-h-96 overflow-y-auto",
          isMobile ? "w-screen max-w-sm" : "w-64"
        )}
        align="start"
      >
        {/* Organizations overview link */}
        <DropdownMenuItem className="font-semibold" asChild>
          <Link href="/" className="cursor-pointer">
            Organizations
          </Link>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        {/* All organizations list */}
        <div className="max-h-36 overflow-y-auto">
          {organizations && organizations.map((org) => (
            <DropdownMenuItem
              key={org.id}
              asChild
            >
              <Link
                href={
                  org.id === currentOrganization.id
                    ? pathname
                    : getSmartRedirectUrl({
                        currentPath: pathname,
                        targetOrgSlug: generateCompositeSlug(org.name, org.id),
                        targetOrgId: org.id,
                        targetOrgName: org.name
                      })
                }
                className="flex cursor-pointer justify-between"
                onClick={(e) => {
                  if (org.id === currentOrganization.id) {
                    e.preventDefault()
                    return
                  }
                  setIsOrgLoading(true)
                }}
              >
                <span
                  className="max-w-36 overflow-hidden overflow-ellipsis whitespace-nowrap"
                  title={org.name}
                >
                  {org.name}
                </span>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 hover:bg-background -my-1 ml-4"
                  aria-label={`Open ${org.name} settings`}
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    router.push(`/organizations/${org.id}/settings`)
                  }}
                >
                  <Settings className="h-3 w-3" />
                  <span className="sr-only">Open {org.name} settings</span>
                </Button>
              </Link>
            </DropdownMenuItem>
          ))}
        </div>

        <DropdownMenuSeparator />

        {/* Create new organization */}
        <DropdownMenuItem asChild>
          <Button
            variant="ghost"
            size="sm"
            className="h-8 w-full text-sm font-normal justify-start"
            asChild
          >
            <Link href="/organizations/create">
              <Plus className="mr-1.5 h-4 w-4" aria-hidden="true" />
              New Organization
            </Link>
          </Button>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}