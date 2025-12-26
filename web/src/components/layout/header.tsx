'use client'

import React from 'react'
import { cn } from '@/lib/utils'
import { Separator } from '@/components/ui/separator'
import { SidebarTrigger } from '@/components/ui/sidebar'

interface HeaderProps extends React.HTMLAttributes<HTMLElement> {
  fixed?: boolean
  ref?: React.Ref<HTMLElement>
}

export const Header = ({
  className,
  fixed,
  children,
  ...props
}: HeaderProps) => {
  const [offset, setOffset] = React.useState(0)

  React.useEffect(() => {
    const onScroll = () => {
      setOffset(document.body.scrollTop || document.documentElement.scrollTop)
    }

    // Add scroll listener to the body
    document.addEventListener('scroll', onScroll, { passive: true })

    // Clean up the event listener on unmount
    return () => document.removeEventListener('scroll', onScroll)
  }, [])

  return (
    <header
      className={cn(
        'bg-background relative h-12 border-b',
        fixed && 'header-fixed peer/header sticky top-0 z-50 w-[inherit]',
        offset > 10 && fixed && 'after:absolute after:inset-0 after:bg-background/20 after:backdrop-blur-lg after:-z-10',
        offset > 10 && fixed ? 'shadow-sm' : 'shadow-none',
        className
      )}
      {...props}
    >
      <div className='relative flex h-full items-center gap-2 px-3 py-2 sm:gap-3 sm:px-4'>
        <SidebarTrigger variant='outline' className='scale-125 sm:scale-100' />
        <Separator orientation='vertical' className='h-6' />
        {children}
      </div>
    </header>
  )
}

Header.displayName = 'Header'
