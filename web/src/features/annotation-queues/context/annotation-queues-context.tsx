'use client'

import React, { useMemo, useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { useProjectOnly } from '@/features/projects'
import type { AnnotationQueue } from '../types'

export type AnnotationQueuesDialogType = 'create' | 'edit' | 'delete' | 'add-items'

type AnnotationQueuesContextType = {
  open: AnnotationQueuesDialogType | null
  setOpen: (str: AnnotationQueuesDialogType | null) => void
  currentRow: AnnotationQueue | null
  setCurrentRow: React.Dispatch<React.SetStateAction<AnnotationQueue | null>>
  projectId: string | undefined
  projectSlug?: string
}

const AnnotationQueuesContext = React.createContext<AnnotationQueuesContextType | null>(null)

interface AnnotationQueuesProviderProps {
  children: React.ReactNode
  projectSlug?: string
}

export function AnnotationQueuesProvider({ children, projectSlug }: AnnotationQueuesProviderProps) {
  const [open, setOpen] = useDialogState<AnnotationQueuesDialogType>(null)
  const [currentRow, setCurrentRow] = useState<AnnotationQueue | null>(null)
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
    <AnnotationQueuesContext value={contextValue}>
      {children}
    </AnnotationQueuesContext>
  )
}

export const useAnnotationQueues = () => {
  const context = React.useContext(AnnotationQueuesContext)

  if (!context) {
    throw new Error('useAnnotationQueues must be used within <AnnotationQueuesProvider>')
  }

  return context
}
