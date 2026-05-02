import * as React from 'react'
import { Save } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import type {
  Prompt,
  PromptVersion,
  CreateVersionRequest,
  PromptTemplate,
  ModelConfig,
} from '../../types'
import { PromptEditor } from '../prompt-editor'
import { JsonConfigEditor } from './json-config-editor'
import { SaveVersionDialog } from './save-version-dialog'

interface PromptEditPanelProps {
  prompt: Prompt
  sourceVersion: PromptVersion | null
  currentVariables: string[]
  isRestoreFlow: boolean
  onSave: (data: CreateVersionRequest) => Promise<void>
  onCancel: () => void
  isSaving?: boolean
}

function VariablesBar({ variables }: { variables: string[] }) {
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
              void navigator.clipboard.writeText(`{{${variable}}}`)
            }}
          >
            {`{{${variable}}}`}
          </Badge>
        ))}
      </div>
    </div>
  )
}

export function PromptEditPanel({
  prompt,
  sourceVersion,
  currentVariables,
  isRestoreFlow,
  onSave,
  onCancel,
  isSaving,
}: PromptEditPanelProps) {
  const initialVersion: PromptVersion = sourceVersion || {
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

  const [editedTemplate, setEditedTemplate] = React.useState<PromptTemplate>(
    initialVersion.template,
  )
  const [editedConfig, setEditedConfig] = React.useState<Record<string, unknown> | null>(
    (initialVersion.config as Record<string, unknown>) || null,
  )
  const [isConfigValid, setIsConfigValid] = React.useState(true)
  const [showSaveDialog, setShowSaveDialog] = React.useState(false)

  const sourceVersionId = sourceVersion?.id
  React.useEffect(() => {
    if (sourceVersion) {
      setEditedTemplate(sourceVersion.template)
      setEditedConfig((sourceVersion.config as Record<string, unknown>) || null)
      setIsConfigValid(true)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sourceVersionId])

  const handleTemplateChange = React.useCallback((template: PromptTemplate) => {
    setEditedTemplate(template)
  }, [])

  const handleConfigChange = React.useCallback(
    (config: Record<string, unknown> | null, isValid: boolean) => {
      setEditedConfig(config)
      setIsConfigValid(isValid)
    },
    [],
  )

  const handleSave = React.useCallback(
    async (commitMessage?: string) => {
      const defaultMessage = isRestoreFlow
        ? `Restored from version ${sourceVersion?.version || prompt.version}`
        : undefined

      await onSave({
        template: editedTemplate,
        config: (editedConfig as ModelConfig | undefined) ?? undefined,
        commit_message: commitMessage || defaultMessage,
      })
    },
    [editedTemplate, editedConfig, isRestoreFlow, sourceVersion, prompt.version, onSave],
  )

  const canSave = isConfigValid

  return (
    <div className="flex h-full flex-col bg-background">
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

      <VariablesBar variables={currentVariables} />

      <ScrollArea className="flex-1">
        <div className="p-4 space-y-6">
          <PromptEditor
            type={prompt.type}
            template={editedTemplate}
            onChange={handleTemplateChange}
            variables={currentVariables}
          />

          <div className="border-t" />

          <JsonConfigEditor
            config={editedConfig}
            onChange={handleConfigChange}
          />
        </div>
      </ScrollArea>

      <SaveVersionDialog
        open={showSaveDialog}
        onOpenChange={setShowSaveDialog}
        onSave={handleSave}
        isSaving={isSaving || false}
      />
    </div>
  )
}
