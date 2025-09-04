'use client'

import * as React from 'react'
import { LucideIcon } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { cn } from '@/lib/utils'

interface MetricCardProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string
  value: string | number
  description?: string
  icon?: LucideIcon | React.ComponentType<{ className?: string }>
  trend?: {
    value: number
    label: string
    direction: 'up' | 'down'
  }
  loading?: boolean
  error?: string
}

export function MetricCard({
  title,
  value,
  description,
  icon: Icon,
  trend,
  loading = false,
  error,
  className,
  ...props
}: MetricCardProps) {
  if (loading) {
    return (
      <Card className={cn('animate-pulse', className)} {...props}>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <div className='h-4 w-24 bg-muted rounded' />
          <div className='h-4 w-4 bg-muted rounded' />
        </CardHeader>
        <CardContent>
          <div className='h-8 w-32 bg-muted rounded mb-2' />
          <div className='h-3 w-40 bg-muted rounded' />
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className={cn('border-destructive', className)} {...props}>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium text-destructive'>
            {title}
          </CardTitle>
          {Icon && <Icon className='h-4 w-4 text-destructive' />}
        </CardHeader>
        <CardContent>
          <div className='text-sm text-destructive'>
            Error loading metric
          </div>
          <p className='text-xs text-muted-foreground mt-1'>{error}</p>
        </CardContent>
      </Card>
    )
  }

  const formatValue = (val: string | number) => {
    if (typeof val === 'number') {
      // Format large numbers with commas
      return val.toLocaleString()
    }
    return val
  }

  const getTrendColor = (direction: 'up' | 'down') => {
    return direction === 'up' ? 'text-green-600' : 'text-red-600'
  }

  const getTrendIcon = (direction: 'up' | 'down') => {
    return direction === 'up' ? '↗' : '↘'
  }

  return (
    <Card className={className} {...props}>
      <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
        <CardTitle className='text-sm font-medium'>{title}</CardTitle>
        {Icon && <Icon className='h-4 w-4 text-muted-foreground' />}
      </CardHeader>
      <CardContent>
        <div className='text-2xl font-bold'>{formatValue(value)}</div>
        {(description || trend) && (
          <div className='flex items-center justify-between'>
            {description && (
              <p className='text-xs text-muted-foreground flex-1'>
                {description}
              </p>
            )}
            {trend && (
              <p className={cn('text-xs flex items-center', getTrendColor(trend.direction))}>
                <span className='mr-1'>{getTrendIcon(trend.direction)}</span>
                {trend.value}% {trend.label}
              </p>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}