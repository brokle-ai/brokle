import { Metadata } from 'next'
import { NotificationsForm, ContentSection } from '@/features/settings'

export const metadata: Metadata = {
  title: 'Notifications',
  description: 'Configure your notification preferences.',
}

export default function NotificationsPage() {
  return (
    <ContentSection
      title="Notifications"
      description="Configure how you receive notifications."
    >
      <NotificationsForm />
    </ContentSection>
  )
}