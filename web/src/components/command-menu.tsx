'use client'

import React from 'react'
import { useRouter } from 'next/navigation'
import { ArrowRight, ChevronRight, Laptop, Moon, Sun } from 'lucide-react'
import { useSearch } from '@/context/search-context'
import { useTheme } from '@/context/theme-context'
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command'
import { ScrollArea } from './ui/scroll-area'
import { useNavigationContext } from '@/hooks/use-navigation-context'
import { processNavigation } from '@/lib/navigation/process-routes'
import { ROUTES } from '@/lib/navigation/routes'

/**
 * Safe wrapper to use navigation context
 * Returns null if WorkspaceProvider is not available (e.g., on auth pages)
 */
function useSafeNavigationContext() {
  try {
    return useNavigationContext()
  } catch {
    // Not in workspace context (auth pages, etc.)
    return null
  }
}

export function CommandMenu() {
  const router = useRouter()
  const { setTheme } = useTheme()
  const { open, setOpen } = useSearch()
  const navigationContext = useSafeNavigationContext()

  // Process navigation only if context available
  const flatNavigation = React.useMemo(() => {
    if (!navigationContext) return []

    const { flatNavigation } = processNavigation({
      routes: ROUTES,
      context: navigationContext.context,
      permissions: navigationContext.permissions,
      featureFlags: navigationContext.featureFlags,
      isPermissionsLoading: navigationContext.isPermissionsLoading,
    })

    return flatNavigation
  }, [navigationContext])

  const runCommand = React.useCallback(
    (command: () => unknown) => {
      setOpen(false)
      command()
    },
    [setOpen]
  )

  return (
    <CommandDialog modal open={open} onOpenChange={setOpen}>
      <CommandInput placeholder='Type a command or search...' />
      <CommandList>
        <ScrollArea type='hover' className='h-72 pe-1'>
          <CommandEmpty>No results found.</CommandEmpty>
          {flatNavigation.length > 0 && (
            <>
              <CommandGroup heading='Navigation'>
                {flatNavigation.map((route) => (
                  <CommandItem
                    key={route.pathname}
                    value={route.title}
                    onSelect={() => {
                      runCommand(() => router.push(route.url))
                    }}
                  >
                    <div className='flex size-4 items-center justify-center'>
                      {route.icon ? <route.icon className='size-4' /> : <ArrowRight className='text-muted-foreground/80 size-2' />}
                    </div>
                    {route.title}
                  </CommandItem>
                ))}
              </CommandGroup>
              <CommandSeparator />
            </>
          )}
          <CommandSeparator />
          <CommandGroup heading='Theme'>
            <CommandItem onSelect={() => runCommand(() => setTheme('light'))}>
              <Sun /> <span>Light</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => setTheme('dark'))}>
              <Moon className='scale-90' />
              <span>Dark</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => setTheme('system'))}>
              <Laptop />
              <span>System</span>
            </CommandItem>
          </CommandGroup>
        </ScrollArea>
      </CommandList>
    </CommandDialog>
  )
}