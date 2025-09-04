'use client'

import * as React from 'react'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

type StatusVariant = 
  | 'online'
  | 'offline'
  | 'pending'
  | 'success'
  | 'error'
  | 'warning'
  | 'info'

interface StatusIndicatorProps extends React.HTMLAttributes<HTMLDivElement> {
  status: StatusVariant
  label?: string
  showDot?: boolean
  size?: 'sm' | 'md' | 'lg'
  variant?: 'badge' | 'dot' | 'pill'
}

const statusConfig = {
  online: {
    color: 'bg-green-500',
    textColor: 'text-green-700',
    bgColor: 'bg-green-50',
    borderColor: 'border-green-200',
    label: 'Online',
  },
  offline: {
    color: 'bg-gray-500',
    textColor: 'text-gray-700',
    bgColor: 'bg-gray-50',
    borderColor: 'border-gray-200',
    label: 'Offline',
  },
  pending: {
    color: 'bg-yellow-500',
    textColor: 'text-yellow-700',
    bgColor: 'bg-yellow-50',
    borderColor: 'border-yellow-200',
    label: 'Pending',
  },
  success: {
    color: 'bg-green-500',
    textColor: 'text-green-700',
    bgColor: 'bg-green-50',
    borderColor: 'border-green-200',
    label: 'Success',
  },
  error: {
    color: 'bg-red-500',
    textColor: 'text-red-700',
    bgColor: 'bg-red-50',
    borderColor: 'border-red-200',
    label: 'Error',
  },
  warning: {
    color: 'bg-yellow-500',
    textColor: 'text-yellow-700',
    bgColor: 'bg-yellow-50',
    borderColor: 'border-yellow-200',
    label: 'Warning',
  },
  info: {
    color: 'bg-blue-500',
    textColor: 'text-blue-700',
    bgColor: 'bg-blue-50',
    borderColor: 'border-blue-200',
    label: 'Info',
  },
}

const sizeConfig = {
  sm: {
    dot: 'h-2 w-2',
    text: 'text-xs',
    padding: 'px-2 py-1',
  },
  md: {
    dot: 'h-3 w-3',
    text: 'text-sm',
    padding: 'px-3 py-1',
  },
  lg: {
    dot: 'h-4 w-4',
    text: 'text-base',
    padding: 'px-4 py-2',
  },
}

export function StatusIndicator({
  status,
  label,
  showDot = true,
  size = 'md',
  variant = 'badge',
  className,
  ...props
}: StatusIndicatorProps) {
  const config = statusConfig[status]
  const sizeConf = sizeConfig[size]
  const displayLabel = label || config.label

  if (variant === 'dot') {
    return (
      <div
        className={cn('flex items-center gap-2', className)}
        {...props}
      >
        <div
          className={cn(
            'rounded-full',
            sizeConf.dot,
            config.color,
            status === 'online' && 'animate-pulse'
          )}
        />
        {displayLabel && (
          <span className={cn('font-medium', sizeConf.text, config.textColor)}>
            {displayLabel}
          </span>
        )}
      </div>
    )
  }

  if (variant === 'pill') {
    return (
      <div
        className={cn(
          'inline-flex items-center gap-2 rounded-full border',
          sizeConf.padding,
          config.bgColor,
          config.borderColor,
          className
        )}
        {...props}
      >
        {showDot && (
          <div
            className={cn(
              'rounded-full',
              sizeConf.dot,
              config.color,
              status === 'online' && 'animate-pulse'
            )}
          />
        )}
        <span className={cn('font-medium', sizeConf.text, config.textColor)}>
          {displayLabel}
        </span>
      </div>
    )
  }

  // Default badge variant
  const badgeVariant = (() => {
    switch (status) {
      case 'success':
      case 'online':
        return 'default'
      case 'error':
      case 'offline':
        return 'destructive'
      case 'warning':
      case 'pending':
        return 'secondary'
      case 'info':
        return 'outline'
      default:
        return 'secondary'
    }
  })()

  return (
    <Badge variant={badgeVariant} className={cn('gap-1', className)} {...props}>
      {showDot && (
        <div
          className={cn(
            'rounded-full',
            sizeConf.dot,
            config.color,
            status === 'online' && 'animate-pulse'
          )}
        />
      )}
      {displayLabel}
    </Badge>
  )
}