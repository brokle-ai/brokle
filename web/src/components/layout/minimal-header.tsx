import { ProfileDropdown } from '@/components/profile-dropdown'
import { BrokleLogo } from '@/components/ui/brokle-logo'
import Link from 'next/link'

export function MinimalHeader() {
  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-16 items-center justify-between px-6">
        <Link href="/" className="group">
          <BrokleLogo variant="full" size="sm" />
        </Link>

        <ProfileDropdown currentProject={null} />
      </div>
    </header>
  )
}
