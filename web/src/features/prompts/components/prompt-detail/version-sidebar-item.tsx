'use client'

import * as React from 'react'
import { formatDistanceToNow, format } from 'date-fns'
import { ArrowLeftRight, RotateCcw, Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import type { PromptVersion } from '../../types'
import { LabelBadge } from '../label-badge'

// ============================================================================
// Types
// ============================================================================

interface VersionSidebarItemProps {
  version: PromptVersion
  isSelected: boolean
  isLatest: boolean
  isCompareFrom: boolean
  protectedLabels: string[]
  onClick: () => void
  onCompareClick: (e: React.MouseEvent) => void
  onRestoreClick?: (e: React.MouseEvent) => void
  onCopyIdClick: (e: React.MouseEvent) => void
}

// ============================================================================
// CopyIdButton Component
// ============================================================================

function CopyIdButton({
  versionId,
  onClick,
}: {
  versionId: string
  onClick: (e: React.MouseEvent) => void
}) {
  const [copied, setCopied] = React.useState(false)

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    navigator.clipboard.writeText(versionId)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
    onClick(e)
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6"
          onClick={handleClick}
        >
          {copied ? (
            <Check className="h-3 w-3 text-green-500" />
          ) : (
            <Copy className="h-3 w-3" />
          )}
        </Button>
      </TooltipTrigger>
      <TooltipContent side="top">
        <p className="text-xs">{copied ? 'Copied!' : 'Copy version ID'}</p>
      </TooltipContent>
    </Tooltip>
  )
}

// ============================================================================
// Main VersionSidebarItem Component
// ============================================================================

export function VersionSidebarItem({
  version,
  isSelected,
  isLatest,
  isCompareFrom,
  protectedLabels,
  onClick,
  onCompareClick,
  onRestoreClick,
  onCopyIdClick,
}: VersionSidebarItemProps) {
  const createdDate = new Date(version.created_at)
  const relativeTime = formatDistanceToNow(createdDate, { addSuffix: true })
  const absoluteTime = format(createdDate, 'PPpp')

  // Truncate commit message to ~50 chars for sidebar
  const truncatedMessage = version.commit_message
    ? version.commit_message.length > 50
      ? version.commit_message.slice(0, 50) + '...'
      : version.commit_message
    : null

  return (
    <div
      className={cn(
        'group relative pl-10 pr-2 py-2 cursor-pointer transition-colors rounded-md',
        isSelected
          ? 'bg-accent'
          : 'hover:bg-accent/50'
      )}
      onClick={onClick}
    >
      {/* Timeline dot */}
      <div
        className={cn(
          'absolute left-3 top-4 h-3 w-3 rounded-full border-2 bg-background transition-colors',
          isSelected
            ? 'border-primary bg-primary'
            : isLatest
            ? 'border-green-500'
            : 'border-muted-foreground/50'
        )}
      />

      {/* Content */}
      <div className="space-y-1">
        {/* Header row: version badge + labels */}
        <div className="flex items-center gap-1.5 flex-wrap">
          <Badge
            variant={isLatest ? 'default' : 'secondary'}
            className={cn(
              'font-mono text-xs h-5 px-1.5',
              isCompareFrom && 'ring-2 ring-primary ring-offset-1'
            )}
          >
            #{version.version}
          </Badge>
          {version.labels.slice(0, 2).map((label) => (
            <LabelBadge
              key={label}
              label={label}
              isProtected={protectedLabels.includes(label)}
              className="text-xs h-5 px-1.5"
            />
          ))}
          {version.labels.length > 2 && (
            <span className="text-xs text-muted-foreground">
              +{version.labels.length - 2}
            </span>
          )}
        </div>

        {/* Commit message (truncated) */}
        {truncatedMessage && (
          <p className="text-xs text-muted-foreground line-clamp-1">
            {truncatedMessage}
          </p>
        )}

        {/* Timestamp */}
        <p
          className="text-xs text-muted-foreground/70"
          title={absoluteTime}
        >
          {relativeTime}
        </p>
      </div>

      {/* Hover actions */}
      <div
        className={cn(
          'absolute right-1 top-1/2 -translate-y-1/2 flex items-center gap-0.5',
          'opacity-0 group-hover:opacity-100 transition-opacity'
        )}
      >
        {/* Compare button */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant={isCompareFrom ? 'default' : 'ghost'}
              size="icon"
              className="h-6 w-6"
              onClick={onCompareClick}
            >
              <ArrowLeftRight className="h-3 w-3" />
            </Button>
          </TooltipTrigger>
          <TooltipContent side="top">
            <p className="text-xs">
              {isCompareFrom ? 'Cancel compare' : 'Compare versions'}
            </p>
          </TooltipContent>
        </Tooltip>

        {/* Restore button (only for non-latest versions) */}
        {!isLatest && onRestoreClick && (
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6"
                onClick={onRestoreClick}
              >
                <RotateCcw className="h-3 w-3" />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="top">
              <p className="text-xs">Restore this version</p>
            </TooltipContent>
          </Tooltip>
        )}

        {/* Copy ID button */}
        <CopyIdButton versionId={version.id} onClick={onCopyIdClick} />
      </div>
    </div>
  )
}
