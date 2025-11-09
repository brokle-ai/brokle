import { RouteSection, RouteGroup, type Route } from './types'
import {
  Grid2X2,
  Settings,
  FolderOpen,
  Home,
} from 'lucide-react'

export const ROUTES: Route[] = [
  // ========================================
  // ROOT CONTEXT (1 route)
  // ========================================
  {
    title: 'Dashboard',
    pathname: '/',
    icon: Grid2X2,
    section: RouteSection.Main,
    show: ({ currentProject, currentOrganization }) =>
      !currentProject && !currentOrganization,
  },

  // ========================================
  // ORGANIZATION CONTEXT (2 routes)
  // ========================================
  {
    title: 'Projects',
    pathname: '/organizations/[orgSlug]',
    icon: Grid2X2,
    section: RouteSection.Main,
    show: ({ currentOrgSlug, currentProjectSlug, pathname }) =>
      !!currentOrgSlug &&
      !currentProjectSlug &&
      !pathname.startsWith('/settings'),
  },
  {
    title: 'Settings',
    pathname: '/organizations/[orgSlug]/settings',
    icon: Settings,
    section: RouteSection.Main,
    rbacScope: 'settings:read',
    show: ({ currentProjectSlug, pathname }) =>
      !currentProjectSlug &&
      !pathname.startsWith('/settings'),
  },

  // ========================================
  // PROJECT CONTEXT (6 routes)
  // ========================================

  // Project Group (1 route)
  {
    title: 'Overview',
    pathname: '/projects/[projectSlug]',
    icon: FolderOpen,
    section: RouteSection.Main,
    rbacScope: 'projects:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },

  // Settings Group (1 route)
  {
    title: 'Settings',
    pathname: '/projects/[projectSlug]/settings',
    icon: Settings,
    section: RouteSection.Main,
    group: RouteGroup.Settings,
    rbacScope: 'settings:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },

  // ========================================
  // USER SETTINGS CONTEXT (1 route)
  // ========================================
  {
    title: 'Home',
    pathname: '/',
    icon: Home,
    section: RouteSection.Main,
    show: ({ pathname }) => pathname.startsWith('/settings'),
  },
]

// Total: 6 routes
// - Root: 1 (Dashboard)
// - Organization: 2 (Projects, Settings)
// - Project: 2 (Overview, Settings)
// - User Settings: 1 (Home - back to dashboard)
