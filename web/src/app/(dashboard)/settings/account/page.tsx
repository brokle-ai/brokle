import { Metadata } from 'next'
import { AccountForm } from '@/components/settings/account-form'
import { ContentSection } from '@/components/settings/content-section'

export const metadata: Metadata = {
  title: 'Account',
  description: 'Manage your account settings.',
}

export default function AccountPage() {
  return (
    <ContentSection
      title="Account"
      description="Update your account settings. Set your preferred language and timezone."
    >
      <AccountForm />
    </ContentSection>
  )
}