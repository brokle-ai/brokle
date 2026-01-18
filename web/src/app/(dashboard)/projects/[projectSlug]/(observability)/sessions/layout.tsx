import type { ReactNode } from 'react'

/**
 * Sessions layout component.
 *
 * This file creates a route segment for the sessions page, which enables
 * nuqs to properly manage URL search parameters with shallow routing
 * within this segment.
 */
export default function SessionsLayout({
  children,
}: {
  children: ReactNode
}) {
  return children
}
