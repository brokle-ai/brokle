'use client'

import React, { useMemo, useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { useProjectOnly } from '@/features/projects'
import type { EvaluationRule } from '../types'

export type RulesDialogType = 'create' | 'edit' | 'delete'

type RulesContextType = {
  open: RulesDialogType | null
  setOpen: (str: RulesDialogType | null) => void
  currentRow: EvaluationRule | null
  setCurrentRow: React.Dispatch<React.SetStateAction<EvaluationRule | null>>
  projectId: string | undefined
  orgId: string | undefined
  projectSlug?: string
}

const RulesContext = React.createContext<RulesContextType | null>(null)

interface RulesProviderProps {
  children: React.ReactNode
  projectSlug?: string
}

export function RulesProvider({ children, projectSlug }: RulesProviderProps) {
  const [open, setOpen] = useDialogState<RulesDialogType>(null)
  const [currentRow, setCurrentRow] = useState<EvaluationRule | null>(null)
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
    <RulesContext value={contextValue}>
      {children}
    </RulesContext>
  )
}

export function useRules() {
  const context = React.useContext(RulesContext)

  if (!context) {
    throw new Error('useRules must be used within <RulesProvider>')
  }

  return context
}
