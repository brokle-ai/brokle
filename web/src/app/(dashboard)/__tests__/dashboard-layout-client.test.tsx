// Regression test for the dashboard cold-load redirect bug.
//
// Bug history: dashboard-layout-client.tsx used to read
// `useAuthStore(s => s.isAuthenticated)` BEFORE its `useState(() =>
// hydrate(...))` initializer ran. The Zustand selector captured the
// initial `false`, the `useEffect(() => router.replace(SIGNIN))`
// closure locked that stale value, and every cold load bounced the
// user to /signin even though the server-side DAL had already proven
// the session valid.
//
// Fix shape: deleted the client-side cold-load redirect entirely.
// Auth is enforced by proxy.ts + lib/auth/dal.ts on the server;
// cross-tab logout / runtime expiry are handled reactively in
// components/providers.tsx (auth:session-expired + storage events).
//
// What this test guards against:
//   1. The hydrate seed still runs on first render — descendants
//      that read `useAuthStore.getState()` see a populated user.
//   2. No navigation side-effect fires on a valid cold load — neither
//      `redirect()` from next/navigation nor `useRouter().replace`.

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render } from '@testing-library/react'

import type {
  User,
  OrganizationWithProjects,
} from '@/features/authentication'

// next/navigation: capture replace/push/redirect to assert they're
// never called by the layout on a valid cold load. The component
// shouldn't import any of these any more (the offending useRouter
// import was deleted), but other code paths in the rendered tree
// might pull them in indirectly.
const { replace, push, redirect } = vi.hoisted(() => ({
  replace: vi.fn(),
  push: vi.fn(),
  redirect: vi.fn(),
}))

vi.mock('next/navigation', () => ({
  useRouter: () => ({ replace, push, prefetch: vi.fn(), back: vi.fn() }),
  usePathname: () => '/projects/test',
  useSearchParams: () => new URLSearchParams(),
  redirect,
}))

// Stub out the heavy child tree — we're testing this layout's own
// hydration + redirect contract, not the workspace context, sidebar,
// or query layer. Each replacement is an identity wrapper so the
// initialUser/initialOrganizations props still flow visibly.
vi.mock('@/context/workspace-context', () => ({
  WorkspaceProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useWorkspace: () => ({ error: null, user: null, organizations: [] }),
}))
vi.mock('@/components/layout/authenticated-layout', () => ({
  AuthenticatedLayout: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))
vi.mock('@/components/ui/sidebar', () => ({
  SidebarProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  SidebarInset: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))
vi.mock('@/components/layout/app-sidebar', () => ({
  AppSidebar: () => null,
}))
vi.mock('@/components/errors/workspace-error-page', () => ({
  WorkspaceErrorPage: () => null,
}))
vi.mock('@/hooks/use-navigation-context', () => ({
  useNavigationContext: () => ({
    context: {},
    permissions: {},
    featureFlags: {},
    isPermissionsLoading: false,
    user: null,
    isLoading: false,
  }),
}))
vi.mock('@/lib/navigation/process-routes', () => ({
  processNavigation: () => ({ mainNavigation: [], secondaryNavigation: [] }),
}))

// Imports below run AFTER the mocks above (vi.mock is hoisted but
// these statements use the mocked modules, so order matters
// semantically even if not syntactically).
import { DashboardLayoutClient } from '../dashboard-layout-client'
import { useAuthStore } from '@/features/authentication'

const fakeUser: User = {
  id: 'user-1',
  email: 'a@b.com',
  firstName: 'Test',
  lastName: 'User',
  isEmailVerified: true,
  defaultOrganizationId: 'org-1',
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
} as unknown as User

const fakeOrgs: OrganizationWithProjects[] = [
  {
    id: 'org-1',
    name: 'Test Org',
    plan: 'free',
    projects: [],
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  } as unknown as OrganizationWithProjects,
]

describe('DashboardLayoutClient', () => {
  beforeEach(() => {
    replace.mockClear()
    push.mockClear()
    redirect.mockClear()
    useAuthStore.getState().clearAuth()
  })

  it('hydrates the auth store on first render with the server-fetched user', () => {
    render(
      <DashboardLayoutClient
        defaultOpen
        initialUser={fakeUser}
        initialOrganizations={fakeOrgs}
      >
        <div data-testid="child" />
      </DashboardLayoutClient>,
    )

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.user?.id).toBe('user-1')
    expect(state.organization?.id).toBe('org-1')
  })

  it('does NOT redirect to /signin on a valid cold load', () => {
    render(
      <DashboardLayoutClient
        defaultOpen
        initialUser={fakeUser}
        initialOrganizations={fakeOrgs}
      >
        <div data-testid="child" />
      </DashboardLayoutClient>,
    )

    expect(replace).not.toHaveBeenCalled()
    expect(push).not.toHaveBeenCalled()
    expect(redirect).not.toHaveBeenCalled()
  })
})
