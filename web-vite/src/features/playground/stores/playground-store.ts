import { create } from 'zustand'
import { registerSessionReset } from '@/lib/auth/session'
import type {
  ChatMessage,
  ModelConfig,
  RunHistoryEntry,
  ToolCall,
} from '../types'
import { createMessage } from '../types'

export interface PlaygroundWindow {
  id: string
  name: string

  messages: ChatMessage[]

  loadedFromPromptId: string | null
  loadedFromPromptName: string | null
  loadedFromPromptVersionId: string | null
  loadedFromPromptVersionNumber: number | null
  loadedTemplate: string | null

  loadedFromSpanId: string | null
  loadedFromSpanName: string | null
  loadedFromTraceId: string | null

  variables: Record<string, string>
  config: ModelConfig | null
  createTrace: boolean

  lastOutput: string | null
  lastMetrics: {
    model?: string
    prompt_tokens?: number
    completion_tokens?: number
    total_tokens?: number
    cost?: number
    ttft_ms?: number
    total_duration_ms?: number
  } | null
  lastToolCalls: ToolCall[] | null

  runHistory: RunHistoryEntry[]

  isExecuting: boolean

  lastSavedSnapshot: string | null
}

interface PlaygroundState {
  currentSessionId: string | null

  windows: PlaygroundWindow[]

  sharedVariables: Record<string, string>
  useSharedVariables: boolean

  isExecutingAll: boolean

  setCurrentSessionId: (sessionId: string | null) => void

  addWindow: () => void
  removeWindow: (index: number) => void
  updateWindow: (index: number, updates: Partial<PlaygroundWindow>) => void
  duplicateWindow: (index: number) => void
  renameWindow: (index: number, name: string) => void
  setLastSavedSnapshot: (index: number, snapshot: string | null) => void
  setAllSavedSnapshots: () => void

  addRunToHistory: (
    index: number,
    entry: Omit<RunHistoryEntry, 'id' | 'timestamp' | 'isStale'>,
  ) => void
  restoreFromHistory: (windowIndex: number, historyId: string) => void
  clearWindowHistory: (index: number) => void
  markHistoryAsStale: (index: number) => void

  setSharedVariables: (variables: Record<string, string>) => void
  toggleSharedVariables: () => void

  setWindowExecuting: (index: number, isExecuting: boolean) => void
  setWindowOutput: (
    index: number,
    output: string,
    metrics: PlaygroundWindow['lastMetrics'],
    inputSnapshot?: {
      messages: ChatMessage[]
      variables: Record<string, string>
      config: ModelConfig | null
    },
    toolCalls?: ToolCall[] | null,
  ) => void
  setExecutingAll: (isExecuting: boolean) => void

  clearAll: () => void

  loadWindowsFromSession: (
    windowsData: Array<{
      messages: ChatMessage[]
      variables?: Record<string, string>
      config?: ModelConfig | null
      loadedFromPromptId?: string | null
      loadedFromPromptName?: string | null
      loadedFromPromptVersionId?: string | null
      loadedFromPromptVersionNumber?: number | null
      loadedTemplate?: string | null
    }>,
  ) => void

  unlinkPrompt: (windowIndex: number) => void

  loadFromPrompt: (promptData: {
    messages: ChatMessage[]
    config?: ModelConfig | null
    loadedFromPromptId: string
    loadedFromPromptName: string
    loadedFromPromptVersionId?: string
    loadedFromPromptVersionNumber?: number
    loadedTemplate?: string
  }) => void

  loadFromSpan: (spanData: {
    messages: ChatMessage[]
    config?: ModelConfig | null
    loadedFromSpanId: string
    loadedFromSpanName: string
    loadedFromTraceId: string
  }) => void

  unlinkSpan: (windowIndex: number) => void
}

