'use client'

import React, { useMemo } from 'react'
import { useProjectOnly } from '@/features/projects'

type ScoresContextType = {
  projectId: string | undefined
  projectSlug?: string
}

const ScoresContext = React.createContext<ScoresContextType | null>(null)

interface ScoresProviderProps {
  children: React.ReactNode
  projectSlug?: string
}

export function ScoresProvider({ children, projectSlug }: ScoresProviderProps) {
  const { currentProject } = useProjectOnly()

  const contextValue = useMemo(
    () => ({
      projectId: currentProject?.id,
      projectSlug,
    }),
    [currentProject?.id, projectSlug]
  )

  return (
    <ScoresContext value={contextValue}>
      {children}
    </ScoresContext>
  )
}

export function useScores() {
  const context = React.useContext(ScoresContext)

  if (!context) {
    throw new Error('useScores must be used within <ScoresProvider>')
  }

  return context
}
