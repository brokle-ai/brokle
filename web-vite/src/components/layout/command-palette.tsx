import * as React from 'react'
import { useNavigate, useParams } from '@tanstack/react-router'
import {
  Activity,
  ArrowRight,
  BarChart3,
  Boxes,
  CheckSquare,
  CreditCard,
  Database,
  FileText,
  FlaskConical,
  Laptop,
  LayoutDashboard,
  LogOut,
  MessageSquare,
  Moon,
  Plug,
  Search,
  Settings,
  Sun,
  Users,
} from 'lucide-react'
import { useTheme } from '@/context/theme-context'
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'
import { rawFetch } from '@/lib/api/client'

interface CommandRoute {
  title: string
  url: string
  icon: React.ComponentType<{ className?: string }>
  group: string
}

interface QuickAction {
  title: string
  icon: React.ComponentType<{ className?: string }>
  run: () => void
}

interface PaletteContext {
  open: boolean
  setOpen: (v: boolean) => void
  toggle: () => void
}

const PaletteCtx = React.createContext<PaletteContext | null>(null)

// Mounted at the document root by `<AuthenticatedLayout>` — owns the
// ⌘K / Ctrl+K shortcut + the dialog itself. Routes/quick-actions are
// derived from the current `:orgId/:projectId` so the palette only ever
// surfaces destinations a user can actually navigate to.
export function CommandPaletteProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = React.useState(false)
  const toggle = React.useCallback(() => setOpen((v) => !v), [])

  React.useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        toggle()
      }
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [toggle])

  const ctx = React.useMemo<PaletteContext>(
    () => ({ open, setOpen, toggle }),
    [open, toggle],
  )

  return (
    <PaletteCtx.Provider value={ctx}>
      {children}
      <CommandPalette />
    </PaletteCtx.Provider>
  )
}

export function useCommandPalette(): PaletteContext {
  const ctx = React.useContext(PaletteCtx)
  if (!ctx) {
    throw new Error('useCommandPalette must be used inside CommandPaletteProvider')
  }
  return ctx
}

function buildRoutes(orgId: string, projectId: string): CommandRoute[] {
  const p = `/o/${orgId}/p/${projectId}`
  return [
    { title: 'Overview', url: p, icon: LayoutDashboard, group: 'Observability' },
    { title: 'Traces', url: `${p}/traces`, icon: Activity, group: 'Observability' },
    { title: 'Sessions', url: `${p}/sessions`, icon: MessageSquare, group: 'Observability' },
    { title: 'Dashboards', url: `${p}/dashboards`, icon: BarChart3, group: 'Observability' },
    { title: 'Datasets', url: `${p}/datasets`, icon: Database, group: 'Evaluation' },
    { title: 'Evaluators', url: `${p}/evaluators`, icon: CheckSquare, group: 'Evaluation' },
    { title: 'Experiments', url: `${p}/experiments`, icon: FlaskConical, group: 'Evaluation' },
    { title: 'Prompts', url: `${p}/prompts`, icon: FileText, group: 'Prompts' },
    { title: 'Members', url: `${p}/settings/members`, icon: Users, group: 'Settings' },
    { title: 'AI Providers', url: `${p}/settings/ai-providers`, icon: Plug, group: 'Settings' },
    { title: 'Project Settings', url: `${p}/settings`, icon: Settings, group: 'Settings' },
    { title: 'Organization Settings', url: `/o/${orgId}/settings`, icon: Settings, group: 'Settings' },
    { title: 'Billing', url: `/o/${orgId}/billing`, icon: CreditCard, group: 'Settings' },
  ]
}

async function signOut(): Promise<void> {
  try {
    await rawFetch('/api/v1/auth/logout', { method: 'POST' })
  } catch {
    // Best-effort — local state still clears below.
  } finally {
    useAuthStore.getState().expireSession()
    window.location.href = '/signin'
  }
}

