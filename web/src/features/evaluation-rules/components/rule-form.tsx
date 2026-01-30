'use client'

import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Slider } from '@/components/ui/slider'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Separator } from '@/components/ui/separator'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import { Bot, Code, Regex } from 'lucide-react'
import { RuleFilterBuilder } from './rule-filter-builder'
import { VariableMappingEditor } from './variable-mapping-editor'
import { LLMConfigPanel, type LLMConfig } from './llm-config-panel'
import { OutputSchemaBuilder } from './output-schema-builder'
import type {
  EvaluationRule,
  CreateEvaluationRuleRequest,
  LLMScorerConfig,
  BuiltinScorerConfig,
  RegexScorerConfig,
  FilterClause,
  VariableMap,
  OutputField,
} from '../types'

// Default output schema for new rules
const DEFAULT_OUTPUT_SCHEMA: OutputField[] = [
  {
    name: 'score',
    type: 'numeric',
    description: 'Quality score from 0 to 1',
    min_value: 0,
    max_value: 1,
  },
]

function extractScorerDefaults(rule?: EvaluationRule) {
  if (!rule) {
    return {
      llm_credential_id: '',
      llm_model: '',
      llm_system_prompt: '',
      llm_user_prompt: '',
      llm_temperature: 0,
      llm_output_schema: DEFAULT_OUTPUT_SCHEMA,
      builtin_scorer_name: 'contains' as const,
      regex_pattern: '',
      regex_score_name: 'regex_match',
      regex_match_score: 1,
      regex_no_match_score: 0,
    }
  }

  if (rule.scorer_type === 'llm') {
    const config = rule.scorer_config as LLMScorerConfig
    const systemMsg = config.messages?.find((m) => m.role === 'system')?.content ?? ''
    const userMsg = config.messages?.find((m) => m.role === 'user')?.content ?? ''
    return {
      llm_credential_id: config.credential_id ?? '',
      llm_model: config.model ?? '',
      llm_system_prompt: systemMsg,
      llm_user_prompt: userMsg,
      llm_temperature: config.temperature ?? 0,
      llm_output_schema: config.output_schema ?? DEFAULT_OUTPUT_SCHEMA,
      builtin_scorer_name: 'contains' as const,
      regex_pattern: '',
      regex_score_name: 'regex_match',
      regex_match_score: 1,
      regex_no_match_score: 0,
    }
  }

  if (rule.scorer_type === 'builtin') {
    const config = rule.scorer_config as BuiltinScorerConfig
    return {
      llm_credential_id: '',
      llm_model: '',
      llm_system_prompt: '',
      llm_user_prompt: '',
      llm_temperature: 0,
      llm_output_schema: DEFAULT_OUTPUT_SCHEMA,
      builtin_scorer_name: config.scorer_name ?? 'contains',
      regex_pattern: '',
      regex_score_name: 'regex_match',
      regex_match_score: 1,
      regex_no_match_score: 0,
    }
  }

  if (rule.scorer_type === 'regex') {
    const config = rule.scorer_config as RegexScorerConfig
    return {
      llm_credential_id: '',
      llm_model: '',
      llm_system_prompt: '',
      llm_user_prompt: '',
      llm_temperature: 0,
      llm_output_schema: DEFAULT_OUTPUT_SCHEMA,
      builtin_scorer_name: 'contains' as const,
      regex_pattern: config.pattern ?? '',
      regex_score_name: config.score_name ?? 'regex_match',
      regex_match_score: config.match_score ?? 1,
      regex_no_match_score: config.no_match_score ?? 0,
    }
  }

  return {
    llm_credential_id: '',
    llm_model: '',
    llm_system_prompt: '',
    llm_user_prompt: '',
    llm_temperature: 0,
    llm_output_schema: DEFAULT_OUTPUT_SCHEMA,
    builtin_scorer_name: 'contains' as const,
    regex_pattern: '',
    regex_score_name: 'regex_match',
    regex_match_score: 1,
    regex_no_match_score: 0,
  }
}

// Schema for filter clause
const filterClauseSchema = z.object({
  field: z.string(),
  operator: z.enum(['equals', 'not_equals', 'contains', 'gt', 'lt', 'gte', 'lte', 'is_empty', 'is_not_empty']),
  value: z.unknown(),
})

