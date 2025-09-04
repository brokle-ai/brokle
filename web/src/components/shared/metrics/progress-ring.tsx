'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'

interface ProgressRingProps extends React.HTMLAttributes<HTMLDivElement> {
  progress: number // 0-100
  size?: 'sm' | 'md' | 'lg' | 'xl'
  strokeWidth?: number
  showValue?: boolean
  label?: string
  color?: 'primary' | 'secondary' | 'success' | 'warning' | 'destructive'
}

const sizeMap = {
  sm: { size: 60, text: 'text-xs' },
  md: { size: 80, text: 'text-sm' },
  lg: { size: 120, text: 'text-base' },
  xl: { size: 160, text: 'text-lg' },
}

const colorMap = {
  primary: 'stroke-primary',
  secondary: 'stroke-secondary',
  success: 'stroke-green-500',
  warning: 'stroke-yellow-500',
  destructive: 'stroke-destructive',
}

export function ProgressRing({
  progress,
  size = 'md',
  strokeWidth = 8,
  showValue = true,
  label,
  color = 'primary',
  className,
  ...props
}: ProgressRingProps) {
  const { size: diameter, text: textSize } = sizeMap[size]
  const radius = (diameter - strokeWidth) / 2
  const circumference = 2 * Math.PI * radius
  const strokeDasharray = circumference
  const strokeDashoffset = circumference - (progress / 100) * circumference

  // Ensure progress is between 0 and 100
  const normalizedProgress = Math.min(100, Math.max(0, progress))

  return (
    <div
      className={cn('relative inline-flex items-center justify-center', className)}
      style={{ width: diameter, height: diameter }}
      {...props}
    >
      <svg
        width={diameter}
        height={diameter}
        className='transform -rotate-90'
        viewBox={`0 0 ${diameter} ${diameter}`}
      >
        {/* Background circle */}
        <circle
          cx={diameter / 2}
          cy={diameter / 2}
          r={radius}
          fill='none'
          stroke='currentColor'
          strokeWidth={strokeWidth}
          className='text-muted stroke-current opacity-20'
        />
        {/* Progress circle */}
        <circle
          cx={diameter / 2}
          cy={diameter / 2}
          r={radius}
          fill='none'
          stroke='currentColor'
          strokeWidth={strokeWidth}
          strokeLinecap='round'
          strokeDasharray={strokeDasharray}
          strokeDashoffset={strokeDashoffset}
          className={cn('transition-all duration-300 ease-in-out', colorMap[color])}
        />
      </svg>
      
      {/* Center content */}
      <div className='absolute inset-0 flex flex-col items-center justify-center'>
        {showValue && (
          <span className={cn('font-semibold', textSize)}>
            {Math.round(normalizedProgress)}%
          </span>
        )}
        {label && (
          <span className={cn('text-muted-foreground', 
            size === 'sm' ? 'text-xs' : 'text-xs'
          )}>
            {label}
          </span>
        )}
      </div>
    </div>
  )
}