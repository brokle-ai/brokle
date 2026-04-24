import { useMemo, useCallback } from 'react'
import { useQueryStates, parseAsString } from 'nuqs'
import { useQuery } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type { Variable, VariableValues, WidgetViewType } from '../types'

interface VariableOptionsResponse {
  values: string[]
}

async function fetchVariableOptions(
  projectId: string,
  view: WidgetViewType,
  dimension: string,
  limit = 100,
): Promise<string[]> {
  const search = new URLSearchParams({
    view,
    dimension,
    limit: String(limit),
  })
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/dashboards/variable-options?${search.toString()}`,
    { method: 'GET' },
  )
  const body = (await resp.json()) as VariableOptionsResponse
  return body.values ?? []
}

export interface UseDashboardVariablesReturn {
  values: VariableValues
  setValue: (name: string, value: unknown) => void
  setValues: (values: VariableValues) => void
  resetToDefaults: () => void
  hasActiveVariables: boolean
  getVariableOptions: (variable: Variable) => string[]
  isLoadingOptions: boolean
}

export function useDashboardVariables(
  projectId: string | undefined,
  variables: Variable[] | undefined,
): UseDashboardVariablesReturn {
  const parserConfig = useMemo(() => {
    if (!variables || variables.length === 0)
      return {} as Record<string, typeof parseAsString>

    return variables.reduce(
      (acc, variable) => {
        acc[`var_${variable.name}`] = parseAsString
        return acc
      },
      {} as Record<string, typeof parseAsString>,
    )
  }, [variables])

  const [urlParams, setUrlParams] = useQueryStates(parserConfig)

  const values = useMemo((): VariableValues => {
    if (!variables) return {}

    return variables.reduce((acc, variable) => {
      const urlKey = `var_${variable.name}`
      const urlValue = urlParams[urlKey]

      if (urlValue !== null && urlValue !== undefined) {
        if (variable.type === 'number') {
          acc[variable.name] =
            parseFloat(urlValue) || (variable.default as number | undefined) || 0
        } else if (variable.multi && typeof urlValue === 'string') {
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

  const queryVariables = useMemo(() => {
    return variables?.filter((v) => v.type === 'query' && v.query_config) ?? []
  }, [variables])

  const { data: queryOptions, isLoading: isLoadingOptions } = useQuery({
    queryKey: [
      'variable-options',
      projectId,
      queryVariables.map((v) => v.name),
    ],
    queryFn: async () => {
      if (!projectId || queryVariables.length === 0) return {}

      const results: Record<string, string[]> = {}
      await Promise.all(
        queryVariables.map(async (variable) => {
          if (!variable.query_config) return
          try {
            const options = await fetchVariableOptions(
              projectId,
              variable.query_config.view,
              variable.query_config.dimension,
              variable.query_config.limit,
            )
            results[variable.name] = options
          } catch {
            results[variable.name] = []
          }
        }),
      )
      return results
    },
    enabled: !!projectId && queryVariables.length > 0,
    staleTime: 60_000,
    gcTime: 5 * 60_000,
  })

  const getVariableOptions = useCallback(
    (variable: Variable): string[] => {
      if (variable.type === 'select' && variable.options) {
        return variable.options
      }
      if (variable.type === 'query' && queryOptions) {
        return queryOptions[variable.name] ?? []
      }
      return []
    },
    [queryOptions],
  )

  const setValue = useCallback(
    (name: string, value: unknown) => {
      const urlKey = `var_${name}`

      if (value === null || value === undefined || value === '') {
        setUrlParams({ [urlKey]: null })
      } else if (Array.isArray(value)) {
        setUrlParams({ [urlKey]: value.join(',') || null })
      } else {
        setUrlParams({ [urlKey]: String(value) })
      }
    },
    [setUrlParams],
  )

  const setValues = useCallback(
    (newValues: VariableValues) => {
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
    },
    [setUrlParams],
  )

  const resetToDefaults = useCallback(() => {
    if (!variables) return
    const urlUpdates: Record<string, null> = {}
    for (const variable of variables) {
      urlUpdates[`var_${variable.name}`] = null
    }
    setUrlParams(urlUpdates)
  }, [variables, setUrlParams])

  const hasActiveVariables = useMemo(() => {
    if (!variables) return false
    return variables.some((variable) => {
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

export const variableQueryKeys = {
  all: ['variables'] as const,
  options: (projectId: string, variables: string[]) =>
    [...variableQueryKeys.all, 'options', projectId, variables] as const,
}
