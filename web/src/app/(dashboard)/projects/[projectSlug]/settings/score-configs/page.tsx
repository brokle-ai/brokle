'use client'

import { useState } from 'react'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ContentSection } from '@/features/settings'
import { ScoreConfigsSection } from '@/features/scores'
import { useWorkspace } from '@/context/workspace-context'

export default function ScoreConfigsSettingsPage() {
  const { currentProject } = useWorkspace()
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  if (!currentProject) {
    return null
  }

  return (
    <ContentSection
      title="Score Configs"
      description="Define validation rules and schemas for evaluation scores."
      action={
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Config
        </Button>
      }
    >
      <ScoreConfigsSection
        projectId={currentProject.id}
        createDialogOpen={createDialogOpen}
        onCreateDialogOpenChange={setCreateDialogOpen}
      />
    </ContentSection>
  )
}
