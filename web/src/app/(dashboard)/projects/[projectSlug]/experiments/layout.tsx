import type { ReactNode } from 'react'

/**
 * Experiments layout component.
 *
 * This file creates a route segment for the experiments page, which enables
 * nuqs to properly manage URL search parameters with shallow routing
 * within this segment.
 */
export default function ExperimentsLayout({
  children,
}: {
  children: ReactNode
}) {
  return children
}
