'use client'

import { useCallback, useMemo } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { AlertCircle, Settings2, Eye } from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { ViewSelector } from './view-selector'
import { MeasureSelector } from './measure-selector'
import { DimensionSelector } from './dimension-selector'
import { FilterBuilder } from './filter-builder'
import { QueryPreview } from './query-preview'
import { useViewDefinitions } from '../../hooks/use-widget-queries'
import type { WidgetQuery, WidgetViewType, QueryFilter } from '../../types'

interface QueryBuilderProps {
  query?: WidgetQuery
  onQueryChange: (query: WidgetQuery) => void
  disabled?: boolean
}

const DEFAULT_QUERY: WidgetQuery = {
  view: 'traces',
  measures: [],
  dimensions: [],
  filters: [],
}

export function QueryBuilder({
  query = DEFAULT_QUERY,
  onQueryChange,
  disabled,
}: QueryBuilderProps) {
  const { data: viewDefinitionsResponse, isLoading, error } = useViewDefinitions()
  const viewDefinitions = viewDefinitionsResponse?.views

  const currentViewDef = useMemo(() => {
    if (!viewDefinitions || !query.view) return undefined
    return viewDefinitions[query.view]
  }, [viewDefinitions, query.view])

  const handleViewChange = useCallback(
    (view: WidgetViewType) => {
      // When view changes, reset measures and dimensions
      onQueryChange({
        ...query,
        view,
        measures: [],
        dimensions: [],
        filters: [],
      })
    },
    [query, onQueryChange]
  )

  const handleMeasuresChange = useCallback(
    (measures: string[]) => {
      onQueryChange({
        ...query,
        measures,
      })
    },
    [query, onQueryChange]
  )

  const handleDimensionsChange = useCallback(
    (dimensions: string[]) => {
      onQueryChange({
        ...query,
        dimensions,
      })
    },
    [query, onQueryChange]
  )

  const handleFiltersChange = useCallback(
    (filters: QueryFilter[]) => {
      onQueryChange({
        ...query,
        filters,
      })
    },
    [query, onQueryChange]
  )

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load view definitions. Please try again later.
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <div className="space-y-4">
      <Tabs defaultValue="builder">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="builder" className="flex items-center gap-2">
            <Settings2 className="h-4 w-4" />
            Query Builder
          </TabsTrigger>
          <TabsTrigger value="preview" className="flex items-center gap-2">
            <Eye className="h-4 w-4" />
            Preview
          </TabsTrigger>
        </TabsList>

        <TabsContent value="builder" className="space-y-4 mt-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-base">Data Source</CardTitle>
              <CardDescription>
                Choose where to fetch data from
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ViewSelector
                value={query.view}
                onValueChange={handleViewChange}
                viewDefinitions={viewDefinitions}
                disabled={disabled || isLoading}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-base">Metrics</CardTitle>
              <CardDescription>
                Select measures to aggregate and dimensions to group by
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <MeasureSelector
                measures={currentViewDef?.measures}
                value={query.measures ?? []}
                onValueChange={handleMeasuresChange}
                disabled={disabled || isLoading || !query.view}
              />

              <DimensionSelector
                dimensions={currentViewDef?.dimensions}
                value={query.dimensions ?? []}
                onValueChange={handleDimensionsChange}
                disabled={disabled || isLoading || !query.view}
                maxSelections={3}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-base">Filters</CardTitle>
              <CardDescription>
                Filter data based on specific conditions
              </CardDescription>
            </CardHeader>
            <CardContent>
              <FilterBuilder
                dimensions={currentViewDef?.dimensions}
                filters={query.filters ?? []}
                onFiltersChange={handleFiltersChange}
                disabled={disabled || isLoading || !query.view}
              />
            </CardContent>
          </Card>

          {/* Validation feedback */}
          {query.view && query.measures?.length === 0 && (
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Select at least one measure to display data in the widget.
              </AlertDescription>
            </Alert>
          )}
        </TabsContent>

        <TabsContent value="preview" className="mt-4">
          <QueryPreview query={query} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
