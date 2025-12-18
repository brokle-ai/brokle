'use client'

import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { Play, X, Copy, Loader2, CloudOff, Cloud, Settings2 } from 'lucide-react'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Slider } from '@/components/ui/slider'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { useProjectOnly } from '@/features/projects'
import { usePlaygroundStore, createContentSnapshot, type PlaygroundWindow } from '../stores/playground-store'
import type { WindowState, ChatMessage } from '../types'
import { createMessage } from '../types'
import { useStreaming } from '../hooks/use-streaming'
import { useUpdateSessionMutation } from '../hooks/use-playground-queries'
import { MessageEditor } from './message-editor'
import { LoadPromptDropdown } from './load-prompt-dropdown'
import { extractVariablesFromMessages } from './variable-editor'
import { ModelSelector } from './model-selector'
import { ToolbarRow } from './toolbar-row'
import { SaveAsPromptDialog, type PromptSavedData } from './save-as-prompt-dialog'
import { StreamingOutput } from './streaming-output'
import type { ExecuteRequest } from '../types'

interface PlaygroundWindowProps {
  index: number
  sessionId?: string
}

const AUTO_SAVE_DELAY = 1500

const CONFIG_PRESETS = {
  creative: {
    temperature: 1.0,
    top_p: 1.0,
    frequency_penalty: 0.5,
    presence_penalty: 0.5,
  },
  balanced: {
    temperature: 0.7,
    top_p: 0.9,
    frequency_penalty: 0.0,
    presence_penalty: 0.0,
  },
  precise: {
    temperature: 0.3,
    top_p: 0.5,
    frequency_penalty: 0.0,
    presence_penalty: 0.0,
  },
}

// Helper to convert store windows to backend WindowState format
const buildWindowsPayload = (windows: PlaygroundWindow[]): WindowState[] => {
  return windows.map((w) => ({
    template: { messages: w.messages },
    variables: w.variables,
    config: w.config || undefined,
    // Include prompt linking metadata so it persists across saves
    loadedFromPromptId: w.loadedFromPromptId || undefined,
    loadedFromPromptName: w.loadedFromPromptName || undefined,
    loadedFromPromptVersionId: w.loadedFromPromptVersionId || undefined,
    loadedFromPromptVersionNumber: w.loadedFromPromptVersionNumber || undefined,
    loadedTemplate: w.loadedTemplate || undefined,
  }))
}

