import { Link } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import type { ReactNode } from 'react'

// TanStack Router equivalent of the next/link-based PageHeader. The
// web/ version accepts a `backHref` as a raw URL string; we keep that
// API (string href → anchor-style fallback via <Link to=...>) but
// route strings use TanStack's path syntax at call sites.

interface PageHeaderProps {
  title: string
  children?: ReactNode
  backHref?: string
  description?: string | null
  metadata?: ReactNode
  badges?: ReactNode
}

export function PageHeader({
  title,
  children,
  backHref,
  description,
  metadata,
  badges,
}: PageHeaderProps) {
  return (
    <div className="flex items-center justify-between mt-2 mb-2">
      <div className="space-y-1">
        <div className="flex items-center gap-3">
          {backHref && (
            <Link
              to={backHref}
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              <ArrowLeft className="h-5 w-5" />
            </Link>
          )}
          <h1 className="text-lg font-semibold">{title}</h1>
          {badges}
        </div>
        {description && (
          <p className="text-muted-foreground ml-8">{description}</p>
        )}
        {metadata && (
          <div className="text-sm text-muted-foreground ml-8">{metadata}</div>
        )}
      </div>
      {children && <div className="flex items-center gap-2">{children}</div>}
    </div>
  )
}