function CommandPalette() {
  const { open, setOpen } = useCommandPalette()
  const { setTheme } = useTheme()
  const navigate = useNavigate()
  const params = useParams({ strict: false }) as {
    orgId?: string
    projectId?: string
  }

  const routes: CommandRoute[] = React.useMemo(() => {
    if (!params.orgId || !params.projectId) return []
    return buildRoutes(params.orgId, params.projectId)
  }, [params.orgId, params.projectId])

  const grouped = React.useMemo(() => {
    const out: Record<string, CommandRoute[]> = {}
    for (const r of routes) {
      if (!out[r.group]) out[r.group] = []
      out[r.group]!.push(r)
    }
    return out
  }, [routes])

  const runCommand = React.useCallback(
    (command: () => void) => {
      setOpen(false)
      command()
    },
    [setOpen],
  )

  const quickActions: QuickAction[] = React.useMemo(() => {
    if (!params.orgId || !params.projectId) return []
    const p = `/o/${params.orgId}/p/${params.projectId}`
    return [
      {
        title: 'Create dashboard',
        icon: BarChart3,
        run: () => {
          void navigate({ to: `${p}/dashboards` })
        },
      },
      {
        title: 'Create dataset',
        icon: Database,
        run: () => {
          void navigate({ to: `${p}/datasets` })
        },
      },
      {
        title: 'Create prompt',
        icon: FileText,
        run: () => {
          void navigate({ to: `${p}/prompts` })
        },
      },
      {
        title: 'Switch organization',
        icon: Boxes,
        run: () => {
          void navigate({ to: '/o' })
        },
      },
    ]
  }, [params.orgId, params.projectId, navigate])

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput placeholder="Type a command or search..." />
      <CommandList>
        <ScrollArea type="hover" className="h-72 pe-1">
          <CommandEmpty>No results found.</CommandEmpty>

          {Object.entries(grouped).map(([group, items]) => (
            <CommandGroup key={group} heading={group}>
              {items.map((route) => {
                const Icon = route.icon
                return (
                  <CommandItem
                    key={route.url}
                    value={`${group} ${route.title}`}
                    onSelect={() =>
                      runCommand(() => {
                        void navigate({ to: route.url })
                      })
                    }
                  >
                    <div className="flex size-4 items-center justify-center">
                      <Icon className="size-4" />
                    </div>
                    <span>{route.title}</span>
                  </CommandItem>
                )
              })}
            </CommandGroup>
          ))}

          {quickActions.length > 0 && (
            <>
              <CommandSeparator />
              <CommandGroup heading="Quick actions">
                {quickActions.map((action) => {
                  const Icon = action.icon
                  return (
                    <CommandItem
                      key={action.title}
                      value={`Quick ${action.title}`}
                      onSelect={() => runCommand(action.run)}
                    >
                      <div className="flex size-4 items-center justify-center">
                        <Icon className="size-4" />
                      </div>
                      <span>{action.title}</span>
                    </CommandItem>
                  )
                })}
                <CommandItem
                  value="Sign out"
                  onSelect={() =>
                    runCommand(() => {
                      void signOut()
                    })
                  }
                >
                  <div className="flex size-4 items-center justify-center">
                    <LogOut className="size-4" />
                  </div>
                  <span>Sign out</span>
                </CommandItem>
              </CommandGroup>
            </>
          )}

          <CommandSeparator />
          <CommandGroup heading="Theme">
            <CommandItem onSelect={() => runCommand(() => setTheme('light'))}>
              <Sun className="size-4" />
              <span>Light</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => setTheme('dark'))}>
              <Moon className="size-4" />
              <span>Dark</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => setTheme('system'))}>
              <Laptop className="size-4" />
              <span>System</span>
            </CommandItem>
          </CommandGroup>
        </ScrollArea>
      </CommandList>
    </CommandDialog>
  )
}

// Topbar trigger — small ghost button with the platform-aware ⌘K hint.
export function CommandPaletteTrigger({ className }: { className?: string }) {
  const { setOpen } = useCommandPalette()
  const [isMac, setIsMac] = React.useState(false)

  React.useEffect(() => {
    setIsMac(typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform))
  }, [])

  return (
    <Button
      variant="outline"
      size="sm"
      onClick={() => setOpen(true)}
      className={cn(
        'h-8 gap-2 px-2 text-muted-foreground hover:text-foreground',
        className,
      )}
    >
      <Search className="size-4" />
      <span className="hidden text-xs sm:inline">Search</span>
      <kbd className="pointer-events-none ml-2 hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground sm:inline-flex">
        <span className="text-xs">{isMac ? '⌘' : 'Ctrl'}</span>K
      </kbd>
    </Button>
  )
}

// `ArrowRight` retained as fallback icon if a future route omits one.
export const _CommandFallbackIcon = ArrowRight
