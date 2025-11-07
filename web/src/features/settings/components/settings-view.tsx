'use client'

import { ProfileForm } from './profile-form'
import { AccountForm } from './account-form'
import { AppearanceForm } from './appearance-form'
import { NotificationsForm } from './notifications-form'
import { DisplayForm } from './display-form'
import { SettingsNav } from './settings-nav'
import { ContentSection } from './content-section'

export function SettingsView() {
  return (
    <div className="grid gap-6 lg:grid-cols-[240px_1fr]">
        <aside>
          <SettingsNav />
        </aside>
        <div className="space-y-6">
          <ContentSection
            title="Profile"
            description="Update your personal information"
          >
            <ProfileForm />
          </ContentSection>

          <ContentSection
            title="Account"
            description="Manage your account settings"
          >
            <AccountForm />
          </ContentSection>

          <ContentSection
            title="Appearance"
            description="Customize how the app looks"
          >
            <AppearanceForm />
          </ContentSection>

          <ContentSection
            title="Notifications"
            description="Configure your notification preferences"
          >
            <NotificationsForm />
          </ContentSection>

          <ContentSection
            title="Display"
            description="Adjust display settings"
          >
            <DisplayForm />
          </ContentSection>
        </div>
      </div>
  )
}
