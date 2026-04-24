import { lazy, Suspense, type ReactNode } from 'react'

// Read-only syntax-highlighted code block using Shiki. Preferred over
// the CM6 editor for pure display — Shiki is smaller, server/build-
// renderable, and doesn't ship an editor runtime.

const Impl = lazy(() => import('./readonly-code-impl'))

export interface ReadonlyCodeProps {
  code: string
  language?: string
  theme?: 'github-light' | 'github-dark' | 'nord'
  className?: string
}

export function ReadonlyCode(props: ReadonlyCodeProps): ReactNode {
  return (
    <Suspense fallback={<pre className={props.className}>{props.code}</pre>}>
      <Impl {...props} />
    </Suspense>
  )
}
