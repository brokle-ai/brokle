'use client'

import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'

interface PageHeaderProps {
  /** Page title */
  title: string
  /** Action buttons or other elements (displayed on the right) */
  children?: React.ReactNode
  /** Back navigation URL - if provided, shows back arrow */
  backHref?: string
  /** Optional description shown below title */
  description?: string | null
  /** Optional metadata like "Created 2 hours ago" */
  metadata?: React.ReactNode
  /** Optional badges shown inline with title */
  badges?: React.ReactNode
}

/**
 * PageHeader - Unified header for all pages
 *
 * @example List page
 * <PageHeader title="Datasets">
 *   <CreateDatasetDialog />
 * </PageHeader>
 *
 * @example Detail page
 * <PageHeader
 *   title={dataset.name}
 *   backHref={`/projects/${projectSlug}/datasets`}
 *   description={dataset.description}
 *   metadata={`Created ${formatDistanceToNow(dataset.created_at)}`}
 * >
 *   <Button onClick={handleEdit}>Edit</Button>
 * </PageHeader>
 */
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
              href={backHref}
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
