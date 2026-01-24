'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { AlertCircle, Plus, Calculator, Bot } from 'lucide-react'
import { useExperimentWizard } from '../../../context/experiment-wizard-context'
import { EvaluatorList } from './evaluator-list'
import { AddEvaluatorDialog } from './add-evaluator-dialog'
import type { EvaluatorCategory } from '../../../types'

export function EvaluatorsStep() {
  const { state, validationState, shouldShowStepErrors } = useExperimentWizard()
  const { evaluatorState } = state
  const validation = validationState.step3
  const showErrors = shouldShowStepErrors(3)

  const [addDialogOpen, setAddDialogOpen] = useState(false)
  const [selectedCategory, setSelectedCategory] = useState<EvaluatorCategory>('heuristics')

  const heuristicEvaluators = evaluatorState.evaluators.filter(
    (e) => e.scorer_type === 'builtin' || e.scorer_type === 'regex'
  )
  const llmEvaluators = evaluatorState.evaluators.filter((e) => e.scorer_type === 'llm')

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex-row items-start justify-between space-y-0">
          <div className="space-y-1.5">
            <CardTitle className="text-lg">Evaluators</CardTitle>
            <CardDescription>
              Add evaluators to automatically score your experiment outputs.
            </CardDescription>
          </div>
          <Button onClick={() => setAddDialogOpen(true)} size="sm">
            <Plus className="mr-2 h-4 w-4" />
            Add Evaluator
          </Button>
        </CardHeader>
        <CardContent>
          <Tabs
            value={selectedCategory}
            onValueChange={(v) => setSelectedCategory(v as EvaluatorCategory)}
          >
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="heuristics" className="flex items-center gap-2">
                <Calculator className="h-4 w-4" />
                Heuristics
                {heuristicEvaluators.length > 0 && (
                  <span className="ml-1 rounded-full bg-primary/20 px-2 py-0.5 text-xs">
                    {heuristicEvaluators.length}
                  </span>
                )}
              </TabsTrigger>
              <TabsTrigger value="llm_judges" className="flex items-center gap-2">
                <Bot className="h-4 w-4" />
                LLM Judges
                {llmEvaluators.length > 0 && (
                  <span className="ml-1 rounded-full bg-primary/20 px-2 py-0.5 text-xs">
                    {llmEvaluators.length}
                  </span>
                )}
              </TabsTrigger>
            </TabsList>
            <TabsContent value="heuristics" className="mt-4">
              <EvaluatorList
                evaluators={heuristicEvaluators}
                emptyMessage="No heuristic evaluators added yet. Add built-in scorers or regex patterns."
              />
            </TabsContent>
            <TabsContent value="llm_judges" className="mt-4">
              <EvaluatorList
                evaluators={llmEvaluators}
                emptyMessage="No LLM evaluators added yet. Add AI-powered judges to score outputs."
              />
            </TabsContent>
          </Tabs>

          {showErrors && validation.errors.find((e) => e.field === 'evaluators') && (
            <p className="text-sm text-destructive mt-4">
              {validation.errors.find((e) => e.field === 'evaluators')?.message}
            </p>
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

      <AddEvaluatorDialog
        open={addDialogOpen}
        onOpenChange={setAddDialogOpen}
        defaultCategory={selectedCategory}
      />
    </div>
  )
}
