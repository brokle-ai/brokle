'use client'

import { SearchProvider } from '@/context/search-context'
import { OrgProvider } from '@/context/org-context'
import SkipToMain from '@/components/skip-to-main'
import type { User } from '@/types/auth'

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
      <OrgProvider>
        <SkipToMain />
        {children}
      </OrgProvider>
    </SearchProvider>
  )
}
