'use client'

import { useOrganization } from '@/context/organization-context'
import { UsageAnalyticsDashboard } from '@/components/analytics/usage-analytics-dashboard'
import { TabsContent } from '@/components/ui/tabs'

export default function ProjectAnalyticsSettingsPage() {
  const { currentProject, currentOrganization } = useOrganization()

  if (!currentProject || !currentOrganization) {
    return null
  }

  return (
    <TabsContent value="analytics" className="space-y-6">
      <UsageAnalyticsDashboard 
        organizationId={currentOrganization.id}
        projectId={currentProject.id}
      />
    </TabsContent>
  )
}