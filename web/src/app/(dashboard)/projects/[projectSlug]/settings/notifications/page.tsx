'use client'

import { NotificationsForm, ContentSection } from '@/features/settings'

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
