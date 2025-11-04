'use client'

import { useState, useMemo } from 'react'
import { useRouter } from 'next/navigation'
import { 
  FolderOpen, 
  Plus, 
  Search, 
  Filter, 
  MoreVertical, 
  Play, 
  Pause, 
  Archive, 
  Settings,
  Trash2,
  Copy,
  ExternalLink,
  CheckSquare,
  Square
} from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { CreateProjectModal } from './create-project-modal'
import { BulkActionsBar } from './bulk-actions-bar'
import { buildProjectUrl } from '@/lib/utils/slug-utils'
import { cn } from '@/lib/utils'
import type { Project, ProjectStatus, ProjectEnvironment } from '@/types/organization'

interface ProjectGridProps {
  className?: string
  showCreateButton?: boolean
}

export function ProjectGrid({ className, showCreateButton = true }: ProjectGridProps) {
  const router = useRouter()
  const { currentOrganization, projects, isLoadingProjects } = useWorkspace()
  
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<ProjectStatus | 'all'>('all')
  const [environmentFilter, setEnvironmentFilter] = useState<ProjectEnvironment | 'all'>('all')
  const [sortBy, setSortBy] = useState<'name' | 'created' | 'requests' | 'cost'>('name')
  const [selectedProjects, setSelectedProjects] = useState<Project[]>([])
  const [bulkSelectMode, setBulkSelectMode] = useState(false)

  const filteredAndSortedProjects = useMemo(() => {
    // Ensure projects is always an array
    const safeProjects = projects || []
    const filtered = safeProjects.filter(project => {
      const matchesSearch = project.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           project.description?.toLowerCase().includes(searchTerm.toLowerCase())
      const matchesStatus = statusFilter === 'all' || project.status === statusFilter
      const matchesEnvironment = environmentFilter === 'all' || project.environment === environmentFilter
      
      return matchesSearch && matchesStatus && matchesEnvironment
    })

    // Sort projects
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name)
        case 'created':
          return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        case 'requests':
          return (b.metrics.total_requests || 0) - (a.metrics.total_requests || 0)
        case 'cost':
          return (b.metrics.total_cost || 0) - (a.metrics.total_cost || 0)
        default:
          return 0
      }
    })

    return filtered
  }, [projects, searchTerm, statusFilter, environmentFilter, sortBy])

  const getStatusColor = (status: ProjectStatus) => {
    switch (status) {
      case 'active':
        return 'bg-green-500'
      case 'inactive':
        return 'bg-yellow-500'
      case 'archived':
        return 'bg-gray-500'
      default:
        return 'bg-gray-500'
    }
  }

  const getEnvironmentColor = (environment: ProjectEnvironment) => {
    switch (environment) {
      case 'production':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
      case 'staging':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
      case 'development':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
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

  const handleProjectAction = (action: string, project: Project) => {
    // These would typically call API endpoints
    console.log(`Action: ${action} on project:`, project.name)
    // TODO: Implement project actions
  }

  const toggleProjectSelection = (project: Project) => {
    const isSelected = selectedProjects.some(p => p.id === project.id)
    if (isSelected) {
      setSelectedProjects(selectedProjects.filter(p => p.id !== project.id))
    } else {
      setSelectedProjects([...selectedProjects, project])
    }
  }

  const selectAllProjects = () => {
    if (selectedProjects.length === filteredAndSortedProjects.length) {
      setSelectedProjects([])
    } else {
      setSelectedProjects(filteredAndSortedProjects)
    }
  }

  const clearSelection = () => {
    setSelectedProjects([])
    setBulkSelectMode(false)
  }

  const handleProjectsUpdated = () => {
    // Refresh projects data would be handled here
    // For now we'll just clear selection
    clearSelection()
  }

  if (!currentOrganization) {
    return null
  }

  return (
    <div className={cn("space-y-6", className)}>
      {/* Bulk Actions Bar */}
      <BulkActionsBar 
        selectedProjects={selectedProjects}
        onClearSelection={clearSelection}
        onProjectsUpdated={handleProjectsUpdated}
      />

      {/* Header */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-foreground">Projects</h2>
          <p className="text-muted-foreground">
            Manage and monitor your AI projects in {currentOrganization.name}
          </p>
        </div>
        
        <div className="flex items-center gap-2">
          {filteredAndSortedProjects.length > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setBulkSelectMode(!bulkSelectMode)}
            >
              <CheckSquare className="mr-2 h-4 w-4" />
              {bulkSelectMode ? 'Exit Select' : 'Bulk Select'}
            </Button>
          )}
          
          {showCreateButton && (
            <CreateProjectModal 
              trigger={
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  New Project
                </Button>
              }
            />
          )}
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        {bulkSelectMode && filteredAndSortedProjects.length > 0 && (
          <Button
            variant="outline"
            size="sm"
            onClick={selectAllProjects}
            className="w-fit"
          >
            {selectedProjects.length === filteredAndSortedProjects.length ? (
              <CheckSquare className="mr-2 h-4 w-4" />
            ) : (
              <Square className="mr-2 h-4 w-4" />
            )}
            {selectedProjects.length === filteredAndSortedProjects.length 
              ? 'Deselect All' 
              : `Select All (${filteredAndSortedProjects.length})`}
          </Button>
        )}
        
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <Input
            placeholder="Search projects..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
        
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-full sm:w-40">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="inactive">Inactive</SelectItem>
            <SelectItem value="archived">Archived</SelectItem>
          </SelectContent>
        </Select>

        <Select value={environmentFilter} onValueChange={setEnvironmentFilter}>
          <SelectTrigger className="w-full sm:w-40">
            <SelectValue placeholder="Environment" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Environments</SelectItem>
            <SelectItem value="production">Production</SelectItem>
            <SelectItem value="staging">Staging</SelectItem>
            <SelectItem value="development">Development</SelectItem>
          </SelectContent>
        </Select>

        <Select value={sortBy} onValueChange={setSortBy}>
          <SelectTrigger className="w-full sm:w-40">
            <SelectValue placeholder="Sort by" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="name">Name</SelectItem>
            <SelectItem value="created">Created Date</SelectItem>
            <SelectItem value="requests">Requests</SelectItem>
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
              className={cn(
                "cursor-pointer hover:shadow-md transition-all duration-200 group",
                selectedProjects.some(p => p.id === project.id) && "ring-2 ring-primary",
                bulkSelectMode && "cursor-default"
              )}
              onClick={(e) => {
                if (bulkSelectMode) {
                  e.preventDefault()
                  toggleProjectSelection(project)
                } else {
                  handleProjectClick(project)
                }
              }}
            >
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3 flex-1 min-w-0">
                    {bulkSelectMode && (
                      <div 
                        className="flex items-center"
                        onClick={(e) => {
                          e.stopPropagation()
                          toggleProjectSelection(project)
                        }}
                      >
                        {selectedProjects.some(p => p.id === project.id) ? (
                          <CheckSquare className="h-5 w-5 text-primary" />
                        ) : (
                          <Square className="h-5 w-5 text-muted-foreground hover:text-primary transition-colors" />
                        )}
                      </div>
                    )}
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
                  
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                      <Button variant="ghost" size="sm">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuLabel>Actions</DropdownMenuLabel>
                      <DropdownMenuItem onClick={(e) => {
                        e.stopPropagation()
                        handleProjectClick(project)
                      }}>
                        <ExternalLink className="mr-2 h-4 w-4" />
                        Open Project
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={(e) => {
                        e.stopPropagation()
                        handleProjectAction('settings', project)
                      }}>
                        <Settings className="mr-2 h-4 w-4" />
                        Settings
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={(e) => {
                        e.stopPropagation()
                        handleProjectAction('duplicate', project)
                      }}>
                        <Copy className="mr-2 h-4 w-4" />
                        Duplicate
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      {project.status === 'active' ? (
                        <DropdownMenuItem onClick={(e) => {
                          e.stopPropagation()
                          handleProjectAction('pause', project)
                        }}>
                          <Pause className="mr-2 h-4 w-4" />
                          Pause Project
                        </DropdownMenuItem>
                      ) : (
                        <DropdownMenuItem onClick={(e) => {
                          e.stopPropagation()
                          handleProjectAction('activate', project)
                        }}>
                          <Play className="mr-2 h-4 w-4" />
                          Activate Project
                        </DropdownMenuItem>
                      )}
                      <DropdownMenuItem onClick={(e) => {
                        e.stopPropagation()
                        handleProjectAction('archive', project)
                      }}>
                        <Archive className="mr-2 h-4 w-4" />
                        Archive
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem 
                        className="text-destructive"
                        onClick={(e) => {
                          e.stopPropagation()
                          handleProjectAction('delete', project)
                        }}
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        Delete Project
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </CardHeader>
              
              <CardContent className="space-y-4">
                <div className="flex items-center gap-2">
                  <Badge className={cn("text-xs", getEnvironmentColor(project.environment))}>
                    {project.environment}
                  </Badge>
                  {project.updated_at && (
                    <span className="text-xs text-muted-foreground">
                      Updated {new Date(project.updated_at).toLocaleDateString()}
                    </span>
                  )}
                </div>

                {project.description && (
                  <CardDescription className="line-clamp-2">
                    {project.description}
                  </CardDescription>
                )}

                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <div className="font-medium text-foreground">
                      {formatNumber(project.metrics.requests_today)}
                    </div>
                    <div className="text-muted-foreground text-xs">
                      Requests today
                    </div>
                  </div>
                  <div>
                    <div className="font-medium text-foreground">
                      ${project.metrics.cost_today.toFixed(2)}
                    </div>
                    <div className="text-muted-foreground text-xs">
                      Cost today
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <div className="font-medium text-foreground">
                      {project.metrics.avg_latency}ms
                    </div>
                    <div className="text-muted-foreground text-xs">
                      Avg latency
                    </div>
                  </div>
                  <div>
                    <div className="font-medium text-foreground">
                      {(project.metrics.error_rate * 100).toFixed(2)}%
                    </div>
                    <div className="text-muted-foreground text-xs">
                      Error rate
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
              {searchTerm || statusFilter !== 'all' || environmentFilter !== 'all' 
                ? 'No projects match your filters' 
                : 'No projects yet'}
            </h3>
            <p className="text-muted-foreground mb-6">
              {searchTerm || statusFilter !== 'all' || environmentFilter !== 'all'
                ? 'Try adjusting your search or filters to find what you\'re looking for.'
                : 'Create your first project to start using the AI platform'}
            </p>
            {(!searchTerm && statusFilter === 'all' && environmentFilter === 'all') && showCreateButton && (
              <CreateProjectModal 
                trigger={
                  <Button>
                    <Plus className="mr-2 h-4 w-4" />
                    Create Project
                  </Button>
                }
              />
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}