import { Metadata } from 'next'
import { DisplayForm } from '@/components/settings/display-form'
import { ContentSection } from '@/components/settings/content-section'

export const metadata: Metadata = {
  title: 'Display',
  description: 'Configure display and layout preferences.',
}

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