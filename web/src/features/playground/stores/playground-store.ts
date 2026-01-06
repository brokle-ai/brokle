import { create } from 'zustand'
import type { ChatMessage, ModelConfig } from '../types'
import { createMessage } from '../types'

export interface PlaygroundWindow {
  id: string

  // Messages are always the primary data (chat only, no text mode)
  messages: ChatMessage[]

  // Prompt linking (Opik-style - track prompt AND version)
  loadedFromPromptId: string | null        // Which prompt was loaded (if any)
  loadedFromPromptName: string | null      // Display name for "Linked: X"
  loadedFromPromptVersionId: string | null // Version ID (ULID) for precise tracking
  loadedFromPromptVersionNumber: number | null // Version number (e.g., 5) for display
  loadedTemplate: string | null            // Original template JSON for change detection

  // Shared
  variables: Record<string, string>
  config: ModelConfig | null
  createTrace: boolean

  // Last execution (ephemeral - not persisted to DB)
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

  // Execution state (UI only)
  isExecuting: boolean

  // Content snapshot for dirty detection (JSON string of saveable content)
  // isDirty is computed from: currentSnapshot !== lastSavedSnapshot
  lastSavedSnapshot: string | null
}

interface PlaygroundState {
  // Current session ID (from URL)
  currentSessionId: string | null

  // Windows (up to 3)
  windows: PlaygroundWindow[]

  // Shared variables across windows
  sharedVariables: Record<string, string>
  useSharedVariables: boolean

  // Global execution state
  isExecutingAll: boolean

  // Actions - Session
  setCurrentSessionId: (sessionId: string | null) => void

  // Actions - Windows
  addWindow: () => void
  removeWindow: (index: number) => void
  updateWindow: (index: number, updates: Partial<PlaygroundWindow>) => void
  duplicateWindow: (index: number) => void
  setLastSavedSnapshot: (index: number, snapshot: string | null) => void
  setAllSavedSnapshots: () => void

  // Actions - Variables
  setSharedVariables: (variables: Record<string, string>) => void
  toggleSharedVariables: () => void

  // Actions - Execution
  setWindowExecuting: (index: number, isExecuting: boolean) => void
  setWindowOutput: (index: number, output: string, metrics: PlaygroundWindow['lastMetrics']) => void
  setExecutingAll: (isExecuting: boolean) => void

  // Reset
  clearAll: () => void

  // Session Loading - atomic multi-window initialization
  loadWindowsFromSession: (windowsData: Array<{
    messages: ChatMessage[]
    variables?: Record<string, string>
    config?: ModelConfig | null
    loadedFromPromptId?: string | null
    loadedFromPromptName?: string | null
    loadedFromPromptVersionId?: string | null
    loadedFromPromptVersionNumber?: number | null
    loadedTemplate?: string | null
  }>) => void

  // Prompt Linking - unlink a prompt from a window
  unlinkPrompt: (windowIndex: number) => void

  // Load prompt directly into store (for "Try in Playground" feature)
  // This is the new in-memory approach - no session creation until save
  loadFromPrompt: (promptData: {
    messages: ChatMessage[]
    config?: ModelConfig | null
    loadedFromPromptId: string
    loadedFromPromptName: string
    loadedFromPromptVersionId?: string
    loadedFromPromptVersionNumber?: number
    loadedTemplate?: string
  }) => void
}

const createEmptyWindow = (): PlaygroundWindow => ({
  id: crypto.randomUUID(),
  messages: [
    createMessage('system', ''),
    createMessage('user', ''),
  ],
  loadedFromPromptId: null,
  loadedFromPromptName: null,
  loadedFromPromptVersionId: null,
  loadedFromPromptVersionNumber: null,
  loadedTemplate: null,
  variables: {},
  config: null,
  createTrace: false, // Default OFF for playground (ephemeral)
  lastOutput: null,
  lastMetrics: null,
  isExecuting: false,
  lastSavedSnapshot: null, // null = never saved, isDirty computed from comparison
})

/**
 * Creates a JSON snapshot of window's saveable content for dirty comparison.
 * This is the industry standard approach (Notion, Google Docs):
 * isDirty = currentSnapshot !== lastSavedSnapshot
 */
export const createContentSnapshot = (window: PlaygroundWindow): string => {
  // Strip IDs from messages for comparison (IDs are for drag-drop, not content)
  const messagesForSnapshot = window.messages.map(({ role, content }) => ({ role, content }))
  return JSON.stringify({
    messages: messagesForSnapshot,
    variables: window.variables,
    config: window.config,
  })
}

/**
 * Computes whether a window has unsaved changes.
 * Returns false if never saved (no point saving empty state).
 */
