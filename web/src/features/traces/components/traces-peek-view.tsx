'use client'

import { useEffect, useRef } from 'react'
import { useSearchParams } from 'next/navigation'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import { X, Maximize2, ExternalLink } from 'lucide-react'
import { usePeekNavigation } from '../hooks/use-peek-navigation'
import { useDetailNavigation } from '../hooks/use-detail-navigation'
import { PeekViewTraceDetail } from './peek-view-trace-detail'
import { DetailPageNav } from './detail-page-nav'

/**
 * Peek sheet component for viewing trace details
 * Slides in from right, non-modal, can be expanded to full page
 */
export function TracesPeekView() {
  const searchParams = useSearchParams()
  const peekId = searchParams.get('peek')
  const { closePeek, expandPeek } = usePeekNavigation()
  const { handlePrev, handleNext } = useDetailNavigation()

  // Ref pattern for stable keyboard handlers
  const handlersRef = useRef({ handlePrev, handleNext })

  useEffect(() => {
    handlersRef.current = { handlePrev, handleNext }
  }, [handlePrev, handleNext])

  // Keyboard shortcuts - document level to work with Sheet focus trap
  useEffect(() => {
    if (!peekId) return

    const handleKeyDown = (e: KeyboardEvent) => {
      // Arrow keys for navigation
      if (e.key === 'ArrowLeft' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault()
        handlersRef.current.handlePrev()
      }
      if (e.key === 'ArrowRight' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault()
        handlersRef.current.handleNext()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [peekId])

  if (!peekId) return null

  return (
    <Sheet open={!!peekId} onOpenChange={(open) => !open && closePeek()} modal={false}>
      <SheetContent
        side='right'
        className='flex max-h-full min-h-0 min-w-[60vw] flex-col gap-0 overflow-hidden rounded-l-xl p-0'
        onPointerDownOutside={(e) => {
          // Prevent sheet closure when clicking outside with modal={false}
          e.preventDefault()
        }}
        tabIndex={-1}
      >
        {/* Header */}
        <SheetHeader className='flex min-h-12 flex-row flex-nowrap items-center justify-between space-y-0 px-4 py-2'>
          <SheetTitle className='text-sm font-mono'>
            Trace: {peekId.substring(0, 8)}...
          </SheetTitle>

          <div className='flex items-center gap-2 mr-8'>
            {/* Prev/Next Navigation */}
            <DetailPageNav />

            {/* Expand Buttons */}
            <Button
              variant='ghost'
              size='icon'
              className='h-8 w-8'
              onClick={() => expandPeek(false)}
              title='Open in current tab'
            >
              <Maximize2 className='h-4 w-4' />
            </Button>

            <Button
              variant='ghost'
              size='icon'
              className='h-8 w-8'
              onClick={() => expandPeek(true)}
              title='Open in new tab'
            >
              <ExternalLink className='h-4 w-4' />
            </Button>

          </div>
        </SheetHeader>

        {/* Scrollable Content */}
        <div className='flex-1 overflow-auto p-6'>
          <PeekViewTraceDetail />
        </div>
      </SheetContent>
    </Sheet>
  )
}