const createEmptyWindow = (index?: number): PlaygroundWindow => ({
  id: crypto.randomUUID(),
  name: index !== undefined ? `Window ${index + 1}` : 'New Window',
  messages: [createMessage('system', ''), createMessage('user', '')],
  loadedFromPromptId: null,
  loadedFromPromptName: null,
  loadedFromPromptVersionId: null,
  loadedFromPromptVersionNumber: null,
  loadedTemplate: null,
  loadedFromSpanId: null,
  loadedFromSpanName: null,
  loadedFromTraceId: null,
  variables: {},
  config: null,
  createTrace: false,
  lastOutput: null,
  lastMetrics: null,
  lastToolCalls: null,
  runHistory: [],
  isExecuting: false,
  lastSavedSnapshot: null,
})

/**
 * JSON snapshot of a window's saveable content for dirty-comparison.
 */
export const createContentSnapshot = (window: PlaygroundWindow): string => {
  const messagesForSnapshot = window.messages.map(({ role, content }) => ({
    role,
    content,
  }))
  return JSON.stringify({
    messages: messagesForSnapshot,
    variables: window.variables,
    config: window.config,
  })
}

export const isWindowDirty = (window: PlaygroundWindow): boolean => {
  if (!window.lastSavedSnapshot) return false
  return createContentSnapshot(window) !== window.lastSavedSnapshot
}

