'use client'

import { useState, useCallback } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { datasetsApi } from '../api/datasets-api'
import { datasetQueryKeys } from './use-datasets'
import { chunkCSVContent } from '../utils/csv-parser'
import type { CSVColumnMapping, BulkImportResult } from '../types'

interface ChunkedImportOptions {
  content: string
  columnMapping: CSVColumnMapping
  hasHeader: boolean
  deduplicate: boolean
  maxPayloadSize?: number
  delayBetweenChunks?: number
}

interface ChunkedImportProgress {
  currentChunk: number
  totalChunks: number
  itemsCreated: number
  itemsSkipped: number
  errors: string[]
  failedChunks: number[]
}

interface UseChunkedCsvImportReturn {
  importCsv: (options: ChunkedImportOptions) => Promise<BulkImportResult>
  progress: ChunkedImportProgress | null
  isImporting: boolean
  error: Error | null
  reset: () => void
}

const DEFAULT_MAX_PAYLOAD_SIZE = 500 * 1024 // 500KB
const DEFAULT_DELAY_BETWEEN_CHUNKS = 100 // ms
const DEFAULT_MAX_RETRIES = 3
const INITIAL_RETRY_DELAY = 100 // ms

/**
 * Retry a function with exponential backoff
 */
async function retryWithBackoff<T>(
  fn: () => Promise<T>,
  maxRetries: number = DEFAULT_MAX_RETRIES,
  initialDelay: number = INITIAL_RETRY_DELAY
): Promise<T> {
  let lastError: Error = new Error('Unknown error')
  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      return await fn()
    } catch (err) {
      lastError = err instanceof Error ? err : new Error(String(err))
      if (attempt < maxRetries) {
        const delay = initialDelay * Math.pow(2, attempt)
        await new Promise(resolve => setTimeout(resolve, delay))
      }
    }
  }
  throw lastError
}

/**
 * Hook for chunked CSV import with progress tracking
 * Splits large files into manageable chunks and uploads sequentially
 */
export function useChunkedCsvImport(
  projectId: string,
  datasetId: string
): UseChunkedCsvImportReturn {
  const queryClient = useQueryClient()
  const [progress, setProgress] = useState<ChunkedImportProgress | null>(null)
  const [isImporting, setIsImporting] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const reset = useCallback(() => {
    setProgress(null)
    setIsImporting(false)
    setError(null)
  }, [])

  const importCsv = useCallback(async (options: ChunkedImportOptions): Promise<BulkImportResult> => {
    const {
      content,
      columnMapping,
      hasHeader,
      deduplicate,
      maxPayloadSize = DEFAULT_MAX_PAYLOAD_SIZE,
      delayBetweenChunks = DEFAULT_DELAY_BETWEEN_CHUNKS,
    } = options

    setIsImporting(true)
    setError(null)

    // Split content into chunks
    const chunks = chunkCSVContent(content, hasHeader, maxPayloadSize)
    const totalChunks = chunks.length

    setProgress({
      currentChunk: 0,
      totalChunks,
      itemsCreated: 0,
      itemsSkipped: 0,
      errors: [],
      failedChunks: [],
    })

    const result: BulkImportResult = {
      created: 0,
      skipped: 0,
      errors: [],
    }

    for (let i = 0; i < chunks.length; i++) {
      const chunk = chunks[i]

      // Update progress before upload
      setProgress(prev => prev ? {
        ...prev,
        currentChunk: i + 1,
      } : null)

      try {
        // Upload chunk with retry logic
        const chunkResult = await retryWithBackoff(
          () => datasetsApi.importFromCsv(projectId, datasetId, {
            content: chunk,
            column_mapping: {
              input_column: columnMapping.input_column,
              expected_column: columnMapping.expected_column || undefined,
              metadata_columns: columnMapping.metadata_columns?.length
                ? columnMapping.metadata_columns
                : undefined,
            },
            has_header: hasHeader,
            deduplicate,
          })
        )

        // Accumulate results
        result.created += chunkResult.created
        result.skipped += chunkResult.skipped
        if (chunkResult.errors) {
          result.errors = [...(result.errors || []), ...chunkResult.errors]
        }
      } catch (err) {
        // Track failed chunk but continue with remaining chunks
        const errorMsg = err instanceof Error ? err.message : 'Unknown error'
        result.errors = [...(result.errors || []), `Chunk ${i + 1}: ${errorMsg}`]

        setProgress(prev => prev ? {
          ...prev,
          failedChunks: [...prev.failedChunks, i],
          errors: [...prev.errors, `Chunk ${i + 1} failed after retries`],
        } : null)
        // Continue with next chunk instead of stopping
      }

      // Update progress after upload attempt
      setProgress(prev => prev ? {
        ...prev,
        itemsCreated: result.created,
        itemsSkipped: result.skipped,
        errors: result.errors || [],
      } : null)

      // Add delay between chunks to avoid overwhelming the server
      if (i < chunks.length - 1) {
        await new Promise(resolve => setTimeout(resolve, delayBetweenChunks))
      }
    }

    // Invalidate queries to refresh data
    queryClient.invalidateQueries({
      queryKey: datasetQueryKeys.items(projectId, datasetId),
    })
    queryClient.invalidateQueries({
      queryKey: datasetQueryKeys.detail(projectId, datasetId),
    })

    setIsImporting(false)
    return result
  }, [projectId, datasetId, queryClient])

  return {
    importCsv,
    progress,
    isImporting,
    error,
    reset,
  }
}

/**
 * Standard mutation-based CSV import (for small files)
 * Use this when file size is under maxPayloadSize
 */
export function useCsvImportMutation(projectId: string, datasetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (params: {
      content: string
      columnMapping: CSVColumnMapping
      hasHeader: boolean
      deduplicate: boolean
    }) => {
      return datasetsApi.importFromCsv(projectId, datasetId, {
        content: params.content,
        column_mapping: {
          input_column: params.columnMapping.input_column,
          expected_column: params.columnMapping.expected_column || undefined,
          metadata_columns: params.columnMapping.metadata_columns?.length
            ? params.columnMapping.metadata_columns
            : undefined,
        },
        has_header: params.hasHeader,
        deduplicate: params.deduplicate,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.items(projectId, datasetId),
      })
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.detail(projectId, datasetId),
      })
    },
  })
}
