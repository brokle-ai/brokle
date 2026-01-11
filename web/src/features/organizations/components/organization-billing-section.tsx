'use client'

import { Receipt, BarChart3, Shield } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

import { BillingPage, UsagePage, BudgetsPage } from '@/features/billing'

export function OrganizationBillingSection() {
  const { currentOrganization } = useWorkspace()

  if (!currentOrganization) {
    return null
  }

  // Format projects for the budget page
  const projectOptions = currentOrganization?.projects?.map((p) => ({ id: p.id, name: p.name })) ?? []

  return (
    <Tabs defaultValue="billing" className="space-y-6">
      <TabsList>
        <TabsTrigger value="billing" className="gap-2">
          <Receipt className="h-4 w-4" />
          Billing
        </TabsTrigger>
        <TabsTrigger value="usage" className="gap-2">
          <BarChart3 className="h-4 w-4" />
          Usage
        </TabsTrigger>
        <TabsTrigger value="budgets" className="gap-2">
          <Shield className="h-4 w-4" />
          Budgets
        </TabsTrigger>
      </TabsList>

      {/* Billing Tab */}
      <TabsContent value="billing">
        <BillingPage organizationId={currentOrganization.id} />
      </TabsContent>

      {/* Usage Tab */}
      <TabsContent value="usage">
        <UsagePage organizationId={currentOrganization.id} />
      </TabsContent>

      {/* Budgets Tab */}
      <TabsContent value="budgets">
        <BudgetsPage
          organizationId={currentOrganization.id}
          projects={projectOptions}
        />
      </TabsContent>
    </Tabs>
  )
}
