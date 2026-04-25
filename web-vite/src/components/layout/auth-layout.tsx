import { BrokleLogo } from '@/components/ui/brokle-logo'

interface AuthLayoutProps {
  children: React.ReactNode
}

// Centered-card chrome shared by every page under the (auth) route
// group in web/. Ported verbatim (minus the `'use client'` directive)
// so visual parity holds without pulling web/'s Next.js runtime.
export function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className="container grid h-svh max-w-none items-center justify-center">
      <div className="mx-auto flex w-full flex-col justify-center space-y-2 py-8 sm:w-[480px] sm:p-8">
        <div className="mb-4 flex items-center justify-center">
          <BrokleLogo variant="full" size="md" />
        </div>
        {children}
      </div>
    </div>
  )
}
