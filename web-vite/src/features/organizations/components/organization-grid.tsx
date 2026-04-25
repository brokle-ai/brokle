import { useState, useMemo } from 'react'
import { Link } from '@tanstack/react-router'
import { Building2, Plus, Search } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { PageHeader } from '@/components/layout/page-header'
import { CreateOrganizationDialog } from './create-organization-dialog'
import type { Organization } from '../queries'

interface OrganizationGridProps {
  organizations: Organization[]
  className?: string
}

const PLAN_LABEL: Record<Organization['subscription_plan'], string> = {
  free: 'Free',
  pro: 'Pro',
  business: 'Business',
  enterprise: 'Enterprise',
}

const PLAN_VARIANT: Record<
  Organization['subscription_plan'],
  'secondary' | 'default' | 'outline'
> = {
  free: 'outline',
  pro: 'secondary',
  business: 'default',
  enterprise: 'default',
}

function formatDate(iso: string): string {
  try {
    return new Intl.DateTimeFormat(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    }).format(new Date(iso))
  } catch {
    return iso
  }
}

function initialsFromName(name: string): string {
  const parts = name.trim().split(/\s+/).slice(0, 2)
  return parts.map((p) => p.charAt(0).toUpperCase()).join('') || '?'
}

export function OrganizationGrid({
  organizations,
  className,
}: OrganizationGridProps) {
  const [searchTerm, setSearchTerm] = useState('')
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  const filtered = useMemo(() => {
    if (!searchTerm) return organizations
    const term = searchTerm.toLowerCase()
    return organizations.filter(
      (org) =>
        org.name.toLowerCase().includes(term) ||
        org.slug.toLowerCase().includes(term),
    )
  }, [organizations, searchTerm])

  return (
    <>
      <PageHeader title="Organizations">
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Organization
        </Button>
        <CreateOrganizationDialog
          open={createDialogOpen}
          onOpenChange={setCreateDialogOpen}
        />
      </PageHeader>

      <div
        className={cn(
          '-mx-4 flex-1 overflow-auto px-4 py-1 space-y-6',
          className,
        )}
      >
        <div className="relative max-w-md">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <Input
            placeholder="Search organizations..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>

        {filtered.length > 0 ? (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {filtered.map((org) => (
              <Link
                key={org.id}
                to="/o/$orgId"
                params={{ orgId: org.id }}
                className="block"
              >
                <Card className="cursor-pointer group h-full">
                  <CardHeader className="pb-3">
                    <div className="flex items-center gap-3">
                      <div className="bg-muted flex size-10 items-center justify-center rounded-lg group-hover:bg-primary/10 transition-colors text-sm font-medium">
                        {initialsFromName(org.name) || (
                          <Building2 className="size-5" />
                        )}
                      </div>
                      <div className="flex-1 min-w-0">
                        <CardTitle className="text-lg truncate">
                          {org.name}
                        </CardTitle>
                        <div className="text-xs text-muted-foreground truncate">
                          {org.slug}
                        </div>
                      </div>
                    </div>
                  </CardHeader>

                  <CardContent className="space-y-3">
                    <div className="flex items-center gap-2">
                      <Badge variant={PLAN_VARIANT[org.subscription_plan]}>
                        {PLAN_LABEL[org.subscription_plan]}
                      </Badge>
                    </div>
                    <CardDescription className="text-xs">
                      Created {formatDate(org.created_at)}
                    </CardDescription>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        ) : searchTerm ? (
          <Card className="text-center py-12">
            <CardContent>
              <Building2 className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
              <h3 className="text-lg font-medium mb-2">
                No organizations match your search
              </h3>
              <p className="text-muted-foreground">
                Try a different search term.
              </p>
            </CardContent>
          </Card>
        ) : (
          <Card className="text-center py-12">
            <CardContent>
              <Building2 className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
              <h3 className="text-lg font-medium mb-2">
                No organizations yet
              </h3>
              <p className="text-muted-foreground mb-6">
                Create your first organization to start managing AI projects
                and team members.
              </p>
              <Button onClick={() => setCreateDialogOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Create Your First Organization
              </Button>
            </CardContent>
          </Card>
        )}
      </div>
    </>
  )
}
