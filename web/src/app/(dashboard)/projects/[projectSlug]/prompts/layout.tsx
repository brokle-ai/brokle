import type { ReactNode } from 'react'

/**
 * Prompts layout component.
 *
 * This file creates a route segment for the prompts page, which enables
 * nuqs to properly manage URL search parameters with shallow routing
 * within this segment.
 */
export default function PromptsLayout({
  children,
}: {
  children: ReactNode
}) {
  return children
}
