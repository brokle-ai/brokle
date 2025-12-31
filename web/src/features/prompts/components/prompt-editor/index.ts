/**
 * Prompt Editor Components
 *
 * This module exports all components for the prompt template editor,
 * including Monaco-based editor with syntax highlighting, validation,
 * and preview functionality.
 */

// Main editor components
export { AdvancedTemplateEditor, SimpleTemplateEditor } from './AdvancedTemplateEditor'
export { MonacoTemplateEditor, useMonacoEditorRef } from './MonacoTemplateEditor'
export { DialectSelector, getDialectLabel, getDialectDescription } from './DialectSelector'

// Variable components
export {
  VariablePanel,
  VariableInputPanel,
  generateSampleValues,
} from './VariablePanel'

// Preview components
export {
  TemplatePreview,
  ValidationStatus,
  SyntaxErrorList,
} from './TemplatePreview'

// Legacy components (for backward compatibility)
export { PromptTemplateInput } from './PromptTemplateInput'
export { ChatMessageEditor } from './ChatMessageEditor'
export { PromptEditorToolbar } from './PromptEditorToolbar'
export { VariableBadge, VariableList } from './VariableExtractor'

// Dialect definitions
export {
  registerAllTemplateLanguages,
  getLanguageIdForDialect,
  MUSTACHE_LANGUAGE_ID,
  JINJA2_LANGUAGE_ID,
  TEMPLATE_LIGHT_THEME,
  TEMPLATE_DARK_THEME,
  getTemplateTheme,
} from './dialects'
