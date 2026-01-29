'use client'

import React, { useMemo, useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { useProjectOnly } from '@/features/projects'
import type { Evaluator } from '../types'

export type EvaluatorsDialogType = 'create' | 'edit' | 'delete'

type EvaluatorsContextType = {
  open: EvaluatorsDialogType | null
  setOpen: (str: EvaluatorsDialogType | null) => void
  currentRow: Evaluator | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Evaluator | null>>
  projectId: string | undefined
  orgId: string | undefined
  projectSlug?: string
}

const EvaluatorsContext = React.createContext<EvaluatorsContextType | null>(null)

interface EvaluatorsProviderProps {
  children: React.ReactNode
  projectSlug?: string
}

export function EvaluatorsProvider({ children, projectSlug }: EvaluatorsProviderProps) {
  const [open, setOpen] = useDialogState<EvaluatorsDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Evaluator | null>(null)
  const { currentProject } = useProjectOnly()

  const contextValue = useMemo(
    () => ({
      open,
      setOpen,
      currentRow,
      setCurrentRow,
      projectId: currentProject?.id,
      orgId: currentProject?.organizationId,
      projectSlug,
    }),
    [open, setOpen, currentRow, currentProject?.id, currentProject?.organizationId, projectSlug]
  )

  return (
    <EvaluatorsContext value={contextValue}>
      {children}
    </EvaluatorsContext>
  )
}

export function useEvaluators() {
  const context = React.useContext(EvaluatorsContext)

  if (!context) {
    throw new Error('useEvaluators must be used within <EvaluatorsProvider>')
  }

  return context
}
