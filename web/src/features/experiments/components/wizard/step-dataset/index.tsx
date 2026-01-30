'use client'

import { useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { AlertCircle, Info, Wand2 } from 'lucide-react'
import { useExperimentWizard } from '../../../context/experiment-wizard-context'
import { useDatasetFieldsQuery } from '../../../hooks/use-wizard-queries'
import { DatasetSelector } from './dataset-selector'
import { VariableMappingTable } from './variable-mapping-table'

export function DatasetStep() {
  const { projectId, state, setDatasetFields, autoMapVariables, validationState, shouldShowStepErrors } =
    useExperimentWizard()
  const { configState, datasetState } = state
  const validation = validationState.step2
  const showErrors = shouldShowStepErrors(2)

  // Fetch dataset fields when a dataset is selected
  const { data: datasetFields, isLoading: isLoadingFields } = useDatasetFieldsQuery(
    projectId,
    datasetState.datasetId ?? undefined
  )

  // Update context when fields are loaded
  useEffect(() => {
    if (datasetFields) {
      setDatasetFields(datasetFields)
    }
  }, [datasetFields, setDatasetFields])

  const hasVariables = configState.promptVariables.length > 0
  const hasDataset = !!datasetState.datasetId
  const hasFields = !!datasetState.datasetFields

  return (
    <div className="space-y-6">
      {/* Dataset Selection */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Select Dataset</CardTitle>
          <CardDescription>
            Choose a dataset containing the test data for your experiment.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <DatasetSelector />
          {showErrors && validation.errors.find((e) => e.field === 'datasetId') && (
            <p className="text-sm text-destructive">
              {validation.errors.find((e) => e.field === 'datasetId')?.message}
            </p>
          )}
        </CardContent>
      </Card>

      {/* Variable Mapping */}
      {hasDataset && hasVariables && (
        <Card>
          <CardHeader className="flex-row items-start justify-between space-y-0">
            <div className="space-y-1.5">
              <CardTitle className="text-lg">Variable Mapping</CardTitle>
              <CardDescription>
                Map prompt template variables to dataset fields.
              </CardDescription>
            </div>
            {hasFields && (
              <Button
                variant="outline"
                size="sm"
                onClick={autoMapVariables}
                className="shrink-0"
              >
                <Wand2 className="mr-2 h-4 w-4" />
                Auto-Map
              </Button>
            )}
          </CardHeader>
          <CardContent>
            {isLoadingFields ? (
              <div className="flex items-center justify-center py-8">
                <div className="text-sm text-muted-foreground">Loading dataset fields...</div>
              </div>
            ) : hasFields ? (
              <VariableMappingTable />
            ) : (
              <Alert>
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  Could not load dataset fields. Please try selecting a different dataset.
                </AlertDescription>
              </Alert>
            )}
            {showErrors && validation.errors.find((e) => e.field === 'variableMapping') && (
              <p className="text-sm text-destructive mt-2">
                {validation.errors.find((e) => e.field === 'variableMapping')?.message}
              </p>
            )}
          </CardContent>
        </Card>
      )}

      {/* No Variables Notice */}
      {hasDataset && !hasVariables && (
        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            The selected prompt has no template variables. The same prompt will be used for all
            dataset items without any variable substitution.
          </AlertDescription>
        </Alert>
      )}

      {/* Validation Summary */}
      {showErrors && !validation.isValid && validation.errors.length > 0 && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Please fix the following errors before continuing:
            <ul className="list-disc list-inside mt-1">
              {validation.errors.map((error, i) => (
                <li key={i}>{error.message}</li>
              ))}
            </ul>
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}
