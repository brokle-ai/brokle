import { useState, useMemo } from 'react'
import { Link } from '@tanstack/react-router'
import { FolderOpen, Plus, Search } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { PageHeader } from '@/components/layout/page-header'
import { CreateProjectDialog } from '@/features/projects/components/create-project-dialog'
import type { Project } from '@/features/projects/queries'

interface ProjectGridProps {
  orgId: string
  projects: Project[]
  className?: string
  showCreateButton?: boolean
}

type StatusFilter = 'all' | 'active' | 'archived'
type SortKey = 'name' | 'created'

export function ProjectGrid({
  orgId,
  projects,
  className,
  showCreateButton = true,
}: ProjectGridProps) {
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [sortBy, setSortBy] = useState<SortKey>('name')
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  const filteredAndSorted = useMemo(() => {
    const filtered = projects.filter((project) => {
      const term = searchTerm.toLowerCase()
      const matchesSearch =
        project.name.toLowerCase().includes(term) ||
        (project.description?.toLowerCase().includes(term) ?? false)
      const matchesStatus =
        statusFilter === 'all' || project.status === statusFilter
      return matchesSearch && matchesStatus
    })

    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name)
        case 'created':
          return (
            new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
          )
        default:
          return 0
      }
    })

    return filtered
  }, [projects, searchTerm, statusFilter, sortBy])

  const getStatusColor = (status: Project['status']) => {
    switch (status) {
      case 'active':
        return 'bg-green-500'
      case 'archived':
        return 'bg-gray-500'
      default:
        return 'bg-green-500'
    }
  }

  const hasFilters = Boolean(searchTerm) || statusFilter !== 'all'

  return (
    <>
      <PageHeader title="Projects">
        {showCreateButton && (
          <>
            <Button onClick={() => setCreateDialogOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              New Project
            </Button>
            <CreateProjectDialog
              organizationId={orgId}
              open={createDialogOpen}
              onOpenChange={setCreateDialogOpen}
            />
          </>
        )}
      </PageHeader>

      <div
        className={cn(
          '-mx-4 flex-1 overflow-auto px-4 py-1 space-y-6',
          className,
        )}
      >
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
            <Input
              placeholder="Search projects..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>

          <Select
            value={statusFilter}
            onValueChange={(value) => setStatusFilter(value as StatusFilter)}
          >
            <SelectTrigger className="w-full sm:w-40">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="archived">Archived</SelectItem>
            </SelectContent>
          </Select>

          <Select
            value={sortBy}
            onValueChange={(value) => setSortBy(value as SortKey)}
          >
            <SelectTrigger className="w-full sm:w-40">
              <SelectValue placeholder="Sort by" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="name">Name</SelectItem>
              <SelectItem value="created">Created Date</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {filteredAndSorted.length > 0 ? (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {filteredAndSorted.map((project) => (
              <Link
                key={project.id}
                to="/o/$orgId/p/$projectId"
                params={{ orgId, projectId: project.id }}
                className="block"
              >
                <Card className="cursor-pointer group h-full">
                  <CardHeader className="pb-3">
                    <div className="flex items-center gap-3">
                      <div className="bg-muted flex size-10 items-center justify-center rounded-lg group-hover:bg-primary/10 transition-colors">
                        <FolderOpen className="size-5 group-hover:text-primary transition-colors" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <CardTitle className="text-lg truncate">
                          {project.name}
                        </CardTitle>
                        <div className="flex items-center gap-2 mt-1">
                          <div
                            className={cn(
                              'w-2 h-2 rounded-full',
                              getStatusColor(project.status),
                            )}
                          />
                          <span className="text-sm capitalize text-muted-foreground">
                            {project.status ?? 'active'}
                          </span>
                        </div>
                      </div>
                    </div>
                  </CardHeader>

                  <CardContent className="space-y-2">
                    {project.description ? (
                      <CardDescription className="line-clamp-2">
                        {project.description}
                      </CardDescription>
                    ) : (
                      <CardDescription className="italic text-muted-foreground/70">
                        No description
                      </CardDescription>
                    )}
                    <div className="text-xs text-muted-foreground pt-2">
                      {project.slug}
                    </div>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        ) : (
          <Card className="text-center py-12">
            <CardContent>
              <FolderOpen className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
              <h3 className="text-lg font-medium mb-2">
                {hasFilters
                  ? 'No projects match your filters'
                  : 'No projects yet'}
              </h3>
              <p className="text-muted-foreground mb-6">
                {hasFilters
                  ? "Try adjusting your search or filters to find what you're looking for."
                  : 'Create your first project to start observing your AI applications'}
              </p>
              {!hasFilters && showCreateButton && (
                <Button onClick={() => setCreateDialogOpen(true)}>
                  <Plus className="mr-2 h-4 w-4" />
                  Create Project
                </Button>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </>
  )
}
