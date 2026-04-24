import { createFileRoute } from '@tanstack/react-router'
import { useState, useRef, useCallback } from 'react'
import { z } from 'zod'
import {
  Plus,
  RotateCcw,
  PlayCircle,
  Bookmark,
  FolderOpen,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PlaygroundWindow } from '@/features/playground/components/playground-window'
import { SaveSessionDialog } from '@/features/playground/components/save-session-dialog'
import { SavedSessionsSidebar } from '@/features/playground/components/saved-sessions-sidebar'
import { SharedVariablesPanel } from '@/features/playground/components/shared-variables-panel'
import { usePlaygroundStore } from '@/features/playground/stores/playground-store'
import { usePlaygroundKeyboard } from '@/features/playground/hooks/use-playground-keyboard'

// The playground is a leaf route that keeps its state in Zustand and
// never depends on URL search params — any hostile value is silently
// dropped via `.catch({})`.
const searchSchema = z.object({}).catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/playground',
)({
  validateSearch: searchSchema,
  component: PlaygroundPage,
})

function PlaygroundPage() {
  const { orgId, projectId } = Route.useParams()

  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [saveDialogOpen, setSaveDialogOpen] = useState(false)

  const windows = usePlaygroundStore((s) => s.windows)
  const addWindow = usePlaygroundStore((s) => s.addWindow)
  const clearAll = usePlaygroundStore((s) => s.clearAll)
  const isExecutingAll = usePlaygroundStore((s) => s.isExecutingAll)
  const setExecutingAll = usePlaygroundStore((s) => s.setExecutingAll)

  const windowExecuteRefs = useRef<Map<number, () => Promise<void>>>(new Map())

  const registerWindowExecute = useCallback(
    (index: number, executeFn: () => Promise<void>) => {
      windowExecuteRefs.current.set(index, executeFn)
    },
    [],
  )

  const unregisterWindowExecute = useCallback((index: number) => {
    windowExecuteRefs.current.delete(index)
  }, [])

  const handleExecuteAll = useCallback(async () => {
    setExecutingAll(true)

    try {
      const executeFns = Array.from(windowExecuteRefs.current.entries())
        .sort(([a], [b]) => a - b)
        .map(([, fn]) => fn)

      await Promise.all(executeFns.map((fn) => fn()))
    } catch (error) {
      console.error('Execute all failed:', error)
    } finally {
      setExecutingAll(false)
    }
  }, [setExecutingAll])

  const handleNewSession = useCallback(() => {
    clearAll()
  }, [clearAll])

  usePlaygroundKeyboard({
    onExecuteAll: windows.length > 1 ? handleExecuteAll : undefined,
    onNewWindow: addWindow,
    enabled: true,
  })

  return (
    <main className="mx-auto w-full max-w-[1600px] p-6">
      <header className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Playground</h1>
          <p className="text-sm text-muted-foreground">
            Prototype prompts with live streaming execution.
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setSidebarOpen(true)}>
            <FolderOpen className="mr-2 h-4 w-4" />
            Sessions
          </Button>
          <Button variant="outline" onClick={() => setSaveDialogOpen(true)}>
            <Bookmark className="mr-2 h-4 w-4" />
            Save Session
          </Button>
          <Button variant="outline" onClick={clearAll}>
            <RotateCcw className="mr-2 h-4 w-4" />
            Reset
          </Button>
        </div>
      </header>

      <div className="space-y-6 mt-6">
        <SharedVariablesPanel />

        <div className="flex items-center justify-between">
          <Button
            variant="outline"
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

        {windows.length <= 3 ? (
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
                projectId={projectId}
                orgId={orgId}
                onRegisterExecute={(fn) => registerWindowExecute(index, fn)}
                onUnregisterExecute={() => unregisterWindowExecute(index)}
              />
            ))}
          </div>
        ) : (
          <div className="flex flex-nowrap overflow-x-auto gap-4 pb-4 min-h-[calc(100vh-280px)] snap-x snap-mandatory">
            {windows.map((_, index) => (
              <div
                key={windows[index].id}
                className="flex-none w-80 sm:w-96 lg:w-[420px] xl:w-[480px] snap-start"
              >
                <PlaygroundWindow
                  index={index}
                  projectId={projectId}
                  orgId={orgId}
                  onRegisterExecute={(fn) => registerWindowExecute(index, fn)}
                  onUnregisterExecute={() => unregisterWindowExecute(index)}
                />
              </div>
            ))}
          </div>
        )}
      </div>

      <SavedSessionsSidebar
        open={sidebarOpen}
        onOpenChange={setSidebarOpen}
        projectId={projectId}
        onNewSession={handleNewSession}
      />

      <SaveSessionDialog
        open={saveDialogOpen}
        onOpenChange={setSaveDialogOpen}
        projectId={projectId}
      />
    </main>
  )
}
