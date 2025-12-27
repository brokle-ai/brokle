import type * as Monaco from 'monaco-editor'

/**
 * Jinja2 language definition for Monaco Editor.
 * Supports {{ variable }}, {% statement %}, {# comment #},
 * filters ({{ var|filter }}), and control structures.
 */

export const JINJA2_LANGUAGE_ID = 'jinja2-template'

export const jinja2LanguageConfiguration: Monaco.languages.LanguageConfiguration = {
  comments: {
    blockComment: ['{#', '#}'],
  },
  brackets: [
    ['{{', '}}'],
    ['{%', '%}'],
    ['{#', '#}'],
    ['(', ')'],
    ['[', ']'],
  ],
  autoClosingPairs: [
    { open: '{{', close: '}}' },
    { open: '{%', close: '%}' },
    { open: '{#', close: '#}' },
    { open: '(', close: ')' },
    { open: '[', close: ']' },
    { open: '"', close: '"' },
    { open: "'", close: "'" },
  ],
  surroundingPairs: [
    { open: '{{', close: '}}' },
    { open: '{%', close: '%}' },
    { open: '{#', close: '#}' },
    { open: '(', close: ')' },
    { open: '[', close: ']' },
    { open: '"', close: '"' },
    { open: "'", close: "'" },
  ],
  folding: {
    markers: {
      start: /\{%\s*(if|for|block|macro|call|filter)\b/,
      end: /\{%\s*end(if|for|block|macro|call|filter)\b/,
    },
  },
}

const JINJA2_FILTERS = [
  'abs', 'attr', 'batch', 'capitalize', 'center', 'count', 'default', 'd',
  'dictsort', 'escape', 'e', 'filesizeformat', 'first', 'float', 'forceescape',
  'format', 'groupby', 'indent', 'int', 'join', 'last', 'length', 'list',
  'lower', 'map', 'max', 'min', 'pprint', 'random', 'reject', 'rejectattr',
  'replace', 'reverse', 'round', 'safe', 'select', 'selectattr', 'slice',
  'sort', 'string', 'striptags', 'sum', 'title', 'trim', 'truncate', 'unique',
  'upper', 'urlencode', 'urlize', 'wordcount', 'wordwrap', 'xmlattr',
]

const JINJA2_TESTS = [
  'callable', 'defined', 'divisibleby', 'eq', 'equalto', 'escaped', 'even',
  'false', 'ge', 'gt', 'greaterthan', 'in', 'iterable', 'le', 'lower',
  'lt', 'lessthan', 'mapping', 'ne', 'none', 'number', 'odd', 'sameas',
  'sequence', 'string', 'true', 'undefined', 'upper',
]

const JINJA2_KEYWORDS = [
  'if', 'else', 'elif', 'endif', 'for', 'endfor', 'in', 'not', 'and', 'or',
  'block', 'endblock', 'extends', 'include', 'import', 'from', 'as', 'with',
  'endwith', 'macro', 'endmacro', 'call', 'endcall', 'filter', 'endfilter',
  'set', 'endset', 'raw', 'endraw', 'autoescape', 'endautoescape',
  'true', 'false', 'none', 'is', 'recursive', 'scoped',
]

export const jinja2TokensProvider: Monaco.languages.IMonarchLanguage = {
  defaultToken: '',
  tokenPostfix: '.jinja2',

  keywords: JINJA2_KEYWORDS,
  filters: JINJA2_FILTERS,
  tests: JINJA2_TESTS,

  operators: [
    '+', '-', '*', '/', '//', '%', '**',
    '==', '!=', '<', '>', '<=', '>=',
    '~', '|', '.', '[', ']', '(', ')',
  ],

  tokenizer: {
    root: [
      [/\{#/, { token: 'comment.jinja2', next: '@comment' }],
      [/\{%-?/, { token: 'delimiter.jinja2.statement', next: '@statement' }],
      [/\{\{-?/, { token: 'delimiter.jinja2.expression', next: '@expression' }],
      [/[^{]+/, 'text'],
      [/\{(?![{%#])/, 'text'],
    ],

    comment: [
      [/[^#]+/, 'comment.jinja2'],
      [/#\}/, { token: 'comment.jinja2', next: '@pop' }],
      [/#/, 'comment.jinja2'],
    ],

    statement: [
      [/-?%\}/, { token: 'delimiter.jinja2.statement', next: '@pop' }],
      { include: '@jinja2Common' },
    ],

    expression: [
      [/-?\}\}/, { token: 'delimiter.jinja2.expression', next: '@pop' }],
      { include: '@jinja2Common' },
    ],

    jinja2Common: [
      [/\s+/, 'white'],
      [/\d+\.\d+/, 'number.float.jinja2'],
      [/\d+/, 'number.jinja2'],
      [/"([^"\\]|\\.)*"/, 'string.jinja2'],
      [/'([^'\\]|\\.)*'/, 'string.jinja2'],
      [/\|/, { token: 'operator.pipe.jinja2', next: '@filter' }],
      [
        /[a-zA-Z_][a-zA-Z0-9_]*/,
        {
          cases: {
            '@keywords': 'keyword.jinja2',
            '@default': 'variable.jinja2',
          },
        },
      ],
      [/[+\-*/%]/, 'operator.jinja2'],
      [/[<>=!]=?/, 'operator.jinja2'],
      [/~/, 'operator.jinja2'],
      [/[[\]()]/, 'delimiter.jinja2'],
      [/\./, 'delimiter.dot.jinja2'],
      [/,/, 'delimiter.comma.jinja2'],
      [/:/, 'delimiter.colon.jinja2'],
    ],

    filter: [
      [/\s+/, 'white'],
      [
        /[a-zA-Z_][a-zA-Z0-9_]*/,
        {
          cases: {
            '@filters': { token: 'function.filter.jinja2', next: '@pop' },
            '@default': { token: 'function.filter.jinja2', next: '@pop' },
          },
        },
      ],
      [/./, { token: '@rematch', next: '@pop' }],
    ],
  },
}

export function registerJinja2Language(monaco: typeof Monaco): void {
  monaco.languages.register({
    id: JINJA2_LANGUAGE_ID,
    extensions: ['.jinja', '.jinja2', '.j2'],
    aliases: ['Jinja2', 'Jinja', 'jinja2'],
    mimetypes: ['text/x-jinja2', 'text/jinja2'],
  })
  monaco.languages.setLanguageConfiguration(
    JINJA2_LANGUAGE_ID,
    jinja2LanguageConfiguration
  )
  monaco.languages.setMonarchTokensProvider(
    JINJA2_LANGUAGE_ID,
    jinja2TokensProvider
  )
}

export function getJinja2CompletionItems(
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
      documentation: `Insert {{ ${variable} }}`,
      range,
    })
  }

  for (const filter of JINJA2_FILTERS) {
    items.push({
      label: filter,
      kind: 1,
      insertText: filter,
      detail: 'Jinja2 filter',
      documentation: `Apply ${filter} filter`,
      range,
    })
  }

  for (const test of JINJA2_TESTS) {
    items.push({
      label: `is ${test}`,
      kind: 13,
      insertText: `is ${test}`,
      detail: 'Jinja2 test',
      documentation: `Test if value ${test}`,
      range,
    })
  }

  items.push(
    {
      label: 'if',
      kind: 14,
      insertText: '{% if ${1:condition} %}\n\t$0\n{% endif %}',
      insertTextRules: 4,
      detail: 'If block',
      documentation: 'Conditional rendering',
      range,
    },
    {
      label: 'if-else',
      kind: 14,
      insertText: '{% if ${1:condition} %}\n\t${2:then}\n{% else %}\n\t${3:else}\n{% endif %}',
      insertTextRules: 4,
      detail: 'If-else block',
      documentation: 'Conditional with else branch',
      range,
    },
    {
      label: 'if-elif-else',
      kind: 14,
      insertText: '{% if ${1:condition1} %}\n\t${2:then1}\n{% elif ${3:condition2} %}\n\t${4:then2}\n{% else %}\n\t${5:else}\n{% endif %}',
      insertTextRules: 4,
      detail: 'If-elif-else block',
      documentation: 'Conditional with elif and else branches',
      range,
    },
    {
      label: 'for',
      kind: 14,
      insertText: '{% for ${1:item} in ${2:items} %}\n\t{{ ${1:item} }}\n{% endfor %}',
      insertTextRules: 4,
      detail: 'For loop',
      documentation: 'Iterate over a sequence',
      range,
    },
    {
      label: 'for-else',
      kind: 14,
      insertText: '{% for ${1:item} in ${2:items} %}\n\t{{ ${1:item} }}\n{% else %}\n\t${3:No items found}\n{% endfor %}',
      insertTextRules: 4,
      detail: 'For loop with else',
      documentation: 'Iterate with fallback for empty sequence',
      range,
    },
    {
      label: 'block',
      kind: 14,
      insertText: '{% block ${1:name} %}\n\t$0\n{% endblock %}',
      insertTextRules: 4,
      detail: 'Block definition',
      documentation: 'Define a template block for inheritance',
      range,
    },
    {
      label: 'extends',
      kind: 14,
      insertText: '{% extends "${1:base.html}" %}',
      insertTextRules: 4,
      detail: 'Extends',
      documentation: 'Extend a parent template',
      range,
    },
    {
      label: 'include',
      kind: 14,
      insertText: '{% include "${1:partial.html}" %}',
      insertTextRules: 4,
      detail: 'Include',
      documentation: 'Include another template',
      range,
    },
    {
      label: 'macro',
      kind: 14,
      insertText: '{% macro ${1:name}(${2:args}) %}\n\t$0\n{% endmacro %}',
      insertTextRules: 4,
      detail: 'Macro definition',
      documentation: 'Define a reusable macro',
      range,
    },
    {
      label: 'set',
      kind: 14,
      insertText: '{% set ${1:variable} = ${2:value} %}',
      insertTextRules: 4,
      detail: 'Set variable',
      documentation: 'Assign a value to a variable',
      range,
    },
    {
      label: 'with',
      kind: 14,
      insertText: '{% with ${1:variable} = ${2:value} %}\n\t$0\n{% endwith %}',
      insertTextRules: 4,
      detail: 'With block',
      documentation: 'Create a scoped variable',
      range,
    },
    {
      label: 'filter',
      kind: 14,
      insertText: '{% filter ${1:filtername} %}\n\t$0\n{% endfilter %}',
      insertTextRules: 4,
      detail: 'Filter block',
      documentation: 'Apply a filter to a block of content',
      range,
    },
    {
      label: 'raw',
      kind: 14,
      insertText: '{% raw %}\n\t$0\n{% endraw %}',
      insertTextRules: 4,
      detail: 'Raw block',
      documentation: 'Output content without processing',
      range,
    },
    {
      label: 'comment',
      kind: 14,
      insertText: '{# ${1:comment} #}',
      insertTextRules: 4,
      detail: 'Comment',
      documentation: 'Add a comment (not rendered)',
      range,
    }
  )

  return items
}

export function createJinja2CompletionProvider(
  monaco: typeof Monaco,
  getVariables: () => string[]
): Monaco.languages.CompletionItemProvider {
  return {
    triggerCharacters: ['{', '%', '|', '.'],
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

      const lastExprOpen = textUntilPosition.lastIndexOf('{{')
      const lastExprClose = textUntilPosition.lastIndexOf('}}')
      const lastStmtOpen = textUntilPosition.lastIndexOf('{%')
      const lastStmtClose = textUntilPosition.lastIndexOf('%}')

      if (lastExprOpen > lastExprClose) {
        const afterPipe = textUntilPosition.slice(lastExprOpen).includes('|')
        if (afterPipe) {
          return {
            suggestions: JINJA2_FILTERS.map((filter) => ({
              label: filter,
              kind: 1,
              insertText: filter,
              detail: 'Jinja2 filter',
              range,
            })),
          }
        }

        const variables = getVariables()
        return {
          suggestions: getJinja2CompletionItems(variables, range),
        }
      }

      if (lastStmtOpen > lastStmtClose) {
        return {
          suggestions: JINJA2_KEYWORDS.map((kw) => ({
            label: kw,
            kind: 13,
            insertText: kw,
            detail: 'Jinja2 keyword',
            range,
          })),
        }
      }

      return {
        suggestions: [
          {
            label: '{{',
            kind: 14,
            insertText: '{{ ${1:variable} }}',
            insertTextRules: 4,
            detail: 'Expression',
            documentation: 'Insert an expression',
            range,
          },
          {
            label: '{%',
            kind: 14,
            insertText: '{% ${1:statement} %}',
            insertTextRules: 4,
            detail: 'Statement',
            documentation: 'Insert a statement',
            range,
          },
          {
            label: '{#',
            kind: 14,
            insertText: '{# ${1:comment} #}',
            insertTextRules: 4,
            detail: 'Comment',
            documentation: 'Insert a comment',
            range,
          },
        ],
      }
    },
  }
}
