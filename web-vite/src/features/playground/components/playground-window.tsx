import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import {
  Play,
  X,
  Copy,
  Loader2,
  CloudOff,
  Cloud,
  Settings2,
  FlaskConical,
  Pencil,
  Check,
} from 'lucide-react'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { TooltipProvider } from '@/components/ui/tooltip'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  usePlaygroundStore,
  createContentSnapshot,
  type PlaygroundWindow as PlaygroundWindowState,
} from '../stores/playground-store'
import type {
  WindowState,
  ChatMessage,
  ModelConfig,
  ParameterKey,
  ExecuteRequest,
} from '../types'
import {
  createMessage,
  PARAMETER_DEFINITIONS,
  PROVIDER_PARAMETER_SUPPORT,
  getEnabledModelConfig,
} from '../types'
import { useStreaming } from '../hooks/use-streaming'
import { useUpdateSessionMutation } from '../hooks/use-playground-queries'
import { MessageEditor } from './message-editor'
import { LoadPromptDropdown } from './load-prompt-dropdown'
import { extractVariablesFromMessages } from './variable-editor'
import { ModelSelector } from './model-selector'
import { ToolbarRow } from './toolbar-row'
import {
  SaveAsPromptDialog,
  type PromptSavedData,
} from './save-as-prompt-dialog'
import { StreamingOutput } from './streaming-output'
import { ParameterControl } from './parameter-control'
import { RunHistoryPanel } from './run-history-panel'

interface PlaygroundWindowProps {
  index: number
  projectId: string
  orgId: string
  sessionId?: string
  onRegisterExecute?: (executeFn: () => Promise<void>) => void
  onUnregisterExecute?: () => void
}

const AUTO_SAVE_DELAY = 1500

// Convert store windows → backend WindowState payload.
const buildWindowsPayload = (
  windows: PlaygroundWindowState[],
): WindowState[] => {
  return windows.map((w) => ({
    template: { messages: w.messages },
    variables: w.variables,
    config: w.config || undefined,
    loadedFromPromptId: w.loadedFromPromptId || undefined,
    loadedFromPromptName: w.loadedFromPromptName || undefined,
    loadedFromPromptVersionId: w.loadedFromPromptVersionId || undefined,
    loadedFromPromptVersionNumber:
      w.loadedFromPromptVersionNumber || undefined,
    loadedTemplate: w.loadedTemplate || undefined,
  }))
}

