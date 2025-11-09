'use client'

import * as React from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { ChevronRight, Home, Building2, FolderOpen, Settings, Users, CreditCard } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { generateCompositeSlug, extractIdFromCompositeSlug } from '@/lib/utils/slug-utils'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { cn } from '@/lib/utils'

interface BreadcrumbsProps {
  className?: string
}

export function Breadcrumbs({ className }: BreadcrumbsProps) {
  const pathname = usePathname()
  const { currentOrganization, currentProject } = useWorkspace()

  const getBreadcrumbItems = () => {
    const items = []
    const segments = pathname.split('/').filter(Boolean)

    // Home
    items.push({
      href: '/',
      label: 'Organizations',
      icon: Home,
      isActive: segments.length === 0,
    })

    // Organization
    if (currentOrganization) {
      const orgCompositeSlug = generateCompositeSlug(currentOrganization.name, currentOrganization.id)
      items.push({
        href: `/${orgCompositeSlug}`,
        label: currentOrganization.name,
        icon: Building2,
        isActive: segments.length === 1,
      })

      // Check for organization-level pages
      if (segments.length >= 2) {
        const secondSegment = segments[1]
        
        if (secondSegment === 'settings') {
          items.push({
            href: `/${orgCompositeSlug}/settings`,
            label: 'Settings',
            icon: Settings,
            isActive: segments.length === 2,
          })

          // Settings sub-pages
          if (segments.length >= 3) {
            const settingsPage = segments[2]
            const settingsPages: Record<string, { label: string; icon: React.ElementType }> = {
              'organization': { label: 'Organization', icon: Building2 },
              'members': { label: 'Members', icon: Users },
              'billing': { label: 'Billing', icon: CreditCard },
            }

            if (settingsPages[settingsPage]) {
              items.push({
                href: `/${orgCompositeSlug}/settings/${settingsPage}`,
                label: settingsPages[settingsPage].label,
                icon: settingsPages[settingsPage].icon,
                isActive: true,
              })
            }
          }
        } else if (secondSegment === 'projects') {
          items.push({
            href: `/${orgCompositeSlug}/projects`,
            label: 'Projects',
            icon: FolderOpen,
            isActive: true,
          })
        } else if (currentProject) {
          // Check if second segment matches project by extracting ID
          const secondSegmentId = extractIdFromCompositeSlug(secondSegment)
          if (secondSegmentId === currentProject.id) {
            // Project
            const projectCompositeSlug = generateCompositeSlug(currentProject.name, currentProject.id)
            items.push({
              href: `/${orgCompositeSlug}/${projectCompositeSlug}`,
              label: currentProject.name,
              icon: FolderOpen,
            isActive: segments.length === 2,
          })

          // Project sub-pages
          if (segments.length >= 3) {
            const projectPage = segments[2]
            const projectPages: Record<string, string> = {
              'tasks': 'Tasks',
              'traces': 'Traces',
              'settings': 'Settings',
            }

            if (projectPages[projectPage]) {
              items.push({
                href: `/${orgCompositeSlug}/${projectCompositeSlug}/${projectPage}`,
                label: projectPages[projectPage],
                icon: undefined,
                isActive: true,
              })
            }
          }
          }
        }
      }
    }

    return items
  }

  const breadcrumbItems = getBreadcrumbItems()

  if (breadcrumbItems.length <= 1) {
    return null // Don't show breadcrumbs for root level
  }

  return (
    <Breadcrumb className={cn("flex items-center", className)}>
      <BreadcrumbList className="flex items-center">
        {breadcrumbItems.map((item, index) => (
          <React.Fragment key={item.href}>
            <BreadcrumbItem className="flex items-center">
              {item.isActive ? (
                <BreadcrumbPage className="flex items-center gap-1.5 text-sm font-medium">
                  {item.icon && <item.icon className="h-4 w-4" />}
                  {item.label}
                </BreadcrumbPage>
              ) : (
                <BreadcrumbLink asChild>
                  <Link href={item.href} className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors">
                    {item.icon && <item.icon className="h-4 w-4" />}
                    {item.label}
                  </Link>
                </BreadcrumbLink>
              )}
            </BreadcrumbItem>
            {index < breadcrumbItems.length - 1 && (
              <BreadcrumbSeparator className="flex items-center">
                <ChevronRight className="h-4 w-4" />
              </BreadcrumbSeparator>
            )}
          </React.Fragment>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  )
}