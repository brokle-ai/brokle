import { Metadata } from 'next'
import { AppearanceForm } from '@/components/settings/appearance-form'
import { ContentSection } from '@/components/settings/content-section'

export const metadata: Metadata = {
  title: 'Appearance',
  description: 'Customize the appearance of the dashboard.',
}

export default function AppearancePage() {
  return (
    <ContentSection
      title="Appearance"
      description="Customize the appearance of the dashboard. Automatically switch between day and night themes."
    >
      <AppearanceForm />
    </ContentSection>
  )
}