import type * as Monaco from 'monaco-editor'

/**
 * Mustache language definition for Monaco Editor.
 * Supports {{variable}}, {{#section}}...{{/section}}, {{^inverted}},
 * {{>partial}}, {{!comment}}, and {{{raw}}} syntax.
 */

export const MUSTACHE_LANGUAGE_ID = 'mustache-template'

export const mustacheLanguageConfiguration: Monaco.languages.LanguageConfiguration = {
  comments: {
    blockComment: ['{{!', '}}'],
  },
  brackets: [
    ['{{', '}}'],
    ['{{#', '}}'],
    ['{{/', '}}'],
    ['{{^', '}}'],
    ['{{{', '}}}'],
  ],
  autoClosingPairs: [
    { open: '{{', close: '}}' },
    { open: '{{{', close: '}}}' },
    { open: '{{#', close: '}}' },
    { open: '{{^', close: '}}' },
    { open: '{{>', close: '}}' },
    { open: '{{!', close: '}}' },
    { open: '"', close: '"' },
    { open: "'", close: "'" },
  ],
  surroundingPairs: [
    { open: '{{', close: '}}' },
    { open: '{{{', close: '}}}' },
    { open: '"', close: '"' },
    { open: "'", close: "'" },
  ],
  folding: {
    markers: {
      start: /\{\{#/,
      end: /\{\{\//,
    },
  },
}

export const mustacheTokensProvider: Monaco.languages.IMonarchLanguage = {
  defaultToken: '',
  tokenPostfix: '.mustache',

  brackets: [
    { open: '{{', close: '}}', token: 'delimiter.mustache' },
    { open: '{{{', close: '}}}', token: 'delimiter.mustache.unescaped' },
  ],

  tokenizer: {
    root: [
      [/\{\{\{/, { token: 'delimiter.mustache.unescaped', next: '@unescapedVariable' }],
      [/\{\{!/, { token: 'comment.mustache', next: '@comment' }],
      [/\{\{#/, { token: 'delimiter.mustache.section', next: '@sectionStart' }],
      [/\{\{\//, { token: 'delimiter.mustache.section', next: '@sectionEnd' }],
      [/\{\{\^/, { token: 'delimiter.mustache.inverted', next: '@invertedSection' }],
      [/\{\{>/, { token: 'delimiter.mustache.partial', next: '@partial' }],
      [/\{\{/, { token: 'delimiter.mustache', next: '@variable' }],
      [/[^{]+/, 'text'],
      [/\{(?!\{)/, 'text'],
    ],

    variable: [
      [/\s+/, 'white'],
      [/[a-zA-Z_][a-zA-Z0-9_]*/, 'variable.mustache'],
      [/\./, 'delimiter.dot'],
      [/\}\}/, { token: 'delimiter.mustache', next: '@pop' }],
    ],

    unescapedVariable: [
      [/\s+/, 'white'],
      [/[a-zA-Z_][a-zA-Z0-9_]*/, 'variable.mustache.unescaped'],
      [/\./, 'delimiter.dot'],
      [/\}\}\}/, { token: 'delimiter.mustache.unescaped', next: '@pop' }],
    ],

    comment: [
      [/[^}]+/, 'comment.mustache'],
      [/\}\}/, { token: 'comment.mustache', next: '@pop' }],
      [/\}/, 'comment.mustache'],
    ],

    sectionStart: [
      [/\s+/, 'white'],
      [/[a-zA-Z_][a-zA-Z0-9_]*/, 'keyword.mustache.section'],
      [/\./, 'delimiter.dot'],
      [/\}\}/, { token: 'delimiter.mustache.section', next: '@pop' }],
    ],

    sectionEnd: [
      [/\s+/, 'white'],
      [/[a-zA-Z_][a-zA-Z0-9_]*/, 'keyword.mustache.section'],
      [/\./, 'delimiter.dot'],
      [/\}\}/, { token: 'delimiter.mustache.section', next: '@pop' }],
    ],

    invertedSection: [
      [/\s+/, 'white'],
      [/[a-zA-Z_][a-zA-Z0-9_]*/, 'keyword.mustache.inverted'],
      [/\./, 'delimiter.dot'],
      [/\}\}/, { token: 'delimiter.mustache.inverted', next: '@pop' }],
    ],

    partial: [
      [/\s+/, 'white'],
      [/[a-zA-Z_][a-zA-Z0-9_./\-]*/, 'string.mustache.partial'],
      [/\}\}/, { token: 'delimiter.mustache.partial', next: '@pop' }],
    ],
  },
}

export function registerMustacheLanguage(monaco: typeof Monaco): void {
  monaco.languages.register({
    id: MUSTACHE_LANGUAGE_ID,
    extensions: ['.mustache', '.hbs', '.handlebars'],
    aliases: ['Mustache', 'Handlebars', 'mustache'],
    mimetypes: ['text/x-mustache', 'text/x-handlebars-template'],
  })
  monaco.languages.setLanguageConfiguration(
    MUSTACHE_LANGUAGE_ID,
    mustacheLanguageConfiguration
  )
  monaco.languages.setMonarchTokensProvider(
    MUSTACHE_LANGUAGE_ID,
    mustacheTokensProvider
  )
}

export function getMustacheCompletionItems(
  variables: string[],
  range: Monaco.IRange
): Monaco.languages.CompletionItem[] {
  const items: Monaco.languages.CompletionItem[] = []

  for (const variable of variables) {
    items.push({
      label: variable,
      kind: 5,
      insertText: variable,
      detail: 'Template variable',
      documentation: `Insert {{${variable}}}`,
      range,
    })
  }

  items.push(
    {
      label: 'section',
      kind: 14,
      insertText: '{{#${1:section}}}\n\t$0\n{{/${1:section}}}',
      insertTextRules: 4,
      detail: 'Section block',
      documentation: 'Create a section block that renders if the value is truthy',
      range,
    },
    {
      label: 'inverted',
      kind: 14,
      insertText: '{{^${1:section}}}\n\t$0\n{{/${1:section}}}',
      insertTextRules: 4,
      detail: 'Inverted section block',
      documentation: 'Create an inverted section that renders if the value is falsy',
      range,
    },
    {
      label: 'each',
      kind: 14,
      insertText: '{{#each ${1:items}}}\n\t{{this}}\n{{/each}}',
      insertTextRules: 4,
      detail: 'Each loop',
      documentation: 'Iterate over an array',
      range,
    },
    {
      label: 'if',
      kind: 14,
      insertText: '{{#if ${1:condition}}}\n\t$0\n{{/if}}',
      insertTextRules: 4,
      detail: 'If block',
      documentation: 'Conditional rendering',
      range,
    },
    {
      label: 'unless',
      kind: 14,
      insertText: '{{#unless ${1:condition}}}\n\t$0\n{{/unless}}',
      insertTextRules: 4,
      detail: 'Unless block',
      documentation: 'Render if condition is falsy',
      range,
    },
    {
      label: 'partial',
      kind: 14,
      insertText: '{{> ${1:partialName}}}',
      insertTextRules: 4,
      detail: 'Partial include',
      documentation: 'Include a partial template',
      range,
    },
    {
      label: 'comment',
      kind: 14,
      insertText: '{{! ${1:comment} }}',
      insertTextRules: 4,
      detail: 'Comment',
      documentation: 'Add a comment (not rendered)',
      range,
    },
    {
      label: 'raw',
      kind: 14,
      insertText: '{{{${1:variable}}}}',
      insertTextRules: 4,
      detail: 'Raw/unescaped output',
      documentation: 'Output without HTML escaping',
      range,
    }
  )

  return items
}

export function createMustacheCompletionProvider(
  monaco: typeof Monaco,
  getVariables: () => string[]
): Monaco.languages.CompletionItemProvider {
  return {
    triggerCharacters: ['{', '#', '^', '>', '/'],
    provideCompletionItems: (model, position) => {
      const word = model.getWordUntilPosition(position)
      const range: Monaco.IRange = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn,
      }

      const textUntilPosition = model.getValueInRange({
        startLineNumber: position.lineNumber,
        startColumn: 1,
        endLineNumber: position.lineNumber,
        endColumn: position.column,
      })

      const lastOpen = textUntilPosition.lastIndexOf('{{')
      const lastClose = textUntilPosition.lastIndexOf('}}')

      if (lastOpen > lastClose) {
        const variables = getVariables()
        return {
          suggestions: getMustacheCompletionItems(variables, range),
        }
      }

      return {
        suggestions: [
          {
            label: '{{',
            kind: 14,
            insertText: '{{${1:variable}}}',
            insertTextRules: 4,
            detail: 'Variable',
            documentation: 'Insert a variable',
            range,
          },
          {
            label: '{{#',
            kind: 14,
            insertText: '{{#${1:section}}}\n\t$0\n{{/${1:section}}}',
            insertTextRules: 4,
            detail: 'Section',
            documentation: 'Create a section block',
            range,
          },
        ],
      }
    },
  }
}
