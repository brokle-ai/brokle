import { useState, useCallback, useRef } from 'react'
import { config } from '@/lib/config'
import { getCookie } from '@/lib/utils/cookies'
import type { StreamChunk, StreamMetrics, ExecuteRequest, ChatMessage, ModelConfig } from '../types'

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
  onEnd?: (content: string, metrics: StreamMetrics, capturedInputs: CapturedInputs | null) => void
  onError?: (error: string) => void
}

export const useStreaming = (options?: UseStreamingOptions) => {
  const [isStreaming, setIsStreaming] = useState(false)
  const [content, setContent] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [metrics, setMetrics] = useState<StreamMetrics>({})
  const [capturedInputs, setCapturedInputs] = useState<CapturedInputs | null>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

  const stream = useCallback(
    async (request: ExecuteRequest, fullConfig?: ModelConfig | null) => {
      // CRITICAL: Capture inputs BEFORE any async operations
      // This ensures history entries reflect inputs at execution time, not when streaming ends
      //
      // NOTE: We capture fullConfig (from windowState.config) for history, NOT request.config_overrides.
      // - config_overrides is filtered by getEnabledModelConfig() for the API (only enabled params)
      // - fullConfig contains the complete UI state including _enabled flags and disabled param values
      // This allows history restore to fully recreate the config panel state.
      const inputSnapshot: CapturedInputs = {
        messages: 'messages' in request.template
          ? request.template.messages.map(m => ({ ...m }))
          : [],
        variables: { ...request.variables },
        config: fullConfig !== undefined
          ? (fullConfig ? { ...fullConfig } : null)
          : (request.config_overrides ? { ...request.config_overrides } : null),
      }
      setCapturedInputs(inputSnapshot)

      setIsStreaming(true)
      setContent('')
      setError(null)
      setMetrics({})

      abortControllerRef.current = new AbortController()

      try {
        const headers: HeadersInit = {
          'Content-Type': 'application/json',
        }

        const csrfToken = getCookie('csrf_token')
        if (csrfToken) {
          headers['X-CSRF-Token'] = csrfToken
        }

        const response = await fetch(`${config.api.baseUrl}/api/v1/playground/stream`, {
          method: 'POST',
          headers,
          body: JSON.stringify(request),
          credentials: 'include',
          signal: abortControllerRef.current.signal,
        })

        if (!response.ok) {
          const errorData = await response.json().catch(() => null)
          const message = errorData?.error?.message || response.statusText
          throw new Error(`HTTP ${response.status}: ${message}`)
        }

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
          const lines = text.split('\n').filter((line) => line.startsWith('data:'))

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
                    options?.onEnd?.(accumulatedContent, chunk.metrics, inputSnapshot)
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
    [options]
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
