'use client'

import React, { useMemo, useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { type Trace } from '../data/schema'

type TracesDialogType = 'delete'

type TracesContextType = {
  open: TracesDialogType | null
  setOpen: (str: TracesDialogType | null) => void
  currentRow: Trace | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Trace | null>>
  projectSlug?: string
  currentPageTraceIds: string[]
  setCurrentPageTraceIds: React.Dispatch<React.SetStateAction<string[]>>
  /** Current page trace objects for session grouping */
  currentPageTraces: Trace[]
  setCurrentPageTraces: React.Dispatch<React.SetStateAction<Trace[]>>
}

const TracesContext = React.createContext<TracesContextType | null>(null)

interface TracesProviderProps {
  children: React.ReactNode
  projectSlug?: string
}

export function TracesProvider({ children, projectSlug }: TracesProviderProps) {
  const [open, setOpen] = useDialogState<TracesDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Trace | null>(null)
  const [currentPageTraceIds, setCurrentPageTraceIds] = useState<string[]>([])
  const [currentPageTraces, setCurrentPageTraces] = useState<Trace[]>([])

  const contextValue = useMemo(
    () => ({
      open,
      setOpen,
      currentRow,
      setCurrentRow,
      projectSlug,
      currentPageTraceIds,
      setCurrentPageTraceIds,
      currentPageTraces,
      setCurrentPageTraces,
    }),
    [open, currentRow, projectSlug, currentPageTraceIds, currentPageTraces]
  )

  return (
    <TracesContext value={contextValue}>
      {children}
    </TracesContext>
  )
}

export const useTraces = () => {
  const tracesContext = React.useContext(TracesContext)

  if (!tracesContext) {
    throw new Error('useTraces has to be used within <TracesProvider>')
  }

  return tracesContext
}
