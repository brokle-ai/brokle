'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import { useMutation } from '@tanstack/react-query'
import {
  previewTemplate,
  detectDialect,
  type PreviewTemplateRequest,
  type PreviewTemplateResponse,
  type DetectDialectRequest,
} from '../api/prompts-api'
import type { PromptType, TemplateDialect, TextTemplate, ChatTemplate } from '../types'
import { useTemplateValidation } from './use-template-validation'

interface UseTemplatePreviewOptions {
  projectId: string
  type: PromptType
  dialect?: TemplateDialect
  debounceMs?: number
  enabled?: boolean
}

interface UseTemplatePreviewReturn {
  isCompiling: boolean
  compiled: TextTemplate | ChatTemplate | null
  detectedDialect: TemplateDialect | null
  error: string | null
  preview: (template: TextTemplate | ChatTemplate, variables: Record<string, unknown>) => void
  reset: () => void
}

/**
 * Hook for template preview/compilation with debouncing.
 *
 * Features:
 * - Debounced preview (default 300ms)
 * - Compiles template with provided variables
 * - Returns compiled result
 */
export function useTemplatePreview({
  projectId,
  type,
  dialect,
  debounceMs = 300,
  enabled = true,
}: UseTemplatePreviewOptions): UseTemplatePreviewReturn {
  const [compiled, setCompiled] = useState<TextTemplate | ChatTemplate | null>(null)
  const [detectedDialect, setDetectedDialect] = useState<TemplateDialect | null>(null)
  const [error, setError] = useState<string | null>(null)

  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null)

  // Preview mutation
  const previewMutation = useMutation({
    mutationFn: async (data: PreviewTemplateRequest) => {
      return previewTemplate(projectId, data)
    },
    onSuccess: (response: PreviewTemplateResponse) => {
      setCompiled(response.compiled)
      setDetectedDialect(response.dialect)
      setError(null)
    },
    onError: (err: unknown) => {
      const apiError = err as { message?: string }
      setError(apiError?.message || 'Failed to compile template')
      setCompiled(null)
    },
  })

  // Debounced preview function
  const preview = useCallback(
    (template: TextTemplate | ChatTemplate, variables: Record<string, unknown>) => {
      if (!enabled || !projectId) return

      // Clear existing timer
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }

      // Set new debounce timer
      debounceTimerRef.current = setTimeout(() => {
        previewMutation.mutate({
          template,
          type,
          variables,
          dialect: dialect === 'auto' ? undefined : dialect,
        })
      }, debounceMs)
    },
    [projectId, type, dialect, debounceMs, enabled, previewMutation]
  )

  // Reset state
  const reset = useCallback(() => {
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current)
    }
    setCompiled(null)
    setDetectedDialect(null)
    setError(null)
  }, [])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }
    }
  }, [])

  return {
    isCompiling: previewMutation.isPending,
    compiled,
    detectedDialect,
    error,
    preview,
    reset,
  }
}

/**
 * Simpler hook for immediate preview (no debouncing).
 * Use for explicit "preview" button clicks.
 */
export function usePreviewTemplateMutation(projectId: string) {
  return useMutation({
    mutationFn: async (data: PreviewTemplateRequest) => {
      return previewTemplate(projectId, data)
    },
  })
}

/**
 * Hook for dialect detection.
 */
export function useDetectDialectMutation(projectId: string) {
  return useMutation({
    mutationFn: async (data: DetectDialectRequest) => {
      return detectDialect(projectId, data)
    },
  })
}

/**
 * Combined hook for template editing with live preview.
 *
 * Features:
 * - Validates template on change
 * - Previews with sample variables
 * - Auto-detects dialect
 */
interface UseTemplateEditorOptions {
  projectId: string
  type: PromptType
  initialDialect?: TemplateDialect
  validationDebounceMs?: number
  previewDebounceMs?: number
}

interface UseTemplateEditorReturn {
  // Dialect
  dialect: TemplateDialect
  setDialect: (dialect: TemplateDialect) => void
  detectedDialect: TemplateDialect | null

  // Validation
  isValidating: boolean
  isValid: boolean | null
  errors: Array<{ line: number; column: number; message: string; code: string }>
  warnings: Array<{ line: number; column: number; message: string; code: string }>
  variables: string[]

  // Preview
  isCompiling: boolean
  compiled: TextTemplate | ChatTemplate | null
  previewError: string | null

  // Actions
  updateTemplate: (template: TextTemplate | ChatTemplate) => void
  updatePreviewVariables: (variables: Record<string, unknown>) => void
  reset: () => void
}

export function useTemplateEditor({
  projectId,
  type,
  initialDialect = 'auto',
  validationDebounceMs = 500,
  previewDebounceMs = 300,
}: UseTemplateEditorOptions): UseTemplateEditorReturn {
  const [dialect, setDialect] = useState<TemplateDialect>(initialDialect)
  const [currentTemplate, setCurrentTemplate] = useState<TextTemplate | ChatTemplate | null>(null)
  const [previewVariables, setPreviewVariables] = useState<Record<string, unknown>>({})

  // Validation hook - uses the actual validation API
  const validation = useTemplateValidation({
    projectId,
    type,
    dialect,
    debounceMs: validationDebounceMs,
    enabled: true,
  })

  // Preview hook
  const preview = useTemplatePreview({
    projectId,
    type,
    dialect,
    debounceMs: previewDebounceMs,
    enabled: true,
  })

  // Update template (triggers both validation and preview)
  const updateTemplate = useCallback(
    (template: TextTemplate | ChatTemplate) => {
      setCurrentTemplate(template)
      // Trigger validation
      validation.validate(template)
      // Trigger preview if we have variables
      if (Object.keys(previewVariables).length > 0) {
        preview.preview(template, previewVariables)
      }
    },
    [previewVariables, preview, validation]
  )

  // Update preview variables
  const updatePreviewVariables = useCallback(
    (variables: Record<string, unknown>) => {
      setPreviewVariables(variables)
      if (currentTemplate) {
        preview.preview(currentTemplate, variables)
      }
    },
    [currentTemplate, preview]
  )

  // Reset all state
  const reset = useCallback(() => {
    setCurrentTemplate(null)
    setPreviewVariables({})
    preview.reset()
    validation.reset()
  }, [preview, validation])

  return {
    dialect,
    setDialect,
    detectedDialect: validation.detectedDialect ?? preview.detectedDialect,

    isValidating: validation.isValidating,
    isValid: validation.isValid,
    errors: validation.errors,
    warnings: validation.warnings,
    variables: validation.variables,

    isCompiling: preview.isCompiling,
    compiled: preview.compiled,
    previewError: preview.error,

    updateTemplate,
    updatePreviewVariables,
    reset,
  }
}
