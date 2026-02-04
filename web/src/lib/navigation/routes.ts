import { RouteSection, RouteGroup, type Route } from './types'
import {
  Grid2X2,
  Settings,
  FolderOpen,
  Home,
  Activity,
  FileText,
  FlaskConical,
  Database,
  BarChart3,
  Scale,
  LayoutDashboard,
  ClipboardCheck,
} from 'lucide-react'

export const ROUTES: Route[] = [
  // ========================================
  // ROOT CONTEXT
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
  // ORGANIZATION CONTEXT
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
  // PROJECT CONTEXT
  // ========================================

  // Project Group
  {
    title: 'Overview',
    pathname: '/projects/[projectSlug]',
    icon: FolderOpen,
    section: RouteSection.Main,
    rbacScope: 'projects:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },
  {
    title: 'Dashboards',
    pathname: '/projects/[projectSlug]/dashboards',
    icon: LayoutDashboard,
    section: RouteSection.Main,
    rbacScope: 'projects:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },

  // Observability Group
  {
    title: 'Traces',
    pathname: '/projects/[projectSlug]/traces',
    icon: Activity,
    section: RouteSection.Main,
    group: RouteGroup.Observability,
    rbacScope: 'projects:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },
  {
    title: 'Prompts',
    pathname: '/projects/[projectSlug]/prompts',
    icon: FileText,
    section: RouteSection.Main,
    group: RouteGroup.Observability,
    rbacScope: 'projects:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },
  {
    title: 'Playground',
    pathname: '/projects/[projectSlug]/playground',
    icon: FlaskConical,
    section: RouteSection.Main,
    group: RouteGroup.Observability,
    rbacScope: 'projects:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },

  // Evaluations Group (ordered by workflow)
  {
    title: 'Datasets',
    pathname: '/projects/[projectSlug]/datasets',
    icon: Database,
    section: RouteSection.Main,
    group: RouteGroup.Evaluations,
    rbacScope: 'projects:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },
  {
    title: 'Evaluators',
    pathname: '/projects/[projectSlug]/evaluators',
    icon: Scale,
    section: RouteSection.Main,
    group: RouteGroup.Evaluations,
    rbacScope: 'projects:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },
  {
    title: 'Experiments',
    pathname: '/projects/[projectSlug]/experiments',
    icon: FlaskConical,
    section: RouteSection.Main,
    group: RouteGroup.Evaluations,
    rbacScope: 'projects:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },
  {
    title: 'Scores',
    pathname: '/projects/[projectSlug]/scores',
    icon: BarChart3,
    section: RouteSection.Main,
    group: RouteGroup.Evaluations,
    rbacScope: 'projects:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },
  {
    title: 'Annotation Queues',
    pathname: '/projects/[projectSlug]/annotation-queues',
    icon: ClipboardCheck,
    section: RouteSection.Main,
    group: RouteGroup.Evaluations,
    rbacScope: 'projects:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },

  // Other Group
  {
    title: 'Settings',
    pathname: '/projects/[projectSlug]/settings',
    icon: Settings,
    section: RouteSection.Main,
    group: RouteGroup.Other,
    rbacScope: 'settings:read',
    show: ({ currentProjectSlug }) => !!currentProjectSlug,
  },

  // ========================================
  // USER SETTINGS CONTEXT
  // ========================================
  {
    title: 'Home',
    pathname: '/',
    icon: Home,
    section: RouteSection.Main,
    show: ({ pathname }) => pathname.startsWith('/settings'),
  },
]
