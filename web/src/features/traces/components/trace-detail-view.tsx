'use client'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import type { Trace } from '../data/schema'
import { statuses } from '../data/constants'
import { format } from 'date-fns'
import { Clock, DollarSign, Layers, Server, Tag } from 'lucide-react'

interface TraceDetailViewProps {
  trace: Trace
}

function formatDuration(ms: number | undefined): string {
  if (!ms) return '-'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatCost(cost: number | undefined): string {
  if (!cost) return '-'
  return `$${cost.toFixed(6)}`
}

export function TraceDetailView({ trace }: TraceDetailViewProps) {
  const status = statuses.find((s) => s.value === trace.status)
  const StatusIcon = status?.icon

  return (
    <div className='space-y-6'>
      {/* Header */}
      <div>
        <div className='flex items-center gap-3 mb-2'>
          <h2 className='text-2xl font-bold'>{trace.name}</h2>
          {StatusIcon && (
            <div className='flex items-center gap-2'>
              <StatusIcon className='h-4 w-4' />
              <span className='text-sm'>{status.label}</span>
            </div>
          )}
        </div>
        <p className='text-sm text-muted-foreground font-mono'>{trace.id}</p>
      </div>

      {/* Metrics Grid */}
      <div className='grid grid-cols-2 md:grid-cols-4 gap-4'>
        <Card>
          <CardHeader className='pb-2'>
            <CardTitle className='text-sm font-medium text-muted-foreground flex items-center gap-2'>
              <Clock className='h-4 w-4' />
              Duration
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className='text-2xl font-bold'>{formatDuration(trace.durationMs)}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className='pb-2'>
            <CardTitle className='text-sm font-medium text-muted-foreground flex items-center gap-2'>
              <DollarSign className='h-4 w-4' />
              Cost
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className='text-2xl font-bold'>{formatCost(trace.cost)}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className='pb-2'>
            <CardTitle className='text-sm font-medium text-muted-foreground flex items-center gap-2'>
              <Layers className='h-4 w-4' />
              Observations
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className='text-2xl font-bold'>{trace.observationCount}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className='pb-2'>
            <CardTitle className='text-sm font-medium text-muted-foreground flex items-center gap-2'>
              <Server className='h-4 w-4' />
              Tokens
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className='text-2xl font-bold'>{trace.tokens?.toLocaleString() || '-'}</div>
          </CardContent>
        </Card>
      </div>

      {/* Metadata Card */}
      <Card>
        <CardHeader>
          <CardTitle>Trace Metadata</CardTitle>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='grid grid-cols-2 gap-4'>
            <div>
              <div className='text-sm font-medium text-muted-foreground'>Start Time</div>
              <div className='text-sm'>{format(trace.startTime, 'PPpp')}</div>
            </div>
            {trace.endTime && (
              <div>
                <div className='text-sm font-medium text-muted-foreground'>End Time</div>
                <div className='text-sm'>{format(trace.endTime, 'PPpp')}</div>
              </div>
            )}
            {trace.environment && (
              <div>
                <div className='text-sm font-medium text-muted-foreground'>Environment</div>
                <Badge variant='outline'>{trace.environment}</Badge>
              </div>
            )}
            {trace.serviceName && (
              <div>
                <div className='text-sm font-medium text-muted-foreground'>Service</div>
                <div className='text-sm'>{trace.serviceName}</div>
              </div>
            )}
            {trace.serviceVersion && (
              <div>
                <div className='text-sm font-medium text-muted-foreground'>Version</div>
                <div className='text-sm font-mono'>{trace.serviceVersion}</div>
              </div>
            )}
          </div>

          {trace.tags && trace.tags.length > 0 && (
            <>
              <Separator />
              <div>
                <div className='text-sm font-medium text-muted-foreground mb-2 flex items-center gap-2'>
                  <Tag className='h-4 w-4' />
                  Tags
                </div>
                <div className='flex flex-wrap gap-2'>
                  {trace.tags.map((tag) => (
                    <Badge key={tag} variant='secondary'>
                      {tag}
                    </Badge>
                  ))}
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Observations Section - Placeholder */}
      <Card>
        <CardHeader>
          <CardTitle>Observations</CardTitle>
        </CardHeader>
        <CardContent>
          <div className='text-sm text-muted-foreground'>
            {trace.observationCount} observation{trace.observationCount !== 1 ? 's' : ''} in this trace
          </div>
          <p className='text-sm text-muted-foreground mt-2'>
            Detailed observation view coming soon...
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
