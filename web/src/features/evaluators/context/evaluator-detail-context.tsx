'use client'

import React, { useMemo, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import useDialogState from '@/hooks/use-dialog-state'
import { useProjectOnly } from '@/features/projects'
import {
  useEvaluatorQuery,
  useDeleteEvaluatorMutation,
  useActivateEvaluatorMutation,
  useDeactivateEvaluatorMutation,
  useTriggerEvaluatorMutation,
} from '../hooks/use-evaluators'
import type { Evaluator } from '../types'

export type EvaluatorDetailDialogType = 'edit' | 'delete'

interface EvaluatorDetailContextType {
  // Data
  evaluator: Evaluator | null
  isLoading: boolean
  error: string | null
  refetch: () => void
  // Dialog state
  open: EvaluatorDetailDialogType | null
  setOpen: (dialog: EvaluatorDetailDialogType | null) => void
  // Mutations
  handleDelete: () => Promise<void>
  handleToggleStatus: () => Promise<void>
  handleTrigger: () => Promise<void>
  isDeleting: boolean
  isToggling: boolean
  isTriggering: boolean
  // Navigation
  projectSlug: string
  projectId: string
  evaluatorId: string
}

const EvaluatorDetailContext = React.createContext<EvaluatorDetailContextType | null>(null)

interface EvaluatorDetailProviderProps {
  children: React.ReactNode
  projectSlug: string
  evaluatorId: string
}

export function EvaluatorDetailProvider({
  children,
  projectSlug,
  evaluatorId,
}: EvaluatorDetailProviderProps) {
  const router = useRouter()
  const [open, setOpen] = useDialogState<EvaluatorDetailDialogType>(null)
  const { currentProject, isLoading: projectLoading } = useProjectOnly()

  const projectId = currentProject?.id ?? ''

  const {
    data: evaluator,
    isLoading: evaluatorLoading,
    error,
    refetch,
  } = useEvaluatorQuery(projectId || undefined, evaluatorId)

  const deleteMutation = useDeleteEvaluatorMutation(projectId)
  const activateMutation = useActivateEvaluatorMutation(projectId)
  const deactivateMutation = useDeactivateEvaluatorMutation(projectId)
  const triggerMutation = useTriggerEvaluatorMutation(projectId, evaluatorId)

  const isLoading = projectLoading || evaluatorLoading

  const handleDelete = useCallback(async () => {
    if (evaluator) {
      await deleteMutation.mutateAsync({
        evaluatorId: evaluator.id,
        evaluatorName: evaluator.name,
      })
      router.push(`/projects/${projectSlug}/evaluators`)
    }
  }, [evaluator, deleteMutation, router, projectSlug])

  const handleToggleStatus = useCallback(async () => {
    if (!evaluator) return
    if (evaluator.status === 'active') {
      await deactivateMutation.mutateAsync({
        evaluatorId: evaluator.id,
        evaluatorName: evaluator.name,
      })
    } else {
      await activateMutation.mutateAsync({
        evaluatorId: evaluator.id,
        evaluatorName: evaluator.name,
      })
    }
  }, [evaluator, activateMutation, deactivateMutation])

  const handleTrigger = useCallback(async () => {
    await triggerMutation.mutateAsync(undefined)
  }, [triggerMutation])

  const contextValue = useMemo(
    () => ({
      evaluator: evaluator ?? null,
      isLoading,
      error: error?.message ?? null,
      refetch,
      open,
      setOpen,
      handleDelete,
      handleToggleStatus,
      handleTrigger,
      isDeleting: deleteMutation.isPending,
      isToggling: activateMutation.isPending || deactivateMutation.isPending,
      isTriggering: triggerMutation.isPending,
      projectSlug,
      projectId,
      evaluatorId,
    }),
    [
      evaluator,
      isLoading,
      error,
      refetch,
      open,
      setOpen,
      handleDelete,
      handleToggleStatus,
      handleTrigger,
      deleteMutation.isPending,
      activateMutation.isPending,
      deactivateMutation.isPending,
      triggerMutation.isPending,
      projectSlug,
      projectId,
      evaluatorId,
    ]
  )

  return (
    <EvaluatorDetailContext value={contextValue}>
      {children}
    </EvaluatorDetailContext>
  )
}

export function useEvaluatorDetail() {
  const context = React.useContext(EvaluatorDetailContext)

  if (!context) {
    throw new Error('useEvaluatorDetail must be used within <EvaluatorDetailProvider>')
  }

  return context
}
