'use client'

import { useState } from 'react'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ContentSection } from '@/features/settings'
import { ProjectAPIKeysSection } from '@/features/projects'

export default function ProjectAPIKeysPage() {
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  return (
    <ContentSection
      title="API Keys"
      description="Manage API keys for accessing your project programmatically."
      action={
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create API Key
        </Button>
      }
    >
      <ProjectAPIKeysSection
        createDialogOpen={createDialogOpen}
        onCreateDialogOpenChange={setCreateDialogOpen}
      />
    </ContentSection>
  )
}
