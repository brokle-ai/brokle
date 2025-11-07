import { type LucideIcon } from 'lucide-react'
import { type ReactNode } from 'react'
import { type Scope } from '@/hooks/rbac/use-has-access'
import { type ProjectSummary, type OrganizationWithProjects } from '@/features/authentication'

export enum RouteSection {
  Main = 'main',
  Secondary = 'secondary',
}

export enum RouteGroup {
  Project = 'Project',
  Observability = 'Observability',
  Settings = 'Settings',
}

export type NavigationContext = {
  currentOrganizationId: string | null
  currentProjectId: string | null
  currentOrgSlug: string | null
  currentProjectSlug: string | null
  pathname: string
  currentProject: ProjectSummary | null
  currentOrganization: OrganizationWithProjects | null
}

export type BadgeConfig =
  | { type: 'static', value: string }
  | { type: 'dynamic', key: string }

export type Route = {
  title: string
  pathname: string
  icon?: LucideIcon
  section: RouteSection
  group?: RouteGroup
  menuNode?: ReactNode
  badge?: BadgeConfig
  rbacScope?: Scope | Scope[]
  featureFlag?: string
  show?: (context: NavigationContext) => boolean
  newTab?: boolean
}

export type ProcessedRoute = Route & {
  url: string
  isActive: boolean
}
