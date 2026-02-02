'use client'

import { DisplayForm, ContentSection } from '@/features/settings'

export default function DisplayPage() {
  return (
    <ContentSection
      title="Display"
      description="Turn items on or off to control what's displayed in the app."
    >
      <DisplayForm />
    </ContentSection>
  )
}
