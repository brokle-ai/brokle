'use client'

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertCircle, FileText, Info } from 'lucide-react'
import { useExperimentWizard } from '../../../context/experiment-wizard-context'
import { PromptSelector } from './prompt-selector'
import { VersionSelector } from './version-selector'

export function ConfigStep() {
  const { state, updateConfigState, validationState, shouldShowStepErrors } = useExperimentWizard()
  const { configState } = state
  const validation = validationState.step1
  const showErrors = shouldShowStepErrors(1)

  return (
    <div className="space-y-6">
      {/* Basic Info */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Basic Information</CardTitle>
          <CardDescription>
            Give your experiment a name and optional description.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="experiment-name">Experiment Name *</Label>
            <Input
              id="experiment-name"
              placeholder="e.g., RAG Evaluation v1"
              value={configState.name}
              onChange={(e) => updateConfigState({ name: e.target.value })}
            />
            {showErrors && validation.errors.find((e) => e.field === 'name') && (
              <p className="text-sm text-destructive">
                {validation.errors.find((e) => e.field === 'name')?.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="experiment-description">Description</Label>
            <Textarea
              id="experiment-description"
              placeholder="Optional: Describe what this experiment tests..."
              value={configState.description}
              onChange={(e) => updateConfigState({ description: e.target.value })}
              rows={3}
            />
          </div>
        </CardContent>
      </Card>

      {/* Prompt Selection */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Prompt Selection</CardTitle>
          <CardDescription>
            Choose the prompt template and version to use for this experiment.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Prompt *</Label>
              <PromptSelector />
              {showErrors && validation.errors.find((e) => e.field === 'promptId') && (
                <p className="text-sm text-destructive">
                  {validation.errors.find((e) => e.field === 'promptId')?.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label>Version *</Label>
              <VersionSelector />
              {showErrors && validation.errors.find((e) => e.field === 'promptVersionId') && (
                <p className="text-sm text-destructive">
                  {validation.errors.find((e) => e.field === 'promptVersionId')?.message}
                </p>
              )}
            </div>
          </div>

          {/* Show prompt variables */}
          {configState.promptVariables.length > 0 && (
            <div className="rounded-md border p-4 bg-muted/30">
              <div className="flex items-center gap-2 text-sm font-medium mb-2">
                <FileText className="h-4 w-4" />
                Template Variables
              </div>
              <div className="flex flex-wrap gap-2">
                {configState.promptVariables.map((variable) => (
                  <Badge key={variable} variant="secondary">
                    {`{{${variable}}}`}
                  </Badge>
                ))}
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                These variables will need to be mapped to dataset fields in the next step.
              </p>
            </div>
          )}

          {configState.selectedPrompt && configState.promptVariables.length === 0 && (
            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription>
                This prompt has no template variables. It will use the same input for all dataset items.
              </AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>

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
