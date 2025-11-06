import { Metadata } from 'next'
import { DisplayForm, ContentSection } from '@/features/settings'

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