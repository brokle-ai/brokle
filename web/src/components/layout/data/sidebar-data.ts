import {
  IconBrowserCheck,
  IconHelp,
  IconLayoutDashboard,
  IconNotification,
  IconPalette,
  IconSettings,
  IconTool,
  IconUserCog,
  IconUsers,
} from '@tabler/icons-react'
import { BarChart3, Database, DollarSign, Building2 } from 'lucide-react'
import { BrokleLogo } from '@/assets/brokle-logo'
import { type SidebarData } from '../types'

// Function to generate context-aware navigation
export function getSidebarData(orgSlug?: string, projectSlug?: string): SidebarData {
  const orgUrl = orgSlug ? `/organizations/${orgSlug}` : ''
  const projectUrl = projectSlug ? `/projects/${projectSlug}` : ''
  
  // Define navigation groups based on context
  const navGroups = []

  // Root level navigation (when no organization is selected)
  if (!orgSlug) {
    navGroups.push({
      title: 'Dashboard',
      items: [
        {
          title: 'Overview',
          url: '/',
          icon: IconLayoutDashboard,
        },
      ],
    })
  }

  // Organization level navigation (when organization is selected)
  if (orgSlug && !projectSlug) {
    navGroups.push(
      {
        title: 'Organization',
        items: [
          {
            title: 'Overview',
            url: orgUrl,
            icon: Building2,
          },
          {
            title: 'Members',
            url: `${orgUrl}/members`,
            icon: IconUsers,
          },
          {
            title: 'Billing',
            url: `${orgUrl}/billing`,
            icon: DollarSign,
          },
          {
            title: 'Settings',
            url: `${orgUrl}/settings`,
            icon: IconSettings,
          },
        ],
      }
    )
  }
      
  // Project level navigation (when project is selected)
  if (projectSlug) {
    navGroups.push(
      {
        title: 'Project',
        items: [
          {
            title: 'Dashboard',
            url: projectUrl,
            icon: IconLayoutDashboard,
          },
          {
            title: 'Analytics',
            url: `${projectUrl}/analytics`,
            icon: BarChart3,
          },
          {
            title: 'Models',
            url: `${projectUrl}/models`,
            icon: Database,
          },
          {
            title: 'Costs',
            url: `${projectUrl}/costs`,
            badge: '$1.2k',
            icon: DollarSign,
          },
          {
            title: 'Settings',
            url: `${projectUrl}/settings`,
            icon: IconSettings,
          },
        ],
      }
    )
  }

  // Settings section (context-aware)
  const settingsItems = []
  
  // Organization settings (only when organization is selected)
  if (orgSlug) {
    settingsItems.push(
      {
        title: 'Organization',
        url: `${orgUrl}/settings`,
        icon: Building2,
      },
      {
        title: 'Members',
        url: `${orgUrl}/members`,
        icon: IconUsers,
      },
      {
        title: 'Billing',
        url: `${orgUrl}/billing`,
        icon: DollarSign,
      }
    )
  }

  // Personal settings (always available)
  settingsItems.push(
    {
      title: 'Personal',
      icon: IconSettings,
      items: [
        {
          title: 'Profile',
          url: '/settings',
          icon: IconUserCog,
        },
        {
          title: 'Account',
          url: '/settings/account',
          icon: IconTool,
        },
        {
          title: 'Appearance',
          url: '/settings/appearance',
          icon: IconPalette,
        },
        {
          title: 'Notifications',
          url: '/settings/notifications',
          icon: IconNotification,
        },
        {
          title: 'Display',
          url: '/settings/display',
          icon: IconBrowserCheck,
        },
      ],
    },
    {
      title: 'Help Center',
      url: '/help-center',
      icon: IconHelp,
    }
  )

  navGroups.push({
    title: 'Settings',
    items: settingsItems,
  })

  return {
    user: {
      name: 'AI Engineer',
      email: 'engineer@company.com',
      avatar: '/avatars/user.jpg',
    },
    teams: [
      {
        name: 'Brokle AI Platform',
        logo: BrokleLogo,
        plan: 'Complete AI Infrastructure',
      },
    ],
    navGroups,
  }
}

// Default sidebar data for non-context pages
export const sidebarData = getSidebarData()
