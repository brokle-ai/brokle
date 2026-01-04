'use client'

import React, { useMemo, useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { useProjectOnly } from '@/features/projects'
import type { Dashboard } from '../types'

export type DashboardsDialogType = 'create' | 'edit' | 'delete'

type DashboardsContextType = {
  open: DashboardsDialogType | null
  setOpen: (str: DashboardsDialogType | null) => void
  currentRow: Dashboard | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Dashboard | null>>
  projectId: string | undefined
  projectSlug?: string
}

const DashboardsContext = React.createContext<DashboardsContextType | null>(null)

interface DashboardsProviderProps {
  children: React.ReactNode
  projectSlug?: string
}

export function DashboardsProvider({ children, projectSlug }: DashboardsProviderProps) {
  const [open, setOpen] = useDialogState<DashboardsDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Dashboard | null>(null)
  const { currentProject } = useProjectOnly()

  const contextValue = useMemo(
    () => ({
      open,
      setOpen,
      currentRow,
      setCurrentRow,
      projectId: currentProject?.id,
      projectSlug,
    }),
    [open, setOpen, currentRow, currentProject?.id, projectSlug]
  )

  return (
    <DashboardsContext value={contextValue}>
      {children}
    </DashboardsContext>
  )
}

export const useDashboards = () => {
  const context = React.useContext(DashboardsContext)

  if (!context) {
    throw new Error('useDashboards must be used within <DashboardsProvider>')
  }

  return context
}
