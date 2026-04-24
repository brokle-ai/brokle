import React, { useMemo, useState } from 'react'
import { useParams } from '@tanstack/react-router'
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
  // web-vite routing lives under `/o/:orgId/p/:projectId/...`. Keep orgId
  // on the context so link-emitting components can render TanStack
  // <Link to=... params={{orgId, projectId, ...}}> without re-reading
  // route params themselves.
  orgId: string | undefined
  // Retained for API shape parity with web/. In web-vite this is the
  // composite `${orgId}/${projectId}` fragment; do NOT interpolate it
  // into URLs — use orgId/projectId + TanStack Link instead.
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
  const params = useParams({ strict: false }) as { orgId?: string }

  const contextValue = useMemo(
    () => ({
      open,
      setOpen,
      currentRow,
      setCurrentRow,
      projectId: currentProject?.id,
      orgId: params.orgId,
      projectSlug,
    }),
    [open, setOpen, currentRow, currentProject?.id, params.orgId, projectSlug],
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
