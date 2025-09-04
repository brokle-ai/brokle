/**
 * Utilities for working with organization and project slugs
 */

/**
 * Generate a URL-friendly slug from a name
 */
export function generateSlug(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
    .replace(/^-+|-+$/g, '')
    .substring(0, 50) // Max length for slugs
}

/**
 * Validate if a slug is properly formatted
 */
export function isValidSlug(slug: string): boolean {
  if (!slug || typeof slug !== 'string') return false
  
  // Must be lowercase, alphanumeric with hyphens, 1-50 characters
  const slugRegex = /^[a-z0-9-]{1,50}$/
  
  // Cannot start or end with hyphen
  if (slug.startsWith('-') || slug.endsWith('-')) return false
  
  // Cannot have consecutive hyphens
  if (slug.includes('--')) return false
  
  return slugRegex.test(slug)
}

/**
 * Check if a slug is available (not taken by existing organizations/projects)
 */
export function isSlugAvailable(
  slug: string,
  existingSlugs: string[]
): boolean {
  return !existingSlugs.includes(slug)
}

/**
 * Generate a unique slug by appending numbers if needed
 */
export function generateUniqueSlug(
  baseName: string,
  existingSlugs: string[]
): string {
  const slug = generateSlug(baseName)
  
  if (isSlugAvailable(slug, existingSlugs)) {
    return slug
  }
  
  let counter = 1
  while (!isSlugAvailable(`${slug}-${counter}`, existingSlugs)) {
    counter++
  }
  
  return `${slug}-${counter}`
}

/**
 * Extract organization and project slugs from pathname
 */
export function parsePathContext(pathname: string): {
  orgSlug?: string
  projectSlug?: string
} {
  if (!pathname || pathname === '/') {
    return {}
  }
  
  // Remove leading slash and split
  const segments = pathname.replace(/^\//, '').split('/')
  
  // Handle different path patterns:
  // /orgSlug -> { orgSlug }
  // /orgSlug/projectSlug -> { orgSlug, projectSlug }
  // /orgSlug/projectSlug/... -> { orgSlug, projectSlug }
  // /orgSlug/settings/... -> { orgSlug } (settings is not a project)
  
  if (segments.length === 0 || segments[0] === '') {
    return {}
  }
  
  const orgSlug = segments[0]
  
  // Check if the first segment is a reserved slug (like 'dashboard', 'auth', etc.)
  if (isReservedSlug(orgSlug)) {
    return {}
  }
  
  // Check if second segment is a special path (not a project)
  const specialPaths = ['settings', 'projects', 'members', 'billing', 'api-keys']
  if (segments.length > 1 && !specialPaths.includes(segments[1])) {
    return { orgSlug, projectSlug: segments[1] }
  }
  
  return { orgSlug }
}

/**
 * Build URL for organization context
 */
export function buildOrgUrl(orgSlug: string, path: string = ''): string {
  const basePath = `/${orgSlug}`
  if (!path || path === '/') return basePath
  
  // Ensure path starts with /
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${basePath}${normalizedPath}`
}

/**
 * Build URL for project context
 */
export function buildProjectUrl(
  orgSlug: string,
  projectSlug: string,
  path: string = ''
): string {
  const basePath = `/${orgSlug}/${projectSlug}`
  if (!path || path === '/') return basePath
  
  // Ensure path starts with /
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${basePath}${normalizedPath}`
}

/**
 * Reserved slugs that cannot be used for organizations or projects
 */
export const RESERVED_SLUGS = [
  'api',
  'auth',
  'admin',
  'www',
  'mail',
  'ftp',
  'localhost',
  'help',
  'support',
  'docs',
  'blog',
  'about',
  'contact',
  'pricing',
  'features',
  'security',
  'privacy',
  'terms',
  'onboarding',
  'settings',
  'dashboard',
  'app',
  'web',
  'mobile',
  'static',
  'assets',
  'cdn',
  'img',
  'images',
  'css',
  'js',
  'javascript',
  'fonts',
]

/**
 * Check if a slug is reserved
 */
export function isReservedSlug(slug: string): boolean {
  return RESERVED_SLUGS.includes(slug.toLowerCase())
}