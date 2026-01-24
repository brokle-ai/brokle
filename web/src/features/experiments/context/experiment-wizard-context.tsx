'use client'

import React, { createContext, useContext, useMemo, useState, useCallback } from 'react'
import { useProjectOnly } from '@/features/projects'
import type { PromptListItem, PromptVersion, ModelConfig } from '@/features/prompts/types'
import type {
  ExperimentWizardState,
  ConfigStepState,
  DatasetStepState,
  EvaluatorStepState,
  DatasetListItem,
  DatasetFieldsResponse,
  ExperimentVariableMapping,
  WizardEvaluator,
  ValidationResult,
  CreateExperimentFromWizardRequest,
  EstimateCostResponse,
} from '../types'
import {
  useCreateFromWizardMutation,
  useEstimateCostMutation,
} from '../hooks/use-wizard-queries'
import type { Experiment } from '../types'

// ============================================================================
// Initial State
// ============================================================================

const initialConfigState: ConfigStepState = {
  name: '',
  description: '',
  promptId: null,
  promptVersionId: null,
  selectedPrompt: null,
  selectedVersion: null,
  modelConfigOverride: null,
  promptVariables: [],
}

const initialDatasetState: DatasetStepState = {
  datasetId: null,
  datasetVersionId: null,
  selectedDataset: null,
  variableMapping: [],
  datasetFields: null,
}

const initialEvaluatorState: EvaluatorStepState = {
  evaluators: [],
}

const initialWizardState: ExperimentWizardState = {
  currentStep: 1,
  completedSteps: new Set<number>(),
  touchedSteps: new Set<number>(),
  configState: initialConfigState,
  datasetState: initialDatasetState,
  evaluatorState: initialEvaluatorState,
}

// ============================================================================
// Context Types
// ============================================================================

interface ExperimentWizardContextValue {
  state: ExperimentWizardState
  projectId: string | undefined

  // Navigation
  goToStep: (step: 1 | 2 | 3 | 4) => void
  nextStep: () => void
  prevStep: () => void
  isStepComplete: (step: number) => boolean
  attemptNextStep: () => boolean // Returns true if successful, marks step as touched

  // Validation
  getStepValidation: (step: number) => ValidationResult
  validationState: {
    step1: ValidationResult
    step2: ValidationResult
    step3: ValidationResult
  }
  shouldShowStepErrors: (step: number) => boolean

  // Step 1: Config State Updates
  updateConfigState: (updates: Partial<ConfigStepState>) => void
  selectPrompt: (prompt: PromptListItem) => void
  selectVersion: (version: PromptVersion) => void

  // Step 2: Dataset State Updates
  updateDatasetState: (updates: Partial<DatasetStepState>) => void
  selectDataset: (dataset: DatasetListItem) => void
  setDatasetFields: (fields: DatasetFieldsResponse) => void
  autoMapVariables: () => void

  // Step 3: Evaluator State Updates
  updateEvaluatorState: (updates: Partial<EvaluatorStepState>) => void
  addEvaluator: (evaluator: Omit<WizardEvaluator, 'id'>) => void
  updateEvaluator: (id: string, updates: Partial<WizardEvaluator>) => void
  removeEvaluator: (id: string) => void
  duplicateEvaluator: (id: string) => void

  // Actions
  submit: (runImmediately: boolean) => Promise<Experiment>
  isSubmitting: boolean
  estimateCost: () => Promise<EstimateCostResponse | null>
  costEstimate: EstimateCostResponse | null
  isEstimating: boolean
  reset: () => void
}

const ExperimentWizardContext = createContext<ExperimentWizardContextValue | null>(null)

// ============================================================================
// Provider
// ============================================================================

interface ExperimentWizardProviderProps {
  children: React.ReactNode
}

