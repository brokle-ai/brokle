'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useParams, usePathname } from 'next/navigation'
import { Settings, Key, Puzzle, Shield, AlertTriangle, Bot } from 'lucide-react'
import { cn } from '@/lib/utils'
import { buttonVariants } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

type ProjectSettingsNavProps = React.HTMLAttributes<HTMLElement>

export function ProjectSettingsNav({ className, ...props }: ProjectSettingsNavProps) {
  const params = useParams()
  const pathname = usePathname()
  const projectSlug = params?.projectSlug as string

  // Build navigation items with dynamic projectSlug
  const sidebarNavItems = [
    {
      title: 'General',
      href: `/projects/${projectSlug}/settings`,
      icon: <Settings size={18} />,
    },
    {
      title: 'API Keys',
      href: `/projects/${projectSlug}/settings/api-keys`,
      icon: <Key size={18} />,
    },
    {
      title: 'AI Providers',
      href: `/projects/${projectSlug}/settings/ai-providers`,
      icon: <Bot size={18} />,
    },
    {
      title: 'Integrations',
      href: `/projects/${projectSlug}/settings/integrations`,
      icon: <Puzzle size={18} />,
    },
    {
      title: 'Security',
      href: `/projects/${projectSlug}/settings/security`,
      icon: <Shield size={18} />,
    },
    {
      title: 'Danger Zone',
      href: `/projects/${projectSlug}/settings/danger`,
      icon: <AlertTriangle size={18} />,
    },
  ]

  const [val, setVal] = useState(pathname ?? sidebarNavItems[0].href)

  const handleSelect = (href: string) => {
    setVal(href)
    // Note: In a real app, you might want to programmatically navigate here
    // For now, we rely on the Link component for navigation
  }

  return (
    <>
      {/* Mobile Select */}
      <div className='p-1 md:hidden'>
        <Select value={val} onValueChange={handleSelect}>
          <SelectTrigger className='h-12 sm:w-48'>
            <SelectValue placeholder='Project Settings' />
          </SelectTrigger>
          <SelectContent>
            {sidebarNavItems.map((item) => (
              <SelectItem key={item.href} value={item.href}>
                <div className='flex gap-x-4 px-2 py-1'>
                  <span className='scale-125'>{item.icon}</span>
                  <span className='text-md'>{item.title}</span>
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Desktop Sidebar */}
      <ScrollArea
        orientation='horizontal'
        type='always'
        className='bg-background hidden w-full min-w-40 px-1 py-2 md:block'
      >
        <nav
          className={cn(
            'flex space-x-2 py-1 lg:flex-col lg:space-y-1 lg:space-x-0',
            className
          )}
          {...props}
        >
          {sidebarNavItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                buttonVariants({ variant: 'ghost' }),
                pathname === item.href
                  ? 'bg-muted hover:bg-accent'
                  : 'hover:bg-accent hover:underline',
                'justify-start'
              )}
            >
              <span className='me-2'>{item.icon}</span>
              {item.title}
            </Link>
          ))}
        </nav>
      </ScrollArea>
    </>
  )
}
