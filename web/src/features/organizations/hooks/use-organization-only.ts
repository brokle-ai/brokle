/**
 * Organization-only convenience hook
 * 
 * Provides a clean interface for components that only need organization-related
 * state and actions, without the complexity of project management.
 */

import { useWorkspace } from '@/context/workspace-context'
import { useRouter } from 'next/navigation'
import type { OrganizationWithProjects } from '@/features/authentication'

export interface OrganizationOnlyContext {
  // State
  organizations: OrganizationWithProjects[]
  currentOrganization: OrganizationWithProjects | null
  isLoading: boolean
  error: string | null

  // Organization Actions
  switchOrganization: (compositeSlug: string) => void

  // Computed Properties
  hasOrganization: boolean
  organizationCount: number
  isOwner: boolean
  isAdmin: boolean
  canCreateOrganization: boolean
}

/**
 * Hook that provides only organization-related functionality
 * 
 * Perfect for:
 * - Organization switchers
 * - Organization settings pages  
 * - Organization creation forms
 * - Components that don't need project state
 * 
 * @example
 * ```tsx
 * function OrganizationSwitcher() {
 *   const { 
 *     organizations, 
 *     currentOrganization, 
 *     switchOrganization,
 *     isLoading 
 *   } = useOrganizationOnly()
 *   
 *   return (
 *     <Select 
 *       value={currentOrganization?.slug}
 *       onChange={switchOrganization}
 *       loading={isLoading}
 *     >
 *       {organizations.map(org => (
 *         <Option key={org.id} value={org.slug}>
 *           {org.name}
 *         </Option>
 *       ))}
 *     </Select>
 *   )
 * }
 * ```
 */
export function useOrganizationOnly(): OrganizationOnlyContext {
  const workspace = useWorkspace()
  const router = useRouter()

  return {
    // State (organization-focused)
    organizations: workspace.organizations,
    currentOrganization: workspace.currentOrganization,
    isLoading: workspace.isLoading,
    error: workspace.error,

    // Organization Actions
    switchOrganization: (compositeSlug: string) => {
      router.push(`/organizations/${compositeSlug}`)
    },

    // Computed Properties
    hasOrganization: workspace.currentOrganization !== null,
    organizationCount: workspace.organizations.length,
    isOwner: workspace.currentOrganization?.role === 'owner',
    isAdmin: workspace.currentOrganization?.role === 'owner' || workspace.currentOrganization?.role === 'admin',
    canCreateOrganization: true,
  }
}

/**
 * Hook for organization settings and management
 * 
 * Provides additional utilities specifically for organization management pages
 * 
 * @example
 * ```tsx
 * function OrganizationSettings() {
 *   const { 
 *     currentOrganization, 
 *     isOwner, 
 *     canManageSettings 
 *   } = useOrganizationManagement()
 *   
 *   if (!canManageSettings) {
 *     return <AccessDenied />
 *   }
 *   
 *   return <OrganizationSettingsForm org={currentOrganization} />
 * }
 * ```
 */
export function useOrganizationManagement() {
  const org = useOrganizationOnly()

  return {
    ...org,

    // Management-specific computed properties
    canManageSettings: org.isAdmin,
    canManageMembers: org.isAdmin,
    canManageBilling: org.isOwner,
    canDeleteOrganization: org.isOwner,

    // Management-specific utilities
    getPlan: () => org.currentOrganization?.plan || 'free',
  }
}

/**
 * Hook for organization selection flows
 * 
 * Optimized for organization selection components like dropdowns and modals
 * 
 * @example
 * ```tsx
 * function OrganizationSelector() {
 *   const { 
 *     availableOrganizations, 
 *     selectedOrgSlug, 
 *     selectOrganization,
 *     hasMultipleOptions
 *   } = useOrganizationSelector()
 *   
 *   if (!hasMultipleOptions) {
 *     return null // No need to show selector
 *   }
 *   
 *   return (
 *     <Dropdown>
 *       {availableOrganizations.map(org => (
 *         <DropdownItem 
 *           key={org.id}
 *           onClick={() => selectOrganization(org.slug)}
 *           selected={selectedOrgSlug === org.slug}
 *         >
 *           {org.name}
 *         </DropdownItem>
 *       ))}
 *     </Dropdown>
 *   )
 * }
 * ```
 */
export function useOrganizationSelector() {
  const org = useOrganizationOnly()

  return {
    availableOrganizations: org.organizations,
    selectedOrgSlug: org.currentOrganization?.compositeSlug || null,
    selectOrganization: org.switchOrganization,
    hasMultipleOptions: org.organizationCount > 1,
    isLoading: org.isLoading,
    hasSelection: org.hasOrganization,
  }
}