'use client'

import * as React from 'react'
import * as SliderPrimitive from '@radix-ui/react-slider'
import { cn } from '@/lib/utils'

interface BipolarSliderProps
  extends Omit<React.ComponentPropsWithoutRef<typeof SliderPrimitive.Root>, 'value' | 'onValueChange'> {
  value: number
  onValueChange: (value: number) => void
  origin?: number // Default: midpoint of min/max
}

const BipolarSlider = React.forwardRef<
  React.ElementRef<typeof SliderPrimitive.Root>,
  BipolarSliderProps
>(({ className, value, onValueChange, min = -2, max = 2, origin, ...props }, ref) => {
  const effectiveOrigin = origin ?? (min + max) / 2
  const range = max - min

  // Calculate percentages for the fill
  const originPercent = ((effectiveOrigin - min) / range) * 100
  const valuePercent = ((value - min) / range) * 100

  // Determine fill position and width
  const fillLeft = Math.min(originPercent, valuePercent)
  const fillWidth = Math.abs(valuePercent - originPercent)

  return (
    <SliderPrimitive.Root
      ref={ref}
      className={cn(
        'relative flex w-full touch-none select-none items-center',
        className
      )}
      value={[value]}
      onValueChange={([v]) => onValueChange(v)}
      min={min}
      max={max}
      {...props}
    >
      <SliderPrimitive.Track className="relative h-2 w-full grow overflow-hidden rounded-full bg-secondary">
        {/* Custom range fill from origin to value */}
        <div
          className="absolute h-full bg-primary"
          style={{
            left: `${fillLeft}%`,
            width: `${fillWidth}%`,
          }}
        />
      </SliderPrimitive.Track>
      <SliderPrimitive.Thumb className="block h-5 w-5 rounded-full border-2 border-primary bg-background ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50" />
    </SliderPrimitive.Root>
  )
})
BipolarSlider.displayName = 'BipolarSlider'

export { BipolarSlider }
