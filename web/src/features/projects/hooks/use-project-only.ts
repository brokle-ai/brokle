'use client'

/**
 * Project-only convenience hook
 * 
 * Provides a clean interface for components that only need project-related
 * state and actions, with minimal organization context.
 */

import { useWorkspace } from '@/context/workspace-context'
import type { WorkspaceError } from '@/context/workspace-errors'
import { useRouter } from 'next/navigation'
import type { OrganizationWithProjects, ProjectSummary } from '@/features/authentication'

export interface ProjectOnlyContext {
  // State
  currentProject: ProjectSummary | null
  projects: ProjectSummary[]
  organization: OrganizationWithProjects | null
  isLoading: boolean
  error: WorkspaceError | null

  // Project Actions
  switchProject: (projectSlug: string) => void

  // Computed Properties
  hasProject: boolean
  projectCount: number
  hasMultipleProjects: boolean
  projectsInCurrentOrg: ProjectSummary[]
}

/**
 * Hook that provides only project-related functionality
 * 
 * Perfect for:
 * - Project switchers
 * - Project dashboards
 * - Project settings pages
 * - Components focused on current project data
 * 
 * @example
 * ```tsx
 * function ProjectDashboard() {
 *   const {
 *     currentProject,
 *     hasProject,
 *     isLoading
 *   } = useProjectOnly()
 *
 *   if (isLoading) return <LoadingSpinner />
 *   if (!hasProject) return <NoProjectSelected />
 *
 *   return (
 *     <div>
 *       <h1>{currentProject.name}</h1>
 *       <MetricsDashboard metrics={currentProject.metrics} />
 *     </div>
 *   )
 * }
 * ```
 */
export function useProjectOnly(): ProjectOnlyContext {
  const workspace = useWorkspace()
  const router = useRouter()

  // Projects from current organization
  const projects = workspace.currentOrganization?.projects || []

  return {
    // State (project-focused)
    currentProject: workspace.currentProject,
    projects: projects,
    organization: workspace.currentOrganization,
    isLoading: workspace.isLoading,
    error: workspace.error,

    // Project Actions
    switchProject: (compositeSlug: string) => {
      router.push(`/projects/${compositeSlug}`)
    },

    // Computed Properties
    hasProject: workspace.currentProject !== null,
    projectCount: projects.length,
    hasMultipleProjects: projects.length > 1,
    projectsInCurrentOrg: projects,
  }
}

/**
 * Hook for project selection flows
 * 
 * Optimized for project selection components like dropdowns and modals
 * 
 * @example
 * ```tsx
 * function ProjectSelector() {
 *   const { 
 *     availableProjects, 
 *     selectedProjectSlug, 
 *     selectProject,
 *     hasMultipleOptions,
 *     canCreateProject
 *   } = useProjectSelector()
 *   
 *   return (
 *     <div>
 *       <Dropdown>
 *         {availableProjects.map(project => (
 *           <DropdownItem
 *             key={project.id}
 *             onClick={() => selectProject(project.compositeSlug)}
 *             selected={selectedProjectSlug === project.compositeSlug}
 *           >
 *             {project.name}
 *           </DropdownItem>
 *         ))}
 *       </Dropdown>
 *       
 *       {canCreateProject && (
 *         <CreateProjectButton />
 *       )}
 *     </div>
 *   )
 * }
 * ```
 */
export function useProjectSelector() {
  const project = useProjectOnly()

  return {
    availableProjects: project.projects,
    selectedProjectSlug: project.currentProject?.compositeSlug || null,
    selectProject: project.switchProject,
    hasMultipleOptions: project.hasMultipleProjects,
    isLoading: project.isLoading,
    hasSelection: project.hasProject,
    canCreateProject: project.organization !== null,
    organizationName: project.organization?.name || 'Unknown',
  }
}

/**
 * Hook for project metrics and monitoring
 * 
 * Provides utilities specifically for project metrics and performance monitoring
 * 
 * @example
 * ```tsx
 * function ProjectMetrics() {
 *   const { 
 *     metrics, 
 *     hasMetrics,
 *     requestsToday,
 *     costToday,
 *     averageLatency,
 *     errorRate,
 *     isHealthy
 *   } = useProjectMetrics()
 *   
 *   return (
 *     <MetricsGrid>
 *       <MetricCard 
 *         title="Requests Today" 
 *         value={requestsToday} 
 *       />
 *       <MetricCard 
 *         title="Cost Today" 
 *         value={`$${costToday}`} 
 *       />
 *       <MetricCard 
 *         title="Avg Latency" 
 *         value={`${averageLatency}ms`} 
 *       />
 *       <MetricCard 
 *         title="Error Rate" 
 *         value={`${errorRate}%`} 
 *         status={isHealthy ? 'good' : 'warning'}
 *       />
 *     </MetricsGrid>
 *   )
 * }
 * ```
 */
export function useProjectMetrics() {
  const project = useProjectOnly()
  const metrics = project.currentProject?.metrics || null

  return {
    // Raw metrics
    metrics,
    hasMetrics: metrics !== null,

    // Individual metric values
    requestsToday: metrics?.requests_today || 0,
    costToday: metrics?.cost_today || 0,
    averageLatency: metrics?.avg_latency || 0,
    errorRate: metrics?.error_rate || 0,

    // Computed health indicators
    isHealthy: metrics ? metrics.error_rate < 5 : true, // < 5% error rate is healthy
    hasActivity: metrics ? metrics.requests_today > 0 : false,
    
    // Formatted values for display
    formattedCost: metrics ? `$${metrics.cost_today.toFixed(2)}` : '$0.00',
    formattedLatency: metrics ? `${metrics.avg_latency.toFixed(0)}ms` : '0ms',
    formattedErrorRate: metrics ? `${metrics.error_rate.toFixed(1)}%` : '0.0%',
  }
}

/**
 * Hook for project information and metadata
 * 
 * Provides utilities for accessing project information without metrics
 * 
 * @example
 * ```tsx
 * function ProjectInfo() {
 *   const {
 *     name,
 *     description,
 *     status,
 *     isActive,
 *     createdAt,
 *     age
 *   } = useProjectInfo()
 *
 *   return (
 *     <ProjectCard>
 *       <h3>{name}</h3>
 *       <p>{description}</p>
 *       <Badge variant={isActive ? 'success' : 'secondary'}>
 *         {status}
 *       </Badge>
 *       <small>Created {age}</small>
 *     </ProjectCard>
 *   )
 * }
 * ```
 */
export function useProjectInfo() {
  const project = useProjectOnly()
  const current = project.currentProject

  return {
    // Basic info
    name: current?.name || '',
    description: current?.description || '',
    status: current?.status || 'inactive',
    compositeSlug: current?.compositeSlug || '',

    // Computed properties
    isActive: current?.status === 'active',

    // Dates
    createdAt: current?.createdAt,
    updatedAt: current?.updatedAt,
    
    // Formatted dates
    age: current ?
      new Date(current.createdAt).toLocaleDateString() :
      'Unknown',
    lastUpdated: current ?
      new Date(current.updatedAt).toLocaleDateString() :
      'Unknown',
  }
}