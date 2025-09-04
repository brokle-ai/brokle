import { Metadata } from 'next'
import { NotificationsForm } from '@/components/settings/notifications-form'
import { ContentSection } from '@/components/settings/content-section'

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