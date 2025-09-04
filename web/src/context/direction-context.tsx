'use client'

import { createContext, useContext, useEffect, useState } from 'react'
import { DirectionProvider as RadixDirectionProvider } from '@radix-ui/react-direction'

export type Direction = 'ltr' | 'rtl'

// Storage key following Next.js conventions
const DIRECTION_STORAGE_KEY = 'brokle-direction-preference'

// Default value to prevent hydration mismatches
const DEFAULT_DIRECTION: Direction = 'ltr'

type DirectionContextType = {
  defaultDir: Direction
  dir: Direction
  setDir: (dir: Direction) => void
  resetDir: () => void
}

const DirectionContext = createContext<DirectionContextType | null>(null)

type DirectionProviderProps = {
  children: React.ReactNode
}

export function DirectionProvider({ children }: DirectionProviderProps) {
  // SSR-safe state initialization
  const [dir, setDirState] = useState<Direction>(DEFAULT_DIRECTION)
  const [mounted, setMounted] = useState(false)

  // Hydration effect
  useEffect(() => {
    setMounted(true)
    
    // Load saved preference after hydration
    try {
      const savedDir = localStorage.getItem(DIRECTION_STORAGE_KEY) as Direction
      if (savedDir && ['ltr', 'rtl'].includes(savedDir)) {
        setDirState(savedDir)
      }
    } catch (error) {
      // Handle localStorage errors gracefully
      console.warn('Failed to load direction preference:', error)
    }
  }, [])

  // Apply direction to document element (client-side only)
  useEffect(() => {
    if (!mounted) return
    
    const htmlElement = document.documentElement
    htmlElement.setAttribute('dir', dir)
    
    // Cleanup function to reset if component unmounts
    return () => {
      htmlElement.setAttribute('dir', DEFAULT_DIRECTION)
    }
  }, [dir, mounted])

  const setDir = (newDir: Direction) => {
    setDirState(newDir)
    
    if (mounted) {
      try {
        localStorage.setItem(DIRECTION_STORAGE_KEY, newDir)
      } catch (error) {
        console.warn('Failed to save direction preference:', error)
      }
    }
  }

  const resetDir = () => {
    setDir(DEFAULT_DIRECTION)
    
    if (mounted) {
      try {
        localStorage.removeItem(DIRECTION_STORAGE_KEY)
      } catch (error) {
        console.warn('Failed to reset direction preference:', error)
      }
    }
  }

  const contextValue: DirectionContextType = {
    defaultDir: DEFAULT_DIRECTION,
    dir,
    setDir,
    resetDir,
  }

  return (
    <DirectionContext.Provider value={contextValue}>
      <RadixDirectionProvider dir={dir}>
        {children}
      </RadixDirectionProvider>
    </DirectionContext.Provider>
  )
}

export const useDirection = () => {
  const context = useContext(DirectionContext)
  if (!context) {
    throw new Error('useDirection must be used within a DirectionProvider')
  }
  return context
}