export const usePlaygroundStore = create<PlaygroundState>()((set, get) => ({
  currentSessionId: null,
  windows: [createEmptyWindow(0)],
  sharedVariables: {},
  useSharedVariables: false,
  isExecutingAll: false,

  setCurrentSessionId: (sessionId) => set({ currentSessionId: sessionId }),

  addWindow: () => {
    const { windows } = get()
    if (windows.length >= 20) return
    const newWindow = createEmptyWindow(windows.length)
    newWindow.lastSavedSnapshot = createContentSnapshot(newWindow)
    set({ windows: [...windows, newWindow] })
  },

  removeWindow: (index) => {
    const { windows } = get()
    if (windows.length <= 1) return
    set({ windows: windows.filter((_, i) => i !== index) })
  },

  updateWindow: (index, updates) => {
    const { windows } = get()
    const newWindows = [...windows]
    newWindows[index] = { ...newWindows[index], ...updates }
    set({ windows: newWindows })
  },

  duplicateWindow: (index) => {
    const { windows } = get()
    if (windows.length >= 20) return
    const win = windows[index]
    const newWindow: PlaygroundWindow = {
      ...win,
      id: crypto.randomUUID(),
      name: `${win.name} (copy)`,
      lastOutput: null,
      lastMetrics: null,
      runHistory: [],
      isExecuting: false,
      lastSavedSnapshot: null,
    }
    newWindow.lastSavedSnapshot = createContentSnapshot(newWindow)
    set({ windows: [...windows, newWindow] })
  },

  renameWindow: (index, name) => {
    const { windows } = get()
    const newWindows = [...windows]
    if (newWindows[index]) {
      newWindows[index] = { ...newWindows[index], name }
    }
    set({ windows: newWindows })
  },

  setLastSavedSnapshot: (index, snapshot) => {
    const { windows } = get()
    const newWindows = [...windows]
    if (newWindows[index]) {
      newWindows[index] = { ...newWindows[index], lastSavedSnapshot: snapshot }
    }
    set({ windows: newWindows })
  },

  setAllSavedSnapshots: () => {
    const { windows } = get()
    const newWindows = windows.map((w) => ({
      ...w,
      lastSavedSnapshot: createContentSnapshot(w),
    }))
    set({ windows: newWindows })
  },

  setSharedVariables: (variables) => set({ sharedVariables: variables }),

  toggleSharedVariables: () =>
    set((state) => ({ useSharedVariables: !state.useSharedVariables })),

  setWindowExecuting: (index, isExecuting) => {
    const { windows } = get()
    const newWindows = [...windows]
    newWindows[index] = { ...newWindows[index], isExecuting }
    set({ windows: newWindows })
  },

  setWindowOutput: (index, output, metrics, inputSnapshot, toolCalls) => {
    const { windows } = get()
    const window = windows[index]
    if (!window) return

    const historyMessages = inputSnapshot?.messages ?? window.messages
    const historyVariables = inputSnapshot?.variables ?? window.variables
    const historyConfig = inputSnapshot?.config ?? window.config

    const historyEntry: RunHistoryEntry = {
      id: crypto.randomUUID(),
      content: output,
      metrics: metrics
        ? {
            prompt_tokens: metrics.prompt_tokens,
            completion_tokens: metrics.completion_tokens,
            total_tokens: metrics.total_tokens,
            cost: metrics.cost,
            latency_ms: metrics.total_duration_ms,
            ttft_ms: metrics.ttft_ms,
            model: metrics.model,
          }
        : null,
      timestamp: new Date().toISOString(),
      isStale: false,
      messages: historyMessages.map((m) => ({ ...m })),
      variables: { ...historyVariables },
      config: historyConfig ? { ...historyConfig } : null,
    }

    const newHistory = [historyEntry, ...window.runHistory].slice(0, 10)

    const newWindows = [...windows]
    newWindows[index] = {
      ...window,
      lastOutput: output,
      lastMetrics: metrics,
      lastToolCalls: toolCalls || null,
      runHistory: newHistory,
      isExecuting: false,
    }
    set({ windows: newWindows })
  },

  setExecutingAll: (isExecuting) => set({ isExecutingAll: isExecuting }),

  addRunToHistory: (index, entry) => {
    const { windows } = get()
    const window = windows[index]
    if (!window) return

    const historyEntry: RunHistoryEntry = {
      ...entry,
      id: crypto.randomUUID(),
      timestamp: new Date().toISOString(),
      isStale: false,
    }

    const newHistory = [historyEntry, ...window.runHistory].slice(0, 10)
    const newWindows = [...windows]
    newWindows[index] = { ...window, runHistory: newHistory }
    set({ windows: newWindows })
  },

  restoreFromHistory: (windowIndex, historyId) => {
    const { windows } = get()
    const window = windows[windowIndex]
    if (!window) return

    const historyEntry = window.runHistory.find((h) => h.id === historyId)
    if (!historyEntry) return

    const newWindows = [...windows]
    newWindows[windowIndex] = {
      ...window,
      messages: historyEntry.messages.map((m) => ({
        ...m,
        id: crypto.randomUUID(),
      })),
      variables: { ...historyEntry.variables },
      config: historyEntry.config ? { ...historyEntry.config } : null,
      lastOutput: historyEntry.content,
      lastMetrics: historyEntry.metrics
        ? {
            model: historyEntry.metrics.model,
            prompt_tokens: historyEntry.metrics.prompt_tokens,
            completion_tokens: historyEntry.metrics.completion_tokens,
            total_tokens: historyEntry.metrics.total_tokens,
            cost: historyEntry.metrics.cost,
            ttft_ms: historyEntry.metrics.ttft_ms,
            total_duration_ms: historyEntry.metrics.latency_ms,
          }
        : null,
    }
    set({ windows: newWindows })
  },

  clearWindowHistory: (index) => {
    const { windows } = get()
    const newWindows = [...windows]
    if (newWindows[index]) {
      newWindows[index] = { ...newWindows[index], runHistory: [] }
    }
    set({ windows: newWindows })
  },

  markHistoryAsStale: (index) => {
    const { windows } = get()
    const window = windows[index]
    if (!window || window.runHistory.length === 0) return

    const newWindows = [...windows]
    newWindows[index] = {
      ...window,
      runHistory: window.runHistory.map((entry) => ({
        ...entry,
        isStale: true,
      })),
    }
    set({ windows: newWindows })
  },

  clearAll: () =>
    set({
      currentSessionId: null,
      windows: [createEmptyWindow(0)],
      sharedVariables: {},
      useSharedVariables: false,
      isExecutingAll: false,
    }),

  loadWindowsFromSession: (windowsData) => {
    const newWindows = windowsData.map((data, index) => {
      const messagesWithIds = data.messages.map((msg) =>
        msg.id ? msg : { ...msg, id: crypto.randomUUID() },
      )
      const window: PlaygroundWindow = {
        ...createEmptyWindow(index),
        messages: messagesWithIds,
        loadedFromPromptId: data.loadedFromPromptId || null,
        loadedFromPromptName: data.loadedFromPromptName || null,
        loadedFromPromptVersionId: data.loadedFromPromptVersionId || null,
        loadedFromPromptVersionNumber:
          data.loadedFromPromptVersionNumber || null,
        loadedTemplate: data.loadedTemplate || null,
        variables: data.variables || {},
        config: data.config || null,
      }
      window.lastSavedSnapshot = createContentSnapshot(window)
      return window
    })
    set({
      windows: newWindows.length > 0 ? newWindows : [createEmptyWindow(0)],
    })
  },

  unlinkPrompt: (windowIndex) => {
    const { windows } = get()
    const newWindows = [...windows]
    if (newWindows[windowIndex]) {
      newWindows[windowIndex] = {
        ...newWindows[windowIndex],
        loadedFromPromptId: null,
        loadedFromPromptName: null,
        loadedFromPromptVersionId: null,
        loadedFromPromptVersionNumber: null,
        loadedTemplate: null,
      }
    }
    set({ windows: newWindows })
  },

  loadFromPrompt: (promptData) => {
    const messagesWithIds = promptData.messages.map((msg) =>
      msg.id ? msg : { ...msg, id: crypto.randomUUID() },
    )

    const window: PlaygroundWindow = {
      ...createEmptyWindow(0),
      name: promptData.loadedFromPromptName || 'Window 1',
      messages: messagesWithIds,
      loadedFromPromptId: promptData.loadedFromPromptId,
      loadedFromPromptName: promptData.loadedFromPromptName,
      loadedFromPromptVersionId: promptData.loadedFromPromptVersionId || null,
      loadedFromPromptVersionNumber:
        promptData.loadedFromPromptVersionNumber || null,
      loadedTemplate: promptData.loadedTemplate || null,
      config: promptData.config || null,
    }

    window.lastSavedSnapshot = createContentSnapshot(window)

    set({
      currentSessionId: null,
      windows: [window],
    })
  },

  loadFromSpan: (spanData) => {
    const messagesWithIds = spanData.messages.map((msg) =>
      msg.id ? msg : { ...msg, id: crypto.randomUUID() },
    )

    const window: PlaygroundWindow = {
      ...createEmptyWindow(0),
      name: spanData.loadedFromSpanName || 'Window 1',
      messages: messagesWithIds,
      loadedFromSpanId: spanData.loadedFromSpanId,
      loadedFromSpanName: spanData.loadedFromSpanName,
      loadedFromTraceId: spanData.loadedFromTraceId,
      config: spanData.config || null,
    }

    window.lastSavedSnapshot = createContentSnapshot(window)

    set({
      currentSessionId: null,
      windows: [window],
    })
  },

  unlinkSpan: (windowIndex) => {
    const { windows } = get()
    const newWindows = [...windows]
    if (newWindows[windowIndex]) {
      newWindows[windowIndex] = {
        ...newWindows[windowIndex],
        loadedFromSpanId: null,
        loadedFromSpanName: null,
        loadedFromTraceId: null,
      }
    }
    set({ windows: newWindows })
  },
}))

// Account-scoped state — wipe when crossing the auth boundary so the
// next user doesn't inherit the previous user's windows / run
// history / shared variables. Module-load side effect (matches
// Zustand's own `create()` pattern of self-initialising on import).
registerSessionReset(() => usePlaygroundStore.getState().clearAll())
