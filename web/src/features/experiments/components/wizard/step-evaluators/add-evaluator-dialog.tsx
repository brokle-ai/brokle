'use client'

import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Calculator, Bot, Regex } from 'lucide-react'
import { useExperimentWizard } from '../../../context/experiment-wizard-context'
import { BUILTIN_SCORERS, type EvaluatorCategory } from '../../../types'
import type { ScorerType, BuiltinScorerConfig, RegexScorerConfig, LLMScorerConfig } from '@/features/evaluators/types'

interface AddEvaluatorDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  defaultCategory?: EvaluatorCategory
}

export function AddEvaluatorDialog({
  open,
  onOpenChange,
  defaultCategory = 'heuristics',
}: AddEvaluatorDialogProps) {
  const { addEvaluator } = useExperimentWizard()

  const [name, setName] = useState('')
  const [scorerType, setScorerType] = useState<ScorerType>(
    defaultCategory === 'llm_judges' ? 'llm' : 'builtin'
  )

  // Builtin config
  const [builtinScorerName, setBuiltinScorerName] = useState<string>('contains')
  const [builtinConfig, setBuiltinConfig] = useState<Record<string, unknown>>({})

  // Regex config
  const [regexPattern, setRegexPattern] = useState('')
  const [regexScoreName, setRegexScoreName] = useState('regex_match')

  // LLM config
  const [llmModel, setLlmModel] = useState('gpt-4o-mini')
  const [llmPrompt, setLlmPrompt] = useState('')

  const resetForm = () => {
    setName('')
    setScorerType(defaultCategory === 'llm_judges' ? 'llm' : 'builtin')
    setBuiltinScorerName('contains')
    setBuiltinConfig({})
    setRegexPattern('')
    setRegexScoreName('regex_match')
    setLlmModel('gpt-4o-mini')
    setLlmPrompt('')
  }

  const handleSubmit = () => {
    let config: BuiltinScorerConfig | RegexScorerConfig | LLMScorerConfig

    switch (scorerType) {
      case 'builtin':
        config = {
          scorer_name: builtinScorerName as BuiltinScorerConfig['scorer_name'],
          config: builtinConfig,
        }
        break
      case 'regex':
        config = {
          pattern: regexPattern,
          score_name: regexScoreName,
          match_score: 1,
          no_match_score: 0,
        }
        break
      case 'llm':
        config = {
          credential_id: '', // Will be set by backend
          model: llmModel,
          messages: [
            { role: 'system', content: 'You are a helpful assistant that evaluates AI outputs.' },
            { role: 'user', content: llmPrompt || 'Evaluate the following output: {{output}}' },
          ],
          temperature: 0,
          response_format: 'json',
          output_schema: [
            { name: 'score', type: 'numeric', min_value: 0, max_value: 1 },
            { name: 'reasoning', type: 'categorical' },
          ],
        }
        break
    }

    addEvaluator({
      name: name || `${scorerType} evaluator`,
      scorer_type: scorerType,
      scorer_config: config,
    })

    resetForm()
    onOpenChange(false)
  }

  const isValid = () => {
    if (!name.trim()) return false

    switch (scorerType) {
      case 'builtin':
        return !!builtinScorerName
      case 'regex':
        return !!regexPattern.trim()
      case 'llm':
        return !!llmModel
      default:
        return false
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Add Evaluator</DialogTitle>
          <DialogDescription>
            Configure an evaluator to score your experiment outputs.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="evaluator-name">Name *</Label>
            <Input
              id="evaluator-name"
              placeholder="e.g., Relevance Check"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <Tabs
            value={scorerType}
            onValueChange={(v) => setScorerType(v as ScorerType)}
          >
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="builtin" className="flex items-center gap-2">
                <Calculator className="h-4 w-4" />
                Builtin
              </TabsTrigger>
              <TabsTrigger value="regex" className="flex items-center gap-2">
                <Regex className="h-4 w-4" />
                Regex
              </TabsTrigger>
              <TabsTrigger value="llm" className="flex items-center gap-2">
                <Bot className="h-4 w-4" />
                LLM
              </TabsTrigger>
            </TabsList>

            <TabsContent value="builtin" className="space-y-4 mt-4">
              <div className="space-y-2">
                <Label>Scorer Type</Label>
                <Select value={builtinScorerName} onValueChange={setBuiltinScorerName}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select scorer..." />
                  </SelectTrigger>
                  <SelectContent>
                    {BUILTIN_SCORERS.map((scorer) => (
                      <SelectItem key={scorer.name} value={scorer.name}>
                        <div className="flex flex-col">
                          <span>{scorer.label}</span>
                          <span className="text-xs text-muted-foreground">
                            {scorer.description}
                          </span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {builtinScorerName === 'contains' && (
                <div className="space-y-2">
                  <Label>Substring to check</Label>
                  <Input
                    placeholder="e.g., yes, thank you"
                    value={(builtinConfig.substring as string) || ''}
                    onChange={(e) =>
                      setBuiltinConfig({ ...builtinConfig, substring: e.target.value })
                    }
                  />
                </div>
              )}

              {builtinScorerName === 'length_check' && (
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Min Length</Label>
                    <Input
                      type="number"
                      placeholder="0"
                      value={(builtinConfig.min_length as number) || ''}
                      onChange={(e) =>
                        setBuiltinConfig({
                          ...builtinConfig,
                          min_length: parseInt(e.target.value) || undefined,
                        })
                      }
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Max Length</Label>
                    <Input
                      type="number"
                      placeholder="1000"
                      value={(builtinConfig.max_length as number) || ''}
                      onChange={(e) =>
                        setBuiltinConfig({
                          ...builtinConfig,
                          max_length: parseInt(e.target.value) || undefined,
                        })
                      }
                    />
                  </div>
                </div>
              )}
            </TabsContent>

            <TabsContent value="regex" className="space-y-4 mt-4">
              <div className="space-y-2">
                <Label>Pattern *</Label>
                <Input
                  placeholder="e.g., \\d{4}-\\d{2}-\\d{2}"
                  value={regexPattern}
                  onChange={(e) => setRegexPattern(e.target.value)}
                  className="font-mono"
                />
                <p className="text-xs text-muted-foreground">
                  Regular expression pattern to match against output.
                </p>
              </div>
              <div className="space-y-2">
                <Label>Score Name</Label>
                <Input
                  placeholder="regex_match"
                  value={regexScoreName}
                  onChange={(e) => setRegexScoreName(e.target.value)}
                />
              </div>
            </TabsContent>

            <TabsContent value="llm" className="space-y-4 mt-4">
              <div className="space-y-2">
                <Label>Model</Label>
                <Select value={llmModel} onValueChange={setLlmModel}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select model..." />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="gpt-4o-mini">GPT-4o Mini</SelectItem>
                    <SelectItem value="gpt-4o">GPT-4o</SelectItem>
                    <SelectItem value="claude-3-haiku-20240307">Claude 3 Haiku</SelectItem>
                    <SelectItem value="claude-3-5-sonnet-20241022">Claude 3.5 Sonnet</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Evaluation Prompt</Label>
                <Textarea
                  placeholder="Evaluate the following output for relevance and accuracy: {{output}}"
                  value={llmPrompt}
                  onChange={(e) => setLlmPrompt(e.target.value)}
                  rows={4}
                />
                <p className="text-xs text-muted-foreground">
                  Use {"{{output}}"} to reference the experiment output.
                </p>
              </div>
            </TabsContent>
          </Tabs>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={!isValid()}>
            Add Evaluator
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