// Schema for variable mapping
const variableMapSchema = z.object({
  variable_name: z.string(),
  source: z.enum(['span_input', 'span_output', 'span_metadata', 'trace_input']),
  json_path: z.string(),
})

// Schema for output field
const outputFieldSchema = z.object({
  name: z.string(),
  type: z.enum(['numeric', 'categorical', 'boolean']),
  description: z.string().optional(),
  min_value: z.number().optional(),
  max_value: z.number().optional(),
  categories: z.array(z.string()).optional(),
})

const ruleFormSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  description: z.string(),
  status: z.enum(['active', 'inactive', 'paused']),
  target_scope: z.enum(['span', 'trace']),
  span_names: z.string(),
  sampling_rate: z.number().min(0).max(1),
  filter: z.array(filterClauseSchema),
  variable_mapping: z.array(variableMapSchema),
  scorer_type: z.enum(['llm', 'builtin', 'regex']),
  llm_credential_id: z.string(),
  llm_model: z.string(),
  llm_system_prompt: z.string(),
  llm_user_prompt: z.string(),
  llm_temperature: z.number().min(0).max(2),
  llm_output_schema: z.array(outputFieldSchema),
  builtin_scorer_name: z.string(),
  regex_pattern: z.string(),
  regex_score_name: z.string(),
  regex_match_score: z.number(),
  regex_no_match_score: z.number(),
})

type RuleFormData = z.infer<typeof ruleFormSchema>

interface RuleFormProps {
  rule?: EvaluationRule
  onSubmit: (data: CreateEvaluationRuleRequest) => void
  onCancel: () => void
  isLoading?: boolean
  orgId?: string
}