export function PlaygroundWindow({ index, sessionId }: PlaygroundWindowProps) {
  const { currentProject } = useProjectOnly()
  const projectId = currentProject?.id || ''

  // Select window state (object reference will change, but we compute isDirty from content)
  const windowState = usePlaygroundStore((s) => s.windows[index])

  // Select actions (stable references)
  const updateWindow = usePlaygroundStore((s) => s.updateWindow)
  const removeWindow = usePlaygroundStore((s) => s.removeWindow)
  const duplicateWindow = usePlaygroundStore((s) => s.duplicateWindow)
  const setWindowOutput = usePlaygroundStore((s) => s.setWindowOutput)
  const setLastSavedSnapshot = usePlaygroundStore((s) => s.setLastSavedSnapshot)
  const unlinkPrompt = usePlaygroundStore((s) => s.unlinkPrompt)
  const windows = usePlaygroundStore((s) => s.windows)

  // CRITICAL FIX: Compute isDirty from content snapshot comparison
  // This prevents infinite re-renders - isDirty only changes when CONTENT actually changes
  const currentSnapshot = useMemo(() => {
    if (!windowState) return ''
    return createContentSnapshot(windowState)
  }, [windowState])

  const isDirty = useMemo(() => {
    if (!windowState?.lastSavedSnapshot) return false // Never saved = not dirty
    return currentSnapshot !== windowState.lastSavedSnapshot
  }, [currentSnapshot, windowState?.lastSavedSnapshot])

  // Detect if linked prompt has been modified (Opik-style change detection)
  const hasUnsavedPromptChanges = useMemo(() => {
    if (!windowState?.loadedFromPromptVersionId || !windowState?.loadedTemplate) {
      return false
    }
    // Compare current messages (without IDs) with original loaded template
    const currentTemplate = JSON.stringify(
      windowState.messages.map(({ role, content }) => ({ role, content }))
    )
    return currentTemplate !== windowState.loadedTemplate
  }, [windowState?.messages, windowState?.loadedTemplate, windowState?.loadedFromPromptVersionId])

  const updateSessionMutation = useUpdateSessionMutation(projectId, sessionId || '')
  const saveTimerRef = useRef<NodeJS.Timeout | null>(null)

  // Refs to hold latest values for use in callbacks (avoids stale closures)
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

  // Auto-save reads from refs to get latest values (avoids dependency loop)
  const autoSave = useCallback(async () => {
    const { isDirty: currentIsDirty, sessionId: currentSessionId, projectId: currentProjectId, windows: allWindows, currentSnapshot: snapshot } = latestValuesRef.current

    if (!currentSessionId || !currentProjectId || !currentIsDirty || !allWindows?.length) return

    try {
      // Send all windows to prevent overwriting between concurrent edits
      await updateSessionMutation.mutateAsync({
        windows: buildWindowsPayload(allWindows),
      })

      setLastSavedSnapshot(index, snapshot)
    } catch (error) {
      console.error('Auto-save failed:', error)
    }
  }, [updateSessionMutation, setLastSavedSnapshot, index])

  // Store autoSave in a ref for stable access in effects (prevents infinite loops)
  const autoSaveRef = useRef(autoSave)
  useEffect(() => {
    autoSaveRef.current = autoSave
  }, [autoSave])

  // Debounced auto-save - uses ref for autoSave to avoid dependency issues
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

  // Save on unmount (may not complete on page unload, but works for navigation)
  useEffect(() => {
    return () => {
      const { isDirty: currentIsDirty, sessionId: currentSessionId } = latestValuesRef.current
      if (currentIsDirty && currentSessionId) {
        autoSaveRef.current()
      }
    }
  }, [])

  // Using globalThis to avoid shadowing with the local 'window' variable
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      const { isDirty: currentIsDirty, sessionId: currentSessionId } = latestValuesRef.current
      if (currentIsDirty && currentSessionId) {
        autoSaveRef.current()
        e.preventDefault()
        e.returnValue = ''
      }
    }

    globalThis.addEventListener('beforeunload', handleBeforeUnload)
    return () => globalThis.removeEventListener('beforeunload', handleBeforeUnload)
  }, [])

  const { stream, abort, isStreaming, content, error, metrics } = useStreaming({
    onEnd: (finalContent, finalMetrics) => {
      setWindowOutput(index, finalContent, finalMetrics)
    },
  })

  const [configOpen, setConfigOpen] = useState(false)
  const currentConfig = windowState?.config || {}

  const handleConfigChange = (field: string, value: number | undefined) => {
    updateWindow(index, {
      config: { ...currentConfig, [field]: value },
    })
  }

  const applyPreset = (preset: keyof typeof CONFIG_PRESETS) => {
    updateWindow(index, {
      config: { ...currentConfig, ...CONFIG_PRESETS[preset] },
    })
  }

  const handlePromptSaved = useCallback((data: PromptSavedData) => {
    const currentTemplate = JSON.stringify(
      windowState?.messages.map(({ role, content }) => ({ role, content })) ?? []
    )

    updateWindow(index, {
      loadedFromPromptId: data.promptId,
      loadedFromPromptName: data.promptName,
      loadedFromPromptVersionId: data.versionId,
      loadedFromPromptVersionNumber: data.versionNumber,
      loadedTemplate: currentTemplate, // Now synced - removes modified indicator
    })
  }, [windowState?.messages, updateWindow, index])

  const handleExecute = async () => {
    if (!windowState) return

    if (!windowState.config?.model) {
      alert('Please select a model')
      return
    }
    if (!windowState.messages.length || !windowState.messages.some(m => m.content.trim())) {
      alert('Please enter at least one message')
      return
    }

    const request: ExecuteRequest = {
      template: { messages: windowState.messages },
      prompt_type: 'chat',
      variables: windowState.variables,
      config_overrides: windowState.config || undefined,
      session_id: sessionId,
      project_id: projectId,
    }

    await stream(request)
  }

  const extractedVariables = extractVariablesFromMessages(windowState?.messages || [])

  const getSaveStatus = () => {
    if (!sessionId) return null
    if (updateSessionMutation.isPending) return 'saving'
    if (isDirty) return 'unsaved'
    return 'saved'
  }

  const saveStatus = getSaveStatus()

  return (
    <Card className="flex flex-col h-full">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="flex items-center gap-2">
          <LoadPromptDropdown
            projectId={projectId}
            selectedPromptName={windowState?.loadedFromPromptName}
            selectedPromptVersionNumber={windowState?.loadedFromPromptVersionNumber}
            hasUnsavedChanges={hasUnsavedPromptChanges}
            onLoad={({ messages, promptId, promptName, promptVersionId, promptVersionNumber, originalTemplate }) => {
              const vars = extractVariablesFromMessages(messages)
              const newVariables = vars.reduce((acc, v) => {
                acc[v] = ''
                return acc
              }, {} as Record<string, string>)
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
            config={windowState?.config ?? null}
            loadedFromPromptId={windowState?.loadedFromPromptId ?? null}
            loadedFromPromptName={windowState?.loadedFromPromptName ?? null}
            loadedFromPromptVersionNumber={windowState?.loadedFromPromptVersionNumber ?? null}
            disabled={isStreaming}
            onSuccess={handlePromptSaved}
          />
          <ModelSelector
            value={windowState?.config?.model}
            onChange={(model) => {
              updateWindow(index, {
                config: { ...(windowState?.config || {}), model },
              })
            }}
            disabled={isStreaming}
            compact
          />
          <Popover open={configOpen} onOpenChange={setConfigOpen}>
            <PopoverTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8" disabled={isStreaming}>
                <Settings2 className="h-4 w-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80" align="start">
              <Tabs defaultValue="params" className="w-full">
                <TabsList className="w-full">
                  <TabsTrigger value="params" className="flex-1 text-xs">
                    Parameters
                  </TabsTrigger>
                  <TabsTrigger value="presets" className="flex-1 text-xs">
                    Presets
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="params" className="space-y-4 mt-4">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-xs">Temperature</Label>
                      <span className="text-xs text-muted-foreground">
                        {currentConfig.temperature?.toFixed(1) ?? '0.7'}
                      </span>
                    </div>
                    <Slider
                      value={[currentConfig.temperature ?? 0.7]}
                      onValueChange={([value]) => handleConfigChange('temperature', value)}
                      min={0}
                      max={2}
                      step={0.1}
                      disabled={isStreaming}
                    />
                  </div>

                  <div className="space-y-2">
                    <Label className="text-xs">Max Tokens</Label>
                    <Input
                      type="number"
                      value={currentConfig.max_tokens || ''}
                      onChange={(e) =>
                        handleConfigChange('max_tokens', e.target.value ? parseInt(e.target.value) : undefined)
                      }
                      placeholder="Default"
                      disabled={isStreaming}
                      className="h-8 text-xs"
                      min={1}
                      max={128000}
                    />
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-xs">Top P</Label>
                      <span className="text-xs text-muted-foreground">
                        {currentConfig.top_p?.toFixed(2) ?? '1.00'}
                      </span>
                    </div>
                    <Slider
                      value={[currentConfig.top_p ?? 1.0]}
                      onValueChange={([value]) => handleConfigChange('top_p', value)}
                      min={0}
                      max={1}
                      step={0.05}
                      disabled={isStreaming}
                    />
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-xs">Frequency Penalty</Label>
                      <span className="text-xs text-muted-foreground">
                        {currentConfig.frequency_penalty?.toFixed(1) ?? '0.0'}
                      </span>
                    </div>
                    <Slider
                      value={[currentConfig.frequency_penalty ?? 0.0]}
                      onValueChange={([value]) => handleConfigChange('frequency_penalty', value)}
                      min={-2}
                      max={2}
                      step={0.1}
                      disabled={isStreaming}
                    />
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-xs">Presence Penalty</Label>
                      <span className="text-xs text-muted-foreground">
                        {currentConfig.presence_penalty?.toFixed(1) ?? '0.0'}
                      </span>
                    </div>
                    <Slider
                      value={[currentConfig.presence_penalty ?? 0.0]}
                      onValueChange={([value]) => handleConfigChange('presence_penalty', value)}
                      min={-2}
                      max={2}
                      step={0.1}
                      disabled={isStreaming}
                    />
                  </div>
                </TabsContent>

                <TabsContent value="presets" className="space-y-2 mt-4">
                  <p className="text-xs text-muted-foreground mb-3">
                    Quick parameter presets
                  </p>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => applyPreset('creative')}
                    disabled={isStreaming}
                    className="w-full justify-start text-xs h-8"
                  >
                    <span className="font-medium">Creative</span>
                    <span className="ml-auto text-muted-foreground">temp=1.0</span>
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => applyPreset('balanced')}
                    disabled={isStreaming}
                    className="w-full justify-start text-xs h-8"
                  >
                    <span className="font-medium">Balanced</span>
                    <span className="ml-auto text-muted-foreground">temp=0.7</span>
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => applyPreset('precise')}
                    disabled={isStreaming}
                    className="w-full justify-start text-xs h-8"
                  >
                    <span className="font-medium">Precise</span>
                    <span className="ml-auto text-muted-foreground">temp=0.3</span>
                  </Button>
                </TabsContent>
              </Tabs>
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
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => duplicateWindow(index)}
            disabled={windows.length >= 20}
            title="Duplicate window"
          >
            <Copy className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => removeWindow(index)}
            disabled={windows.length <= 1}
            title="Remove window"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>

      <CardContent className="flex-1 flex flex-col gap-4 overflow-auto">
        <div className="flex-1 min-h-[200px] overflow-auto">
          <MessageEditor
            messages={windowState?.messages ?? []}
            onChange={(messages: ChatMessage[]) => {
            const vars = extractVariablesFromMessages(messages)
            const newVariables = vars.reduce((acc, v) => {
              acc[v] = windowState?.variables[v] ?? ''
              return acc
            }, {} as Record<string, string>)
            updateWindow(index, { messages, variables: newVariables })
          }}
          />
        </div>

        <ToolbarRow
          variables={windowState?.variables ?? {}}
          extractedVariables={extractedVariables}
          onVariablesChange={(vars) => updateWindow(index, { variables: vars })}
          onAddMessage={() => {
            const newMessages = [...(windowState?.messages ?? []), createMessage('user', '')]
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
            content={isStreaming ? content : windowState?.lastOutput || ''}
            isStreaming={isStreaming}
            error={error}
            metrics={windowState?.lastMetrics || metrics}
            onStop={abort}
          />
        </div>
      </CardContent>
    </Card>
  )
}
