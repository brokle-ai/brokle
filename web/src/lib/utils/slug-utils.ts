/**
 * Slug utilities for composite slugs with embedded IDs
 * Enables cross-organization access with user-friendly URLs
 */

/**
 * Generate a composite slug from name and ID
 * @param name - Human readable name (e.g., "Brokle Technologies")
 * @param id - ULID identifier (e.g., "01K4MZR3ZEXW0QE66DF8DKBEZ3")
 * @returns Composite slug (e.g., "brokle-technologies-01k4mzr3zexw0qe66df8dkbez3")
 * @throws Error if name is empty or ID is invalid
 */
export function generateCompositeSlug(name: string, id: string): string {
  // Validate inputs
  if (!name || !name.trim()) {
    throw new Error('Organization/Project name cannot be empty')
  }

  if (!id || id.length !== 26) {
    throw new Error(`Invalid ULID format: expected 26 characters, got ${id?.length || 0}`)
  }

  // Convert name to slug format
  let nameSlug = name
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '') // Remove special characters
    .replace(/[\s_-]+/g, '-') // Replace spaces and underscores with hyphens
    .replace(/^-+|-+$/g, '') // Remove leading/trailing hyphens

  // Handle edge case: all characters removed (special chars only)
  if (!nameSlug) {
    nameSlug = 'org' // Fallback for names with only special characters
  }

  // Truncate to reasonable length to avoid URL length limits
  if (nameSlug.length > 50) {
    nameSlug = nameSlug.substring(0, 50).replace(/-+$/, '')
  }

  // Convert ID to lowercase for URL
  const lowercaseId = id.toLowerCase()

  return `${nameSlug}-${lowercaseId}`
}

/**
 * Extract the original ID from a composite slug
 * @param compositeSlug - Composite slug (e.g., "brokle-technologies-01k4mzr3zexw0qe66df8dkbez3")
 * @returns Original ULID (e.g., "01K4MZR3ZEXW0QE66DF8DKBEZ3")
 */
export function extractIdFromCompositeSlug(compositeSlug: string): string {
  // ULID format: 26 characters, alphanumeric
  // Extract last 26 characters and convert back to uppercase
  const urlId = compositeSlug.slice(-26)
  
  if (urlId.length !== 26) {
    throw new Error(`Invalid composite slug format: ${compositeSlug}`)
  }

  return urlId.toUpperCase()
}

/**
 * Extract the name slug portion from a composite slug
 * @param compositeSlug - Composite slug (e.g., "brokle-technologies-01k4mzr3zexw0qe66df8dkbez3")
 * @returns Name slug portion (e.g., "brokle-technologies")
 */
export function extractNameSlugFromCompositeSlug(compositeSlug: string): string {
  // Remove the last 27 characters (26 for ID + 1 for hyphen)
  return compositeSlug.slice(0, -27)
}

/**
 * Validate if a string looks like a composite slug
 * @param slug - String to validate
 * @returns True if it appears to be a valid composite slug
 */
export function isValidCompositeSlug(slug: string): boolean {
  // Should end with hyphen + 26 character ULID
  const pattern = /^.+-[0-9A-Za-z]{26}$/
  return pattern.test(slug)
}

/**
 * Check if a slug is a legacy slug (no embedded ID)
 * @param slug - String to check
 * @returns True if it's a legacy slug format
 */
export function isLegacySlug(slug: string): boolean {
  return !isValidCompositeSlug(slug)
}

/**
 * Build URL for organization with composite slug
 * @param name - Organization name
 * @param id - Organization ID  
 * @param path - Optional sub-path
 * @returns Organization URL (e.g., "/organizations/brokle-tech-01k4mzr3zexw0qe66df8dkbez3")
 */
export function buildOrgUrl(name: string, id: string, path: string = ''): string {
  const compositeSlug = generateCompositeSlug(name, id)
  const basePath = `/organizations/${compositeSlug}`
  
  if (!path || path === '/') return basePath
  
  // Ensure path starts with /
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${basePath}${normalizedPath}`
}

/**
 * Build URL for project with composite slug
 * @param name - Project name
 * @param id - Project ID
 * @param path - Optional sub-path
 * @returns Project URL (e.g., "/projects/analytics-platform-01k4mzr4f36gmrf5r21v5fkxvj")
 */
export function buildProjectUrl(name: string, id: string, path: string = ''): string {
  const compositeSlug = generateCompositeSlug(name, id)
  const basePath = `/projects/${compositeSlug}`

  if (!path || path === '/') return basePath

  // Ensure path starts with /
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${basePath}${normalizedPath}`
}

/**
 * Parse pathname to extract organization and project composite slugs
 * @param pathname - URL pathname (e.g., "/organizations/acme-corp-01jcxyz/projects/my-project-01abc")
 * @returns Object with orgSlug and projectSlug (null if not found)
 *
 * @example
 * parsePathContext('/organizations/acme-corp-01jcxyz123abc')
 * // Returns: { orgSlug: 'acme-corp-01jcxyz123abc', projectSlug: null }
 *
 * parsePathContext('/organizations/acme-01jcxyz/projects/api-01abc/settings')
 * // Returns: { orgSlug: 'acme-01jcxyz', projectSlug: 'api-01abc' }
 */
export function parsePathContext(pathname: string): {
  orgSlug: string | null
  projectSlug: string | null
} {
  const segments = pathname.split('/').filter(Boolean)

  // Pattern: /organizations/[compositeSlug]
  const orgIndex = segments.indexOf('organizations')
  const orgSlug = (orgIndex !== -1 && segments[orgIndex + 1])
    ? segments[orgIndex + 1]
    : null

  // Pattern: /projects/[compositeSlug]
  const projectIndex = segments.indexOf('projects')
  const projectSlug = (projectIndex !== -1 && segments[projectIndex + 1])
    ? segments[projectIndex + 1]
    : null

  return { orgSlug, projectSlug }
}

/**
 * Get or generate slug for an organization
 * @param org - Organization object
 * @returns Composite slug (uses org.slug if available, otherwise generates from name + id)
 */
export function getOrgSlug(org: { name: string; id: string; slug?: string }): string {
  return org.slug || generateCompositeSlug(org.name, org.id)
}

/**
 * Get or generate slug for a project
 * @param project - Project object
 * @returns Composite slug (uses project.slug if available, otherwise generates from name + id)
 */
export function getProjectSlug(project: { name: string; id: string; slug?: string }): string {
  return project.slug || generateCompositeSlug(project.name, project.id)
}

