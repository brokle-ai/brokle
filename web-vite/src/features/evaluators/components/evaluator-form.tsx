import { useMemo, useState, type FormEvent } from 'react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { PromptEditor } from '@/editors/prompt-editor'
import type {
  EvaluatorStatus,
  EvaluatorTrigger,
  FilterClause,
  ScorerType,
  TargetScope,
} from '../api/types'
import { EvaluatorFilterBuilder } from './evaluator-filter-builder'

// Form state mixes structured editors and raw JSON. The structured
// path covers `filter` (visual builder) + the basics; raw JSON
// remains for `scorer_config` and `variable_mapping` because
// scorer schemas are scorer-type-discriminated and the variable
// mapping is a v2 port. The backend validates the JSON shapes so we
// trust submit-time syntax + server response.
export interface EvaluatorFormState {
  name: string
  description: string
  scorerType: ScorerType
  targetScope: TargetScope
  triggerType: EvaluatorTrigger
  status: EvaluatorStatus
  samplingRate: string
  spanNamesCsv: string
  scorerConfigJson: string
  filter: FilterClause[]
  variableMappingJson: string
}

export const EMPTY_EVALUATOR_FORM: EvaluatorFormState = {
  name: '',
  description: '',
  scorerType: 'llm',
  targetScope: 'span',
  triggerType: 'automatic',
  status: 'active',
  samplingRate: '1.0',
  spanNamesCsv: '',
  scorerConfigJson: '{\n  \n}\n',
  filter: [],
  variableMappingJson: '[]',
}

interface EvaluatorFormProps {
  mode: 'create' | 'edit'
  initial: EvaluatorFormState
  onSubmit: (state: EvaluatorFormState) => void
  isSubmitting?: boolean
  error?: string | null
  onCancel?: () => void
}

export interface ParsedEvaluatorForm {
  name: string
  description?: string
  scorerType: ScorerType
  targetScope: TargetScope
  triggerType: EvaluatorTrigger
  status: EvaluatorStatus
  samplingRate: number
  spanNames: string[]
  scorerConfig: Record<string, unknown>
  filter: FilterClause[]
  variableMapping: unknown[]
}

export function parseEvaluatorForm(
  state: EvaluatorFormState,
): { ok: true; value: ParsedEvaluatorForm } | { ok: false; message: string } {
  const trimmedDesc = state.description.trim()
  const sampling = Number.parseFloat(state.samplingRate)
  if (!Number.isFinite(sampling) || sampling < 0 || sampling > 1) {
    return {
      ok: false,
      message: 'Sampling rate must be a decimal between 0 and 1.',
    }
  }

  let scorerConfig: unknown
  try {
    scorerConfig = JSON.parse(state.scorerConfigJson || '{}')
  } catch (err) {
    return {
      ok: false,
      message: `Scorer config JSON is invalid: ${
        err instanceof Error ? err.message : 'parse error'
      }`,
    }
  }
  if (
    typeof scorerConfig !== 'object' ||
    scorerConfig === null ||
    Array.isArray(scorerConfig)
  ) {
    return { ok: false, message: 'Scorer config must be a JSON object.' }
  }

  let variableMapping: unknown
  try {
    variableMapping = JSON.parse(state.variableMappingJson || '[]')
  } catch (err) {
    return {
      ok: false,
      message: `Variable mapping JSON is invalid: ${
        err instanceof Error ? err.message : 'parse error'
      }`,
    }
  }
  if (!Array.isArray(variableMapping)) {
    return { ok: false, message: 'Variable mapping must be a JSON array.' }
  }

  const spanNames = state.spanNamesCsv
    .split(',')
    .map((s) => s.trim())
    .filter((s) => s.length > 0)

  return {
    ok: true,
    value: {
      name: state.name.trim(),
      description: trimmedDesc.length > 0 ? trimmedDesc : undefined,
      scorerType: state.scorerType,
      targetScope: state.targetScope,
      triggerType: state.triggerType,
      status: state.status,
      samplingRate: sampling,
      spanNames,
      scorerConfig: scorerConfig as Record<string, unknown>,
      filter: state.filter,
      variableMapping: variableMapping as unknown[],
    },
  }
}

