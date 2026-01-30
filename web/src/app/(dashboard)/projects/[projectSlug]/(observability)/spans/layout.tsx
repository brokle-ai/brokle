import type { ReactNode } from 'react'

/**
 * Spans layout component.
 *
 * This file creates a route segment for the spans page, which enables
 * nuqs to properly manage URL search parameters with shallow routing
 * within this segment.
 */
export default function SpansLayout({
  children,
}: {
  children: ReactNode
}) {
  return children
}
