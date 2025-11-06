import { Metadata } from 'next'
import { ProfileForm, ContentSection } from '@/features/settings'

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