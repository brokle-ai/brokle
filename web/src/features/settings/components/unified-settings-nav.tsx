'use client'

import { useState, useMemo } from 'react'
import Link from 'next/link'
import { useParams, usePathname, useRouter } from 'next/navigation'
import {
  Settings,
  Key,
  Target,
  AlertTriangle,
  Building2,
  Users,
  Bot,
  CreditCard,
  UserCog,
  Wrench,
  Palette,
  Bell,
  Monitor,
} from 'lucide-react'
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

type UnifiedSettingsNavProps = React.HTMLAttributes<HTMLElement>

interface NavItem {
  title: string
  href: string
  icon: React.ReactNode
}

interface NavSection {
  title: string
  items: NavItem[]
}

export function UnifiedSettingsNav({ className, ...props }: UnifiedSettingsNavProps) {
  const params = useParams()
  const pathname = usePathname()
  const router = useRouter()
  const projectSlug = params?.projectSlug as string

  // Build navigation sections with dynamic projectSlug
  const sections: NavSection[] = useMemo(() => [
    {
      title: 'PROJECT',
      items: [
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
          title: 'Score Configs',
          href: `/projects/${projectSlug}/settings/score-configs`,
          icon: <Target size={18} />,
        },
        {
          title: 'Danger Zone',
          href: `/projects/${projectSlug}/settings/danger`,
          icon: <AlertTriangle size={18} />,
        },
      ],
    },
    {
      title: 'ORGANIZATION',
      items: [
        {
          title: 'General',
          href: `/projects/${projectSlug}/settings/organization`,
          icon: <Building2 size={18} />,
        },
        {
          title: 'Members',
          href: `/projects/${projectSlug}/settings/organization/members`,
          icon: <Users size={18} />,
        },
        {
          title: 'AI Providers',
          href: `/projects/${projectSlug}/settings/organization/ai-providers`,
          icon: <Bot size={18} />,
        },
        {
          title: 'Billing',
          href: `/projects/${projectSlug}/settings/organization/billing`,
          icon: <CreditCard size={18} />,
        },
        {
          title: 'Danger Zone',
          href: `/projects/${projectSlug}/settings/organization/danger`,
          icon: <AlertTriangle size={18} />,
        },
      ],
    },
    {
      title: 'ACCOUNT',
      items: [
        {
          title: 'Profile',
          href: `/projects/${projectSlug}/settings/profile`,
          icon: <UserCog size={18} />,
        },
        {
          title: 'Account',
          href: `/projects/${projectSlug}/settings/account`,
          icon: <Wrench size={18} />,
        },
        {
          title: 'Appearance',
          href: `/projects/${projectSlug}/settings/appearance`,
          icon: <Palette size={18} />,
        },
        {
          title: 'Notifications',
          href: `/projects/${projectSlug}/settings/notifications`,
          icon: <Bell size={18} />,
        },
        {
          title: 'Display',
          href: `/projects/${projectSlug}/settings/display`,
          icon: <Monitor size={18} />,
        },
      ],
    },
  ], [projectSlug])

  // Flatten all items for mobile select
  const allItems = useMemo(() =>
    sections.flatMap(section =>
      section.items.map(item => ({
        ...item,
        section: section.title,
      }))
    ),
    [sections]
  )

  const [val, setVal] = useState(pathname ?? sections[0].items[0].href)

  const handleSelect = (href: string) => {
    setVal(href)
    router.push(href)
  }

  return (
    <>
      {/* Mobile Select */}
      <div className='p-1 md:hidden'>
        <Select value={val} onValueChange={handleSelect}>
          <SelectTrigger className='h-12 sm:w-48'>
            <SelectValue placeholder='Settings' />
          </SelectTrigger>
          <SelectContent>
            {sections.map((section, sectionIndex) => (
              <div key={section.title}>
                {sectionIndex > 0 && <div className="h-px bg-border my-1" />}
                <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground">
                  {section.title}
                </div>
                {section.items.map((item) => (
                  <SelectItem key={item.href} value={item.href}>
                    <div className='flex gap-x-4 px-2 py-1'>
                      <span className='scale-125'>{item.icon}</span>
                      <span className='text-md'>{item.title}</span>
                    </div>
                  </SelectItem>
                ))}
              </div>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Desktop Sidebar */}
      <ScrollArea
        orientation='horizontal'
        type='always'
        className='bg-background hidden w-full min-w-44 px-1 py-2 md:block'
      >
        <nav
          className={cn(
            'flex flex-col space-y-4',
            className
          )}
          {...props}
        >
          {sections.map((section) => (
            <div key={section.title} className="space-y-1">
              <h4 className="px-3 text-xs font-semibold text-muted-foreground tracking-wider">
                {section.title}
              </h4>
              <div className="space-y-0.5">
                {section.items.map((item) => (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={cn(
                      buttonVariants({ variant: 'ghost', size: 'sm' }),
                      pathname === item.href
                        ? 'bg-muted hover:bg-accent'
                        : 'hover:bg-accent hover:underline',
                      'w-full justify-start'
                    )}
                  >
                    <span className='me-2'>{item.icon}</span>
                    {item.title}
                  </Link>
                ))}
              </div>
            </div>
          ))}
        </nav>
      </ScrollArea>
    </>
  )
}
