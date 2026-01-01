'use client'

import React, { useMemo, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import useDialogState from '@/hooks/use-dialog-state'
import { useProjectOnly } from '@/features/projects'
import {
  useEvaluationRuleQuery,
  useDeleteEvaluationRuleMutation,
  useActivateEvaluationRuleMutation,
  useDeactivateEvaluationRuleMutation,
  useTriggerEvaluationRuleMutation,
} from '../hooks/use-evaluation-rules'
import type { EvaluationRule } from '../types'

export type RuleDetailDialogType = 'edit' | 'delete'

interface RuleDetailContextType {
  // Data
  rule: EvaluationRule | null
  isLoading: boolean
  error: string | null
  refetch: () => void
  // Dialog state
  open: RuleDetailDialogType | null
  setOpen: (dialog: RuleDetailDialogType | null) => void
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
  ruleId: string
}

const RuleDetailContext = React.createContext<RuleDetailContextType | null>(null)

interface RuleDetailProviderProps {
  children: React.ReactNode
  projectSlug: string
  ruleId: string
}

export function RuleDetailProvider({
  children,
  projectSlug,
  ruleId,
}: RuleDetailProviderProps) {
  const router = useRouter()
  const [open, setOpen] = useDialogState<RuleDetailDialogType>(null)
  const { currentProject, isLoading: projectLoading } = useProjectOnly()

  const projectId = currentProject?.id ?? ''

  const {
    data: rule,
    isLoading: ruleLoading,
    error,
    refetch,
  } = useEvaluationRuleQuery(projectId || undefined, ruleId)

  const deleteMutation = useDeleteEvaluationRuleMutation(projectId)
  const activateMutation = useActivateEvaluationRuleMutation(projectId)
  const deactivateMutation = useDeactivateEvaluationRuleMutation(projectId)
  const triggerMutation = useTriggerEvaluationRuleMutation(projectId, ruleId)

  const isLoading = projectLoading || ruleLoading

  const handleDelete = useCallback(async () => {
    if (rule) {
      await deleteMutation.mutateAsync({
        ruleId: rule.id,
        ruleName: rule.name,
      })
      router.push(`/projects/${projectSlug}/evaluations/rules`)
    }
  }, [rule, deleteMutation, router, projectSlug])

  const handleToggleStatus = useCallback(async () => {
    if (!rule) return
    if (rule.status === 'active') {
      await deactivateMutation.mutateAsync({
        ruleId: rule.id,
        ruleName: rule.name,
      })
    } else {
      await activateMutation.mutateAsync({
        ruleId: rule.id,
        ruleName: rule.name,
      })
    }
  }, [rule, activateMutation, deactivateMutation])

  const handleTrigger = useCallback(async () => {
    await triggerMutation.mutateAsync(undefined)
  }, [triggerMutation])

  const contextValue = useMemo(
    () => ({
      rule: rule ?? null,
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
      ruleId,
    }),
    [
      rule,
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
      ruleId,
    ]
  )

  return (
    <RuleDetailContext value={contextValue}>
      {children}
    </RuleDetailContext>
  )
}

export function useRuleDetail() {
  const context = React.useContext(RuleDetailContext)

  if (!context) {
    throw new Error('useRuleDetail must be used within <RuleDetailProvider>')
  }

  return context
}