export const isWindowDirty = (window: PlaygroundWindow): boolean => {
  if (!window.lastSavedSnapshot) return false // Never saved = not dirty
  return createContentSnapshot(window) !== window.lastSavedSnapshot
}

// Store is now purely in-memory - database is the source of truth
// This store manages UI-only state (execution, dirty flags, etc.)
export const usePlaygroundStore = create<PlaygroundState>()((set, get) => ({
  currentSessionId: null,
  windows: [createEmptyWindow()],
  sharedVariables: {},
  useSharedVariables: false,
  isExecutingAll: false,

  setCurrentSessionId: (sessionId) => set({ currentSessionId: sessionId }),

  addWindow: () => {
    const { windows } = get()
    if (windows.length >= 20) return // Practical limit to prevent memory issues
    const newWindow = createEmptyWindow()
    // Initialize snapshot so isDirty can detect changes
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
    // Note: isDirty is now COMPUTED from lastSavedSnapshot comparison
    // No need to manually track dirty state - just apply updates
    newWindows[index] = {
      ...newWindows[index],
      ...updates,
    }
    set({ windows: newWindows })
  },

  duplicateWindow: (index) => {
    const { windows } = get()
    if (windows.length >= 20) return // Practical limit to prevent memory issues
    const win = windows[index]
    const newWindow: PlaygroundWindow = {
      ...win,
      id: crypto.randomUUID(),
      lastOutput: null,
      lastMetrics: null,
      isExecuting: false,
      lastSavedSnapshot: null, // Will be set below
    }
    // Initialize snapshot so isDirty can detect changes
    newWindow.lastSavedSnapshot = createContentSnapshot(newWindow)
    set({ windows: [...windows, newWindow] })
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
    const newWindows = windows.map(w => ({
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

  setWindowOutput: (index, output, metrics) => {
    const { windows } = get()
    const newWindows = [...windows]
    newWindows[index] = {
      ...newWindows[index],
      lastOutput: output,
      lastMetrics: metrics,
      isExecuting: false,
    }
    set({ windows: newWindows })
  },

  setExecutingAll: (isExecuting) => set({ isExecutingAll: isExecuting }),

  clearAll: () =>
    set({
      currentSessionId: null,
      windows: [createEmptyWindow()],
      sharedVariables: {},
      useSharedVariables: false,
      isExecutingAll: false,
    }),

  // Atomic multi-window initialization from session data
  // This avoids race conditions with addWindow() + updateWindow()
  loadWindowsFromSession: (windowsData) => {
    const newWindows = windowsData.map((data) => {
      // Ensure messages have IDs (migration from old format)
      const messagesWithIds = data.messages.map(msg =>
        msg.id ? msg : { ...msg, id: crypto.randomUUID() }
      )
      const window: PlaygroundWindow = {
        ...createEmptyWindow(),
        messages: messagesWithIds,
        loadedFromPromptId: data.loadedFromPromptId || null,
        loadedFromPromptName: data.loadedFromPromptName || null,
        loadedFromPromptVersionId: data.loadedFromPromptVersionId || null,
        loadedFromPromptVersionNumber: data.loadedFromPromptVersionNumber || null,
        loadedTemplate: data.loadedTemplate || null,
        variables: data.variables || {},
        config: data.config || null,
      }
      // Set snapshot for dirty detection - marks this content as "saved"
      window.lastSavedSnapshot = createContentSnapshot(window)
      return window
    })
    set({ windows: newWindows.length > 0 ? newWindows : [createEmptyWindow()] })
  },

  // Unlink a prompt from a window (keeps content, removes link)
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

  // Load prompt directly into store (for "Try in Playground" feature)
  // This replaces the sessionStorage-based cache transfer
  loadFromPrompt: (promptData) => {
    // Ensure messages have IDs
    const messagesWithIds = promptData.messages.map(msg =>
      msg.id ? msg : { ...msg, id: crypto.randomUUID() }
    )

    const window: PlaygroundWindow = {
      ...createEmptyWindow(),
      messages: messagesWithIds,
      loadedFromPromptId: promptData.loadedFromPromptId,
      loadedFromPromptName: promptData.loadedFromPromptName,
      loadedFromPromptVersionId: promptData.loadedFromPromptVersionId || null,
      loadedFromPromptVersionNumber: promptData.loadedFromPromptVersionNumber || null,
      loadedTemplate: promptData.loadedTemplate || null,
      config: promptData.config || null,
    }

    // Set snapshot for dirty detection
    window.lastSavedSnapshot = createContentSnapshot(window)

    // Replace all windows with this single window containing the prompt
    set({
      currentSessionId: null, // Clear any existing session ID
      windows: [window],
    })
  },
}))
