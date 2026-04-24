import { lazy, Suspense, type ReactNode } from 'react'

// CodeMirror 6 wrapper for editable templates. Replaces Monaco at the
// Jinja2 + JSON editor call sites in web/. Lazy-loaded so the CM6
// bundle (state+view+language+lang-jinja ≈ 120 KB min) stays out of
// the initial route chunk.

const Impl = lazy(() => import('./prompt-editor-impl'))

export interface PromptEditorProps {
  value: string
  onChange?: (next: string) => void
  language?: 'jinja' | 'json' | 'plain'
  readOnly?: boolean
  placeholder?: string
  minHeight?: string
  ariaLabel?: string
}

export function PromptEditor(props: PromptEditorProps): ReactNode {
  return (
    <Suspense fallback={<EditorSkeleton height={props.minHeight ?? '200px'} />}>
      <Impl {...props} />
    </Suspense>
  )
}

function EditorSkeleton({ height }: { height: string }) {
  return (
    <div
      aria-hidden
      className="animate-pulse rounded border bg-muted"
      style={{ minHeight: height }}
    />
  )
}
