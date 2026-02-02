'use client'

import { AccountForm, ContentSection } from '@/features/settings'

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
