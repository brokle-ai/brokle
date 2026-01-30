'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { experimentsApi } from '../api/experiments-api'
import { experimentQueryKeys } from './use-experiments'
import type {
  CreateExperimentFromWizardRequest,
  ValidateStepRequest,
  EstimateCostRequest,
  Experiment,
} from '../types'

// ============================================================================
// Query Keys
// ============================================================================

export const wizardQueryKeys = {
  all: ['experiment-wizard'] as const,
  datasetFields: (projectId: string, datasetId: string) =>
    [...wizardQueryKeys.all, 'dataset-fields', projectId, datasetId] as const,
  experimentConfig: (projectId: string, experimentId: string) =>
    [...wizardQueryKeys.all, 'config', projectId, experimentId] as const,
}

// ============================================================================
// Queries
// ============================================================================

export function useDatasetFieldsQuery(
  projectId: string | undefined,
  datasetId: string | undefined
) {
  return useQuery({
    queryKey: wizardQueryKeys.datasetFields(projectId ?? '', datasetId ?? ''),
    queryFn: () => experimentsApi.getDatasetFields(projectId!, datasetId!),
    enabled: !!projectId && !!datasetId,
    staleTime: 60_000, // Cache for 1 minute
    gcTime: 5 * 60 * 1000,
  })
}

export function useExperimentConfigQuery(
  projectId: string | undefined,
  experimentId: string | undefined
) {
  return useQuery({
    queryKey: wizardQueryKeys.experimentConfig(projectId ?? '', experimentId ?? ''),
    queryFn: () => experimentsApi.getExperimentConfig(projectId!, experimentId!),
    enabled: !!projectId && !!experimentId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

// ============================================================================
// Mutations
// ============================================================================

export function useCreateFromWizardMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateExperimentFromWizardRequest) =>
      experimentsApi.createFromWizard(projectId, data),
    onSuccess: (experiment: Experiment) => {
      // Invalidate experiments list to show the new experiment
      queryClient.invalidateQueries({
        queryKey: experimentQueryKeys.list(projectId),
      })
      toast.success(`Experiment "${experiment.name}" created successfully`)
    },
    onError: (error: Error) => {
      toast.error(`Failed to create experiment: ${error.message}`)
    },
  })
}

export function useValidateStepMutation(projectId: string) {
  return useMutation({
    mutationFn: (data: ValidateStepRequest) =>
      experimentsApi.validateWizardStep(projectId, data),
    // No toast on validation - it's a background operation
  })
}

export function useEstimateCostMutation(projectId: string) {
  return useMutation({
    mutationFn: (data: EstimateCostRequest) =>
      experimentsApi.estimateCost(projectId, data),
    onError: (error: Error) => {
      toast.error(`Failed to estimate cost: ${error.message}`)
    },
  })
}
