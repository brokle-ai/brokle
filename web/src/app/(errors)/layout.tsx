import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Error | Brokle',
  description: 'An error occurred',
}

/**
 * Minimal layout for error pages
 * No authentication, no sidebar, just the error content
 */
export default function ErrorsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return <>{children}</>
}
