'use client'

import { useWorkspace } from '@/context/workspace-context'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { MemberManagement } from './member-management'
import { ProjectGrid } from './project-grid'

export function OrganizationOverview() {
  const { currentOrganization } = useWorkspace()

  if (!currentOrganization) {
    return <div>No organization selected</div>
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Projects</CardTitle>
            <CardDescription>Active projects in this organization</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {currentOrganization.projects?.length || 0}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Members</CardTitle>
            <CardDescription>Team members with access</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {currentOrganization.members?.length || 0}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Plan</CardTitle>
            <CardDescription>Current subscription plan</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold capitalize">
              {currentOrganization.plan || 'Free'}
            </div>
          </CardContent>
        </Card>
      </div>

      <ProjectGrid />
      <MemberManagement />
    </div>
  )
}
