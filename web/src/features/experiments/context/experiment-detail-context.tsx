'use client'

import React, { useMemo, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import useDialogState from '@/hooks/use-dialog-state'
import { useProjectOnly } from '@/features/projects'
import {
  useExperimentQuery,
  useUpdateExperimentMutation,
  useDeleteExperimentMutation,
} from '../hooks/use-experiments'
import type { Experiment, UpdateExperimentRequest } from '../types'

export type ExperimentDetailDialogType = 'edit' | 'delete'

interface ExperimentDetailContextType {
  experiment: Experiment | null
  isLoading: boolean
  error: string | null
  refetch: () => void
  open: ExperimentDetailDialogType | null
  setOpen: (dialog: ExperimentDetailDialogType | null) => void
  handleUpdate: (data: UpdateExperimentRequest) => Promise<void>
  handleDelete: () => Promise<void>
  isUpdating: boolean
  isDeleting: boolean
  projectSlug: string
  projectId: string
  experimentId: string
}

const ExperimentDetailContext = React.createContext<ExperimentDetailContextType | null>(null)

interface ExperimentDetailProviderProps {
  children: React.ReactNode
  projectSlug: string
  experimentId: string
}

export function ExperimentDetailProvider({
  children,
  projectSlug,
  experimentId,
}: ExperimentDetailProviderProps) {
  const router = useRouter()
  const [open, setOpen] = useDialogState<ExperimentDetailDialogType>(null)
  const { currentProject, isLoading: projectLoading } = useProjectOnly()

  const projectId = currentProject?.id ?? ''

  const {
    data: experiment,
    isLoading: experimentLoading,
    error,
    refetch,
  } = useExperimentQuery(projectId, experimentId)

  const updateMutation = useUpdateExperimentMutation(projectId, experimentId)
  const deleteMutation = useDeleteExperimentMutation(projectId)

  const isLoading = projectLoading || experimentLoading

  const handleUpdate = useCallback(
    async (data: UpdateExperimentRequest) => {
      await updateMutation.mutateAsync(data)
      setOpen(null)
    },
    [updateMutation, setOpen]
  )

  const handleDelete = useCallback(async () => {
    if (experiment) {
      await deleteMutation.mutateAsync({
        experimentId: experiment.id,
        experimentName: experiment.name,
      })
      router.push(`/projects/${projectSlug}/experiments`)
    }
  }, [experiment, deleteMutation, router, projectSlug])

  const contextValue = useMemo(
    () => ({
      experiment: experiment ?? null,
      isLoading,
      error: error?.message ?? null,
      refetch,
      open,
      setOpen,
      handleUpdate,
      handleDelete,
      isUpdating: updateMutation.isPending,
      isDeleting: deleteMutation.isPending,
      projectSlug,
      projectId,
      experimentId,
    }),
    [
      experiment,
      isLoading,
      error,
      refetch,
      open,
      setOpen,
      handleUpdate,
      handleDelete,
      updateMutation.isPending,
      deleteMutation.isPending,
      projectSlug,
      projectId,
      experimentId,
    ]
  )

  return (
    <ExperimentDetailContext value={contextValue}>
      {children}
    </ExperimentDetailContext>
  )
}

export function useExperimentDetail() {
  const context = React.useContext(ExperimentDetailContext)

  if (!context) {
    throw new Error('useExperimentDetail must be used within <ExperimentDetailProvider>')
  }

  return context
}
