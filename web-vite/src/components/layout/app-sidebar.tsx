import { Link, useParams, useRouterState } from '@tanstack/react-router'
import {
  Activity,
  BarChart3,
  CheckSquare,
  CreditCard,
  Database,
  FileText,
  FlaskConical,
  LayoutDashboard,
  MessageSquare,
  Settings,
  Users,
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface NavItem {
  label: string
  to: string
  icon: typeof Activity
}

interface NavGroup {
  label: string
  items: NavItem[]
}

// Nav groups map to the tenancy-scoped routes that Phase 1.5 ports
// fill in. Each group is an independent commit.
function buildGroups(orgId: string, projectId: string): NavGroup[] {
  const p = `/o/${orgId}/p/${projectId}`
  return [
    {
      label: 'Observability',
      items: [
        { label: 'Overview', to: `${p}`, icon: LayoutDashboard },
        { label: 'Traces', to: `${p}/traces`, icon: Activity },
        { label: 'Sessions', to: `${p}/sessions`, icon: MessageSquare },
        { label: 'Dashboards', to: `${p}/dashboards`, icon: BarChart3 },
      ],
    },
    {
      label: 'Evaluation',
      items: [
        { label: 'Datasets', to: `${p}/datasets`, icon: Database },
        { label: 'Evaluators', to: `${p}/evaluators`, icon: CheckSquare },
        { label: 'Experiments', to: `${p}/experiments`, icon: FlaskConical },
      ],
    },
    {
      label: 'Prompts',
      items: [{ label: 'Library', to: `${p}/prompts`, icon: FileText }],
    },
    {
      label: 'Settings',
      items: [
        { label: 'Members', to: `${p}/settings/members`, icon: Users },
        { label: 'Project', to: `${p}/settings`, icon: Settings },
        { label: 'Billing', to: `/o/${orgId}/billing`, icon: CreditCard },
      ],
    },
  ]
}

export function AppSidebar() {
  const { orgId, projectId } = useParams({ strict: false }) as {
    orgId?: string
    projectId?: string
  }
  const { location } = useRouterState()
  const pathname = location.pathname

  if (!orgId || !projectId) return null
  const groups = buildGroups(orgId, projectId)

  return (
    <aside className="w-56 shrink-0 border-r bg-sidebar text-sidebar-foreground">
      <div className="flex h-14 items-center border-b px-4">
        <span className="font-semibold">Brokle</span>
      </div>
      <nav className="space-y-4 p-3">
        {groups.map((group) => (
          <div key={group.label}>
            <div className="px-2 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
              {group.label}
            </div>
            <ul className="space-y-0.5">
              {group.items.map((item) => {
                const Icon = item.icon
                const active =
                  pathname === item.to || pathname.startsWith(`${item.to}/`)
                return (
                  <li key={item.to}>
                    <Link
                      to={item.to}
                      className={cn(
                        'flex items-center gap-2 rounded-md px-2 py-1.5 text-sm',
                        active
                          ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                          : 'hover:bg-sidebar-accent/50',
                      )}
                    >
                      <Icon size={16} />
                      <span>{item.label}</span>
                    </Link>
                  </li>
                )
              })}
            </ul>
          </div>
        ))}
      </nav>
    </aside>
  )
}
