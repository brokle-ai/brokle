import type { ScorerType, ScorerConfig, VariableMap } from '@/features/evaluation-rules/types'
import type { ModelConfig, PromptListItem, PromptVersion } from '@/features/prompts/types'

// Note: ExperimentSource is defined in ./index.ts and re-exported via barrel export

// ============================================================================
// VARIABLE MAPPING
// ============================================================================

export type VariableMappingSource =
  | 'dataset_input'
  | 'dataset_expected'
  | 'dataset_metadata'

export interface ExperimentVariableMapping {
  variable_name: string
  source: VariableMappingSource
  field_path: string
  is_auto_mapped: boolean
}

// ============================================================================
// EXPERIMENT EVALUATOR (uses existing scorer types)
// ============================================================================

export interface ExperimentEvaluator {
  name: string
  scorer_type: ScorerType
  scorer_config: ScorerConfig | Record<string, unknown>
  variable_mapping?: VariableMap[]
}

// ============================================================================
// WIZARD EVALUATOR (with client-side ID)
// ============================================================================

export interface WizardEvaluator extends ExperimentEvaluator {
  id: string // Client-side UUID for list keys
}

// ============================================================================
// GROUPED STATE PROPS (Langfuse pattern)
// ============================================================================

export interface ConfigStepState {
  name: string
  description: string
  promptId: string | null
  promptVersionId: string | null
  selectedPrompt: PromptListItem | null
  selectedVersion: PromptVersion | null
  modelConfigOverride: ModelConfig | null
  promptVariables: string[] // Derived from template
}

export interface DatasetStepState {
  datasetId: string | null
  datasetVersionId: string | null
  selectedDataset: DatasetListItem | null
  variableMapping: ExperimentVariableMapping[]
  datasetFields: DatasetFieldsResponse | null
}

export interface EvaluatorStepState {
  evaluators: WizardEvaluator[]
}

export interface ExperimentWizardState {
  currentStep: 1 | 2 | 3 | 4
  completedSteps: Set<number>
  touchedSteps: Set<number> // Steps that user has attempted to proceed from

  // Grouped state props (Langfuse pattern)
  configState: ConfigStepState
  datasetState: DatasetStepState
  evaluatorState: EvaluatorStepState
}

// ============================================================================
// DATASET TYPES (for wizard)
// ============================================================================

export interface DatasetListItem {
  id: string
  name: string
  description?: string
  item_count?: number
  latest_version?: number
  pinned_version?: number
  created_at: string
  updated_at: string
}

// ============================================================================
// FIELD INFO (for variable mapping)
// ============================================================================

export interface FieldInfo {
  path: string
  type: string // string, number, boolean, object, array
  sample_value?: unknown
}

export interface DatasetFieldsResponse {
  input_fields: FieldInfo[]
  expected_fields: FieldInfo[]
  metadata_fields: FieldInfo[]
}

// ============================================================================
// VALIDATION
// ============================================================================

export interface ValidationError {
  field: string
  message: string
}

export interface ValidationResult {
  isValid: boolean
  isComplete: boolean
  errors: ValidationError[]
  warnings: string[]
}

export interface ValidateStepRequest {
  step: number
  data: Record<string, unknown>
}

export interface ValidateStepResponse {
  is_valid: boolean
  errors?: ValidationError[]
  warnings?: string[]
}

// ============================================================================
// COST ESTIMATION
// ============================================================================

export interface CostItem {
  description: string
  estimated_cost: number
  estimated_units: number
  unit_type: string
}

export interface EstimateCostRequest {
  prompt_id: string
  prompt_version_id: string
  dataset_id: string
  dataset_version_id?: string
  evaluators: ExperimentEvaluator[]
}

export interface EstimateCostResponse {
  item_count: number
  estimated_tokens: number
  estimated_cost: number
  currency: string
  cost_breakdown: CostItem[]
}

// ============================================================================
// WIZARD REQUEST
// ============================================================================

export interface CreateExperimentFromWizardRequest {
  // Step 1: Basic Info
  name: string
  description?: string

  // Step 1: Prompt Configuration
  prompt_id: string
  prompt_version_id: string
  model_config_override?: ModelConfig | null

  // Step 2: Dataset Configuration
  dataset_id: string
  dataset_version_id?: string
  variable_mapping: ExperimentVariableMapping[]

  // Step 3: Evaluators
  evaluators: ExperimentEvaluator[]

  // Options
  run_immediately: boolean
}

// ============================================================================
// EXPERIMENT CONFIG (stored config for dashboard-created experiments)
// ============================================================================

export interface ExperimentConfig {
  id: string
  experiment_id: string
  prompt_id: string
  prompt_version_id: string
  model_config?: ModelConfig | null
  dataset_id: string
  dataset_version_id?: string
  variable_mapping: ExperimentVariableMapping[]
  evaluators: ExperimentEvaluator[]
  created_at: string
  updated_at: string
}

// ============================================================================
// EVALUATOR CATEGORIES (Opik pattern)
// ============================================================================

export type EvaluatorCategory = 'heuristics' | 'llm_judges'

export const EVALUATOR_CATEGORIES: Record<
  EvaluatorCategory,
  { label: string; scorerTypes: ScorerType[] }
> = {
  heuristics: {
    label: 'Heuristics',
    scorerTypes: ['builtin', 'regex'],
  },
  llm_judges: {
    label: 'LLM Judges',
    scorerTypes: ['llm'],
  },
}

// Built-in scorer library (W&B Weave pattern)
export const BUILTIN_SCORERS = [
  {
    name: 'contains',
    label: 'Contains',
    description: 'Check if output contains a substring',
    configSchema: { substring: { type: 'string', required: true } },
  },
  {
    name: 'json_valid',
    label: 'JSON Valid',
    description: 'Check if output is valid JSON',
    configSchema: {},
  },
  {
    name: 'length_check',
    label: 'Length Check',
    description: 'Check output length constraints',
    configSchema: {
      min_length: { type: 'number', required: false },
      max_length: { type: 'number', required: false },
    },
  },
  {
    name: 'exact_match',
    label: 'Exact Match',
    description: 'Check if output exactly matches expected',
    configSchema: { case_sensitive: { type: 'boolean', required: false } },
  },
] as const

export type BuiltinScorerName = (typeof BUILTIN_SCORERS)[number]['name']
