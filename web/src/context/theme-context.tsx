'use client'

import { ThemeProvider as NextThemesProvider } from 'next-themes'
import { useTheme as useNextTheme } from 'next-themes'
import { useUIStore } from '@/stores/ui-store'
import { useEffect } from 'react'

type Theme = 'dark' | 'light' | 'system'

type ThemeProviderProps = {
  children: React.ReactNode
  defaultTheme?: Theme
  storageKey?: string
}

export function ThemeProvider({
  children,
  defaultTheme = 'system',
  storageKey = 'brokle-ui-theme',
}: ThemeProviderProps) {
  return (
    <NextThemesProvider
      attribute="class"
      defaultTheme={defaultTheme}
      enableSystem
      storageKey={storageKey}
      disableTransitionOnChange
    >
      <ThemeSync>
        {children}
      </ThemeSync>
    </NextThemesProvider>
  )
}

// Component to sync next-themes with Zustand store
function ThemeSync({ children }: { children: React.ReactNode }) {
  const { theme } = useNextTheme()
  const { setTheme } = useUIStore()

  useEffect(() => {
    if (theme) {
      setTheme(theme as Theme)
    }
  }, [theme, setTheme])

  return <>{children}</>
}

// Hook that uses next-themes but maintains compatibility
export const useTheme = () => {
  const nextTheme = useNextTheme()
  const { setTheme: setZustandTheme } = useUIStore()
  
  const setTheme = (theme: Theme) => {
    nextTheme.setTheme(theme)
    setZustandTheme(theme)
  }

  const resetTheme = () => {
    setTheme('system')
  }

  return {
    theme: nextTheme.theme as Theme,
    setTheme,
    resetTheme,
    systemTheme: nextTheme.systemTheme,
    resolvedTheme: nextTheme.resolvedTheme,
  }
}

// Export alias for compatibility
export const ThemeContextProvider = ThemeProvider
