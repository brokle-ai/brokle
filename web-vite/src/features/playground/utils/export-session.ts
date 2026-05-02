import type { PlaygroundWindow } from '../stores/playground-store'
import type { PlaygroundSession, ModelConfig } from '../types'

export interface ExportedSession {
  version: '1.0'
  exportedAt: string
  session: {
    id: string
    name?: string
    description?: string
    tags: string[]
    createdAt: string
    updatedAt: string
  }
  windows: ExportedWindow[]
  sharedVariables?: Record<string, string>
}

export interface ExportedWindow {
  name: string
  messages: Array<{
    role: 'system' | 'user' | 'assistant'
    content: string
  }>
  variables: Record<string, string>
  config: ModelConfig | null
  linkedPrompt?: {
    id: string
    name: string
    versionId?: string
    versionNumber?: number
  }
  linkedSpan?: {
    spanId: string
    spanName: string
    traceId: string
  }
}

/**
 * Export session and windows to a downloadable JSON file.
 */
export function exportSessionToJSON(
  session: PlaygroundSession,
  windows: PlaygroundWindow[],
  sharedVariables?: Record<string, string>,
): void {
  const exported: ExportedSession = {
    version: '1.0',
    exportedAt: new Date().toISOString(),
    session: {
      id: session.id,
      name: session.name,
      description: session.description,
      tags: session.tags,
      createdAt: session.created_at,
      updatedAt: session.updated_at,
    },
    windows: windows.map((window) => ({
      name: window.name,
      messages: window.messages.map((m) => ({
        role: m.role,
        content: m.content,
      })),
      variables: window.variables,
      config: window.config,
      ...(window.loadedFromPromptId && {
        linkedPrompt: {
          id: window.loadedFromPromptId,
          name: window.loadedFromPromptName || '',
          versionId: window.loadedFromPromptVersionId || undefined,
          versionNumber: window.loadedFromPromptVersionNumber || undefined,
        },
      }),
      ...(window.loadedFromSpanId && {
        linkedSpan: {
          spanId: window.loadedFromSpanId,
          spanName: window.loadedFromSpanName || '',
          traceId: window.loadedFromTraceId || '',
        },
      }),
    })),
    ...(sharedVariables &&
      Object.keys(sharedVariables).length > 0 && { sharedVariables }),
  }

  const blob = new Blob([JSON.stringify(exported, null, 2)], {
    type: 'application/json',
  })
  const url = URL.createObjectURL(blob)

  const sessionName = session.name?.replace(/[^a-z0-9]/gi, '_') || 'session'
  const date = new Date().toISOString().split('T')[0]
  const filename = `playground_${sessionName}_${date}.json`

  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

export function exportWindowsToJSON(
  windows: PlaygroundWindow[],
  sharedVariables?: Record<string, string>,
  name?: string,
): void {
  const exported = {
    version: '1.0',
    exportedAt: new Date().toISOString(),
    windows: windows.map((window) => ({
      name: window.name,
      messages: window.messages.map((m) => ({
        role: m.role,
        content: m.content,
      })),
      variables: window.variables,
      config: window.config,
      ...(window.loadedFromPromptId && {
        linkedPrompt: {
          id: window.loadedFromPromptId,
          name: window.loadedFromPromptName || '',
          versionId: window.loadedFromPromptVersionId || undefined,
          versionNumber: window.loadedFromPromptVersionNumber || undefined,
        },
      }),
      ...(window.loadedFromSpanId && {
        linkedSpan: {
          spanId: window.loadedFromSpanId,
          spanName: window.loadedFromSpanName || '',
          traceId: window.loadedFromTraceId || '',
        },
      }),
    })),
    ...(sharedVariables &&
      Object.keys(sharedVariables).length > 0 && { sharedVariables }),
  }

  const blob = new Blob([JSON.stringify(exported, null, 2)], {
    type: 'application/json',
  })
  const url = URL.createObjectURL(blob)

  const sessionName = name?.replace(/[^a-z0-9]/gi, '_') || 'playground'
  const date = new Date().toISOString().split('T')[0]
  const filename = `${sessionName}_${date}.json`

  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

export function validateImportedSession(
  data: unknown,
): data is ExportedSession {
  if (!data || typeof data !== 'object') return false

  const session = data as Record<string, unknown>

  if (session.version !== '1.0') return false
  if (!Array.isArray(session.windows)) return false

  for (const window of session.windows as unknown[]) {
    if (!window || typeof window !== 'object') return false
    const w = window as Record<string, unknown>
    if (!Array.isArray(w.messages)) return false
    for (const msg of w.messages as unknown[]) {
      if (!msg || typeof msg !== 'object') return false
      const m = msg as Record<string, unknown>
      if (typeof m.role !== 'string' || typeof m.content !== 'string')
        return false
    }
  }

  return true
}

export function parseImportedSession(
  jsonString: string,
): ExportedSession | null {
  try {
    const data = JSON.parse(jsonString)
    if (validateImportedSession(data)) {
      return data
    }
    return null
  } catch {
    return null
  }
}
