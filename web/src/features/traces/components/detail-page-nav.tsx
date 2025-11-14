'use client'

import { Button } from '@/components/ui/button'
import { ChevronUp, ChevronDown } from 'lucide-react'
import { useDetailNavigation } from '../hooks/use-detail-navigation'

/**
 * Navigation buttons for peek view
 * Up/down arrows for prev/next navigation (no position indicator)
 */
export function DetailPageNav() {
  const { canGoPrev, canGoNext, handlePrev, handleNext } = useDetailNavigation()

  return (
    <div className='flex items-center gap-1'>
      <Button
        variant='outline'
        size='icon'
        className='h-8 w-8'
        onClick={handlePrev}
        disabled={!canGoPrev}
        aria-label='Previous trace (K)'
      >
        <ChevronUp className='h-4 w-4' />
      </Button>

      <Button
        variant='outline'
        size='icon'
        className='h-8 w-8'
        onClick={handleNext}
        disabled={!canGoNext}
        aria-label='Next trace (J)'
      >
        <ChevronDown className='h-4 w-4' />
      </Button>
    </div>
  )
}
