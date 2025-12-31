'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import { useMutation } from '@tanstack/react-query'
import {
  validateTemplate,
  type ValidateTemplateRequest,
  type ValidateTemplateResponse,
  type SyntaxError,
  type SyntaxWarning,
} from '../api/prompts-api'
import type { PromptType, TemplateDialect, TextTemplate, ChatTemplate } from '../types'

interface UseTemplateValidationOptions {
  projectId: string
  type: PromptType
  dialect?: TemplateDialect
  debounceMs?: number
  enabled?: boolean
}

interface UseTemplateValidationReturn {
  isValidating: boolean
  isValid: boolean | null
  errors: SyntaxError[]
  warnings: SyntaxWarning[]
  variables: string[]
  detectedDialect: TemplateDialect | null
  validate: (template: TextTemplate | ChatTemplate) => void
  reset: () => void
}

/**
 * Hook for template validation with debouncing.
 *
 * Features:
 * - Debounced validation (default 500ms)
 * - Extracts variables automatically
 * - Detects dialect when not specified
 * - Returns syntax errors and warnings
 */
export function useTemplateValidation({
  projectId,
  type,
  dialect,
  debounceMs = 500,
  enabled = true,
}: UseTemplateValidationOptions): UseTemplateValidationReturn {
  const [isValid, setIsValid] = useState<boolean | null>(null)
  const [errors, setErrors] = useState<SyntaxError[]>([])
  const [warnings, setWarnings] = useState<SyntaxWarning[]>([])
  const [variables, setVariables] = useState<string[]>([])
  const [detectedDialect, setDetectedDialect] = useState<TemplateDialect | null>(null)

  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null)
  const lastTemplateRef = useRef<TextTemplate | ChatTemplate | null>(null)

  // Validation mutation
  const validationMutation = useMutation({
    mutationFn: async (data: ValidateTemplateRequest) => {
      return validateTemplate(projectId, data)
    },
    onSuccess: (response: ValidateTemplateResponse) => {
      setIsValid(response.valid)
      setErrors(response.errors)
      setWarnings(response.warnings)
      setVariables(response.variables)
      setDetectedDialect(response.dialect)
    },
    onError: () => {
      // On error, keep previous state but mark as invalid
      setIsValid(false)
    },
  })

  // Debounced validation function
  const validate = useCallback(
    (template: TextTemplate | ChatTemplate) => {
      if (!enabled || !projectId) return

      lastTemplateRef.current = template

      // Clear existing timer
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }

      // Set new debounce timer
      debounceTimerRef.current = setTimeout(() => {
        validationMutation.mutate({
          template,
          type,
          dialect: dialect === 'auto' ? undefined : dialect,
        })
      }, debounceMs)
    },
    [projectId, type, dialect, debounceMs, enabled, validationMutation]
  )

  // Reset state
  const reset = useCallback(() => {
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current)
    }
    setIsValid(null)
    setErrors([])
    setWarnings([])
    setVariables([])
    setDetectedDialect(null)
    lastTemplateRef.current = null
  }, [])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }
    }
  }, [])

  // Re-validate when dialect changes
  useEffect(() => {
    if (lastTemplateRef.current && enabled) {
      validate(lastTemplateRef.current)
    }
  }, [dialect, enabled]) // eslint-disable-line react-hooks/exhaustive-deps

  return {
    isValidating: validationMutation.isPending,
    isValid,
    errors,
    warnings,
    variables,
    detectedDialect,
    validate,
    reset,
  }
}

/**
 * Simpler hook for immediate validation (no debouncing).
 * Use for explicit "validate" button clicks.
 */
export function useValidateTemplateMutation(projectId: string) {
  return useMutation({
    mutationFn: async (data: ValidateTemplateRequest) => {
      return validateTemplate(projectId, data)
    },
  })
}
