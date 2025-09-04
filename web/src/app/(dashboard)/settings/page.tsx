import { Metadata } from 'next'
import { ProfileForm } from '@/components/settings/profile-form'
import { ContentSection } from '@/components/settings/content-section'

export const metadata: Metadata = {
  title: 'Profile',
  description: 'Manage your profile settings.',
}

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