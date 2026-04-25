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
import { ChatDiffViewer } from './ChatDiffViewer'
import type {
  PromptVersion,
  PromptType,
  TextTemplate,
  ChatTemplate,
  ModelConfig,
} from '../../types'

interface VersionDiffDialogProps {
  isOpen: boolean
  onOpenChange: (open: boolean) => void
  fromVersion: PromptVersion | null
  toVersion: PromptVersion | null
  promptType: PromptType
  promptName?: string
}

function formatConfig(config: ModelConfig | null | undefined): string {
  if (!config) return '{}'
  return JSON.stringify(config, null, 2)
}

function getVariableChanges(
  fromVariables: string[],
  toVariables: string[],
): { added: string[]; removed: string[] } {
  const fromSet = new Set(fromVariables)
  const toSet = new Set(toVariables)
  const added = toVariables.filter((v) => !fromSet.has(v))
  const removed = fromVariables.filter((v) => !toSet.has(v))
  return { added, removed }
}

function VariableChanges({
  added,
  removed,
}: {
  added: string[]
  removed: string[]
}) {
  if (added.length === 0 && removed.length === 0) return null

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

export function VersionDiffDialog({
  isOpen,
  onOpenChange,
  fromVersion,
  toVersion,
  promptType,
  promptName,
}: VersionDiffDialogProps) {
  const { fromText, toText } = useMemo(() => {
    if (!fromVersion || !toVersion || promptType !== 'text') {
      return { fromText: '', toText: '' }
    }
    return {
      fromText: (fromVersion.template as TextTemplate).content ?? '',
      toText: (toVersion.template as TextTemplate).content ?? '',
    }
  }, [fromVersion, toVersion, promptType])

  const { fromMessages, toMessages } = useMemo(() => {
    if (!fromVersion || !toVersion || promptType !== 'chat') {
      return { fromMessages: [], toMessages: [] }
    }
    return {
      fromMessages: (fromVersion.template as ChatTemplate).messages ?? [],
      toMessages: (toVersion.template as ChatTemplate).messages ?? [],
    }
  }, [fromVersion, toVersion, promptType])

  const { fromConfig, toConfig } = useMemo(() => {
    if (!fromVersion || !toVersion) {
      return { fromConfig: '{}', toConfig: '{}' }
    }
    return {
      fromConfig: formatConfig(fromVersion.config),
      toConfig: formatConfig(toVersion.config),
    }
  }, [fromVersion, toVersion])

  const variableChanges = useMemo(() => {
    if (!fromVersion || !toVersion) return { added: [], removed: [] }
    return getVariableChanges(
      fromVersion.variables || [],
      toVersion.variables || [],
    )
  }, [fromVersion, toVersion])

  const hasConfigChanges = fromConfig !== toConfig

  if (!fromVersion || !toVersion) return null

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
            <VariableChanges
              added={variableChanges.added}
              removed={variableChanges.removed}
            />

            <div className="space-y-3">
              <h4 className="text-base font-semibold">Content</h4>
              {promptType === 'chat' ? (
                <ChatDiffViewer
                  fromMessages={fromMessages}
                  toMessages={toMessages}
                  oldLabel={`v${fromVersion.version}`}
                  newLabel={`v${toVersion.version}`}
                  oldSubLabel={fromVersion.commit_message}
                  newSubLabel={toVersion.commit_message}
                />
              ) : (
                <DiffViewer
                  oldString={fromText}
                  newString={toText}
                  oldLabel={`v${fromVersion.version}`}
                  newLabel={`v${toVersion.version}`}
                  oldSubLabel={fromVersion.commit_message}
                  newSubLabel={toVersion.commit_message}
                />
              )}
            </div>

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
