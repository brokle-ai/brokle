/**
 * Project-only convenience hook
 * 
 * Provides a clean interface for components that only need project-related
 * state and actions, with minimal organization context.
 */

import { useOrganization } from '@/context/organization-context'
import type { 
  Organization, 
  Project, 
  CreateProjectData 
} from '@/types/organization'

export interface ProjectOnlyContext {
  // State
  currentProject: Project | null
  projects: Project[]
  organization: Organization | null
  isLoading: boolean
  error: string | null

  // Project Actions
  switchProject: (projectSlug: string) => Promise<void>
  createProject: (data: CreateProjectData) => Promise<Project>

  // Utilities
  getProjectsByOrg: (orgSlug: string) => Project[]

  // Computed Properties
  hasProject: boolean
  projectCount: number
  hasMultipleProjects: boolean
  projectsInCurrentOrg: Project[]
  currentProjectMetrics: Project['metrics'] | null
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
 *     currentProjectMetrics,
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
 *       <MetricsDashboard metrics={currentProjectMetrics} />
 *     </div>
 *   )
 * }
 * ```
 */
export function useProjectOnly(): ProjectOnlyContext {
  const context = useOrganization()

  return {
    // State (project-focused)
    currentProject: context.currentProject,
    projects: context.projects,
    organization: context.currentOrganization, // Minimal org context for project actions
    isLoading: context.isLoading,
    error: context.error,

    // Project Actions (no organization actions)
    switchProject: context.switchProject,
    createProject: context.createProject,

    // Utilities (project-focused)
    getProjectsByOrg: context.getProjectsByOrg,

    // Computed Properties
    hasProject: context.currentProject !== null,
    projectCount: context.projects.length,
    hasMultipleProjects: context.projects.length > 1,
    projectsInCurrentOrg: context.projects,
    currentProjectMetrics: context.currentProject?.metrics || null,
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
 *             onClick={() => selectProject(project.slug)}
 *             selected={selectedProjectSlug === project.slug}
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
    selectedProjectSlug: project.currentProject?.slug || null,
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
  const metrics = project.currentProjectMetrics

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
 *     environment,
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
    environment: current?.environment || 'development',
    status: current?.status || 'inactive',
    slug: current?.slug || '',
    
    // Computed properties
    isActive: current?.status === 'active',
    isDevelopment: current?.environment === 'development',
    isProduction: current?.environment === 'production',
    
    // Dates
    createdAt: current?.created_at,
    updatedAt: current?.updated_at,
    
    // Formatted dates
    age: current ? 
      new Date(current.created_at).toLocaleDateString() : 
      'Unknown',
    lastUpdated: current ? 
      new Date(current.updated_at).toLocaleDateString() : 
      'Unknown',
  }
}