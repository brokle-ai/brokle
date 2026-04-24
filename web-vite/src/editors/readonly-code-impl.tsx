import { useEffect, useState } from 'react'
import { codeToHtml } from 'shiki'
import type { ReadonlyCodeProps } from './readonly-code'

// Shiki rendering is async (grammar + theme load on demand). We render
// the raw code immediately for first paint and swap in the highlighted
// HTML once Shiki resolves — prevents CLS while preserving copyable
// text throughout. No client-side runtime to hydrate; the output is
// static HTML.

export default function ReadonlyCodeImpl({
  code,
  language = 'text',
  theme = 'github-light',
  className,
}: ReadonlyCodeProps) {
  const [html, setHtml] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    codeToHtml(code, { lang: language, theme })
      .then((rendered) => {
        if (!cancelled) setHtml(rendered)
      })
      .catch(() => {
        // Unknown lang or runtime failure — keep the fallback text.
        if (!cancelled) setHtml(null)
      })
    return () => {
      cancelled = true
    }
  }, [code, language, theme])

  if (html) {
    return (
      <div
        className={className}
        // Shiki emits trusted, escaped HTML.
        dangerouslySetInnerHTML={{ __html: html }}
      />
    )
  }
  return <pre className={className}>{code}</pre>
}
