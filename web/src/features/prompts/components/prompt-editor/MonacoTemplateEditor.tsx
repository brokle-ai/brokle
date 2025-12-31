'use client'

import { useCallback, useEffect, useRef, useState } from 'react'
import Editor, { type Monaco, type OnMount } from '@monaco-editor/react'
import type * as MonacoEditor from 'monaco-editor'
import { useTheme } from 'next-themes'
import { Loader2Icon, AlertCircleIcon } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { TemplateDialect } from '@/features/prompts/types'
import {
  registerAllTemplateLanguages,
  getLanguageIdForDialect,
  getTemplateTheme,
  createMustacheCompletionProvider,
  createJinja2CompletionProvider,
  MUSTACHE_LANGUAGE_ID,
  JINJA2_LANGUAGE_ID,
} from './dialects'

interface SyntaxError {
  line: number
  column: number
  message: string
  code: string
}

interface MonacoTemplateEditorProps {
  value: string
  onChange: (value: string) => void
  dialect: TemplateDialect
  variables?: string[]
  errors?: SyntaxError[]
  readOnly?: boolean
  height?: string | number
  className?: string
  placeholder?: string
  onEditorReady?: (editor: MonacoEditor.editor.IStandaloneCodeEditor) => void
}

// Track if languages have been registered globally
let languagesRegistered = false

/**
 * MonacoTemplateEditor - A Monaco-based editor for template content.
 *
 * Features:
 * - Syntax highlighting for Mustache and Jinja2
 * - Auto-completion for variables and template syntax
 * - Error squiggles for validation errors
 * - Light/dark theme support
 */
export function MonacoTemplateEditor({
  value,
  onChange,
  dialect,
  variables = [],
  errors = [],
  readOnly = false,
  height = 300,
  className,
  placeholder,
  onEditorReady,
}: MonacoTemplateEditorProps) {
  const { resolvedTheme } = useTheme()
  const isDarkMode = resolvedTheme === 'dark'
  const editorRef = useRef<MonacoEditor.editor.IStandaloneCodeEditor | null>(null)
  const monacoRef = useRef<Monaco | null>(null)
  const [isEditorReady, setIsEditorReady] = useState(false)

  // Get the current variables (for completion provider)
  const variablesRef = useRef<string[]>(variables)
  variablesRef.current = variables

  // Handle editor mount
  const handleEditorMount: OnMount = useCallback(
    (editor, monaco) => {
      editorRef.current = editor
      monacoRef.current = monaco

      // Register languages only once globally
      if (!languagesRegistered) {
        registerAllTemplateLanguages(monaco)
        languagesRegistered = true
      }

      // Register completion providers for the current session
      const getVariables = () => variablesRef.current

      // Mustache completion
      monaco.languages.registerCompletionItemProvider(
        MUSTACHE_LANGUAGE_ID,
        createMustacheCompletionProvider(monaco, getVariables)
      )

      // Jinja2 completion
      monaco.languages.registerCompletionItemProvider(
        JINJA2_LANGUAGE_ID,
        createJinja2CompletionProvider(monaco, getVariables)
      )

      setIsEditorReady(true)
      onEditorReady?.(editor)
    },
    [onEditorReady]
  )

  // Update error markers when errors change
  useEffect(() => {
    if (!monacoRef.current || !editorRef.current) return

    const model = editorRef.current.getModel()
    if (!model) return

    const markers: MonacoEditor.editor.IMarkerData[] = errors.map((error) => ({
      startLineNumber: error.line,
      startColumn: error.column,
      endLineNumber: error.line,
      endColumn: error.column + 1,
      message: error.message,
      severity: monacoRef.current!.MarkerSeverity.Error,
      source: 'template-validation',
      code: error.code,
    }))

    monacoRef.current.editor.setModelMarkers(model, 'template-validation', markers)
  }, [errors])

  // Update language when dialect changes
  useEffect(() => {
    if (!monacoRef.current || !editorRef.current) return

    const model = editorRef.current.getModel()
    if (!model) return

    const languageId = getLanguageIdForDialect(dialect)
    monacoRef.current.editor.setModelLanguage(model, languageId)
  }, [dialect])

  // Update theme when theme changes
  useEffect(() => {
    if (!monacoRef.current) return

    const theme = getTemplateTheme(isDarkMode)
    monacoRef.current.editor.setTheme(theme)
  }, [isDarkMode])

  // Handle value changes
  const handleChange = useCallback(
    (newValue: string | undefined) => {
      onChange(newValue ?? '')
    },
    [onChange]
  )

  // Calculate the language ID for initial render
  const languageId = getLanguageIdForDialect(dialect)
  const theme = getTemplateTheme(isDarkMode)

  return (
    <div className={cn('relative rounded-md border overflow-hidden', className)}>
      <Editor
        height={height}
        language={languageId}
        theme={theme}
        value={value}
        onChange={handleChange}
        onMount={handleEditorMount}
        loading={
          <div className="flex items-center justify-center h-full bg-background">
            <Loader2Icon className="size-6 animate-spin text-muted-foreground" />
          </div>
        }
        options={{
          readOnly,
          minimap: { enabled: false },
          lineNumbers: 'on',
          lineNumbersMinChars: 3,
          folding: true,
          foldingStrategy: 'indentation',
          wordWrap: 'on',
          scrollBeyondLastLine: false,
          automaticLayout: true,
          tabSize: 2,
          insertSpaces: true,
          fontSize: 13,
          fontFamily: 'var(--font-mono), monospace',
          padding: { top: 8, bottom: 8 },
          renderLineHighlight: 'line',
          selectionHighlight: true,
          occurrencesHighlight: 'singleFile',
          quickSuggestions: {
            other: true,
            comments: false,
            strings: true,
          },
          suggestOnTriggerCharacters: true,
          acceptSuggestionOnEnter: 'on',
          snippetSuggestions: 'inline',
          placeholder: placeholder,
          overviewRulerBorder: false,
          hideCursorInOverviewRuler: true,
          scrollbar: {
            vertical: 'auto',
            horizontal: 'auto',
            verticalScrollbarSize: 10,
            horizontalScrollbarSize: 10,
          },
          fixedOverflowWidgets: true,
        }}
      />

      {/* Error indicator */}
      {errors.length > 0 && (
        <div className="absolute bottom-2 right-2 flex items-center gap-1 text-xs text-destructive bg-destructive/10 px-2 py-1 rounded">
          <AlertCircleIcon className="size-3" />
          <span>
            {errors.length} {errors.length === 1 ? 'error' : 'errors'}
          </span>
        </div>
      )}

      {/* Placeholder when empty */}
      {!value && placeholder && !isEditorReady && (
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none text-muted-foreground text-sm">
          {placeholder}
        </div>
      )}
    </div>
  )
}

/**
 * Get the editor instance ref for external control.
 */
export function useMonacoEditorRef() {
  const editorRef = useRef<MonacoEditor.editor.IStandaloneCodeEditor | null>(null)
  return editorRef
}
