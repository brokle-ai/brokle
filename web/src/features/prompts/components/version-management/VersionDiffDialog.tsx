'use client'

import { useMemo } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { DiffViewer } from './DiffViewer'
import type {
  PromptVersion,
  PromptType,
  TextTemplate,
  ChatTemplate,
  ModelConfig,
} from '../../types'

// ============================================================================
// Types
// ============================================================================

interface VersionDiffDialogProps {
  isOpen: boolean
  onOpenChange: (open: boolean) => void
  fromVersion: PromptVersion | null
  toVersion: PromptVersion | null
  promptType: PromptType
  promptName?: string
}

// ============================================================================
// Helpers
// ============================================================================

/**
 * Format template for diff display
 * Text templates show content directly, chat templates show pretty-printed JSON
 */
function formatTemplate(
  template: TextTemplate | ChatTemplate,
  type: PromptType
): string {
  if (type === 'text') {
    return (template as TextTemplate).content || ''
  }
  // Chat template - format as pretty JSON
  const chatTemplate = template as ChatTemplate
  return JSON.stringify(chatTemplate.messages || [], null, 2)
}

/**
 * Format config for diff display
 */
function formatConfig(config: ModelConfig | null | undefined): string {
  if (!config) return '{}'
  return JSON.stringify(config, null, 2)
}

/**
 * Calculate variable changes between versions
 */
function getVariableChanges(
  fromVariables: string[],
  toVariables: string[]
): { added: string[]; removed: string[] } {
  const fromSet = new Set(fromVariables)
  const toSet = new Set(toVariables)

  const added = toVariables.filter((v) => !fromSet.has(v))
  const removed = fromVariables.filter((v) => !toSet.has(v))

  return { added, removed }
}

// ============================================================================
// Variable Changes Section
// ============================================================================

interface VariableChangesProps {
  added: string[]
  removed: string[]
}

function VariableChanges({ added, removed }: VariableChangesProps) {
  if (added.length === 0 && removed.length === 0) {
    return null
  }

  return (
    <div className="space-y-3">
      <h4 className="text-base font-semibold">Variables</h4>
      <div className="flex flex-wrap gap-3">
        {added.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-green-600 dark:text-green-400 font-medium">
              Added:
            </span>
            <div className="flex flex-wrap gap-1.5">
              {added.map((v) => (
                <Badge
                  key={v}
                  variant="outline"
                  className="bg-green-100 dark:bg-green-900/30 font-mono text-sm px-2 py-0.5"
                >
                  +{`{{${v}}}`}
                </Badge>
              ))}
            </div>
          </div>
        )}
        {removed.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-red-600 dark:text-red-400 font-medium">
              Removed:
            </span>
            <div className="flex flex-wrap gap-1.5">
              {removed.map((v) => (
                <Badge
                  key={v}
                  variant="outline"
                  className="bg-red-100 dark:bg-red-900/30 font-mono text-sm px-2 py-0.5"
                >
                  -{`{{${v}}}`}
                </Badge>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

// ============================================================================
// Main VersionDiffDialog Component
// ============================================================================

export function VersionDiffDialog({
  isOpen,
  onOpenChange,
  fromVersion,
  toVersion,
  promptType,
  promptName,
}: VersionDiffDialogProps) {
  // Format templates for comparison
  const { fromTemplate, toTemplate } = useMemo(() => {
    if (!fromVersion || !toVersion) {
      return { fromTemplate: '', toTemplate: '' }
    }
    return {
      fromTemplate: formatTemplate(fromVersion.template, promptType),
      toTemplate: formatTemplate(toVersion.template, promptType),
    }
  }, [fromVersion, toVersion, promptType])

  // Format configs for comparison
  const { fromConfig, toConfig } = useMemo(() => {
    if (!fromVersion || !toVersion) {
      return { fromConfig: '{}', toConfig: '{}' }
    }
    return {
      fromConfig: formatConfig(fromVersion.config),
      toConfig: formatConfig(toVersion.config),
    }
  }, [fromVersion, toVersion])

  // Calculate variable changes
  const variableChanges = useMemo(() => {
    if (!fromVersion || !toVersion) {
      return { added: [], removed: [] }
    }
    return getVariableChanges(
      fromVersion.variables || [],
      toVersion.variables || []
    )
  }, [fromVersion, toVersion])

  // Check if configs are different
  const hasConfigChanges = fromConfig !== toConfig

  if (!fromVersion || !toVersion) {
    return null
  }

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-5xl max-h-[90vh] flex flex-col">
        <DialogHeader className="space-y-1">
          <DialogTitle className="text-xl">
            Changes v{fromVersion.version} → v{toVersion.version}
          </DialogTitle>
          {promptName && (
            <DialogDescription className="text-base">
              {promptName}
            </DialogDescription>
          )}
        </DialogHeader>

        <ScrollArea className="flex-1 -mx-6 px-6">
          <div className="space-y-6 py-2">
            {/* Variable Changes */}
            <VariableChanges
              added={variableChanges.added}
              removed={variableChanges.removed}
            />

            {/* Content Diff */}
            <div className="space-y-3">
              <h4 className="text-base font-semibold">Content</h4>
              <DiffViewer
                oldString={fromTemplate}
                newString={toTemplate}
                oldLabel={`v${fromVersion.version}`}
                newLabel={`v${toVersion.version}`}
                oldSubLabel={fromVersion.commit_message}
                newSubLabel={toVersion.commit_message}
              />
            </div>

            {/* Config Diff - always show section */}
            <div className="space-y-3">
              <h4 className="text-base font-semibold">Config</h4>
              {hasConfigChanges ? (
                <DiffViewer
                  oldString={fromConfig}
                  newString={toConfig}
                  oldLabel={`v${fromVersion.version}`}
                  newLabel={`v${toVersion.version}`}
                />
              ) : (
                <p className="text-sm text-muted-foreground">No changes</p>
              )}
            </div>
          </div>
        </ScrollArea>

        <DialogFooter className="mt-4">
          <Button onClick={() => onOpenChange(false)}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
