'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import {
  Plus,
  PlayCircle,
  FolderOpen,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Skeleton } from '@/components/ui/skeleton'
import { PlaygroundWindow } from '@/features/playground/components/playground-window'
import { SavedSessionsSidebar } from '@/features/playground/components/saved-sessions-sidebar'
import { useSessionQuery } from '@/features/playground/hooks/use-playground-queries'
import { usePlaygroundStore } from '@/features/playground/stores/playground-store'
import { useProjectOnly } from '@/features/projects/hooks/use-project-only'
import type { ModelConfig, ChatTemplate, ChatMessage } from '@/features/playground/types'
import { createMessage } from '@/features/playground/types'

/**
 * Saved Session Page - Load from Database
 *
 * This page loads a saved session from the database and allows editing.
 * Changes are persisted back to the database.
 *
 * URL: /projects/{projectSlug}/playground/{sessionId}
 */
export default function PlaygroundSessionPage() {
  const params = useParams<{ projectSlug: string; sessionId: string }>()
  const router = useRouter()

  const { currentProject, isLoading: projectLoading } = useProjectOnly()
  const projectId = currentProject?.id

  const [sidebarOpen, setSidebarOpen] = useState(false)

  const windows = usePlaygroundStore((s) => s.windows)
  const addWindow = usePlaygroundStore((s) => s.addWindow)
  const clearAll = usePlaygroundStore((s) => s.clearAll)
  const isExecutingAll = usePlaygroundStore((s) => s.isExecutingAll)
  const setExecutingAll = usePlaygroundStore((s) => s.setExecutingAll)
  const setCurrentSessionId = usePlaygroundStore((s) => s.setCurrentSessionId)
  const loadWindowsFromSession = usePlaygroundStore((s) => s.loadWindowsFromSession)

  // Refs to track window execute functions for Execute All
  const windowExecuteRefs = useRef<Map<number, () => Promise<void>>>(new Map())

  const registerWindowExecute = useCallback((index: number, executeFn: () => Promise<void>) => {
    windowExecuteRefs.current.set(index, executeFn)
  }, [])

  const unregisterWindowExecute = useCallback((index: number) => {
    windowExecuteRefs.current.delete(index)
  }, [])

  const {
    data: session,
    isLoading: sessionLoading,
    error: sessionError,
  } = useSessionQuery(projectId, params.sessionId, {
    enabled: !!projectId && !!params.sessionId,
  })


  // Ref to track if we've loaded the session into the store
  const hasLoadedRef = useRef(false)

  // Helper to extract messages from various template formats
  const extractMessages = useCallback((template: unknown): ChatMessage[] => {
    if (!template) {
      return [createMessage('system', ''), createMessage('user', '')]
    }
    // Chat template format
    if (typeof template === 'object' && 'messages' in (template as ChatTemplate)) {
      const chatTemplate = template as ChatTemplate
      return chatTemplate.messages.map((msg) =>
        msg.id ? msg : { ...msg, id: crypto.randomUUID() }
      )
    }
    // Text template format (backwards compatibility) - convert to single user message
    if (typeof template === 'object' && 'content' in (template as { content: string })) {
      const textTemplate = template as { content: string }
      return [
        createMessage('user', textTemplate.content || ''),
      ]
    }
    return [createMessage('system', ''), createMessage('user', '')]
  }, [])

  useEffect(() => {
    if (!session || hasLoadedRef.current) return

    setCurrentSessionId(params.sessionId)

    if (session.windows && session.windows.length > 0) {
      loadWindowsFromSession(
        session.windows.map((w) => ({
          messages: extractMessages(w.template),
          variables: w.variables,
          config: (w.config as ModelConfig) || null,
          loadedFromPromptId: w.loadedFromPromptId,
          loadedFromPromptName: w.loadedFromPromptName,
          loadedFromPromptVersionId: w.loadedFromPromptVersionId,
          loadedFromPromptVersionNumber: w.loadedFromPromptVersionNumber,
          loadedTemplate: w.loadedTemplate,
        }))
      )
    }

    hasLoadedRef.current = true
  }, [session, params.sessionId, setCurrentSessionId, loadWindowsFromSession, extractMessages])

  useEffect(() => {
    if (sessionError) {
      router.push(`/projects/${params.projectSlug}/playground`)
    }
  }, [sessionError, router, params.projectSlug])

  useEffect(() => {
    return () => {
      setCurrentSessionId(null)
      hasLoadedRef.current = false
    }
  }, [params.sessionId, setCurrentSessionId])

  const handleExecuteAll = async () => {
    setExecutingAll(true)

    try {
      // Get all registered execute functions and run in parallel
      const executeFns = Array.from(windowExecuteRefs.current.entries())
        .sort(([a], [b]) => a - b) // Sort by index
        .map(([, fn]) => fn)

      // Execute all windows in parallel (each streams independently)
      await Promise.all(executeFns.map(fn => fn()))
    } catch (error) {
      console.error('Execute all failed:', error)
    } finally {
      setExecutingAll(false)
    }
  }

  const handleNewSession = () => {
    clearAll()
    router.push(`/projects/${params.projectSlug}/playground`)
  }

  if (projectLoading || sessionLoading) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <Skeleton className="h-8 w-48" />
                <Skeleton className="h-4 w-64 mt-2" />
              </div>
              <div className="flex gap-2">
                <Skeleton className="h-9 w-24" />
                <Skeleton className="h-9 w-24" />
              </div>
            </div>
            <Skeleton className="h-[600px] w-full" />
          </div>
        </Main>
      </>
    )
  }

  return (
    <>
      <DashboardHeader />
      <Main>
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold tracking-tight">
                {session?.name || 'Playground'}
              </h1>
              <p className="text-muted-foreground">
                {session?.description || 'Saved session'}
              </p>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setSidebarOpen(true)}
            >
              <FolderOpen className="mr-2 h-4 w-4" />
              Sessions
            </Button>
          </div>

          <div className="flex items-center justify-between">
            <Button
              variant="outline"
              size="sm"
              onClick={addWindow}
              disabled={windows.length >= 20}
            >
              <Plus className="mr-2 h-4 w-4" />
              Add Window
            </Button>

            {windows.length > 1 && (
              <Button onClick={handleExecuteAll} disabled={isExecutingAll}>
                <PlayCircle className="mr-2 h-4 w-4" />
                Execute All
              </Button>
            )}
          </div>

          {/* Braintrust-style: Grid for 1-3 windows, horizontal scroll for 4+ */}
          {windows.length <= 3 ? (
            // 1-3 windows: Equal width grid, fills available space
            <div
              className={`grid gap-4 min-h-[calc(100vh-280px)] ${
                windows.length === 1
                  ? 'grid-cols-1'
                  : windows.length === 2
                    ? 'grid-cols-2'
                    : 'grid-cols-3'
              }`}
            >
              {windows.map((_, index) => (
                <PlaygroundWindow
                  key={windows[index].id}
                  index={index}
                  sessionId={params.sessionId}
                  onRegisterExecute={(fn) => registerWindowExecute(index, fn)}
                  onUnregisterExecute={() => unregisterWindowExecute(index)}
                />
              ))}
            </div>
          ) : (
            // 4+ windows: Fixed widths with horizontal scroll
            <div className="flex flex-nowrap overflow-x-auto gap-4 pb-4 min-h-[calc(100vh-280px)] snap-x snap-mandatory scrollbar-thin scrollbar-thumb-muted-foreground/20 scrollbar-track-transparent">
              {windows.map((_, index) => (
                <div
                  key={windows[index].id}
                  className="flex-none w-80 sm:w-96 lg:w-[420px] xl:w-[480px] snap-start"
                >
                  <PlaygroundWindow
                    index={index}
                    sessionId={params.sessionId}
                    onRegisterExecute={(fn) => registerWindowExecute(index, fn)}
                    onUnregisterExecute={() => unregisterWindowExecute(index)}
                  />
                </div>
              ))}
            </div>
          )}

        </div>
      </Main>

      {projectId && (
        <SavedSessionsSidebar
          open={sidebarOpen}
          onOpenChange={setSidebarOpen}
          projectId={projectId}
          projectSlug={params.projectSlug}
          currentSessionId={params.sessionId}
          onNewSession={handleNewSession}
        />
      )}

    </>
  )
}
