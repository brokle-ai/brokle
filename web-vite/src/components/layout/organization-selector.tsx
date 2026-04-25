import { useState } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { ChevronDown, Plus, Settings } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'
import { organizationListQueryOptions } from '@/features/organizations/queries'
import { useIsMobile } from '@/hooks/use-mobile'

interface OrganizationSelectorProps {
  currentOrgId: string
  className?: string
}

// Combobox-style org switcher. Reads the org list from the query cache
// (already primed by the `/o/$orgId` route loader) and navigates to
// `/o/{newOrgId}` on selection — the org index route resolves to a
// project or renders the picker.
export function OrganizationSelector({
  currentOrgId,
  className,
}: OrganizationSelectorProps) {
  const { data } = useSuspenseQuery(organizationListQueryOptions())
  const navigate = useNavigate()
  const isMobile = useIsMobile()
  const [open, setOpen] = useState(false)

  const current = data.data.find((o) => o.id === currentOrgId)

  if (!current) return null

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger
        className={cn(
          'flex items-center gap-1 [&_svg]:pointer-events-none [&_svg]:shrink-0',
          'text-sm text-primary hover:text-primary/80 transition-colors',
          className,
        )}
      >
        <span className="font-normal">{current.name}</span>
        <ChevronDown className="size-4" />
      </DropdownMenuTrigger>

      <DropdownMenuContent
        className={cn(
          'max-h-96 overflow-y-auto',
          isMobile ? 'w-screen max-w-sm' : 'w-64',
        )}
        align="start"
      >
        <DropdownMenuItem className="font-semibold" asChild>
          <Link to="/o" className="cursor-pointer">
            Organizations
          </Link>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <div className="max-h-36 overflow-y-auto">
          {data.data.map((org) => (
            <DropdownMenuItem key={org.id} asChild>
              <Link
                to="/o/$orgId"
                params={{ orgId: org.id }}
                className="flex cursor-pointer justify-between"
                onClick={(e) => {
                  if (org.id === current.id) {
                    e.preventDefault()
                  }
                }}
              >
                <span
                  className="max-w-36 overflow-hidden overflow-ellipsis whitespace-nowrap"
                  title={org.name}
                >
                  {org.name}
                </span>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 hover:bg-background -my-1 ml-4"
                  aria-label={`Open ${org.name} settings`}
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    // Org-level settings landing page not yet wired;
                    // navigate to the org root and let the picker
                    // resolve to a project with a settings path.
                    void navigate({ to: '/o/$orgId', params: { orgId: org.id } })
                    setOpen(false)
                  }}
                >
                  <Settings className="h-3 w-3" />
                  <span className="sr-only">Open {org.name} settings</span>
                </Button>
              </Link>
            </DropdownMenuItem>
          ))}
        </div>

        <DropdownMenuSeparator />

        <DropdownMenuItem disabled>
          <Plus className="mr-1.5 h-4 w-4" aria-hidden="true" />
          New Organization
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
