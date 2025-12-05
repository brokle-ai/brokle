'use client'

import * as React from 'react'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import {
  ChevronUp,
  ChevronDown,
  Maximize2,
  ExternalLink,
  X,
  Copy,
  Check,
  ListTree,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Trace } from '../../data/schema'

interface PeekSheetHeaderProps {
  trace: Trace
  onPrevious?: () => void
  onNext?: () => void
  onExpand: () => void
  onExpandNewTab: () => void
  onClose: () => void
  hasPrevious: boolean
  hasNext: boolean
  className?: string
}

/**
 * CopyButton - Small inline copy button with feedback
 */
function CopyButton({ value, className }: { value: string; className?: string }) {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Button
      variant='ghost'
      size='icon'
      className={cn('h-5 w-5', className)}
      onClick={handleCopy}
    >
      {copied ? (
        <Check className='h-3 w-3 text-green-500' />
      ) : (
        <Copy className='h-3 w-3 text-muted-foreground hover:text-foreground' />
      )}
    </Button>
  )
}

/**
 * PeekSheetHeader - Header for the trace peek sheet
 * Displays trace identification, badges, and navigation controls
 */
export function PeekSheetHeader({
  trace,
  onPrevious,
  onNext,
  onExpand,
  onExpandNewTab,
  onClose,
  hasPrevious,
  hasNext,
  className,
}: PeekSheetHeaderProps) {
  return (
    <div className={cn('border-b bg-background', className)}>
      {/* Minimal header: ID + Copy + Navigation */}
      <div className='flex items-center justify-between px-4 py-3'>
        <div className='flex items-center gap-2'>
          {/* Trace prefix */}
          <div className='flex items-center gap-1.5'>
            <ListTree className='h-4 w-4' />
            <span className='text-sm font-medium'>Trace</span>
          </div>
          {/* Trace ID with copy */}
          <span className='text-sm font-medium'>
            {trace.trace_id}
          </span>
          <CopyButton value={trace.trace_id} />
        </div>

        {/* Navigation Controls */}
        <div className='flex items-center gap-1'>
          {/* Prev/Next */}
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-8 w-8'
                  onClick={onPrevious}
                  disabled={!hasPrevious}
                >
                  <ChevronUp className='h-4 w-4' />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Previous trace (←)</TooltipContent>
            </Tooltip>
          </TooltipProvider>

          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-8 w-8'
                  onClick={onNext}
                  disabled={!hasNext}
                >
                  <ChevronDown className='h-4 w-4' />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Next trace (→)</TooltipContent>
            </Tooltip>
          </TooltipProvider>

          <div className='w-px h-6 bg-border mx-1' />

          {/* Expand */}
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-8 w-8'
                  onClick={onExpand}
                >
                  <Maximize2 className='h-4 w-4' />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Open in current tab</TooltipContent>
            </Tooltip>
          </TooltipProvider>

          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-8 w-8'
                  onClick={onExpandNewTab}
                >
                  <ExternalLink className='h-4 w-4' />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Open in new tab</TooltipContent>
            </Tooltip>
          </TooltipProvider>

          <div className='w-px h-6 bg-border mx-1' />

          {/* Close */}
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-8 w-8'
                  onClick={onClose}
                >
                  <X className='h-4 w-4' />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Close (Esc)</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>
    </div>
  )
}
