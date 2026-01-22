'use client'

import * as React from 'react'
import { FlaskConical } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { useProjectOnly } from '@/features/projects'
import { usePlaygroundStore } from '@/features/playground/stores/playground-store'
import {
  isLLMSpan,
  parseSpanToMessages,
  extractModelConfig,
  getDisabledReason,
} from '@/features/playground/utils/parse-span-messages'
import type { Span } from '../../data/schema'

interface OpenInPlaygroundButtonProps {
  span: Span
  traceId: string
}

/**
 * Button to open an LLM span in the playground
 * Only enabled for LLM spans with valid input data
 *
 * Design inspired by Langfuse's JumpToPlaygroundButton
 */
export function OpenInPlaygroundButton({ span, traceId }: OpenInPlaygroundButtonProps) {
  const router = useRouter()
  const { currentProject } = useProjectOnly()
  const loadFromSpan = usePlaygroundStore((s) => s.loadFromSpan)

  // Check if this span can be opened in playground
  const disabledReason = getDisabledReason(span)
  const isDisabled = disabledReason !== null

  const handleOpenInPlayground = () => {
    if (isDisabled || !currentProject?.compositeSlug) return

    // Parse messages from span input
    const messages = parseSpanToMessages(span.input)
    if (messages.length === 0) return

    // Extract model config from span attributes
    const config = extractModelConfig(span)

    // Load into playground store
    loadFromSpan({
      messages,
      config,
      loadedFromSpanId: span.span_id,
      loadedFromSpanName: span.span_name,
      loadedFromTraceId: traceId,
    })

    // Navigate to playground
    router.push(`/projects/${currentProject.compositeSlug}/playground`)
  }

  // Only render for LLM spans
  if (!isLLMSpan(span)) {
    return null
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant='ghost'
            size='icon'
            className='h-5 w-5 hover:bg-muted'
            onClick={handleOpenInPlayground}
            disabled={isDisabled}
          >
            <FlaskConical className='h-3 w-3 text-muted-foreground' />
          </Button>
        </TooltipTrigger>
        <TooltipContent side='bottom'>
          {isDisabled ? (
            <p className='text-xs'>{disabledReason}</p>
          ) : (
            <p className='text-xs'>Open in Playground</p>
          )}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
