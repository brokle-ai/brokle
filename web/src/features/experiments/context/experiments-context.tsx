'use client'

import React, { useMemo, useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { useProjectOnly } from '@/features/projects'
import type { Experiment } from '../types'

export type ExperimentsDialogType = 'create' | 'edit' | 'delete'

type ExperimentsContextType = {
  open: ExperimentsDialogType | null
  setOpen: (str: ExperimentsDialogType | null) => void
  currentRow: Experiment | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Experiment | null>>
  projectId: string | undefined
  projectSlug?: string
}

const ExperimentsContext = React.createContext<ExperimentsContextType | null>(null)

interface ExperimentsProviderProps {
  children: React.ReactNode
  projectSlug?: string
}

export function ExperimentsProvider({
  children,
  projectSlug,
}: ExperimentsProviderProps) {
  const [open, setOpen] = useDialogState<ExperimentsDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Experiment | null>(null)
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
    <ExperimentsContext value={contextValue}>
      {children}
    </ExperimentsContext>
  )
}

export function useExperiments() {
  const context = React.useContext(ExperimentsContext)

  if (!context) {
    throw new Error('useExperiments must be used within <ExperimentsProvider>')
  }

  return context
}
