/**
 * Prompt Hooks
 *
 * This module exports all React hooks for the prompts feature,
 * including query hooks, mutation hooks, and template editing hooks.
 */

// Query and mutation hooks
export {
  promptQueryKeys,
  type PromptFilters,
  usePromptsQuery,
  usePromptQuery,
  useVersionsQuery,
  useVersionQuery,
  useVersionDiffQuery,
  useProtectedLabelsQuery,
  useCreatePromptMutation,
  useUpdatePromptMutation,
  useDeletePromptMutation,
  useCreateVersionMutation,
  useSetLabelsMutation,
  useSetProtectedLabelsMutation,
} from './use-prompts-queries'

// Template validation hooks
export {
  useTemplateValidation,
  useValidateTemplateMutation,
} from './use-template-validation'

// Template preview hooks
export {
  useTemplatePreview,
  usePreviewTemplateMutation,
  useDetectDialectMutation,
  useTemplateEditor,
} from './use-template-preview'

// Other hooks
export { useProjectPrompts } from './use-project-prompts'
export { usePromptsTableState, type UsePromptsTableStateReturn } from './use-prompts-table-state'
export {
  usePromptDetailState,
  type UsePromptDetailStateReturn,
  type PromptDetailTab,
} from './use-prompt-detail-state'
export {
  usePromptEditState,
  type UsePromptEditStateReturn,
} from './use-prompt-edit-state'
