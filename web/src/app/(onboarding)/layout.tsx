import { OrgProvider } from '@/context/org-context'
import { MinimalHeader } from '@/components/layout/minimal-header'

export default function OnboardingLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <OrgProvider>
      <div className="min-h-screen bg-background">
        <MinimalHeader />
        <main>{children}</main>
      </div>
    </OrgProvider>
  )
}
