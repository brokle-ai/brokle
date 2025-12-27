/**
 * Monaco Editor language definitions and themes for template dialects.
 * Supports Mustache and Jinja2 template syntax.
 */

export {
  MUSTACHE_LANGUAGE_ID,
  mustacheLanguageConfiguration,
  mustacheTokensProvider,
  registerMustacheLanguage,
  getMustacheCompletionItems,
  createMustacheCompletionProvider,
} from './mustache-language'

export {
  JINJA2_LANGUAGE_ID,
  jinja2LanguageConfiguration,
  jinja2TokensProvider,
  registerJinja2Language,
  getJinja2CompletionItems,
  createJinja2CompletionProvider,
} from './jinja2-language'

export {
  templateLightThemeRules,
  templateDarkThemeRules,
  TEMPLATE_LIGHT_THEME,
  TEMPLATE_DARK_THEME,
  registerTemplateThemes,
  getTemplateTheme,
} from './template-themes'

import type * as Monaco from 'monaco-editor'
import { registerMustacheLanguage } from './mustache-language'
import { registerJinja2Language } from './jinja2-language'
import { registerTemplateThemes } from './template-themes'

/**
 * Register all template languages and themes with Monaco Editor.
 * Call this once during application initialization.
 */
export function registerAllTemplateLanguages(monaco: typeof Monaco): void {
  registerMustacheLanguage(monaco)
  registerJinja2Language(monaco)
  registerTemplateThemes(monaco)
}

export function getLanguageIdForDialect(
  dialect: 'simple' | 'mustache' | 'jinja2' | 'auto'
): string {
  switch (dialect) {
    case 'mustache':
      return 'mustache-template'
    case 'jinja2':
      return 'jinja2-template'
    case 'simple':
    case 'auto':
    default:
      // For simple templates, use plain text or a minimal highlighting
      return 'mustache-template' // Mustache is a superset of simple {{var}} syntax
  }
}
