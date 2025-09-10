/**
 * Slug utilities for composite slugs with embedded IDs
 * Enables cross-organization access with user-friendly URLs
 */

/**
 * Generate a composite slug from name and ID
 * @param name - Human readable name (e.g., "Brokle Technologies")
 * @param id - ULID identifier (e.g., "01K4MZR3ZEXW0QE66DF8DKBEZ3")
 * @returns Composite slug (e.g., "brokle-technologies-01k4mzr3zexw0qe66df8dkbez3")
 */
export function generateCompositeSlug(name: string, id: string): string {
  // Convert name to slug format
  const nameSlug = name
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '') // Remove special characters
    .replace(/[\s_-]+/g, '-') // Replace spaces and underscores with hyphens
    .replace(/^-+|-+$/g, '') // Remove leading/trailing hyphens

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


