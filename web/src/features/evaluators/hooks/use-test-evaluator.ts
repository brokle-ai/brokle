'use client'

import { useState, useCallback } from 'react'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { extractErrorMessage } from '@/lib/api/error-utils'
import { evaluatorsApi } from '../api/evaluators-api'
import type {
  TestEvaluatorRequest,
  TestEvaluatorResponse,
  TestSummary,
  TestExecution,
  EvaluatorPreview,
} from '../types'

export interface TestEvaluatorResult {
  status: 'idle' | 'running' | 'completed' | 'failed'
  summary?: TestSummary
  executions?: TestExecution[]
  evaluatorPreview?: EvaluatorPreview
  error?: string
}

export interface UseTestEvaluatorOptions {
  /** Default sample limit for test runs (default: 5) */
  defaultSampleLimit?: number
  /** Default time range for test runs (default: '24h') */
  defaultTimeRange?: string
}

export interface UseTestEvaluatorReturn {
  /** Current test result state */
  result: TestEvaluatorResult
  /** Start a test run */
  startTest: (options?: TestOptions) => Promise<void>
  /** Reset the test state */
  resetTest: () => void
  /** Whether a test is currently in progress */
  isRunning: boolean
}

export interface TestOptions {
  /** Specific trace ID to test against */
  traceId?: string
  /** Specific span ID to test against */
  spanId?: string
  /** Specific span IDs to test against */
  spanIds?: string[]
  /** Maximum spans to evaluate (default: 5) */
  sampleLimit?: number
  /** Time range for span selection: "1h", "24h", "7d" (default: "24h") */
  timeRange?: string
}

/**
 * Hook for testing evaluators before activation.
 *
 * Uses the test endpoint which performs a dry-run evaluation
 * without persisting results to the database.
 *
 * @example
 * ```tsx
 * const { result, startTest, isRunning } = useTestEvaluator(projectId, evaluatorId)
 *
 * await startTest({ sampleLimit: 5, timeRange: '24h' })
 *
 * if (result.status === 'completed') {
 *   console.log('Summary:', result.summary)
 *   console.log('Executions:', result.executions)
 * }
 * ```
 */
export function useTestEvaluator(
  projectId: string,
  evaluatorId: string,
  options: UseTestEvaluatorOptions = {}
): UseTestEvaluatorReturn {
  const { defaultSampleLimit = 5, defaultTimeRange = '24h' } = options

  const [result, setResult] = useState<TestEvaluatorResult>({
    status: 'idle',
  })

  // Test mutation
  const testMutation = useMutation({
    mutationFn: (request: TestEvaluatorRequest) =>
      evaluatorsApi.testEvaluator(projectId, evaluatorId, request),
    onSuccess: (response: TestEvaluatorResponse) => {
      setResult({
        status: 'completed',
        summary: response.summary,
        executions: response.executions,
        evaluatorPreview: response.evaluator_preview,
      })

      const successCount = response.summary.success_count
      const totalEvaluated = response.summary.evaluated_spans

      toast.success('Test Completed', {
        description: `Successfully evaluated ${totalEvaluated} span${totalEvaluated !== 1 ? 's' : ''} (${successCount} passed).`,
      })
    },
    onError: (error: unknown) => {
      const errorMessage = extractErrorMessage(error, 'Failed to run test')
      setResult({
        status: 'failed',
        error: errorMessage,
      })
      toast.error('Test Failed', {
        description: errorMessage,
      })
    },
  })

  // Start test
  const startTest = useCallback(
    async (testOptions: TestOptions = {}) => {
      // Reset state and set running
      setResult({
        status: 'running',
      })

      // Build test request
      const request: TestEvaluatorRequest = {
        limit: testOptions.sampleLimit ?? defaultSampleLimit,
        time_range: testOptions.timeRange ?? defaultTimeRange,
      }

      if (testOptions.traceId) {
        request.trace_id = testOptions.traceId
      }

      if (testOptions.spanId) {
        request.span_id = testOptions.spanId
      }

      if (testOptions.spanIds && testOptions.spanIds.length > 0) {
        request.span_ids = testOptions.spanIds
      }

      // Execute test
      await testMutation.mutateAsync(request)
    },
    [testMutation, defaultSampleLimit, defaultTimeRange]
  )

  // Reset test
  const resetTest = useCallback(() => {
    setResult({
      status: 'idle',
    })
  }, [])

  return {
    result,
    startTest,
    resetTest,
    isRunning: result.status === 'running',
  }
}
