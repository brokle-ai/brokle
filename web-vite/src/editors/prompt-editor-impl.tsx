import CodeMirror, { type Extension } from '@uiw/react-codemirror'
import { useMemo } from 'react'
import { jinja } from '@codemirror/lang-jinja'
import { json } from '@codemirror/lang-json'
import { EditorView } from '@codemirror/view'
import type { PromptEditorProps } from './prompt-editor'

// Actual CM6 implementation. Imported lazily by prompt-editor.tsx so
// the heavy grammar bundles stay out of the initial chunk.

export default function PromptEditorImpl({
  value,
  onChange,
  language = 'plain',
  readOnly = false,
  placeholder,
  minHeight = '200px',
  ariaLabel,
}: PromptEditorProps) {
  const extensions = useMemo<Extension[]>(() => {
    const exts: Extension[] = [EditorView.lineWrapping]
    if (language === 'jinja') exts.push(jinja())
    if (language === 'json') exts.push(json())
    return exts
  }, [language])

  return (
    <CodeMirror
      value={value}
      onChange={onChange}
      extensions={extensions}
      readOnly={readOnly}
      editable={!readOnly}
      placeholder={placeholder}
      aria-label={ariaLabel}
      minHeight={minHeight}
      basicSetup={{
        lineNumbers: true,
        foldGutter: true,
        highlightActiveLine: !readOnly,
        highlightActiveLineGutter: !readOnly,
        bracketMatching: true,
        closeBrackets: !readOnly,
        autocompletion: !readOnly,
        searchKeymap: true,
      }}
      theme="light"
    />
  )
}
