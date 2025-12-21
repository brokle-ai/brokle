'use client'

import { useState } from 'react'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ContentSection } from '@/features/settings'
import { ProjectIntegrationsSection } from '@/features/projects'

export default function ProjectIntegrationsPage() {
  const [addDialogOpen, setAddDialogOpen] = useState(false)

  return (
    <ContentSection
      title="Integrations"
      description="Connect external services to enhance your project capabilities."
      action={
        <Button onClick={() => setAddDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Integration
        </Button>
      }
    >
      <ProjectIntegrationsSection
        addDialogOpen={addDialogOpen}
        onAddDialogOpenChange={setAddDialogOpen}
      />
    </ContentSection>
  )
}
