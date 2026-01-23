'use client'

import React, { useMemo, useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { useProjectOnly } from '@/features/projects'
import type { Dataset, DatasetWithItemCount } from '../types'

export type DatasetsDialogType = 'create' | 'edit' | 'delete'

type DatasetsContextType = {
  open: DatasetsDialogType | null
  setOpen: (str: DatasetsDialogType | null) => void
  currentRow: Dataset | DatasetWithItemCount | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Dataset | DatasetWithItemCount | null>>
  projectId: string | undefined
  projectSlug?: string
}

const DatasetsContext = React.createContext<DatasetsContextType | null>(null)

interface DatasetsProviderProps {
  children: React.ReactNode
  projectSlug?: string
}

export function DatasetsProvider({ children, projectSlug }: DatasetsProviderProps) {
  const [open, setOpen] = useDialogState<DatasetsDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Dataset | DatasetWithItemCount | null>(null)
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
    <DatasetsContext value={contextValue}>
      {children}
    </DatasetsContext>
  )
}

export const useDatasets = () => {
  const context = React.useContext(DatasetsContext)

  if (!context) {
    throw new Error('useDatasets must be used within <DatasetsProvider>')
  }

  return context
}
