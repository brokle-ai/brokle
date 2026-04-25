import * as React from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import type { Dashboard } from '../types'

export type DashboardsDialogType = 'create' | 'edit' | 'delete'

interface DashboardsContextType {
  open: DashboardsDialogType | null
  setOpen: (str: DashboardsDialogType | null) => void
  currentRow: Dashboard | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Dashboard | null>>
  projectId: string
  orgId: string
}

const DashboardsContext = React.createContext<DashboardsContextType | null>(
  null,
)

interface DashboardsProviderProps {
  children: React.ReactNode
  projectId: string
  orgId: string
}

export function DashboardsProvider({
  children,
  projectId,
  orgId,
}: DashboardsProviderProps) {
  const [open, setOpen] = useDialogState<DashboardsDialogType>(null)
  const [currentRow, setCurrentRow] = React.useState<Dashboard | null>(null)

  const contextValue = React.useMemo<DashboardsContextType>(
    () => ({
      open,
      setOpen,
      currentRow,
      setCurrentRow,
      projectId,
      orgId,
    }),
    [open, setOpen, currentRow, projectId, orgId],
  )

  return (
    <DashboardsContext.Provider value={contextValue}>
      {children}
    </DashboardsContext.Provider>
  )
}

export function useDashboards(): DashboardsContextType {
  const context = React.useContext(DashboardsContext)
  if (!context) {
    throw new Error('useDashboards must be used within <DashboardsProvider>')
  }
  return context
}
