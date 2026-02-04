/**
 * Project persistence utilities for localStorage-based last visited project tracking.
 * Enables smart redirect to last active project on login.
 */

const STORAGE_KEY = 'brokle_last_project_slug'

/**
 * Get the last visited project slug from localStorage
 * @returns The last visited project composite slug, or null if not set/available
 */
export function getLastProjectSlug(): string | null {
  if (typeof window === 'undefined') return null

  try {
    return localStorage.getItem(STORAGE_KEY)
  } catch {
    // localStorage might be disabled or quota exceeded
    return null
  }
}

/**
 * Save the current project slug as the last visited project
 * @param slug - The project composite slug to save
 */
export function setLastProjectSlug(slug: string): void {
  if (typeof window === 'undefined') return

  try {
    localStorage.setItem(STORAGE_KEY, slug)
  } catch {
    // localStorage might be disabled or quota exceeded - fail silently
  }
}

/**
 * Clear the last visited project from localStorage
 * Call this when the project is deleted or user signs out
 */
export function clearLastProjectSlug(): void {
  if (typeof window === 'undefined') return

  try {
    localStorage.removeItem(STORAGE_KEY)
  } catch {
    // localStorage might be disabled - fail silently
  }
}
