'use client'

import * as React from 'react'
import { Link2, ChevronDown, ChevronRight, Copy, Check, ExternalLink } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { cn } from '@/lib/utils'

// ============================================================================
// Types
// ============================================================================

interface LinksListProps {
  traceIds?: string[]
  spanIds?: string[]
  attributes?: string[]
  className?: string
  onLinkClick?: (traceId: string, spanId?: string) => void
}

interface ParsedLink {
  index: number
  traceId: string
  spanId?: string
  attributes: Record<string, any>
}

// ============================================================================
// Copy Button Component
// ============================================================================

function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    await navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Button
      variant='ghost'
      size='icon'
      className='h-5 w-5 opacity-0 group-hover:opacity-100 transition-opacity'
      onClick={handleCopy}
      title='Copy'
    >
      {copied ? (
        <Check className='h-3 w-3 text-green-500' />
      ) : (
        <Copy className='h-3 w-3 text-muted-foreground' />
      )}
    </Button>
  )
}

// ============================================================================
// Link Card - Expandable card for each link
// ============================================================================

interface LinkCardProps {
  link: ParsedLink
  defaultExpanded?: boolean
  onNavigate?: (traceId: string, spanId?: string) => void
}

function LinkCard({ link, defaultExpanded = false, onNavigate }: LinkCardProps) {
  const [isExpanded, setIsExpanded] = React.useState(defaultExpanded)

  const attrCount = Object.keys(link.attributes).length

  const linkJson = React.useMemo(() => {
    return JSON.stringify({
      trace_id: link.traceId,
      span_id: link.spanId,
      attributes: link.attributes,
    }, null, 2)
  }, [link])

  const handleNavigate = (e: React.MouseEvent) => {
    e.stopPropagation()
    onNavigate?.(link.traceId, link.spanId)
  }

  return (
    <div className='rounded-lg border bg-muted/30 border-border'>
      <Collapsible open={isExpanded} onOpenChange={setIsExpanded}>
        <CollapsibleTrigger asChild>
          <div className='group flex items-center gap-2 p-3 cursor-pointer hover:bg-muted/50 transition-colors'>
            {isExpanded ? (
              <ChevronDown className='h-4 w-4 text-muted-foreground flex-shrink-0' />
            ) : (
              <ChevronRight className='h-4 w-4 text-muted-foreground flex-shrink-0' />
            )}

            <Link2 className='h-4 w-4 text-blue-500 flex-shrink-0' />

            <div className='flex-1 min-w-0'>
              <div className='flex items-center gap-2 flex-wrap'>
                <span className='text-xs text-muted-foreground'>Trace:</span>
                <code className='text-xs font-mono bg-background/50 px-1.5 py-0.5 rounded border'>
                  {link.traceId.substring(0, 8)}...
                </code>
                {link.spanId && (
                  <>
                    <span className='text-xs text-muted-foreground'>Span:</span>
                    <code className='text-xs font-mono bg-background/50 px-1.5 py-0.5 rounded border'>
                      {link.spanId.substring(0, 8)}...
                    </code>
                  </>
                )}
              </div>
            </div>

            {attrCount > 0 && (
              <Badge variant='secondary' className='text-xs flex-shrink-0'>
                {attrCount} attr{attrCount !== 1 ? 's' : ''}
              </Badge>
            )}

            {onNavigate && (
              <Button
                variant='ghost'
                size='icon'
                className='h-6 w-6 flex-shrink-0'
                onClick={handleNavigate}
                title='Navigate to linked span'
              >
                <ExternalLink className='h-3 w-3' />
              </Button>
            )}

            <CopyButton value={linkJson} />
          </div>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <div className='px-3 pb-3 pt-1 border-t border-border/50 space-y-3'>
            {/* Full IDs */}
            <div className='space-y-1.5'>
              <div className='flex items-start gap-2'>
                <span className='text-xs text-muted-foreground font-mono min-w-[80px] flex-shrink-0'>
                  trace_id:
                </span>
                <code className='text-xs font-mono text-foreground break-all'>
                  {link.traceId}
                </code>
              </div>
              {link.spanId && (
                <div className='flex items-start gap-2'>
                  <span className='text-xs text-muted-foreground font-mono min-w-[80px] flex-shrink-0'>
                    span_id:
                  </span>
                  <code className='text-xs font-mono text-foreground break-all'>
                    {link.spanId}
                  </code>
                </div>
              )}
            </div>

            {/* Attributes */}
            {attrCount > 0 && (
              <div className='space-y-1.5 pt-2 border-t border-border/30'>
                <span className='text-xs font-medium text-muted-foreground'>Attributes:</span>
                {Object.entries(link.attributes).map(([key, value]) => (
                  <div key={key} className='flex items-start gap-2'>
                    <span className='text-xs text-muted-foreground font-mono min-w-[100px] flex-shrink-0'>
                      {key}:
                    </span>
                    <span className='text-xs font-mono text-foreground break-all'>
                      {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>
        </CollapsibleContent>
      </Collapsible>
    </div>
  )
}

// ============================================================================
// LinksList - Main Component
// ============================================================================

/**
 * LinksList - Display OTEL span links
 *
 * Features:
 * - List of span links with trace_id and span_id
 * - Clickable to navigate to linked trace/span
 * - Expandable to show link attributes
 * - Copy functionality
 */
export function LinksList({
  traceIds,
  spanIds,
  attributes,
  className,
  onLinkClick,
}: LinksListProps) {
  // Parse links into structured format
  const links: ParsedLink[] = React.useMemo(() => {
    if (!traceIds || traceIds.length === 0) return []

    return traceIds.map((traceId, index) => {
      let parsedAttrs: Record<string, any> = {}
      if (attributes && attributes[index]) {
        try {
          parsedAttrs = JSON.parse(attributes[index])
        } catch {
          // Keep empty if parse fails
        }
      }

      return {
        index,
        traceId,
        spanId: spanIds?.[index],
        attributes: parsedAttrs,
      }
    })
  }, [traceIds, spanIds, attributes])

  if (links.length === 0) {
    return (
      <div className={cn('py-6 text-center', className)}>
        <p className='text-sm text-muted-foreground italic'>No links</p>
      </div>
    )
  }

  return (
    <div className={cn('space-y-2', className)}>
      {/* Summary header */}
      <div className='flex items-center gap-2 py-1'>
        <Link2 className='h-4 w-4 text-blue-500' />
        <span className='text-sm font-medium'>{links.length} Link{links.length !== 1 ? 's' : ''}</span>
      </div>

      {/* Link cards */}
      <div className='space-y-2'>
        {links.map((link) => (
          <LinkCard
            key={`${link.traceId}-${link.index}`}
            link={link}
            defaultExpanded={links.length <= 3}
            onNavigate={onLinkClick}
          />
        ))}
      </div>
    </div>
  )
}
