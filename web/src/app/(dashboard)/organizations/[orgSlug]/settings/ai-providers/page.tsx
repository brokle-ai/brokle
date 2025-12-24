'use client'

import { useState } from 'react'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ContentSection } from '@/features/settings'
import { AIProvidersSettings } from '@/features/ai-providers'
import { useWorkspace } from '@/context/workspace-context'

export default function OrganizationAIProvidersPage() {
  const { currentOrganization } = useWorkspace()
  const [addDialogOpen, setAddDialogOpen] = useState(false)

  if (!currentOrganization) {
    return null
  }

  return (
    <ContentSection
      title="AI Providers"
      description="Configure API credentials for AI model providers. These credentials are shared across all projects in this organization."
      action={
        <Button onClick={() => setAddDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Provider
        </Button>
      }
    >
      <AIProvidersSettings
        orgId={currentOrganization.id}
        addDialogOpen={addDialogOpen}
        onAddDialogOpenChange={setAddDialogOpen}
      />
    </ContentSection>
  )
}
