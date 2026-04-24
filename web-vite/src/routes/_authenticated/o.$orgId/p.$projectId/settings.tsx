import { createFileRoute, Link, Outlet } from '@tanstack/react-router'

// Settings shell — owns the tab strip + outlet. The index route
// (`settings/index.tsx`) renders the project card; `settings/members`
// and `settings/api-keys` render the list views.
export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/settings',
)({
  component: SettingsShell,
})

function SettingsShell() {
  const { orgId, projectId } = Route.useParams()

  return (
    <main className="mx-auto max-w-7xl space-y-6 p-6">
      <header>
        <h1 className="text-2xl font-semibold">Settings</h1>
        <p className="text-sm text-muted-foreground">
          Manage project configuration, members, and API keys.
        </p>
      </header>

      <nav className="flex gap-2 border-b">
        <TabLink to="/o/$orgId/p/$projectId/settings" params={{ orgId, projectId }}>
          General
        </TabLink>
        <TabLink
          to="/o/$orgId/p/$projectId/settings/members"
          params={{ orgId, projectId }}
        >
          Members
        </TabLink>
        <TabLink
          to="/o/$orgId/p/$projectId/settings/api-keys"
          params={{ orgId, projectId }}
        >
          API Keys
        </TabLink>
      </nav>

      <Outlet />
    </main>
  )
}

// Small helper so tab-style Links render an underline on active routes.
// Using TanStack Router's `activeProps` keeps routing state out of
// component props and mirrors the exemplar pattern in o.$orgId.tsx.
function TabLink(props: {
  to: string
  params: { orgId: string; projectId: string }
  children: React.ReactNode
}) {
  return (
    <Link
      to={props.to}
      params={props.params}
      className="px-3 py-2 text-sm text-muted-foreground hover:text-foreground"
      activeProps={{
        className:
          'px-3 py-2 text-sm font-medium text-foreground border-b-2 border-foreground -mb-px',
      }}
      activeOptions={{ exact: true }}
    >
      {props.children}
    </Link>
  )
}
