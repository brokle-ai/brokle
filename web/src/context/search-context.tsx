'use client'

import React from 'react'
import { useUIStore } from '@/stores/ui-store'
import { CommandMenu } from '@/components/command-menu'

interface Props {
  children: React.ReactNode
}

export function SearchProvider({ children }: Props) {
  const { setSearchOpen } = useUIStore()

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setSearchOpen(true)
      }
    }
    document.addEventListener('keydown', down)
    return () => document.removeEventListener('keydown', down)
  }, [setSearchOpen])

  return (
    <>
      {children}
      <CommandMenu />
    </>
  )
}

// Hook for backward compatibility and easier usage
export const useSearch = () => {
  const { searchOpen, setSearchOpen } = useUIStore()
  return { open: searchOpen, setOpen: setSearchOpen }
}
