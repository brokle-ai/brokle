import type * as Monaco from 'monaco-editor'

/**
 * Theme definitions for Mustache and Jinja2 syntax highlighting.
 * Provides both light and dark theme variants.
 */

// Color palette for template syntax
const COLORS = {
  light: {
    // Mustache colors
    delimiter: '#0000FF', // Blue for {{ }}
    variable: '#001080', // Dark blue for variables
    section: '#AF00DB', // Purple for sections
    comment: '#008000', // Green for comments
    partial: '#A31515', // Red for partials
    unescaped: '#E50000', // Bright red for raw/unescaped

    // Jinja2 colors
    keyword: '#AF00DB', // Purple for keywords
    filter: '#795E26', // Brown for filters
    string: '#A31515', // Red for strings
    number: '#098658', // Green for numbers
    operator: '#000000', // Black for operators

    // Common
    text: '#000000',
    background: '#FFFFFF',
  },
  dark: {
    // Mustache colors
    delimiter: '#569CD6', // Light blue for {{ }}
    variable: '#9CDCFE', // Cyan for variables
    section: '#C586C0', // Pink for sections
    comment: '#6A9955', // Green for comments
    partial: '#CE9178', // Orange for partials
    unescaped: '#F14C4C', // Red for raw/unescaped

    // Jinja2 colors
    keyword: '#C586C0', // Pink for keywords
    filter: '#DCDCAA', // Yellow for filters
    string: '#CE9178', // Orange for strings
    number: '#B5CEA8', // Light green for numbers
    operator: '#D4D4D4', // Light gray for operators

    // Common
    text: '#D4D4D4',
    background: '#1E1E1E',
  },
}

export const templateLightThemeRules: Monaco.editor.ITokenThemeRule[] = [
  // Mustache tokens
  { token: 'delimiter.mustache', foreground: COLORS.light.delimiter },
  { token: 'delimiter.mustache.unescaped', foreground: COLORS.light.unescaped },
  { token: 'delimiter.mustache.section', foreground: COLORS.light.section },
  { token: 'delimiter.mustache.inverted', foreground: COLORS.light.section },
  { token: 'delimiter.mustache.partial', foreground: COLORS.light.partial },
  { token: 'variable.mustache', foreground: COLORS.light.variable },
  { token: 'variable.mustache.unescaped', foreground: COLORS.light.unescaped },
  { token: 'keyword.mustache.section', foreground: COLORS.light.section, fontStyle: 'bold' },
  { token: 'keyword.mustache.inverted', foreground: COLORS.light.section, fontStyle: 'bold' },
  { token: 'comment.mustache', foreground: COLORS.light.comment, fontStyle: 'italic' },
  { token: 'string.mustache.partial', foreground: COLORS.light.partial },

  // Jinja2 tokens
  { token: 'delimiter.jinja2.statement', foreground: COLORS.light.keyword },
  { token: 'delimiter.jinja2.expression', foreground: COLORS.light.delimiter },
  { token: 'comment.jinja2', foreground: COLORS.light.comment, fontStyle: 'italic' },
  { token: 'keyword.jinja2', foreground: COLORS.light.keyword, fontStyle: 'bold' },
  { token: 'variable.jinja2', foreground: COLORS.light.variable },
  { token: 'function.filter.jinja2', foreground: COLORS.light.filter },
  { token: 'string.jinja2', foreground: COLORS.light.string },
  { token: 'number.jinja2', foreground: COLORS.light.number },
  { token: 'number.float.jinja2', foreground: COLORS.light.number },
  { token: 'operator.jinja2', foreground: COLORS.light.operator },
  { token: 'operator.pipe.jinja2', foreground: COLORS.light.filter },
  { token: 'delimiter.jinja2', foreground: COLORS.light.operator },
  { token: 'delimiter.dot.jinja2', foreground: COLORS.light.operator },
  { token: 'delimiter.comma.jinja2', foreground: COLORS.light.operator },
  { token: 'delimiter.colon.jinja2', foreground: COLORS.light.operator },

  // Plain text
  { token: 'text', foreground: COLORS.light.text },
]

export const templateDarkThemeRules: Monaco.editor.ITokenThemeRule[] = [
  // Mustache tokens
  { token: 'delimiter.mustache', foreground: COLORS.dark.delimiter },
  { token: 'delimiter.mustache.unescaped', foreground: COLORS.dark.unescaped },
  { token: 'delimiter.mustache.section', foreground: COLORS.dark.section },
  { token: 'delimiter.mustache.inverted', foreground: COLORS.dark.section },
  { token: 'delimiter.mustache.partial', foreground: COLORS.dark.partial },
  { token: 'variable.mustache', foreground: COLORS.dark.variable },
  { token: 'variable.mustache.unescaped', foreground: COLORS.dark.unescaped },
  { token: 'keyword.mustache.section', foreground: COLORS.dark.section, fontStyle: 'bold' },
  { token: 'keyword.mustache.inverted', foreground: COLORS.dark.section, fontStyle: 'bold' },
  { token: 'comment.mustache', foreground: COLORS.dark.comment, fontStyle: 'italic' },
  { token: 'string.mustache.partial', foreground: COLORS.dark.partial },

  // Jinja2 tokens
  { token: 'delimiter.jinja2.statement', foreground: COLORS.dark.keyword },
  { token: 'delimiter.jinja2.expression', foreground: COLORS.dark.delimiter },
  { token: 'comment.jinja2', foreground: COLORS.dark.comment, fontStyle: 'italic' },
  { token: 'keyword.jinja2', foreground: COLORS.dark.keyword, fontStyle: 'bold' },
  { token: 'variable.jinja2', foreground: COLORS.dark.variable },
  { token: 'function.filter.jinja2', foreground: COLORS.dark.filter },
  { token: 'string.jinja2', foreground: COLORS.dark.string },
  { token: 'number.jinja2', foreground: COLORS.dark.number },
  { token: 'number.float.jinja2', foreground: COLORS.dark.number },
  { token: 'operator.jinja2', foreground: COLORS.dark.operator },
  { token: 'operator.pipe.jinja2', foreground: COLORS.dark.filter },
  { token: 'delimiter.jinja2', foreground: COLORS.dark.operator },
  { token: 'delimiter.dot.jinja2', foreground: COLORS.dark.operator },
  { token: 'delimiter.comma.jinja2', foreground: COLORS.dark.operator },
  { token: 'delimiter.colon.jinja2', foreground: COLORS.dark.operator },

  // Plain text
  { token: 'text', foreground: COLORS.dark.text },
]

export const TEMPLATE_LIGHT_THEME = 'template-light'
export const TEMPLATE_DARK_THEME = 'template-dark'

export function registerTemplateThemes(monaco: typeof Monaco): void {
  monaco.editor.defineTheme(TEMPLATE_LIGHT_THEME, {
    base: 'vs',
    inherit: true,
    rules: templateLightThemeRules,
    colors: {
      'editor.background': COLORS.light.background,
    },
  })

  monaco.editor.defineTheme(TEMPLATE_DARK_THEME, {
    base: 'vs-dark',
    inherit: true,
    rules: templateDarkThemeRules,
    colors: {
      'editor.background': COLORS.dark.background,
    },
  })
}

export function getTemplateTheme(isDarkMode: boolean): string {
  return isDarkMode ? TEMPLATE_DARK_THEME : TEMPLATE_LIGHT_THEME
}
