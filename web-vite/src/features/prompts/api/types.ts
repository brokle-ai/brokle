// Re-export canonical types from the feature types module. Kept as a
// legacy shim because a couple of route files import from here.

export type {
  PromptType,
  TemplateDialect,
  PromptLabelInfo,
  ChatMessage,
  ModelConfig,
  TextTemplate,
  ChatTemplate,
  PromptTemplate,
  PromptListItem,
  Prompt,
  PromptVersion,
  VersionDiff,
  PromptListResponse,
  CreatePromptRequest,
  UpdatePromptRequest,
  CreateVersionRequest,
} from '../types'

import type { PromptTemplate, TextTemplate, PromptType } from '../types'

// Form-local state for the shared prompt form.
export interface PromptFormState {
  name: string
  type: PromptType
  tagsCsv: string
  body: string
  variablesCsv: string
  commitMessage: string
}

export function isTextTemplate(t: PromptTemplate): t is TextTemplate {
  return typeof (t as TextTemplate).content === 'string'
}

export function templateToBody(t: PromptTemplate): string {
  if (isTextTemplate(t)) return t.content
  return JSON.stringify(t, null, 2)
}
