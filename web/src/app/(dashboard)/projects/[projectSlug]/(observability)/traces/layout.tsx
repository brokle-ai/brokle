import type { ReactNode } from 'react'

/**
 * Traces layout component.
 *
 * This file creates a route segment for the traces page, which enables
 * nuqs to properly manage URL search parameters with shallow routing
 * within this segment.
 */
export default function TracesLayout({
  children,
}: {
  children: ReactNode
}) {
  return children
}