export function EvaluatorForm({
  mode,
  initial,
  onSubmit,
  isSubmitting,
  error,
  onCancel,
}: EvaluatorFormProps) {
  const [state, setState] = useState<EvaluatorFormState>(initial)
  const [clientError, setClientError] = useState<string | null>(null)

  function update<K extends keyof EvaluatorFormState>(
    key: K,
    value: EvaluatorFormState[K],
  ) {
    setState((prev) => ({ ...prev, [key]: value }))
  }

  const displayError = useMemo(
    () => clientError ?? error ?? null,
    [clientError, error],
  )

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const parsed = parseEvaluatorForm(state)
    if (!parsed.ok) {
      setClientError(parsed.message)
      return
    }
    setClientError(null)
    onSubmit(state)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Basics</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="evaluator-name">Name</Label>
            <Input
              id="evaluator-name"
              value={state.name}
              onChange={(e) => update('name', e.target.value)}
              disabled={isSubmitting}
              required
              maxLength={100}
              placeholder="helpfulness-llm-judge"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="evaluator-description">Description</Label>
            <Textarea
              id="evaluator-description"
              value={state.description}
              onChange={(e) => update('description', e.target.value)}
              disabled={isSubmitting}
              rows={2}
            />
          </div>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="evaluator-kind">Kind</Label>
              <Select
                value={state.scorerType}
                onValueChange={(v) => update('scorerType', v as ScorerType)}
                disabled={isSubmitting}
              >
                <SelectTrigger id="evaluator-kind">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="llm">LLM-as-judge</SelectItem>
                  <SelectItem value="builtin">Builtin</SelectItem>
                  <SelectItem value="regex">Regex</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="evaluator-scope">Target scope</Label>
              <Select
                value={state.targetScope}
                onValueChange={(v) => update('targetScope', v as TargetScope)}
                disabled={isSubmitting}
              >
                <SelectTrigger id="evaluator-scope">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="span">Span</SelectItem>
                  <SelectItem value="trace">Trace</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="evaluator-trigger">Trigger</Label>
              <Select
                value={state.triggerType}
                onValueChange={(v) =>
                  update('triggerType', v as EvaluatorTrigger)
                }
                disabled={isSubmitting}
              >
                <SelectTrigger id="evaluator-trigger">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="automatic">Automatic</SelectItem>
                  <SelectItem value="manual">Manual</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="evaluator-sampling">Sampling rate (0–1)</Label>
              <Input
                id="evaluator-sampling"
                value={state.samplingRate}
                onChange={(e) => update('samplingRate', e.target.value)}
                disabled={isSubmitting}
                placeholder="1.0"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="evaluator-spans">Span names (optional)</Label>
              <Input
                id="evaluator-spans"
                value={state.spanNamesCsv}
                onChange={(e) => update('spanNamesCsv', e.target.value)}
                disabled={isSubmitting}
                placeholder="comma-separated, e.g. llm.generate, chain.run"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Scorer config</CardTitle>
          <CardDescription>
            JSON object. For LLM judges include{' '}
            <code className="font-mono text-xs">credential_id</code>,{' '}
            <code className="font-mono text-xs">model</code>,{' '}
            <code className="font-mono text-xs">messages</code>, and{' '}
            <code className="font-mono text-xs">output_schema</code>. The
            backend validates the shape against{' '}
            <code className="font-mono text-xs">scorer_type</code>.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <PromptEditor
            value={state.scorerConfigJson}
            onChange={(next) => update('scorerConfigJson', next)}
            language="json"
            readOnly={isSubmitting}
            minHeight="260px"
            ariaLabel="Scorer config JSON"
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Filter</CardTitle>
          <CardDescription>
            Build clause-by-clause targeting rules. Empty matches every
            span/trace in scope.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <EvaluatorFilterBuilder
            value={state.filter}
            onChange={(next) => update('filter', next)}
            disabled={isSubmitting}
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Variable mapping</CardTitle>
          <CardDescription>
            JSON array mapping template variables to span fields.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Textarea
            value={state.variableMappingJson}
            onChange={(e) => update('variableMappingJson', e.target.value)}
            disabled={isSubmitting}
            rows={6}
            className="font-mono text-xs"
            aria-label="Variable mapping JSON"
          />
        </CardContent>
      </Card>

      {displayError ? (
        <div className="rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
          {displayError}
        </div>
      ) : null}

      <div className="flex items-center justify-end gap-2">
        {onCancel ? (
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
        ) : null}
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting
            ? 'Saving…'
            : mode === 'edit'
              ? 'Save changes'
              : 'Create evaluator'}
        </Button>
      </div>
    </form>
  )
}
