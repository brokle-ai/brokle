import { useState, useCallback, useRef } from 'react'
import { rawFetch } from '@/lib/api/client'
import type {
  StreamChunk,
  StreamMetrics,
  ExecuteRequest,
  ChatMessage,
  ModelConfig,
} from '../types'

/**
 * Captured inputs at the start of execution.
 * Used to create accurate history entries even if user edits during streaming.
 */
export interface CapturedInputs {
  messages: ChatMessage[]
  variables: Record<string, string>
  config: ModelConfig | null
}

interface UseStreamingOptions {
  onStart?: () => void
  onContent?: (content: string) => void
  onEnd?: (
    content: string,
    metrics: StreamMetrics,
    capturedInputs: CapturedInputs | null,
  ) => void
  onError?: (error: string) => void
}

export const useStreaming = (options?: UseStreamingOptions) => {
  const [isStreaming, setIsStreaming] = useState(false)
  const [content, setContent] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [metrics, setMetrics] = useState<StreamMetrics>({})
  const [capturedInputs, setCapturedInputs] = useState<CapturedInputs | null>(
    null,
  )
  const abortControllerRef = useRef<AbortController | null>(null)

  const stream = useCallback(
    async (request: ExecuteRequest, fullConfig?: ModelConfig | null) => {
      // CRITICAL: Capture inputs BEFORE any async operations so history
      // entries reflect inputs at execution time, not stream-end time.
      const inputSnapshot: CapturedInputs = {
        messages:
          'messages' in request.template
            ? request.template.messages.map((m) => ({ ...m }))
            : [],
        variables: { ...request.variables },
        config: fullConfig ? { ...fullConfig } : null,
      }
      setCapturedInputs(inputSnapshot)

      setIsStreaming(true)
      setContent('')
      setError(null)
      setMetrics({})

      abortControllerRef.current = new AbortController()

      try {
        // rawFetch handles CSRF, auth, refresh-on-401 consistently with the
        // rest of the dashboard plane. Streaming still works because fetch's
        // Response.body is an async iterable regardless of middleware.
        const response = await rawFetch('/api/v1/playground/stream', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(request),
          signal: abortControllerRef.current.signal,
        })

        const reader = response.body?.getReader()
        if (!reader) {
          throw new Error('No response body')
        }

        const decoder = new TextDecoder()
        let accumulatedContent = ''

        options?.onStart?.()

        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          const text = decoder.decode(value, { stream: true })
          const lines = text
            .split('\n')
            .filter((line) => line.startsWith('data:'))

          for (const line of lines) {
            const data = line.startsWith('data: ') ? line.slice(6) : line.slice(5)
            if (!data || data === '[DONE]') continue

            try {
              const chunk: StreamChunk = JSON.parse(data)

              switch (chunk.type) {
                case 'start':
                  break

                case 'content':
                  if (chunk.content) {
                    accumulatedContent += chunk.content
                    setContent(accumulatedContent)
                    options?.onContent?.(accumulatedContent)
                  }
                  break

                case 'end':
                  setIsStreaming(false)
                  break

                case 'metrics':
                  if (chunk.metrics) {
                    setMetrics(chunk.metrics)
                    options?.onEnd?.(
                      accumulatedContent,
                      chunk.metrics,
                      inputSnapshot,
                    )
                  }
                  break

                case 'error':
                  if (chunk.error) {
                    setError(chunk.error)
                    options?.onError?.(chunk.error)
                  }
                  setIsStreaming(false)
                  break
              }
            } catch (e) {
              console.error('Failed to parse SSE chunk:', e)
            }
          }
        }
      } catch (err) {
        if ((err as Error).name !== 'AbortError') {
          const message = (err as Error).message
          setError(message)
          options?.onError?.(message)
        }
      } finally {
        setIsStreaming(false)
        abortControllerRef.current = null
      }
    },
    [options],
  )

  const abort = useCallback(() => {
    abortControllerRef.current?.abort()
    setIsStreaming(false)
  }, [])

  const reset = useCallback(() => {
    setContent('')
    setError(null)
    setMetrics({})
    setCapturedInputs(null)
  }, [])

  return {
    stream,
    abort,
    reset,
    isStreaming,
    content,
    error,
    metrics,
    capturedInputs,
  }
}
