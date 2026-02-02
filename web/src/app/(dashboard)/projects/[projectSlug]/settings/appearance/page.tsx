'use client'

import { AppearanceForm, ContentSection } from '@/features/settings'

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
