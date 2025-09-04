/**
 * Organization-only convenience hook
 * 
 * Provides a clean interface for components that only need organization-related
 * state and actions, without the complexity of project management.
 */

import { useOrganization } from '@/context/organization-context'
import type { Organization, CreateOrganizationData, OrganizationRole } from '@/types/organization'

export interface OrganizationOnlyContext {
  // State
  organizations: Organization[]
  currentOrganization: Organization | null
  isLoading: boolean
  error: string | null

  // Organization Actions
  switchOrganization: (orgSlug: string) => Promise<void>
  createOrganization: (data: CreateOrganizationData) => Promise<Organization>

  // Utilities
  hasAccess: (orgSlug: string) => boolean
  getUserRole: (orgSlug: string) => OrganizationRole | null

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
  const context = useOrganization()

  return {
    // State (organization-focused)
    organizations: context.organizations,
    currentOrganization: context.currentOrganization,
    isLoading: context.isLoading,
    error: context.error,

    // Organization Actions (no project actions)
    switchOrganization: context.switchOrganization,
    createOrganization: context.createOrganization,

    // Utilities (organization-focused)
    hasAccess: (orgSlug: string) => context.hasAccess(orgSlug),
    getUserRole: context.getUserRole,

    // Computed Properties
    hasOrganization: context.currentOrganization !== null,
    organizationCount: context.organizations.length,
    isOwner: context.currentOrganization 
      ? context.getUserRole(context.currentOrganization.slug) === 'owner'
      : false,
    isAdmin: context.currentOrganization
      ? ['owner', 'admin'].includes(context.getUserRole(context.currentOrganization.slug) || '')
      : false,
    canCreateOrganization: true, // User can always create organizations
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
    getMemberCount: () => org.currentOrganization?.members.length || 0,
    getPlan: () => org.currentOrganization?.plan || 'free',
    getBillingEmail: () => org.currentOrganization?.billing_email,
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
    selectedOrgSlug: org.currentOrganization?.slug || null,
    selectOrganization: org.switchOrganization,
    hasMultipleOptions: org.organizationCount > 1,
    isLoading: org.isLoading,
    hasSelection: org.hasOrganization,
  }
}