export function PlaygroundWindow({
  index,
  projectId,
  orgId,
  sessionId,
  onRegisterExecute,
  onUnregisterExecute,
}: PlaygroundWindowProps) {
  // Window slice from the store.
  const windowState = usePlaygroundStore((s) => s.windows[index])

  // Action handles (stable refs).
  const updateWindow = usePlaygroundStore((s) => s.updateWindow)
  const removeWindow = usePlaygroundStore((s) => s.removeWindow)
  const duplicateWindow = usePlaygroundStore((s) => s.duplicateWindow)
  const renameWindow = usePlaygroundStore((s) => s.renameWindow)
  const setWindowOutput = usePlaygroundStore((s) => s.setWindowOutput)
  const setLastSavedSnapshot = usePlaygroundStore(
    (s) => s.setLastSavedSnapshot,
  )
  const unlinkPrompt = usePlaygroundStore((s) => s.unlinkPrompt)
  const unlinkSpan = usePlaygroundStore((s) => s.unlinkSpan)
  const restoreFromHistory = usePlaygroundStore((s) => s.restoreFromHistory)
  const clearWindowHistory = usePlaygroundStore((s) => s.clearWindowHistory)
  const markHistoryAsStale = usePlaygroundStore((s) => s.markHistoryAsStale)
  const windows = usePlaygroundStore((s) => s.windows)

  // Dirty state derived from content snapshot (prevents render loops).
  const currentSnapshot = useMemo(() => {
    if (!windowState) return ''
    return createContentSnapshot(windowState)
  }, [windowState])

  const isDirty = useMemo(() => {
    if (!windowState?.lastSavedSnapshot) return false
    return currentSnapshot !== windowState.lastSavedSnapshot
  }, [currentSnapshot, windowState?.lastSavedSnapshot])

  // Opik-style "modified since load" indicator for linked prompts.
  const hasUnsavedPromptChanges = useMemo(() => {
    if (
      !windowState?.loadedFromPromptVersionId ||
      !windowState?.loadedTemplate
    ) {
      return false
    }
    const currentTemplate = JSON.stringify(
      windowState.messages.map(({ role, content }) => ({ role, content })),
    )
    return currentTemplate !== windowState.loadedTemplate
  }, [
    windowState?.messages,
    windowState?.loadedTemplate,
    windowState?.loadedFromPromptVersionId,
  ])

  const updateSessionMutation = useUpdateSessionMutation(
    projectId,
    sessionId || '',
  )
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Latest-value refs to avoid stale closures in timers / unmount handlers.
  const latestValuesRef = useRef({
    isDirty,
    currentSnapshot,
    sessionId,
    projectId,
    windowState,
    windows,
  })

  useEffect(() => {
    latestValuesRef.current = {
      isDirty,
      currentSnapshot,
      sessionId,
      projectId,
      windowState,
      windows,
    }
  })

  const autoSave = useCallback(async () => {
    const {
      isDirty: currentIsDirty,
      sessionId: currentSessionId,
      projectId: currentProjectId,
      windows: allWindows,
      currentSnapshot: snapshot,
    } = latestValuesRef.current

    if (
      !currentSessionId ||
      !currentProjectId ||
      !currentIsDirty ||
      !allWindows?.length
    )
      return

    try {
      await updateSessionMutation.mutateAsync({
        windows: buildWindowsPayload(allWindows),
      })

      setLastSavedSnapshot(index, snapshot)
    } catch (error) {
      console.error('Auto-save failed:', error)
    }
  }, [updateSessionMutation, setLastSavedSnapshot, index])

  const autoSaveRef = useRef(autoSave)
  useEffect(() => {
    autoSaveRef.current = autoSave
  }, [autoSave])

  useEffect(() => {
    if (!isDirty || !sessionId) return

    if (saveTimerRef.current) {
      clearTimeout(saveTimerRef.current)
    }

    saveTimerRef.current = setTimeout(() => {
      autoSaveRef.current()
    }, AUTO_SAVE_DELAY)

    return () => {
      if (saveTimerRef.current) {
        clearTimeout(saveTimerRef.current)
      }
    }
  }, [isDirty, sessionId])

  useEffect(() => {
    return () => {
      const { isDirty: currentIsDirty, sessionId: currentSessionId } =
        latestValuesRef.current
      if (currentIsDirty && currentSessionId) {
        autoSaveRef.current()
      }
    }
  }, [])

  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      const { isDirty: currentIsDirty, sessionId: currentSessionId } =
        latestValuesRef.current
      if (currentIsDirty && currentSessionId) {
        autoSaveRef.current()
        e.preventDefault()
        e.returnValue = ''
      }
    }

    globalThis.addEventListener('beforeunload', handleBeforeUnload)
    return () =>
      globalThis.removeEventListener('beforeunload', handleBeforeUnload)
  }, [])

  const { stream, abort, isStreaming, content, error, metrics } = useStreaming(
    {
      projectId,
      onEnd: (finalContent, finalMetrics, capturedInputs) => {
        setWindowOutput(
          index,
          finalContent,
          finalMetrics,
          capturedInputs ?? undefined,
        )
      },
    },
  )

  const [configOpen, setConfigOpen] = useState(false)
  const currentConfig: ModelConfig = windowState?.config || {}

  const [isEditingName, setIsEditingName] = useState(false)
  const [editedName, setEditedName] = useState(windowState?.name || '')
  const nameInputRef = useRef<HTMLInputElement>(null)

  const prevInputsSnapshotRef = useRef<string>('')

  useEffect(() => {
    const currentInputsSnapshot = JSON.stringify({
      messages:
        windowState?.messages?.map(({ role, content }) => ({
          role,
          content,
        })) ?? [],
      variables: windowState?.variables ?? {},
      config: windowState?.config ?? null,
    })
    if (
      prevInputsSnapshotRef.current &&
      prevInputsSnapshotRef.current !== currentInputsSnapshot
    ) {
      markHistoryAsStale(index)
    }
    prevInputsSnapshotRef.current = currentInputsSnapshot
  }, [
    windowState?.messages,
    windowState?.variables,
    windowState?.config,
    markHistoryAsStale,
    index,
  ])

  useEffect(() => {
    if (isEditingName && nameInputRef.current) {
      nameInputRef.current.focus()
      nameInputRef.current.select()
    }
  }, [isEditingName])

  const handleNameSave = useCallback(() => {
    const trimmedName = editedName.trim()
    if (trimmedName && trimmedName !== windowState?.name) {
      renameWindow(index, trimmedName)
    } else {
      setEditedName(windowState?.name || '')
    }
    setIsEditingName(false)
  }, [editedName, windowState?.name, renameWindow, index])

  const handleNameKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        handleNameSave()
      } else if (e.key === 'Escape') {
        setEditedName(windowState?.name || '')
        setIsEditingName(false)
      }
    },
    [handleNameSave, windowState?.name],
  )

  const handleConfigChange = (
    field: string,
    value: number | boolean | undefined,
  ) => {
    const newConfig: ModelConfig = { ...currentConfig, [field]: value }

    if (
      typeof value === 'boolean' &&
      value === true &&
      field.endsWith('_enabled')
    ) {
      const paramKey = field.replace('_enabled', '') as ParameterKey
      const def = PARAMETER_DEFINITIONS.find((d) => d.key === paramKey)
      if (def && newConfig[paramKey] === undefined) {
        ;(newConfig as Record<string, number | boolean | undefined>)[paramKey] =
          def.defaultValue
      }
    }

    updateWindow(index, { config: newConfig })
  }

  const handlePromptSaved = useCallback(
    (data: PromptSavedData) => {
      const currentTemplate = JSON.stringify(
        windowState?.messages.map(({ role, content }) => ({
          role,
          content,
        })) ?? [],
      )

      updateWindow(index, {
        loadedFromPromptId: data.promptId,
        loadedFromPromptName: data.promptName,
        loadedFromPromptVersionId: data.versionId,
        loadedFromPromptVersionNumber: data.versionNumber,
        loadedTemplate: currentTemplate,
      })
    },
    [windowState?.messages, updateWindow, index],
  )

  const handleExecute = useCallback(async () => {
    if (!windowState) return

    if (!windowState.config?.model) {
      alert('Please select a model')
      return
    }
    if (
      !windowState.messages.length ||
      !windowState.messages.some((m) => m.content.trim())
    ) {
      alert('Please enter at least one message')
      return
    }

    const filteredConfig = getEnabledModelConfig(windowState.config)

    const request: ExecuteRequest = {
      template: { messages: windowState.messages },
      prompt_type: 'chat',
      variables: windowState.variables,
      config_overrides: filteredConfig,
      session_id: sessionId,
      project_id: projectId,
    }

    await stream(request, windowState.config ?? null)
  }, [windowState, sessionId, projectId, stream])

  useEffect(() => {
    if (onRegisterExecute) {
      onRegisterExecute(handleExecute)
    }
    return () => {
      if (onUnregisterExecute) {
        onUnregisterExecute()
      }
    }
  }, [handleExecute, onRegisterExecute, onUnregisterExecute])

  const extractedVariables = extractVariablesFromMessages(
    windowState?.messages || [],
  )

  const getSaveStatus = () => {
    if (!sessionId) return null
    if (updateSessionMutation.isPending) return 'saving'
    if (isDirty) return 'unsaved'
    return 'saved'
  }

  const saveStatus = getSaveStatus()

  return (
    <TooltipProvider>
      <Card className="flex flex-col h-full">
        <CardHeader className="flex flex-col gap-2 space-y-0 pb-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              {isEditingName ? (
                <div className="flex items-center gap-1">
                  <Input
                    ref={nameInputRef}
                    value={editedName}
                    onChange={(e) => setEditedName(e.target.value)}
                    onBlur={handleNameSave}
                    onKeyDown={handleNameKeyDown}
                    className="h-7 w-40 text-sm font-medium"
                    maxLength={50}
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={handleNameSave}
                  >
                    <Check className="h-3.5 w-3.5" />
                  </Button>
                </div>
              ) : (
                <button
                  onClick={() => {
                    setEditedName(windowState?.name || '')
                    setIsEditingName(true)
                  }}
                  className="flex items-center gap-1.5 text-sm font-medium hover:text-primary transition-colors group"
                  title="Click to rename"
                >
                  <span className="max-w-[200px] truncate">
                    {windowState?.name || 'Window'}
                  </span>
                  <Pencil className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                </button>
              )}
            </div>
            <div className="flex items-center gap-2">
              <RunHistoryPanel
                history={windowState?.runHistory ?? []}
                onRestore={(id) => restoreFromHistory(index, id)}
                onClear={() => clearWindowHistory(index)}
                disabled={isStreaming}
              />
              <Button
                variant="outline"
                size="icon"
                onClick={() => duplicateWindow(index)}
                disabled={windows.length >= 20}
                title="Duplicate window"
                aria-label="Duplicate window"
                className="h-8 w-8"
              >
                <Copy className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="icon"
                onClick={() => removeWindow(index)}
                disabled={windows.length <= 1}
                title="Close window"
                aria-label="Close window"
                className="h-8 w-8"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </div>

          <div className="flex items-center gap-2 flex-wrap">
            <LoadPromptDropdown
              projectId={projectId}
              selectedPromptName={windowState?.loadedFromPromptName}
              selectedPromptVersionNumber={
                windowState?.loadedFromPromptVersionNumber
              }
              hasUnsavedChanges={hasUnsavedPromptChanges}
              onLoad={({
                messages,
                promptId,
                promptName,
                promptVersionId,
                promptVersionNumber,
                originalTemplate,
              }) => {
                const vars = extractVariablesFromMessages(messages)
                const newVariables = vars.reduce(
                  (acc, v) => {
                    acc[v] = ''
                    return acc
                  },
                  {} as Record<string, string>,
                )
                updateWindow(index, {
                  messages,
                  variables: newVariables,
                  loadedFromPromptId: promptId,
                  loadedFromPromptName: promptName,
                  loadedFromPromptVersionId: promptVersionId,
                  loadedFromPromptVersionNumber: promptVersionNumber,
                  loadedTemplate: originalTemplate,
                })
              }}
              onUnlink={() => unlinkPrompt(index)}
              disabled={isStreaming}
            />
            <SaveAsPromptDialog
              projectId={projectId}
              messages={windowState?.messages ?? []}
              loadedFromPromptId={windowState?.loadedFromPromptId ?? null}
              loadedFromPromptName={windowState?.loadedFromPromptName ?? null}
              loadedFromPromptVersionNumber={
                windowState?.loadedFromPromptVersionNumber ?? null
              }
              disabled={isStreaming}
              onSuccess={handlePromptSaved}
            />
            <ModelSelector
              value={windowState?.config?.model}
              credentialId={windowState?.config?.credential_id}
              onChange={(model, provider, credentialId) => {
                updateWindow(index, {
                  config: {
                    ...(windowState?.config || {}),
                    model,
                    provider,
                    credential_id: credentialId,
                  },
                })
              }}
              disabled={isStreaming}
              compact
              orgId={orgId}
            />
            <Popover open={configOpen} onOpenChange={setConfigOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  size="icon"
                  className="h-8 w-8"
                  disabled={isStreaming}
                  aria-label="Configure parameters"
                >
                  <Settings2 className="h-4 w-4" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-80" align="start">
                <div className="space-y-4">
                  <h4 className="text-sm font-medium">Model parameters</h4>
                  {PARAMETER_DEFINITIONS.map((def) => {
                    const provider = currentConfig.provider || 'openai'
                    const supportedParams =
                      PROVIDER_PARAMETER_SUPPORT[provider] ||
                      PROVIDER_PARAMETER_SUPPORT.openai
                    const providerSupported = supportedParams.includes(def.key)
                    const enabledKey =
                      `${def.key}_enabled` as keyof ModelConfig

                    return (
                      <ParameterControl
                        key={def.key}
                        definition={def}
                        value={currentConfig[def.key] as number | undefined}
                        enabled={
                          (currentConfig[enabledKey] as boolean) ?? false
                        }
                        onValueChange={(v) => handleConfigChange(def.key, v)}
                        onEnabledChange={(e) =>
                          handleConfigChange(enabledKey, e)
                        }
                        providerSupported={providerSupported}
                        disabled={isStreaming}
                      />
                    )
                  })}
                </div>
              </PopoverContent>
            </Popover>
            {saveStatus && (
              <span className="flex items-center gap-1 text-xs text-muted-foreground">
                {saveStatus === 'saving' && (
                  <>
                    <Loader2 className="h-3 w-3 animate-spin" />
                    <span>Saving...</span>
                  </>
                )}
                {saveStatus === 'unsaved' && (
                  <>
                    <CloudOff className="h-3 w-3" />
                    <span>Unsaved</span>
                  </>
                )}
                {saveStatus === 'saved' && (
                  <>
                    <Cloud className="h-3 w-3 text-green-500" />
                    <span className="text-green-500">Saved</span>
                  </>
                )}
              </span>
            )}
            {windowState?.loadedFromSpanId && (
              <Badge variant="secondary" className="text-xs gap-1 pl-2">
                <FlaskConical className="h-3 w-3" />
                <span
                  className="max-w-[120px] truncate"
                  title={windowState.loadedFromSpanName || undefined}
                >
                  {windowState.loadedFromSpanName || 'Span'}
                </span>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-4 w-4 ml-1 hover:bg-muted-foreground/20"
                  onClick={() => unlinkSpan(index)}
                  title="Unlink from span"
                >
                  <X className="h-3 w-3" />
                </Button>
              </Badge>
            )}
          </div>
        </CardHeader>

        <CardContent className="flex-1 flex flex-col gap-4 overflow-auto">
          <div className="flex-1 min-h-[200px] overflow-auto">
            <MessageEditor
              messages={windowState?.messages ?? []}
              onChange={(messages: ChatMessage[]) => {
                const vars = extractVariablesFromMessages(messages)
                const newVariables = vars.reduce(
                  (acc, v) => {
                    acc[v] = windowState?.variables[v] ?? ''
                    return acc
                  },
                  {} as Record<string, string>,
                )
                updateWindow(index, { messages, variables: newVariables })
              }}
            />
          </div>

          <ToolbarRow
            variables={windowState?.variables ?? {}}
            extractedVariables={extractedVariables}
            onVariablesChange={(vars) => updateWindow(index, { variables: vars })}
            onAddMessage={() => {
              const newMessages = [
                ...(windowState?.messages ?? []),
                createMessage('user', ''),
              ]
              updateWindow(index, { messages: newMessages })
            }}
            disabled={isStreaming}
          />

          <Button
            onClick={handleExecute}
            disabled={isStreaming}
            className="w-full"
            size="lg"
          >
            {isStreaming ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Executing...
              </>
            ) : (
              <>
                <Play className="mr-2 h-4 w-4" />
                Execute
              </>
            )}
          </Button>

          <div className="flex-1 min-h-[200px]">
            <StreamingOutput
              content={
                isStreaming ? content : windowState?.lastOutput || ''
              }
              isStreaming={isStreaming}
              error={error}
              metrics={windowState?.lastMetrics || metrics}
              onStop={abort}
            />
          </div>
        </CardContent>
      </Card>
    </TooltipProvider>
  )
}
