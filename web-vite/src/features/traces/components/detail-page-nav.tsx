import { ChevronUp, ChevronDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useDetailNavigation } from '../hooks/use-detail-navigation'

// Compact chevron prev/next pair — mounted inside the peek sheet
// header. Rendered on the full-page detail route is also possible
// when we wire it up to the route-provided trace ID, but today the
// only caller is the peek sheet's header.
export function DetailPageNav() {
  const { canGoPrev, canGoNext, handlePrev, handleNext } =
    useDetailNavigation()

  return (
    <div className="flex items-center gap-1">
      <Button
        variant="outline"
        size="icon"
        className="h-8 w-8"
        onClick={handlePrev}
        disabled={!canGoPrev}
        aria-label="Previous trace"
      >
        <ChevronUp className="h-4 w-4" />
      </Button>
      <Button
        variant="outline"
        size="icon"
        className="h-8 w-8"
        onClick={handleNext}
        disabled={!canGoNext}
        aria-label="Next trace"
      >
        <ChevronDown className="h-4 w-4" />
      </Button>
    </div>
  )
}
