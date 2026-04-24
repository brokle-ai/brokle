import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import type { ScoreConfig } from '@/features/scores/api/types'

// A single row in the review form — one input widget per configured
// score. NUMERIC → number input (respects min/max if present),
// CATEGORICAL → select of the config's `categories`, BOOLEAN → Yes/No
// toggle. The widget dispatches on `config.type`; unsupported types
// render nothing rather than fall back to a generic number so upstream
// type drift surfaces as a visible gap rather than silent miscoding.
//
// The `value` shape mirrors the wire-level ScoreSubmission.value —
// `number | string | boolean` — so the parent can forward each entry
// to `CompleteItemRequest.scores` without further coercion.

export interface ScoreInputValue {
  value: number | string | boolean | null
  comment: string
}

interface ScoreInputProps {
  config: ScoreConfig
  value: ScoreInputValue
  onChange: (next: ScoreInputValue) => void
  disabled?: boolean
}

export function ScoreInput({
  config,
  value,
  onChange,
  disabled,
}: ScoreInputProps) {
  const setValue = (v: number | string | boolean | null) =>
    onChange({ ...value, value: v })
  const setComment = (c: string) => onChange({ ...value, comment: c })

  return (
    <div className="space-y-2 rounded-md border p-3">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <Label className="text-sm font-medium">{config.name}</Label>
          {config.description ? (
            <p className="mt-0.5 text-xs text-muted-foreground">
              {config.description}
            </p>
          ) : null}
        </div>
        <span className="shrink-0 text-[10px] uppercase tracking-wide text-muted-foreground">
          {config.type.toLowerCase()}
        </span>
      </div>

      {config.type === 'NUMERIC' ? (
        <NumericField
          value={typeof value.value === 'number' ? value.value : null}
          onChange={setValue}
          min={config.min_value}
          max={config.max_value}
          disabled={disabled}
        />
      ) : null}

      {config.type === 'CATEGORICAL' ? (
        <CategoricalField
          value={typeof value.value === 'string' ? value.value : null}
          onChange={setValue}
          categories={config.categories ?? []}
          disabled={disabled}
        />
      ) : null}

      {config.type === 'BOOLEAN' ? (
        <BooleanField
          value={typeof value.value === 'boolean' ? value.value : null}
          onChange={setValue}
          disabled={disabled}
        />
      ) : null}

      <Textarea
        value={value.comment}
        onChange={(e) => setComment(e.target.value)}
        placeholder="Optional comment..."
        rows={2}
        className="resize-none text-sm"
        disabled={disabled}
      />
    </div>
  )
}

interface NumericFieldProps {
  value: number | null
  onChange: (v: number | null) => void
  min: number | undefined
  max: number | undefined
  disabled?: boolean
}

function NumericField({
  value,
  onChange,
  min,
  max,
  disabled,
}: NumericFieldProps) {
  return (
    <div className="space-y-1">
      <Input
        type="number"
        inputMode="decimal"
        value={value ?? ''}
        onChange={(e) => {
          const raw = e.target.value
          if (raw === '') {
            onChange(null)
            return
          }
          const parsed = Number(raw)
          onChange(Number.isFinite(parsed) ? parsed : null)
        }}
        min={min}
        max={max}
        step="any"
        placeholder={
          min !== undefined && max !== undefined
            ? `${min} - ${max}`
            : 'Enter a number'
        }
        disabled={disabled}
      />
      {min !== undefined && max !== undefined ? (
        <p className="text-[11px] text-muted-foreground">
          Range: {min} to {max}
        </p>
      ) : null}
    </div>
  )
}

interface CategoricalFieldProps {
  value: string | null
  onChange: (v: string | null) => void
  categories: string[]
  disabled?: boolean
}

function CategoricalField({
  value,
  onChange,
  categories,
  disabled,
}: CategoricalFieldProps) {
  if (categories.length === 0) {
    return (
      <p className="text-xs italic text-muted-foreground">
        No categories configured.
      </p>
    )
  }
  return (
    <Select
      value={value ?? ''}
      onValueChange={(v) => onChange(v || null)}
      disabled={disabled}
    >
      <SelectTrigger>
        <SelectValue placeholder="Select a category..." />
      </SelectTrigger>
      <SelectContent>
        {categories.map((c) => (
          <SelectItem key={c} value={c}>
            {c}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}

interface BooleanFieldProps {
  value: boolean | null
  onChange: (v: boolean | null) => void
  disabled?: boolean
}

function BooleanField({ value, onChange, disabled }: BooleanFieldProps) {
  return (
    <div className="flex gap-2">
      <Button
        type="button"
        variant={value === true ? 'default' : 'outline'}
        className={cn('flex-1')}
        onClick={() => onChange(true)}
        disabled={disabled}
      >
        Yes
      </Button>
      <Button
        type="button"
        variant={value === false ? 'default' : 'outline'}
        className={cn('flex-1')}
        onClick={() => onChange(false)}
        disabled={disabled}
      >
        No
      </Button>
    </div>
  )
}
