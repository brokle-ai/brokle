'use client'

import { Badge } from '@/components/ui/badge'
import { Clock, DollarSign, Hash, Boxes, Bot, Cpu, Layers } from 'lucide-react'
import { cn } from '@/lib/utils'
import { formatDuration, formatCost } from '../../utils/format-helpers'

interface MetadataBadgeRowProps {
  environment?: string
  duration?: number // nanoseconds
  cost?: number | string
  inputTokens?: number
  outputTokens?: number
  totalTokens?: number
  version?: string
  modelName?: string
  providerName?: string
  level?: string
  className?: string
}

/**
 * Get environment badge color based on env name
 */
function getEnvColor(env: string): string {
  const envLower = env.toLowerCase()
  if (envLower === 'production' || envLower === 'prod') {
    return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
  }
  if (envLower === 'staging' || envLower === 'stage') {
    return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
  }
  if (envLower === 'development' || envLower === 'dev') {
    return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
  }
  if (envLower === 'test') {
    return 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400'
  }
  return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400'
}

/**
 * Format tokens display with input/output breakdown
 */
function formatTokens(
  inputTokens?: number,
  outputTokens?: number,
  totalTokens?: number
): string | null {
  if (inputTokens && outputTokens) {
    return `${inputTokens.toLocaleString()}→${outputTokens.toLocaleString()}`
  }
  if (totalTokens) {
    return totalTokens.toLocaleString()
  }
  if (inputTokens) {
    return `${inputTokens.toLocaleString()}→?`
  }
  if (outputTokens) {
    return `?→${outputTokens.toLocaleString()}`
  }
  return null
}

/**
 * Get level badge color based on level name
 */
function getLevelColor(level: string): string {
  const levelLower = level.toLowerCase()
  if (levelLower === 'error' || levelLower === 'fatal') {
    return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
  }
  if (levelLower === 'warning' || levelLower === 'warn') {
    return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
  }
  if (levelLower === 'debug') {
    return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400'
  }
  return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
}

/**
 * MetadataBadgeRow - Displays key trace/span metrics as a row of badges
 */
export function MetadataBadgeRow({
  environment,
  duration,
  cost,
  inputTokens,
  outputTokens,
  totalTokens,
  version,
  modelName,
  providerName,
  level,
  className,
}: MetadataBadgeRowProps) {
  const formattedDuration = duration ? formatDuration(duration) : null
  const formattedCost = formatCost(cost)
  const formattedTokens = formatTokens(inputTokens, outputTokens, totalTokens)

  return (
    <div className={cn('flex flex-wrap items-center gap-2', className)}>
      {/* Environment Badge */}
      {environment && (
        <Badge
          variant='secondary'
          className={cn('text-xs font-medium', getEnvColor(environment))}
        >
          {environment}
        </Badge>
      )}

      {/* Level Badge */}
      {level && (
        <Badge
          variant='secondary'
          className={cn('text-xs font-medium gap-1', getLevelColor(level))}
        >
          <Layers className='h-3 w-3' />
          {level}
        </Badge>
      )}

      {/* Provider Badge */}
      {providerName && (
        <Badge variant='outline' className='text-xs font-mono gap-1'>
          <Cpu className='h-3 w-3' />
          {providerName}
        </Badge>
      )}

      {/* Model Badge */}
      {modelName && (
        <Badge variant='outline' className='text-xs font-mono gap-1'>
          <Bot className='h-3 w-3' />
          {modelName}
        </Badge>
      )}

      {/* Latency Badge */}
      {formattedDuration && formattedDuration !== '-' && (
        <Badge variant='outline' className='text-xs font-mono gap-1'>
          <Clock className='h-3 w-3' />
          {formattedDuration}
        </Badge>
      )}

      {/* Cost Badge */}
      {formattedCost && formattedCost !== '-' && (
        <Badge variant='outline' className='text-xs font-mono gap-1'>
          <DollarSign className='h-3 w-3' />
          {formattedCost.replace('$', '')}
        </Badge>
      )}

      {/* Tokens Badge */}
      {formattedTokens && (
        <Badge variant='outline' className='text-xs font-mono gap-1'>
          <Hash className='h-3 w-3' />
          {formattedTokens}
        </Badge>
      )}

      {/* Version Badge */}
      {version && (
        <Badge variant='outline' className='text-xs font-mono gap-1'>
          <Boxes className='h-3 w-3' />
          {version}
        </Badge>
      )}
    </div>
  )
}
