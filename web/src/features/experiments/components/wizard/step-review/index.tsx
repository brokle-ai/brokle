'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import {
  AlertCircle,
  CheckCircle2,
  FileText,
  Database,
  Calculator,
  DollarSign,
  Loader2,
  Play,
  Save,
} from 'lucide-react'
import { useExperimentWizard } from '../../../context/experiment-wizard-context'
import { toast } from 'sonner'

interface ReviewStepProps {
  onSuccess: () => void
}

export function ReviewStep({ onSuccess }: ReviewStepProps) {
  const {
    state,
    submit,
    isSubmitting,
    estimateCost,
    costEstimate,
    isEstimating,
    validationState,
  } = useExperimentWizard()
  const { configState, datasetState, evaluatorState } = state

  const [runImmediately, setRunImmediately] = useState(true)

  const allValid =
    validationState.step1.isValid &&
    validationState.step2.isValid &&
    validationState.step3.isValid

  const handleSubmit = async () => {
    try {
      await submit(runImmediately)
      toast.success('Experiment Created', {
        description: runImmediately
          ? 'Your experiment has been created and is now running.'
          : 'Your experiment has been created as a draft.',
      })
      onSuccess()
    } catch (error) {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Experiment', {
        description: apiError?.message || 'Could not create experiment. Please try again.',
      })
    }
  }

  const handleEstimateCost = async () => {
    await estimateCost()
  }

  return (
    <div className="space-y-6">
      {/* Configuration Summary */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            <CardTitle className="text-lg">Configuration</CardTitle>
            {validationState.step1.isValid ? (
              <CheckCircle2 className="h-4 w-4 text-green-500 ml-auto" />
            ) : (
              <AlertCircle className="h-4 w-4 text-yellow-500 ml-auto" />
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-muted-foreground">Name:</span>
              <p className="font-medium">{configState.name || '—'}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Description:</span>
              <p className="font-medium">{configState.description || '—'}</p>
            </div>
          </div>
          <Separator />
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-muted-foreground">Prompt:</span>
              <p className="font-medium">{configState.selectedPrompt?.name || '—'}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Version:</span>
              <p className="font-medium">
                {configState.selectedVersion ? `v${configState.selectedVersion.version}` : '—'}
              </p>
            </div>
          </div>
          {configState.promptVariables.length > 0 && (
            <>
              <Separator />
              <div>
                <span className="text-muted-foreground text-sm">Variables:</span>
                <div className="flex flex-wrap gap-1 mt-1">
                  {configState.promptVariables.map((v) => (
                    <Badge key={v} variant="secondary" className="text-xs">
                      {`{{${v}}}`}
                    </Badge>
                  ))}
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Dataset Summary */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            <CardTitle className="text-lg">Dataset</CardTitle>
            {validationState.step2.isValid ? (
              <CheckCircle2 className="h-4 w-4 text-green-500 ml-auto" />
            ) : (
              <AlertCircle className="h-4 w-4 text-yellow-500 ml-auto" />
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-muted-foreground">Dataset:</span>
              <p className="font-medium">{datasetState.selectedDataset?.name || '—'}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Items:</span>
              <p className="font-medium">
                {datasetState.selectedDataset?.item_count ?? '—'}
              </p>
            </div>
          </div>
          {datasetState.variableMapping.length > 0 && (
            <>
              <Separator />
              <div>
                <span className="text-muted-foreground text-sm">Variable Mapping:</span>
                <div className="mt-2 space-y-1">
                  {datasetState.variableMapping.map((m) => (
                    <div key={m.variable_name} className="flex items-center gap-2 text-sm">
                      <Badge variant="secondary">{`{{${m.variable_name}}}`}</Badge>
                      <span className="text-muted-foreground">→</span>
                      <span className="font-mono text-xs">
                        {m.source}.{m.field_path}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Evaluators Summary */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <Calculator className="h-5 w-5" />
            <CardTitle className="text-lg">Evaluators</CardTitle>
            {validationState.step3.isValid ? (
              <CheckCircle2 className="h-4 w-4 text-green-500 ml-auto" />
            ) : (
              <AlertCircle className="h-4 w-4 text-yellow-500 ml-auto" />
            )}
          </div>
        </CardHeader>
        <CardContent>
          {evaluatorState.evaluators.length > 0 ? (
            <div className="space-y-2">
              {evaluatorState.evaluators.map((e) => (
                <div
                  key={e.id}
                  className="flex items-center justify-between text-sm p-2 rounded-md bg-muted/50"
                >
                  <span className="font-medium">{e.name}</span>
                  <Badge variant="outline">{e.scorer_type}</Badge>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No evaluators configured</p>
          )}
        </CardContent>
      </Card>

      {/* Cost Estimate */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <DollarSign className="h-5 w-5" />
              <CardTitle className="text-lg">Cost Estimate</CardTitle>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={handleEstimateCost}
              disabled={isEstimating || !allValid}
            >
              {isEstimating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Estimating...
                </>
              ) : (
                'Estimate'
              )}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {costEstimate ? (
            <div className="space-y-3">
              <div className="grid grid-cols-3 gap-4 text-sm">
                <div>
                  <span className="text-muted-foreground">Items:</span>
                  <p className="font-medium">{costEstimate.item_count}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">Est. Tokens:</span>
                  <p className="font-medium">{costEstimate.estimated_tokens.toLocaleString()}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">Est. Cost:</span>
                  <p className="font-medium">
                    {costEstimate.currency}
                    {costEstimate.estimated_cost.toFixed(4)}
                  </p>
                </div>
              </div>
              {costEstimate.cost_breakdown.length > 0 && (
                <>
                  <Separator />
                  <div className="text-xs text-muted-foreground">
                    <p className="font-medium mb-1">Breakdown:</p>
                    {costEstimate.cost_breakdown.map((item, i) => (
                      <div key={i} className="flex justify-between">
                        <span>{item.description}</span>
                        <span>
                          {costEstimate.currency}
                          {item.estimated_cost.toFixed(4)}
                        </span>
                      </div>
                    ))}
                  </div>
                </>
              )}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              Click &quot;Estimate&quot; to calculate the estimated cost of running this experiment.
            </p>
          )}
        </CardContent>
      </Card>

      {/* Run Options */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="run-immediately">Run immediately</Label>
              <p className="text-sm text-muted-foreground">
                Start processing dataset items right after creation
              </p>
            </div>
            <Switch
              id="run-immediately"
              checked={runImmediately}
              onCheckedChange={setRunImmediately}
            />
          </div>
        </CardContent>
      </Card>

      {/* Validation Errors */}
      {!allValid && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Please fix all errors before creating the experiment. Go back to the relevant steps
            to complete the required fields.
          </AlertDescription>
        </Alert>
      )}

      {/* Submit Button */}
      <div className="flex justify-end">
        <Button
          size="lg"
          onClick={handleSubmit}
          disabled={!allValid || isSubmitting}
        >
          {isSubmitting ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Creating...
            </>
          ) : runImmediately ? (
            <>
              <Play className="mr-2 h-4 w-4" />
              Create & Run Experiment
            </>
          ) : (
            <>
              <Save className="mr-2 h-4 w-4" />
              Create Experiment
            </>
          )}
        </Button>
      </div>
    </div>
  )
}
