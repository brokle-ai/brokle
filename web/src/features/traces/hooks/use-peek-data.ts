'use client'

import { useEffect, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import { traces } from '../data/traces'
import type { Trace } from '../data/schema'

/**
 * Hook to fetch trace details for peek view
 * Only fetches when peek parameter exists in URL
 */
export function usePeekData() {
  const searchParams = useSearchParams()
  const peekId = searchParams.get('peek')

  const [trace, setTrace] = useState<Trace | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    if (!peekId) {
      setTrace(null)
      setError(null)
      return
    }

    const fetchTrace = async () => {
      setIsLoading(true)
      setError(null)

      try {
        // TODO: Replace with real API call when backend ready
        // const result = await getTraceById(projectSlug, peekId)

        // MOCK: Find trace from mock data
        await new Promise((resolve) => setTimeout(resolve, 300)) // Simulate network delay
        const foundTrace = traces.find((t) => t.id === peekId)

        if (!foundTrace) {
          throw new Error('Trace not found')
        }

        setTrace(foundTrace)
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to load trace'))
        setTrace(null)
      } finally {
        setIsLoading(false)
      }
    }

    fetchTrace()
  }, [peekId])

  return {
    trace,
    isLoading,
    error,
    peekId,
  }
}
