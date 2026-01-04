'use client'

import { useMemo, useCallback, useEffect } from 'react'
import { useQueryStates, parseAsString } from 'nuqs'
import { useQuery } from '@tanstack/react-query'
import { BrokleAPIClient } from '@/lib/api/core/client'
import type { Variable, VariableValues, WidgetViewType } from '../types'

const client = new BrokleAPIClient('/api')

/**
 * Fetch dynamic options for query-type variables
 */
async function fetchVariableOptions(
  projectId: string,
  view: WidgetViewType,
  dimension: string,
  limit: number = 100
): Promise<string[]> {
  const response = await client.get<{ values: string[] }>(
    `/v1/projects/${projectId}/dashboards/variable-options`,
    {
      view,
      dimension,
      limit: String(limit),
    }
  )
  return response.values || []
}

export interface UseDashboardVariablesReturn {
  /** Current variable values (from URL or defaults) */
  values: VariableValues
  /** Set a single variable value */
  setValue: (name: string, value: unknown) => void
  /** Set multiple variable values at once */
  setValues: (values: VariableValues) => void
  /** Reset all variables to their defaults */
  resetToDefaults: () => void
  /** Whether any variable has a non-default value */
  hasActiveVariables: boolean
  /** Get options for a specific variable (static or query-based) */
  getVariableOptions: (variable: Variable) => string[]
  /** Loading state for query-based variable options */
  isLoadingOptions: boolean
}

/**
 * Hook to manage dashboard variable values with URL persistence.
 *
 * Variables are stored in URL as `var_[name]=[value]` for easy sharing and bookmarking.
 * Supports static options (select type) and dynamic options (query type).
 *
 * @example
 * ```tsx
 * const { values, setValue } = useDashboardVariables(
 *   projectId,
 *   dashboard.config.variables
 * )
 *
 * // Use values in query execution
 * executeDashboardQueries(projectId, dashboardId, { variable_values: values })
 * ```
 */
export function useDashboardVariables(
  projectId: string | undefined,
  variables: Variable[] | undefined
): UseDashboardVariablesReturn {
  // Build parser config for all variables
  const parserConfig = useMemo(() => {
    if (!variables || variables.length === 0) return {}

    return variables.reduce((acc, variable) => {
      // Use var_ prefix to avoid URL param conflicts
      acc[`var_${variable.name}`] = parseAsString
      return acc
    }, {} as Record<string, typeof parseAsString>)
  }, [variables])

  const [urlParams, setUrlParams] = useQueryStates(parserConfig)

  // Build default values map
  const defaults = useMemo((): VariableValues => {
    if (!variables) return {}

    return variables.reduce((acc, variable) => {
      if (variable.default !== undefined) {
        acc[variable.name] = variable.default
      }
      return acc
    }, {} as VariableValues)
  }, [variables])

  // Merge URL params with defaults
  const values = useMemo((): VariableValues => {
    if (!variables) return {}

    return variables.reduce((acc, variable) => {
      const urlKey = `var_${variable.name}`
      const urlValue = urlParams[urlKey]

      if (urlValue !== null && urlValue !== undefined) {
        // Parse URL value based on variable type
        if (variable.type === 'number') {
          acc[variable.name] = parseFloat(urlValue) || variable.default || 0
        } else if (variable.multi && typeof urlValue === 'string') {
          // Multi-select values are comma-separated
          acc[variable.name] = urlValue.split(',').filter(Boolean)
        } else {
          acc[variable.name] = urlValue
        }
      } else if (variable.default !== undefined) {
        acc[variable.name] = variable.default
      }
      return acc
    }, {} as VariableValues)
  }, [variables, urlParams])

  // Fetch options for query-type variables
  const queryVariables = useMemo(() => {
    return variables?.filter(v => v.type === 'query' && v.query_config) || []
  }, [variables])

  const { data: queryOptions, isLoading: isLoadingOptions } = useQuery({
    queryKey: ['variable-options', projectId, queryVariables.map(v => v.name)],
    queryFn: async () => {
      if (!projectId || queryVariables.length === 0) return {}

      const results: Record<string, string[]> = {}

      // Fetch options for each query variable in parallel
      await Promise.all(
        queryVariables.map(async (variable) => {
          if (!variable.query_config) return

          try {
            const options = await fetchVariableOptions(
              projectId,
              variable.query_config.view,
              variable.query_config.dimension,
              variable.query_config.limit
            )
            results[variable.name] = options
          } catch {
            results[variable.name] = []
          }
        })
      )

      return results
    },
    enabled: !!projectId && queryVariables.length > 0,
    staleTime: 60_000, // 1 minute - variable options don't change frequently
    gcTime: 5 * 60_000, // 5 minutes
  })

  // Get options for a variable
  const getVariableOptions = useCallback((variable: Variable): string[] => {
    if (variable.type === 'select' && variable.options) {
      return variable.options
    }
    if (variable.type === 'query' && queryOptions) {
      return queryOptions[variable.name] || []
    }
    return []
  }, [queryOptions])

  // Set a single variable value
  const setValue = useCallback((name: string, value: unknown) => {
    const urlKey = `var_${name}`

    if (value === null || value === undefined || value === '') {
      setUrlParams({ [urlKey]: null })
    } else if (Array.isArray(value)) {
      // Multi-select: join with commas
      setUrlParams({ [urlKey]: value.join(',') || null })
    } else {
      setUrlParams({ [urlKey]: String(value) })
    }
  }, [setUrlParams])

  // Set multiple variable values
  const setValues = useCallback((newValues: VariableValues) => {
    const urlUpdates: Record<string, string | null> = {}

    for (const [name, value] of Object.entries(newValues)) {
      const urlKey = `var_${name}`

      if (value === null || value === undefined || value === '') {
        urlUpdates[urlKey] = null
      } else if (Array.isArray(value)) {
        urlUpdates[urlKey] = value.join(',') || null
      } else {
        urlUpdates[urlKey] = String(value)
      }
    }

    setUrlParams(urlUpdates)
  }, [setUrlParams])

  // Reset all variables to defaults
  const resetToDefaults = useCallback(() => {
    if (!variables) return

    const urlUpdates: Record<string, null> = {}
    for (const variable of variables) {
      urlUpdates[`var_${variable.name}`] = null
    }
    setUrlParams(urlUpdates)
  }, [variables, setUrlParams])

  // Check if any variable differs from its default
  const hasActiveVariables = useMemo(() => {
    if (!variables) return false

    return variables.some(variable => {
      const currentValue = values[variable.name]
      const defaultValue = variable.default

      if (currentValue === undefined && defaultValue === undefined) return false
      if (Array.isArray(currentValue) && Array.isArray(defaultValue)) {
        return JSON.stringify(currentValue) !== JSON.stringify(defaultValue)
      }
      return currentValue !== defaultValue
    })
  }, [variables, values])

  return {
    values,
    setValue,
    setValues,
    resetToDefaults,
    hasActiveVariables,
    getVariableOptions,
    isLoadingOptions,
  }
}

/**
 * Query key factory for variable-related queries
 */
export const variableQueryKeys = {
  all: ['variables'] as const,
  options: (projectId: string, variables: string[]) =>
    [...variableQueryKeys.all, 'options', projectId, variables] as const,
}
