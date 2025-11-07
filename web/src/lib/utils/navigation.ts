/**
 * Active state detection for navigation items
 *
 * Uses EXACT matching only - only the current page is highlighted.
 * Parent routes are NOT highlighted when viewing child routes.
 *
 * Examples:
 * - /projects/123 matches /projects/123 (exact) ✅
 * - /projects/123/analytics matches /projects/123/analytics (exact) ✅
 * - /projects/123/analytics/reports does NOT match /projects/123/analytics ❌
 * - /projects/123/settings/api-keys does NOT match /projects/123/settings ❌
 * - /projects/123/settings/api-keys matches /projects/123/settings/api-keys (exact) ✅
 */
export function isNavItemActive(currentPath: string, routePath: string): boolean {
  // Exact match only - no prefix matching
  return currentPath === routePath
}
