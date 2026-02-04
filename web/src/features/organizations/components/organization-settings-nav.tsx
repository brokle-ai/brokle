'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useParams, usePathname } from 'next/navigation'
import { Settings, Users, CreditCard, Shield, Code, AlertTriangle, Bot } from 'lucide-react'
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

type OrganizationSettingsNavProps = React.HTMLAttributes<HTMLElement>

export function OrganizationSettingsNav({ className, ...props }: OrganizationSettingsNavProps) {
  const params = useParams()
  const pathname = usePathname()
  const orgSlug = params?.orgSlug as string

  // Build navigation items with dynamic orgSlug
  const sidebarNavItems = [
    {
      title: 'General',
      href: `/organizations/${orgSlug}/settings`,
      icon: <Settings size={18} />,
    },
    {
      title: 'Members',
      href: `/organizations/${orgSlug}/settings/members`,
      icon: <Users size={18} />,
    },
    {
      title: 'AI Providers',
      href: `/organizations/${orgSlug}/settings/ai-providers`,
      icon: <Bot size={18} />,
    },
    {
      title: 'Billing',
      href: `/organizations/${orgSlug}/settings/billing`,
      icon: <CreditCard size={18} />,
    },
    // {
    //   title: 'Security',
    //   href: `/organizations/${orgSlug}/settings/security`,
    //   icon: <Shield size={18} />,
    // },
    // {
    //   title: 'Advanced',
    //   href: `/organizations/${orgSlug}/settings/advanced`,
    //   icon: <Code size={18} />,
    // },
    {
      title: 'Danger Zone',
      href: `/organizations/${orgSlug}/settings/danger`,
      icon: <AlertTriangle size={18} />,
    },
  ]

  const [val, setVal] = useState(pathname ?? sidebarNavItems[0].href)

  const handleSelect = (href: string) => {
    setVal(href)
  }

  return (
    <>
      {/* Mobile Select */}
      <div className='p-1 md:hidden'>
        <Select value={val} onValueChange={handleSelect}>
          <SelectTrigger className='h-12 sm:w-48'>
            <SelectValue placeholder='Organization Settings' />
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
