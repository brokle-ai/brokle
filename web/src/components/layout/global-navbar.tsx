'use client'

import React from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useAuth } from '@/hooks/auth/use-auth'
import { BrokleLogo } from '@/assets/brokle-logo'
import { OrgSwitcher } from './org-switcher'
import { ProjectSwitcher } from './project-switcher'
import { NavUser } from './nav-user'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'

// Mock data - in real app these would come from API/context
const mockOrganizations = [
  { id: '1', name: 'Acme Corp', slug: 'acme-corp', plan: 'pro' },
  { id: '2', name: 'Tech Startup', slug: 'tech-startup', plan: 'free' },
]

const mockProjects = [
  { id: '1', name: 'AI Assistant', slug: 'ai-assistant', description: 'Customer support chatbot' },
  { id: '2', name: 'Analytics Engine', slug: 'analytics-engine', description: 'Data processing pipeline' },
]

export function GlobalNavbar() {
  const pathname = usePathname()
  const { user, isLoading } = useAuth()

  // Determine current context from URL
  const isOrgRoute = pathname.startsWith('/organizations/')
  const isProjectRoute = pathname.startsWith('/projects/')
  
  // Extract slugs from URL
  const orgSlug = isOrgRoute ? pathname.split('/')[2] : null
  const projectSlug = isProjectRoute ? pathname.split('/')[2] : null
  
  const currentOrganization = mockOrganizations.find(org => org.slug === orgSlug)
  const currentProject = mockProjects.find(project => project.slug === projectSlug)

  if (isLoading) {
    return (
      <nav className="h-16 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="flex h-16 items-center px-4">
          <div className="flex items-center gap-2">
            <BrokleLogo className="h-6 w-6" />
            <span className="text-lg font-semibold">Brokle</span>
          </div>
          <div className="ml-auto">
            <div className="h-8 w-8 animate-pulse rounded-full bg-muted"></div>
          </div>
        </div>
      </nav>
    )
  }

  if (!user) {
    return null
  }

  return (
    <nav className="h-16 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-16 items-center px-4">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
          <BrokleLogo className="h-6 w-6" />
          <span className="text-lg font-semibold">Brokle</span>
        </Link>

        <Separator orientation="vertical" className="mx-4 h-6" />

        {/* Navigation Links */}
        <div className="flex items-center gap-1">
          <Button 
            variant={pathname === '/' ? 'secondary' : 'ghost'} 
            size="sm" 
            asChild
          >
            <Link href="/">Dashboard</Link>
          </Button>
          
          <Button 
            variant={pathname.startsWith('/organizations') ? 'secondary' : 'ghost'} 
            size="sm" 
            asChild
          >
            <Link href="/organizations">Organizations</Link>
          </Button>
          
          <Button 
            variant={pathname.startsWith('/projects') ? 'secondary' : 'ghost'} 
            size="sm" 
            asChild
          >
            <Link href="/projects">Projects</Link>
          </Button>
        </div>

        {/* Context Switchers */}
        <div className="ml-auto flex items-center gap-4">
          {/* Organization Switcher - show when on org routes or when org is available */}
          {(isOrgRoute || mockOrganizations.length > 0) && (
            <OrgSwitcher 
              currentOrganization={currentOrganization}
              organizations={mockOrganizations}
            />
          )}

          {/* Project Switcher - show when on project routes or when projects are available */}
          {(isProjectRoute || mockProjects.length > 0) && (
            <ProjectSwitcher 
              currentProject={currentProject}
              projects={mockProjects}
            />
          )}

          <Separator orientation="vertical" className="h-6" />

          {/* User Menu */}
          <NavUser 
            user={{
              name: user.name,
              email: user.email,
              avatar: '', // Add avatar URL when available
            }} 
          />
        </div>
      </div>
    </nav>
  )
}