export function RuleForm({
  rule,
  onSubmit,
  onCancel,
  isLoading,
  orgId,
}: RuleFormProps) {
  const scorerDefaults = extractScorerDefaults(rule)

  const form = useForm<RuleFormData>({
    resolver: zodResolver(ruleFormSchema),
    defaultValues: {
      name: rule?.name ?? '',
      description: rule?.description ?? '',
      status: rule?.status ?? 'inactive',
      target_scope: rule?.target_scope ?? 'span',
      span_names: rule?.span_names?.join(', ') ?? '',
      sampling_rate: rule?.sampling_rate ?? 1,
      filter: (rule?.filter ?? []) as FilterClause[],
      variable_mapping: (rule?.variable_mapping ?? []) as VariableMap[],
      scorer_type: rule?.scorer_type ?? 'llm',
      ...scorerDefaults,
    },
  })

  const scorerType = useWatch({ control: form.control, name: 'scorer_type' })
  const samplingRate = useWatch({ control: form.control, name: 'sampling_rate' })
  const llmUserPrompt = useWatch({ control: form.control, name: 'llm_user_prompt' })

  const handleFormSubmit = (data: RuleFormData) => {
    let scorer_config: CreateEvaluationRuleRequest['scorer_config']

    if (data.scorer_type === 'llm') {
      scorer_config = {
        credential_id: data.llm_credential_id,
        model: data.llm_model,
        messages: [
          ...(data.llm_system_prompt ? [{ role: 'system' as const, content: data.llm_system_prompt }] : []),
          { role: 'user' as const, content: data.llm_user_prompt },
        ],
        temperature: data.llm_temperature,
        response_format: 'json' as const,
        output_schema: data.llm_output_schema.map((field) => ({
          ...field,
          type: field.type as 'numeric' | 'categorical' | 'boolean',
        })),
      }
    } else if (data.scorer_type === 'builtin') {
      scorer_config = {
        scorer_name: (data.builtin_scorer_name || 'contains') as 'contains' | 'json_valid' | 'length_check' | 'sentiment' | 'toxicity',
        config: {},
      }
    } else {
      scorer_config = {
        pattern: data.regex_pattern,
        score_name: data.regex_score_name,
        match_score: data.regex_match_score,
        no_match_score: data.regex_no_match_score,
      }
    }

    onSubmit({
      name: data.name,
      description: data.description || undefined,
      status: data.status,
      target_scope: data.target_scope,
      span_names: data.span_names
        ? data.span_names.split(',').map((s) => s.trim()).filter(Boolean)
        : undefined,
      sampling_rate: data.sampling_rate,
      filter: data.filter.length > 0 ? data.filter : undefined,
      variable_mapping: data.variable_mapping.length > 0 ? data.variable_mapping : undefined,
      scorer_type: data.scorer_type,
      scorer_config,
    })
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleFormSubmit)} className="space-y-6">
        <Tabs defaultValue="basic" className="w-full">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="basic">Basic Info</TabsTrigger>
            <TabsTrigger value="targeting">Targeting</TabsTrigger>
            <TabsTrigger value="scorer">Scorer</TabsTrigger>
          </TabsList>

          <TabsContent value="basic" className="space-y-4 mt-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input placeholder="e.g., Response Quality Check" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Describe what this rule evaluates"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="status"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Status</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select status" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="inactive">Inactive (Draft)</SelectItem>
                      <SelectItem value="active">Active (Scoring)</SelectItem>
                      <SelectItem value="paused">Paused</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    Active rules will automatically score matching spans.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </TabsContent>

          <TabsContent value="targeting" className="space-y-4 mt-4">
            <FormField
              control={form.control}
              name="target_scope"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Target Scope</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select scope" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="span">Individual Spans</SelectItem>
                      <SelectItem value="trace">Full Traces</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    Whether to evaluate individual spans or complete traces.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="span_names"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Span Names (Optional)</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="e.g., chat-completion, rag-retrieval"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    Comma-separated list of span names to match. Leave empty to match all spans.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="sampling_rate"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Sampling Rate: {Math.round(samplingRate * 100)}%</FormLabel>
                  <FormControl>
                    <Slider
                      min={0}
                      max={1}
                      step={0.01}
                      value={[field.value]}
                      onValueChange={(values) => field.onChange(values[0])}
                    />
                  </FormControl>
                  <FormDescription>
                    Percentage of matching spans to evaluate. Use lower values for high-volume apps.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Separator className="my-4" />

            {/* Filter Builder */}
            <FormField
              control={form.control}
              name="filter"
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <RuleFilterBuilder
                      value={field.value}
                      onChange={field.onChange}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Separator className="my-4" />

            {/* Variable Mapping Editor - only show for LLM scorer */}
            {scorerType === 'llm' && (
              <FormField
                control={form.control}
                name="variable_mapping"
                render={({ field }) => (
                  <FormItem>
                    <FormControl>
                      <VariableMappingEditor
                        value={field.value}
                        onChange={field.onChange}
                        promptTemplate={llmUserPrompt}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
          </TabsContent>

          <TabsContent value="scorer" className="space-y-4 mt-4">
            <FormField
              control={form.control}
              name="scorer_type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Scorer Type</FormLabel>
                  <FormControl>
                    <RadioGroup
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                      className="grid grid-cols-3 gap-4"
                    >
                      <div>
                        <RadioGroupItem
                          value="llm"
                          id="scorer-llm"
                          className="peer sr-only"
                        />
                        <Label
                          htmlFor="scorer-llm"
                          className="flex flex-col items-start gap-2 rounded-lg border-2 border-muted bg-popover p-4 cursor-pointer transition-colors hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary peer-data-[state=checked]:bg-primary/5 [&:has([data-state=checked])]:border-primary"
                        >
                          <Bot className="h-6 w-6 text-primary" />
                          <div className="font-medium text-sm">LLM</div>
                          <p className="text-xs text-muted-foreground">
                            Use an LLM to evaluate quality, relevance, or custom criteria.
                          </p>
                        </Label>
                      </div>

                      <div>
                        <RadioGroupItem
                          value="builtin"
                          id="scorer-builtin"
                          className="peer sr-only"
                        />
                        <Label
                          htmlFor="scorer-builtin"
                          className="flex flex-col items-start gap-2 rounded-lg border-2 border-muted bg-popover p-4 cursor-pointer transition-colors hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary peer-data-[state=checked]:bg-primary/5 [&:has([data-state=checked])]:border-primary"
                        >
                          <Code className="h-6 w-6 text-primary" />
                          <div className="font-medium text-sm">Built-in</div>
                          <p className="text-xs text-muted-foreground">
                            Pre-built scorers for common checks like JSON validity.
                          </p>
                        </Label>
                      </div>

                      <div>
                        <RadioGroupItem
                          value="regex"
                          id="scorer-regex"
                          className="peer sr-only"
                        />
                        <Label
                          htmlFor="scorer-regex"
                          className="flex flex-col items-start gap-2 rounded-lg border-2 border-muted bg-popover p-4 cursor-pointer transition-colors hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary peer-data-[state=checked]:bg-primary/5 [&:has([data-state=checked])]:border-primary"
                        >
                          <Regex className="h-6 w-6 text-primary" />
                          <div className="font-medium text-sm">Regex</div>
                          <p className="text-xs text-muted-foreground">
                            Pattern matching for detecting specific content.
                          </p>
                        </Label>
                      </div>
                    </RadioGroup>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {scorerType === 'llm' && (
              <div className="space-y-4">
                {/* LLM Config Panel - Model selection and temperature */}
                <LLMConfigPanel
                  orgId={orgId}
                  config={{
                    credential_id: form.watch('llm_credential_id'),
                    model: form.watch('llm_model'),
                    temperature: form.watch('llm_temperature'),
                  }}
                  onChange={(config: LLMConfig) => {
                    form.setValue('llm_credential_id', config.credential_id)
                    form.setValue('llm_model', config.model)
                    form.setValue('llm_temperature', config.temperature)
                  }}
                />

                {/* Prompt Configuration */}
                <div className="space-y-4 border rounded-lg p-4">
                  <h4 className="font-medium">Evaluation Prompt</h4>

                  <FormField
                    control={form.control}
                    name="llm_system_prompt"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>System Prompt (Optional)</FormLabel>
                        <FormControl>
                          <Textarea
                            placeholder="You are an expert evaluator..."
                            className="min-h-[80px]"
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="llm_user_prompt"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>User Prompt</FormLabel>
                        <FormControl>
                          <Textarea
                            placeholder="Evaluate the following response for quality:&#10;&#10;Input: {{input}}&#10;Output: {{output}}&#10;&#10;Rate from 0-1 based on..."
                            className="min-h-[120px] font-mono text-sm"
                            {...field}
                          />
                        </FormControl>
                        <FormDescription>
                          Use {'{{input}}'}, {'{{output}}'}, {'{{metadata}}'} as variables.
                          Configure mappings in the Targeting tab.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                {/* Output Schema Builder */}
                <FormField
                  control={form.control}
                  name="llm_output_schema"
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <OutputSchemaBuilder
                          value={field.value}
                          onChange={field.onChange}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            )}

            {scorerType === 'builtin' && (
              <div className="space-y-4 border rounded-lg p-4">
                <h4 className="font-medium">Built-in Scorer Configuration</h4>

                <FormField
                  control={form.control}
                  name="builtin_scorer_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Scorer</FormLabel>
                      <Select onValueChange={field.onChange} defaultValue={field.value}>
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="Select scorer" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          <SelectItem value="contains">Contains - Check for substring</SelectItem>
                          <SelectItem value="json_valid">JSON Valid - Validate JSON structure</SelectItem>
                          <SelectItem value="length_check">Length Check - Min/max length</SelectItem>
                          <SelectItem value="sentiment">Sentiment - Basic sentiment analysis</SelectItem>
                          <SelectItem value="toxicity">Toxicity - Toxicity detection</SelectItem>
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            )}

            {scorerType === 'regex' && (
              <div className="space-y-4 border rounded-lg p-4">
                <h4 className="font-medium">Regex Scorer Configuration</h4>

                <FormField
                  control={form.control}
                  name="regex_pattern"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Pattern</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="e.g., \b(error|fail|exception)\b"
                          className="font-mono"
                          {...field}
                        />
                      </FormControl>
                      <FormDescription>
                        Regular expression pattern to match against span output.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="regex_score_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Score Name</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="e.g., contains_error"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="grid grid-cols-2 gap-4">
                  <FormField
                    control={form.control}
                    name="regex_match_score"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Match Score</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            step="0.1"
                            value={field.value}
                            onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
                          />
                        </FormControl>
                        <FormDescription>Score when pattern matches.</FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="regex_no_match_score"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>No Match Score</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            step="0.1"
                            value={field.value}
                            onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
                          />
                        </FormControl>
                        <FormDescription>Score when pattern doesn&apos;t match.</FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </div>
            )}
          </TabsContent>
        </Tabs>

        <div className="flex justify-end gap-2 pt-4 border-t">
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit" disabled={isLoading}>
            {isLoading ? 'Saving...' : rule ? 'Update Rule' : 'Create Rule'}
          </Button>
        </div>
      </form>
    </Form>
  )
}
