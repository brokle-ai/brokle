'use client'

import { ProfileForm, ContentSection } from '@/features/settings'

export default function ProfilePage() {
  return (
    <ContentSection
      title="Profile"
      description="This is how others will see you on the site."
    >
      <ProfileForm />
    </ContentSection>
  )
}
