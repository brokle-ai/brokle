'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { TraceDetailView } from '@/features/traces'
import { traces } from '@/features/traces/data/traces'
import { Button } from '@/components/ui/button'
import { ArrowLeft } from 'lucide-react'
import Link from 'next/link'
import type { ProjectParams } from '@/features/organizations'

interface TracePageParams extends ProjectParams {
  traceId: string
}

export default function TraceDetailPage() {
  const params = useParams() as TracePageParams

  // Find trace from mock data
  // TODO: Replace with API call when backend ready
  const trace = traces.find((t) => t.trace_id === params.traceId)

  if (!trace) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className='flex flex-col items-center justify-center py-12'>
            <h2 className='text-2xl font-bold mb-2'>Trace Not Found</h2>
            <p className='text-muted-foreground mb-4'>
              The trace you're looking for doesn't exist or has been deleted.
            </p>
            <Link href={`/projects/${params.projectSlug}/traces`}>
              <Button variant='outline'>
                <ArrowLeft className='mr-2 h-4 w-4' />
                Back to Traces
              </Button>
            </Link>
          </div>
        </Main>
      </>
    )
  }

  return (
    <>
      <DashboardHeader />
      <Main>
        <div className='mb-4'>
          <Link href={`/projects/${params.projectSlug}/traces`}>
            <Button variant='ghost' size='sm'>
              <ArrowLeft className='mr-2 h-4 w-4' />
              Back to Traces
            </Button>
          </Link>
        </div>
        <TraceDetailView trace={trace} />
      </Main>
    </>
  )
}