export function ExperimentWizardProvider({ children }: ExperimentWizardProviderProps) {
  const { currentProject } = useProjectOnly()
  const projectId = currentProject?.id

  // State
  const [state, setState] = useState<ExperimentWizardState>(initialWizardState)
  const [costEstimate, setCostEstimate] = useState<EstimateCostResponse | null>(null)

  // Mutations
  const createMutation = useCreateFromWizardMutation(projectId ?? '')
  const estimateMutation = useEstimateCostMutation(projectId ?? '')

  // ============================================================================
  // Validation Logic
  // ============================================================================

  const getStepValidation = useCallback(
    (step: number): ValidationResult => {
      const errors: { field: string; message: string }[] = []
      const warnings: string[] = []

      switch (step) {
        case 1: {
          const { name, promptId, promptVersionId } = state.configState
          if (!name.trim()) errors.push({ field: 'name', message: 'Experiment name is required' })
          if (!promptId) errors.push({ field: 'promptId', message: 'Please select a prompt' })
          if (!promptVersionId) errors.push({ field: 'promptVersionId', message: 'Please select a prompt version' })
          break
        }
        case 2: {
          const { datasetId, variableMapping } = state.datasetState
          const { promptVariables } = state.configState
          if (!datasetId) errors.push({ field: 'datasetId', message: 'Please select a dataset' })

          // Check all variables are mapped
          const mappedVars = variableMapping.map((m) => m.variable_name)
          const unmappedVars = promptVariables.filter((v) => !mappedVars.includes(v))
          if (unmappedVars.length > 0) {
            errors.push({
              field: 'variableMapping',
              message: `Unmapped variables: ${unmappedVars.join(', ')}`,
            })
          }
          break
        }
        case 3: {
          const { evaluators } = state.evaluatorState
          if (evaluators.length === 0) {
            errors.push({ field: 'evaluators', message: 'At least one evaluator is required' })
          }
          break
        }
      }

      return {
        isValid: errors.length === 0,
        isComplete: errors.length === 0 && warnings.length === 0,
        errors,
        warnings,
      }
    },
    [state]
  )

  const validationState = useMemo(
    () => ({
      step1: getStepValidation(1),
      step2: getStepValidation(2),
      step3: getStepValidation(3),
    }),
    [getStepValidation]
  )

  // ============================================================================
  // Navigation
  // ============================================================================

  const goToStep = useCallback((step: 1 | 2 | 3 | 4) => {
    setState((prev) => ({ ...prev, currentStep: step }))
  }, [])

  const nextStep = useCallback(() => {
    setState((prev) => {
      const newCompleted = new Set(prev.completedSteps)
      const newTouched = new Set(prev.touchedSteps)
      newCompleted.add(prev.currentStep)
      newTouched.add(prev.currentStep)
      return {
        ...prev,
        currentStep: Math.min(prev.currentStep + 1, 4) as 1 | 2 | 3 | 4,
        completedSteps: newCompleted,
        touchedSteps: newTouched,
      }
    })
  }, [])

  const prevStep = useCallback(() => {
    setState((prev) => ({
      ...prev,
      currentStep: Math.max(prev.currentStep - 1, 1) as 1 | 2 | 3 | 4,
    }))
  }, [])

  const isStepComplete = useCallback(
    (step: number) => {
      return state.completedSteps.has(step) || getStepValidation(step).isValid
    },
    [state.completedSteps, getStepValidation]
  )

  const shouldShowStepErrors = useCallback(
    (step: number) => {
      return state.touchedSteps.has(step)
    },
    [state.touchedSteps]
  )

  const attemptNextStep = useCallback(() => {
    // Always mark the current step as touched to show errors
    setState((prev) => {
      const newTouched = new Set(prev.touchedSteps)
      newTouched.add(prev.currentStep)
      return { ...prev, touchedSteps: newTouched }
    })

    // Check if the current step is valid
    const validation = getStepValidation(state.currentStep)
    if (!validation.isValid) {
      return false // Don't proceed, but errors are now shown
    }

    // Proceed to next step
    setState((prev) => {
      const newCompleted = new Set(prev.completedSteps)
      newCompleted.add(prev.currentStep)
      return {
        ...prev,
        currentStep: Math.min(prev.currentStep + 1, 4) as 1 | 2 | 3 | 4,
        completedSteps: newCompleted,
      }
    })
    return true
  }, [state.currentStep, getStepValidation])

  // ============================================================================
  // Step 1: Config State Updates
  // ============================================================================

  const updateConfigState = useCallback((updates: Partial<ConfigStepState>) => {
    setState((prev) => ({
      ...prev,
      configState: { ...prev.configState, ...updates },
    }))
  }, [])

  const selectPrompt = useCallback((prompt: PromptListItem) => {
    setState((prev) => ({
      ...prev,
      configState: {
        ...prev.configState,
        promptId: prompt.id,
        selectedPrompt: prompt,
        promptVersionId: null,
        selectedVersion: null,
        promptVariables: [],
      },
    }))
  }, [])

  const selectVersion = useCallback((version: PromptVersion) => {
    setState((prev) => ({
      ...prev,
      configState: {
        ...prev.configState,
        promptVersionId: version.id,
        selectedVersion: version,
        promptVariables: version.variables || [],
      },
    }))
  }, [])

  // ============================================================================
  // Step 2: Dataset State Updates
  // ============================================================================

  const updateDatasetState = useCallback((updates: Partial<DatasetStepState>) => {
    setState((prev) => ({
      ...prev,
      datasetState: { ...prev.datasetState, ...updates },
    }))
  }, [])

  const selectDataset = useCallback((dataset: DatasetListItem) => {
    setState((prev) => ({
      ...prev,
      datasetState: {
        ...prev.datasetState,
        datasetId: dataset.id,
        selectedDataset: dataset,
        variableMapping: [],
        datasetFields: null,
      },
    }))
  }, [])

  const setDatasetFields = useCallback((fields: DatasetFieldsResponse) => {
    setState((prev) => ({
      ...prev,
      datasetState: {
        ...prev.datasetState,
        datasetFields: fields,
      },
    }))
  }, [])

  const autoMapVariables = useCallback(() => {
    setState((prev) => {
      const { promptVariables } = prev.configState
      const { datasetFields } = prev.datasetState

      if (!datasetFields) return prev

      // Try to auto-map variables based on name matching
      const allFields = [
        ...datasetFields.input_fields.map((f) => ({ ...f, source: 'dataset_input' as const })),
        ...datasetFields.expected_fields.map((f) => ({ ...f, source: 'dataset_expected' as const })),
        ...datasetFields.metadata_fields.map((f) => ({ ...f, source: 'dataset_metadata' as const })),
      ]

      const autoMappings: ExperimentVariableMapping[] = promptVariables.map((varName) => {
        // Try exact match first
        const exactMatch = allFields.find(
          (f) => f.path.toLowerCase() === varName.toLowerCase()
        )
        if (exactMatch) {
          return {
            variable_name: varName,
            source: exactMatch.source,
            field_path: exactMatch.path,
            is_auto_mapped: true,
          }
        }

        // Try partial match
        const partialMatch = allFields.find(
          (f) =>
            f.path.toLowerCase().includes(varName.toLowerCase()) ||
            varName.toLowerCase().includes(f.path.toLowerCase())
        )
        if (partialMatch) {
          return {
            variable_name: varName,
            source: partialMatch.source,
            field_path: partialMatch.path,
            is_auto_mapped: true,
          }
        }

        // Default to first input field
        const defaultField = datasetFields.input_fields[0]
        return {
          variable_name: varName,
          source: 'dataset_input',
          field_path: defaultField?.path || '',
          is_auto_mapped: false,
        }
      })

      return {
        ...prev,
        datasetState: {
          ...prev.datasetState,
          variableMapping: autoMappings,
        },
      }
    })
  }, [])

  // ============================================================================
  // Step 3: Evaluator State Updates
  // ============================================================================

  const updateEvaluatorState = useCallback((updates: Partial<EvaluatorStepState>) => {
    setState((prev) => ({
      ...prev,
      evaluatorState: { ...prev.evaluatorState, ...updates },
    }))
  }, [])

  const addEvaluator = useCallback((evaluator: Omit<WizardEvaluator, 'id'>) => {
    setState((prev) => ({
      ...prev,
      evaluatorState: {
        ...prev.evaluatorState,
        evaluators: [...prev.evaluatorState.evaluators, { ...evaluator, id: crypto.randomUUID() }],
      },
    }))
  }, [])

  const updateEvaluator = useCallback((id: string, updates: Partial<WizardEvaluator>) => {
    setState((prev) => ({
      ...prev,
      evaluatorState: {
        ...prev.evaluatorState,
        evaluators: prev.evaluatorState.evaluators.map((e) =>
          e.id === id ? { ...e, ...updates } : e
        ),
      },
    }))
  }, [])

  const removeEvaluator = useCallback((id: string) => {
    setState((prev) => ({
      ...prev,
      evaluatorState: {
        ...prev.evaluatorState,
        evaluators: prev.evaluatorState.evaluators.filter((e) => e.id !== id),
      },
    }))
  }, [])

  const duplicateEvaluator = useCallback((id: string) => {
    setState((prev) => {
      const original = prev.evaluatorState.evaluators.find((e) => e.id === id)
      if (!original) return prev

      return {
        ...prev,
        evaluatorState: {
          ...prev.evaluatorState,
          evaluators: [
            ...prev.evaluatorState.evaluators,
            { ...original, id: crypto.randomUUID(), name: `${original.name} (Copy)` },
          ],
        },
      }
    })
  }, [])

  // ============================================================================
  // Actions
  // ============================================================================

  const submit = useCallback(
    async (runImmediately: boolean): Promise<Experiment> => {
      const { configState, datasetState, evaluatorState } = state

      const request: CreateExperimentFromWizardRequest = {
        name: configState.name,
        description: configState.description || undefined,
        prompt_id: configState.promptId!,
        prompt_version_id: configState.promptVersionId!,
        model_config_override: configState.modelConfigOverride,
        dataset_id: datasetState.datasetId!,
        dataset_version_id: datasetState.datasetVersionId || undefined,
        variable_mapping: datasetState.variableMapping,
        evaluators: evaluatorState.evaluators.map(({ id, ...rest }) => rest),
        run_immediately: runImmediately,
      }

      return createMutation.mutateAsync(request)
    },
    [state, createMutation]
  )

  const estimateCost = useCallback(async (): Promise<EstimateCostResponse | null> => {
    const { configState, datasetState, evaluatorState } = state

    if (!configState.promptId || !configState.promptVersionId || !datasetState.datasetId) {
      return null
    }

    try {
      const response = await estimateMutation.mutateAsync({
        prompt_id: configState.promptId,
        prompt_version_id: configState.promptVersionId,
        dataset_id: datasetState.datasetId,
        dataset_version_id: datasetState.datasetVersionId || undefined,
        evaluators: evaluatorState.evaluators.map(({ id, ...rest }) => rest),
      })
      setCostEstimate(response)
      return response
    } catch {
      return null
    }
  }, [state, estimateMutation])

  const reset = useCallback(() => {
    setState(initialWizardState)
    setCostEstimate(null)
  }, [])

  // ============================================================================
  // Context Value
  // ============================================================================

  const contextValue = useMemo<ExperimentWizardContextValue>(
    () => ({
      state,
      projectId,

      // Navigation
      goToStep,
      nextStep,
      prevStep,
      isStepComplete,
      attemptNextStep,

      // Validation
      getStepValidation,
      validationState,
      shouldShowStepErrors,

      // Step 1
      updateConfigState,
      selectPrompt,
      selectVersion,

      // Step 2
      updateDatasetState,
      selectDataset,
      setDatasetFields,
      autoMapVariables,

      // Step 3
      updateEvaluatorState,
      addEvaluator,
      updateEvaluator,
      removeEvaluator,
      duplicateEvaluator,

      // Actions
      submit,
      isSubmitting: createMutation.isPending,
      estimateCost,
      costEstimate,
      isEstimating: estimateMutation.isPending,
      reset,
    }),
    [
      state,
      projectId,
      goToStep,
      nextStep,
      prevStep,
      isStepComplete,
      attemptNextStep,
      getStepValidation,
      validationState,
      shouldShowStepErrors,
      updateConfigState,
      selectPrompt,
      selectVersion,
      updateDatasetState,
      selectDataset,
      setDatasetFields,
      autoMapVariables,
      updateEvaluatorState,
      addEvaluator,
      updateEvaluator,
      removeEvaluator,
      duplicateEvaluator,
      submit,
      createMutation.isPending,
      estimateCost,
      costEstimate,
      estimateMutation.isPending,
      reset,
    ]
  )

  return (
    <ExperimentWizardContext.Provider value={contextValue}>
      {children}
    </ExperimentWizardContext.Provider>
  )
}

// ============================================================================
// Hook
// ============================================================================

export function useExperimentWizard() {
  const context = useContext(ExperimentWizardContext)

  if (!context) {
    throw new Error('useExperimentWizard must be used within <ExperimentWizardProvider>')
  }

  return context
}
