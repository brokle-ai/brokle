'use client'

import { useState, useMemo } from 'react'
import { useRouter } from 'next/navigation'
import {
  FolderOpen,
  Plus,
  Search,
} from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { useOrganizationProjects } from '../hooks/use-organization-projects'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { CreateProjectDialog } from '@/features/projects'
import { buildProjectUrl } from '@/lib/utils/slug-utils'
import { cn } from '@/lib/utils'
import { PageHeader } from '@/components/layout/page-header'
import type { Project, ProjectStatus } from '../types'

interface ProjectGridProps {
  className?: string
  showCreateButton?: boolean
}

export function ProjectGrid({ className, showCreateButton = true }: ProjectGridProps) {
  const router = useRouter()
  const { currentOrganization } = useWorkspace()
  const { data } = useOrganizationProjects()

  const projects = data ?? []

  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<ProjectStatus | 'all'>('all')
  const [sortBy, setSortBy] = useState<'name' | 'created' | 'traces' | 'cost'>('name')
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  const filteredAndSortedProjects = useMemo(() => {
    const safeProjects = projects || []
    const filtered = safeProjects.filter(project => {
      const matchesSearch = project.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           project.description?.toLowerCase().includes(searchTerm.toLowerCase())
      const matchesStatus = statusFilter === 'all' || project.status === statusFilter

      return matchesSearch && matchesStatus
    })

    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name)
        case 'created':
          return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        case 'traces':
          return (b.metrics.traces_collected || 0) - (a.metrics.traces_collected || 0)
        case 'cost':
          return (b.metrics.observed_cost || 0) - (a.metrics.observed_cost || 0)
        default:
          return 0
      }
    })

    return filtered
  }, [projects, searchTerm, statusFilter, sortBy])

  const getStatusColor = (status: ProjectStatus) => {
    switch (status) {
      case 'active':
        return 'bg-green-500'
      case 'archived':
        return 'bg-gray-500'
      default:
        return 'bg-gray-500'
    }
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
    return num.toString()
  }

  const handleProjectClick = (project: Project) => {
    const projectUrl = buildProjectUrl(project.name, project.id)
    router.push(projectUrl)
  }

  if (!currentOrganization) {
    return null
  }

  return (
    <>
      <PageHeader title="Projects">
        {showCreateButton && currentOrganization?.id && (
          <>
            <Button onClick={() => setCreateDialogOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              New Project
            </Button>

            <CreateProjectDialog
              organizationId={currentOrganization.id}
              open={createDialogOpen}
              onOpenChange={setCreateDialogOpen}
            />
          </>
        )}
      </PageHeader>

      <div className={cn("-mx-4 flex-1 overflow-auto px-4 py-1 space-y-6", className)}>
        {/* Filters */}
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
        
        <Select value={statusFilter} onValueChange={(value) => setStatusFilter(value as ProjectStatus | 'all')}>
          <SelectTrigger className="w-full sm:w-40">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="archived">Archived</SelectItem>
          </SelectContent>
        </Select>

        <Select value={sortBy} onValueChange={(value) => setSortBy(value as 'name' | 'created' | 'traces' | 'cost')}>
          <SelectTrigger className="w-full sm:w-40">
            <SelectValue placeholder="Sort by" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="name">Name</SelectItem>
            <SelectItem value="created">Created Date</SelectItem>
            <SelectItem value="traces">Traces</SelectItem>
            <SelectItem value="cost">Cost</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Projects Grid */}
      {filteredAndSortedProjects.length > 0 ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredAndSortedProjects.map((project) => (
            <Card
              key={project.id}
              className="cursor-pointer group"
              onClick={() => handleProjectClick(project)}
            >
              <CardHeader className="pb-3">
                <div className="flex items-center gap-3">
                  <div className="bg-muted flex size-10 items-center justify-center rounded-lg group-hover:bg-primary/10 transition-colors">
                    <FolderOpen className="size-5 group-hover:text-primary transition-colors" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <CardTitle className="text-lg truncate">{project.name}</CardTitle>
                    <div className="flex items-center gap-2 mt-1">
                      <div className={cn("w-2 h-2 rounded-full", getStatusColor(project.status))}></div>
                      <span className="text-sm capitalize text-muted-foreground">{project.status}</span>
                    </div>
                  </div>
                </div>
              </CardHeader>
              
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <div className="font-medium text-foreground">
                      {formatNumber(project.metrics.traces_collected)}
                    </div>
                    <div className="text-muted-foreground text-xs">
                      Traces collected
                    </div>
                  </div>
                  <div>
                    <div className="font-medium text-foreground">
                      ${(project.metrics.observed_cost || 0).toFixed(2)}
                    </div>
                    <div className="text-muted-foreground text-xs">
                      AI cost observed
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <div className="font-medium text-foreground">
                      {project.metrics.active_rules}
                    </div>
                    <div className="text-muted-foreground text-xs">
                      Active rules
                    </div>
                  </div>
                  <div>
                    <div className="font-medium text-foreground">
                      {project.metrics.running_experiments}
                    </div>
                    <div className="text-muted-foreground text-xs">
                      Experiments
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <Card className="text-center py-12">
          <CardContent>
            <FolderOpen className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium mb-2">
              {searchTerm || statusFilter !== 'all'
                ? 'No projects match your filters'
                : 'No projects yet'}
            </h3>
            <p className="text-muted-foreground mb-6">
              {searchTerm || statusFilter !== 'all'
                ? 'Try adjusting your search or filters to find what you\'re looking for.'
                : 'Create your first project to start observing your AI applications'}
            </p>
            {(!searchTerm && statusFilter === 'all') && showCreateButton && currentOrganization?.id && (
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