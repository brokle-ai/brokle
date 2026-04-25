import * as React from 'react'
import { Link, useParams, useRouterState } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import {
  Building2,
  ChevronRight,
  CreditCard,
  FolderOpen,
  Home,
  Settings,
  Users,
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { organizationMembershipQueryOptions } from '@/features/organizations/queries'
import { projectMembershipQueryOptions } from '@/features/projects/queries'
import { cn } from '@/lib/utils'

interface BreadcrumbItemData {
  key: string
  href?: string
  label: string
  icon?: LucideIcon
  isActive: boolean
}

interface BreadcrumbsProps {
  className?: string
}

// Route-driven breadcrumbs. Uses tenancy params + pathname segments
// following the `/o/:orgId/p/:projectId/...` shape. The org + project
// names come from the membership queries that the route loaders have
// already warmed into the query cache, so these reads are synchronous
// under <Suspense>.
export function Breadcrumbs({ className }: BreadcrumbsProps) {
  const { orgId, projectId } = useParams({ strict: false }) as {
    orgId?: string
    projectId?: string
  }
  const { location } = useRouterState()
  const pathname = location.pathname

  const items: BreadcrumbItemData[] = [
    { key: 'root', href: '/o', label: 'Organizations', icon: Home, isActive: pathname === '/o' },
  ]

  if (orgId) {
    items.push({
      key: `org-${orgId}`,
      href: `/o/${orgId}`,
      label: '__org__',
      icon: Building2,
      isActive: !projectId && pathname === `/o/${orgId}`,
    })
  }

  if (orgId && projectId) {
    const projectBase = `/o/${orgId}/p/${projectId}`
    items.push({
      key: `project-${projectId}`,
      href: projectBase,
      label: '__project__',
      icon: FolderOpen,
      isActive: pathname === projectBase,
    })

    // Sub-page: pathname tail after the project base.
    const tail = pathname.startsWith(projectBase)
      ? pathname.slice(projectBase.length).replace(/^\/+/, '')
      : ''
    if (tail) {
      const tailSegs = tail.split('/')
      const head = tailSegs[0]!
      const subLabels: Record<string, { label: string; icon?: LucideIcon }> = {
        traces: { label: 'Traces' },
        sessions: { label: 'Sessions' },
        dashboards: { label: 'Dashboards' },
        datasets: { label: 'Datasets' },
        evaluators: { label: 'Evaluators' },
        experiments: { label: 'Experiments' },
        prompts: { label: 'Prompts' },
        settings: { label: 'Settings', icon: Settings },
      }
      const sub = subLabels[head] ?? { label: head }
      items.push({
        key: `sub-${head}`,
        href: `${projectBase}/${head}`,
        label: sub.label,
        icon: sub.icon,
        isActive: tailSegs.length === 1,
      })
      if (tailSegs.length > 1) {
        const deep = tailSegs[1]!
        const deepLabels: Record<string, { label: string; icon?: LucideIcon }> = {
          members: { label: 'Members', icon: Users },
          'ai-providers': { label: 'AI Providers' },
          billing: { label: 'Billing', icon: CreditCard },
          organization: { label: 'Organization', icon: Building2 },
        }
        const d = deepLabels[deep] ?? { label: deep }
        items.push({
          key: `deep-${deep}`,
          href: `${projectBase}/${head}/${deep}`,
          label: d.label,
          icon: d.icon,
          isActive: true,
        })
      }
    }
  }

  if (items.length <= 1) return null

  return (
    <Breadcrumb className={cn('flex items-center', className)}>
      <BreadcrumbList className="flex items-center">
        {items.map((item, index) => (
          <React.Fragment key={item.key}>
            <BreadcrumbItem className="flex items-center">
              {item.isActive || !item.href ? (
                <BreadcrumbPage className="flex items-center gap-1.5 text-sm font-medium">
                  {item.icon ? <item.icon className="h-4 w-4" /> : null}
                  <ResolvedLabel label={item.label} orgId={orgId} projectId={projectId} />
                </BreadcrumbPage>
              ) : (
                <BreadcrumbLink asChild>
                  <Link
                    to={item.href}
                    className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
                  >
                    {item.icon ? <item.icon className="h-4 w-4" /> : null}
                    <ResolvedLabel label={item.label} orgId={orgId} projectId={projectId} />
                  </Link>
                </BreadcrumbLink>
              )}
            </BreadcrumbItem>
            {index < items.length - 1 && (
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

function ResolvedLabel({
  label,
  orgId,
  projectId,
}: {
  label: string
  orgId?: string
  projectId?: string
}) {
  if (label === '__org__' && orgId) return <OrgLabel orgId={orgId} />
  if (label === '__project__' && projectId) return <ProjectLabel projectId={projectId} />
  return <>{label}</>
}

function OrgLabel({ orgId }: { orgId: string }) {
  const { data } = useSuspenseQuery(organizationMembershipQueryOptions(orgId))
  return <>{data.name}</>
}

function ProjectLabel({ projectId }: { projectId: string }) {
  const { data } = useSuspenseQuery(projectMembershipQueryOptions(projectId))
  return <>{data.name}</>
}
