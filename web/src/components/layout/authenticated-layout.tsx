'use client'

import { SearchProvider } from '@/context/search-context'
import SkipToMain from '@/components/skip-to-main'
import type { User } from '@/features/authentication'

interface Props {
  children: React.ReactNode
  serverUser?: User | null
}

export function AuthenticatedLayout({
  children,
  serverUser: _serverUser // eslint-disable-line @typescript-eslint/no-unused-vars
}: Props) {
  return (
    <SearchProvider>
      <SkipToMain />
      {children}
    </SearchProvider>
  )
}
