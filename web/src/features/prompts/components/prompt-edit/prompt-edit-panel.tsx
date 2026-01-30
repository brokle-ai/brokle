'use client'

import * as React from 'react'
import { Save } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import type {
  Prompt,
  PromptVersion,
  TextTemplate,
  ChatTemplate,
  CreateVersionRequest,
} from '../../types'
import { PromptEditor } from '../prompt-editor'
import { JsonConfigEditor } from './json-config-editor'
import { SaveVersionDialog } from './save-version-dialog'

// ============================================================================
// Types
// ============================================================================

interface PromptEditPanelProps {
  prompt: Prompt
  sourceVersion: PromptVersion | null
  currentVariables: string[]
  isRestoreFlow: boolean
  onSave: (data: CreateVersionRequest) => Promise<void>
  onCancel: () => void
  isSaving?: boolean
}

// ============================================================================
// VariablesBar Component
// ============================================================================

interface VariablesBarProps {
  variables: string[]
}

function VariablesBar({ variables }: VariablesBarProps) {
  if (variables.length === 0) return null

  return (
    <div className="flex items-center gap-2 px-4 py-2 border-b bg-muted/30">
      <span className="text-xs text-muted-foreground">Variables:</span>
      <div className="flex flex-wrap gap-1.5">
        {variables.map((variable) => (
          <Badge
            key={variable}
            variant="secondary"
            className="font-mono text-xs h-5 px-1.5 cursor-pointer hover:bg-secondary/80"
            onClick={() => {
              navigator.clipboard.writeText(`{{${variable}}}`)
            }}
          >
            {`{{${variable}}}`}
          </Badge>
        ))}
      </div>
    </div>
  )
}

// ============================================================================
// Main PromptEditPanel Component
// ============================================================================

export function PromptEditPanel({
  prompt,
  sourceVersion,
  currentVariables,
  isRestoreFlow,
  onSave,
  onCancel,
  isSaving,
}: PromptEditPanelProps) {
  // Use source version data if available, otherwise fall back to prompt (latest)
  const initialVersion = sourceVersion || {
    id: prompt.version_id,
    version: prompt.version,
    template: prompt.template,
    config: prompt.config,
    variables: prompt.variables,
    commit_message: prompt.commit_message,
    labels: prompt.labels,
    created_at: prompt.created_at,
    created_by: prompt.created_by,
  }

  // Track edited template locally
  const [editedTemplate, setEditedTemplate] = React.useState<TextTemplate | ChatTemplate>(
    initialVersion.template
  )
  // Track edited config locally
  const [editedConfig, setEditedConfig] = React.useState<Record<string, unknown> | null>(
    (initialVersion.config as Record<string, unknown>) || null
  )
  // Track if config JSON is valid
  const [isConfigValid, setIsConfigValid] = React.useState(true)
  // Save dialog state
  const [showSaveDialog, setShowSaveDialog] = React.useState(false)

  // Reset state when source version changes
  React.useEffect(() => {
    if (sourceVersion) {
      setEditedTemplate(sourceVersion.template)
      setEditedConfig((sourceVersion.config as Record<string, unknown>) || null)
      setIsConfigValid(true)
    }
  }, [sourceVersion?.id])

  // Handle template change
  const handleTemplateChange = React.useCallback(
    (template: TextTemplate | ChatTemplate) => {
      setEditedTemplate(template)
    },
    []
  )

  // Handle config change
  const handleConfigChange = React.useCallback(
    (config: Record<string, unknown> | null, isValid: boolean) => {
      setEditedConfig(config)
      setIsConfigValid(isValid)
    },
    []
  )

  // Handle save with commit message
  const handleSave = React.useCallback(
    async (commitMessage?: string) => {
      // Default commit message for restore flow
      const defaultMessage = isRestoreFlow
        ? `Restored from version ${sourceVersion?.version || prompt.version}`
        : undefined

      await onSave({
        template: editedTemplate,
        config: editedConfig ?? undefined,
        commit_message: commitMessage || defaultMessage,
      })
    },
    [editedTemplate, editedConfig, isRestoreFlow, sourceVersion, prompt.version, onSave]
  )

  // Can save only if config is valid
  const canSave = isConfigValid

  return (
    <div className="flex h-full flex-col bg-background">
      {/* Header */}
      <div className="flex items-center justify-between border-b px-4 py-3">
        <div>
          <h2 className="text-lg font-semibold">
            {isRestoreFlow
              ? `Restore from Version ${sourceVersion?.version || prompt.version}`
              : 'Create New Version'}
          </h2>
          <p className="text-sm text-muted-foreground">
            {isRestoreFlow
              ? 'Review and save to create a new version with this content'
              : `Editing from version ${sourceVersion?.version || prompt.version}`}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={onCancel} disabled={isSaving}>
            Cancel
          </Button>
          <Button
            onClick={() => setShowSaveDialog(true)}
            disabled={!canSave || isSaving}
          >
            <Save className="h-4 w-4 mr-2" />
            Save Version
          </Button>
        </div>
      </div>

      {/* Variables bar */}
      <VariablesBar variables={currentVariables} />

      {/* Editor content */}
      <ScrollArea className="flex-1">
        <div className="p-4 space-y-6">
          <PromptEditor
            type={prompt.type}
            template={editedTemplate}
            onChange={handleTemplateChange}
            variables={currentVariables}
          />

          {/* Divider */}
          <div className="border-t" />

          {/* JSON Config Editor */}
          <JsonConfigEditor
            config={editedConfig}
            onChange={handleConfigChange}
          />
        </div>
      </ScrollArea>

      {/* Save Version Dialog */}
      <SaveVersionDialog
        open={showSaveDialog}
        onOpenChange={setShowSaveDialog}
        onSave={handleSave}
        isSaving={isSaving || false}
      />
    </div>
  )
}
