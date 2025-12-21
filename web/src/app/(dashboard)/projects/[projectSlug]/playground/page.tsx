'use client'

import { useState, useRef, useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import {
  Plus,
  RotateCcw,
  PlayCircle,
  Bookmark,
  FolderOpen,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Skeleton } from '@/components/ui/skeleton'
import { PlaygroundWindow } from '@/features/playground/components/playground-window'
import { SaveSessionDialog } from '@/features/playground/components/save-session-dialog'
import { SavedSessionsSidebar } from '@/features/playground/components/saved-sessions-sidebar'
import { usePlaygroundStore } from '@/features/playground/stores/playground-store'
import { useProjectOnly } from '@/features/projects/hooks/use-project-only'

/**
 * Main Playground Page - In-Memory Only
 *
 * This page renders the playground UI directly from the Zustand store.
 * No session is created until the user explicitly saves.
 *
 * URL: /projects/{projectSlug}/playground
 */
export default function PlaygroundPage() {
  const params = useParams<{ projectSlug: string }>()
  const router = useRouter()

  const { currentProject, isLoading: projectLoading } = useProjectOnly()
  const projectId = currentProject?.id

  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [saveDialogOpen, setSaveDialogOpen] = useState(false)

  // Store state - all in-memory, no database
  const windows = usePlaygroundStore((s) => s.windows)
  const addWindow = usePlaygroundStore((s) => s.addWindow)
  const clearAll = usePlaygroundStore((s) => s.clearAll)
  const isExecutingAll = usePlaygroundStore((s) => s.isExecutingAll)
  const setExecutingAll = usePlaygroundStore((s) => s.setExecutingAll)

  // Refs to track window execute functions for Execute All
  const windowExecuteRefs = useRef<Map<number, () => Promise<void>>>(new Map())

  const registerWindowExecute = useCallback((index: number, executeFn: () => Promise<void>) => {
    windowExecuteRefs.current.set(index, executeFn)
  }, [])

  const unregisterWindowExecute = useCallback((index: number) => {
    windowExecuteRefs.current.delete(index)
  }, [])

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
  }

  if (projectLoading) {
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
              <h1 className="text-2xl font-bold tracking-tight">Playground</h1>
              <p className="text-muted-foreground">
                Test prompts and configurations - changes are in-memory until saved
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setSidebarOpen(true)}
              >
                <FolderOpen className="mr-2 h-4 w-4" />
                Sessions
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setSaveDialogOpen(true)}
              >
                <Bookmark className="mr-2 h-4 w-4" />
                Save Session
              </Button>
              <Button variant="outline" size="sm" onClick={clearAll}>
                <RotateCcw className="mr-2 h-4 w-4" />
                Reset
              </Button>
            </div>
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

          {/* Braintrust-style adaptive layout */}
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
          onNewSession={handleNewSession}
        />
      )}

      {/* Save Session Dialog - now creates session on save */}
      {projectId && (
        <SaveSessionDialog
          open={saveDialogOpen}
          onOpenChange={setSaveDialogOpen}
          projectId={projectId}
          onSuccess={(sessionId) => {
            router.push(`/projects/${params.projectSlug}/playground/${sessionId}`)
          }}
        />
      )}
    </>
  )
}
