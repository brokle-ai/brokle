'use client'

import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { type Trace } from '../data/schema'

type TracesDialogType = 'delete'

type TracesContextType = {
  open: TracesDialogType | null
  setOpen: (str: TracesDialogType | null) => void
  currentRow: Trace | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Trace | null>>
  projectSlug?: string
}

const TracesContext = React.createContext<TracesContextType | null>(null)

interface TracesProviderProps {
  children: React.ReactNode
  projectSlug?: string
}

export function TracesProvider({ children, projectSlug }: TracesProviderProps) {
  const [open, setOpen] = useDialogState<TracesDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Trace | null>(null)

  return (
    <TracesContext value={{ open, setOpen, currentRow, setCurrentRow, projectSlug }}>
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
