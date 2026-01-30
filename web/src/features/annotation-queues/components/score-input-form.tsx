'use client'

import { useEffect } from 'react'
import { Label } from '@/components/ui/label'
import { Slider } from '@/components/ui/slider'
import { Textarea } from '@/components/ui/textarea'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { useScoreConfigsByIdsQuery, type ScoreConfig } from '@/features/scores'
import type { ScoreSubmission } from '../types'

interface ScoreInputFormProps {
  projectId: string
  queueId: string
  scoreConfigIds: string[]
  scores: ScoreSubmission[]
  onScoresChange: (scores: ScoreSubmission[]) => void
}

export function ScoreInputForm({
  projectId,
  queueId,
  scoreConfigIds,
  scores,
  onScoresChange,
}: ScoreInputFormProps) {
  // Fetch actual score configs based on scoreConfigIds
  const { data: scoreConfigs = [], isLoading } = useScoreConfigsByIdsQuery(
    projectId,
    scoreConfigIds
  )

  // Initialize scores if empty
  useEffect(() => {
    if (scores.length === 0 && scoreConfigs.length > 0) {
      const initialScores = scoreConfigs.map((config) => ({
        score_config_id: config.id,
        value: getDefaultValue(config),
        comment: undefined,
      }))
      onScoresChange(initialScores)
    }
  }, [scoreConfigs, scores.length, onScoresChange])

  const updateScore = (configId: string, value: number | string | boolean, comment?: string) => {
    const newScores = scores.map((s) =>
      s.score_config_id === configId ? { ...s, value, comment } : s
    )
    // If score doesn't exist yet, add it
    if (!newScores.find((s) => s.score_config_id === configId)) {
      newScores.push({ score_config_id: configId, value, comment })
    }
    onScoresChange(newScores)
  }

  const getScoreValue = (configId: string): number | string | boolean | undefined => {
    return scores.find((s) => s.score_config_id === configId)?.value
  }

  const getScoreComment = (configId: string): string | undefined => {
    return scores.find((s) => s.score_config_id === configId)?.comment ?? undefined
  }

  // Show loading state while fetching configs
  if (isLoading) {
    return (
      <div className="space-y-6">
        <h4 className="font-medium text-sm">Scores</h4>
        <div className="space-y-3 rounded-lg border p-4">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-6 w-full" />
          <Skeleton className="h-16 w-full" />
        </div>
      </div>
    )
  }

  if (scoreConfigs.length === 0) {
    return (
      <div className="text-sm text-muted-foreground text-center py-4">
        No score configurations assigned to this queue.
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h4 className="font-medium text-sm">Scores</h4>
      {scoreConfigs.map((config) => (
        <ScoreInput
          key={config.id}
          config={config}
          value={getScoreValue(config.id)}
          comment={getScoreComment(config.id)}
          onChange={(value, comment) => updateScore(config.id, value, comment)}
        />
      ))}
    </div>
  )
}

interface ScoreInputProps {
  config: ScoreConfig
  value: number | string | boolean | undefined
  comment?: string
  onChange: (value: number | string | boolean, comment?: string) => void
}

function ScoreInput({ config, value, comment, onChange }: ScoreInputProps) {
  switch (config.type) {
    case 'NUMERIC':
      return (
        <NumericScoreInput
          config={config}
          value={value as number | undefined}
          comment={comment}
          onChange={onChange}
        />
      )
    case 'CATEGORICAL':
      return (
        <CategoricalScoreInput
          config={config}
          value={value as string | undefined}
          comment={comment}
          onChange={onChange}
        />
      )
    case 'BOOLEAN':
      return (
        <BooleanScoreInput
          config={config}
          value={value as boolean | undefined}
          comment={comment}
          onChange={onChange}
        />
      )
  }
}

function NumericScoreInput({
  config,
  value,
  comment,
  onChange,
}: {
  config: ScoreConfig
  value: number | undefined
  comment?: string
  onChange: (value: number, comment?: string) => void
}) {
  const min = config.min_value ?? 0
  const max = config.max_value ?? 10
  const currentValue = value ?? min

  return (
    <div className="space-y-3 rounded-lg border p-4">
      <div className="flex items-center justify-between">
        <div>
          <Label className="font-medium">{config.name}</Label>
          {config.description && (
            <p className="text-xs text-muted-foreground mt-0.5">{config.description}</p>
          )}
        </div>
        <span className="font-mono text-lg font-semibold">{currentValue}</span>
      </div>
      <Slider
        value={[currentValue]}
        min={min}
        max={max}
        step={1}
        onValueChange={([val]) => onChange(val, comment)}
        className="w-full"
      />
      <div className="flex justify-between text-xs text-muted-foreground">
        <span>{min}</span>
        <span>{max}</span>
      </div>
      <Textarea
        placeholder="Optional comment..."
        value={comment ?? ''}
        onChange={(e) => onChange(currentValue, e.target.value || undefined)}
        rows={2}
        className="text-sm"
      />
    </div>
  )
}

/**
 * Smart input selection logic (Opik + Langfuse combined pattern)
 *
 * Uses ToggleGroup for ≤4 categories with short names (≤10 chars each)
 * Uses Select dropdown for many categories or long names
 */
function shouldUseToggleGroup(categories: string[]): boolean {
  const MAX_TOGGLE_CATEGORIES = 4
  const MAX_LABEL_LENGTH = 10 // Opik uses 10 chars

  // Too many categories → use Select
  if (categories.length > MAX_TOGGLE_CATEGORIES) return false

  // Any long category name → use Select (prevents overflow)
  const hasLongNames = categories.some((cat) => cat.length > MAX_LABEL_LENGTH)
  if (hasLongNames) return false

  return true
}

function CategoricalScoreInput({
  config,
  value,
  comment,
  onChange,
}: {
  config: ScoreConfig
  value: string | undefined
  comment?: string
  onChange: (value: string, comment?: string) => void
}) {
  const categories = config.categories ?? ['Good', 'Bad']
  const useToggle = shouldUseToggleGroup(categories)

  return (
    <div className="space-y-3 rounded-lg border p-4">
      <div>
        <Label className="font-medium">{config.name}</Label>
        {config.description && (
          <p className="text-xs text-muted-foreground mt-0.5">{config.description}</p>
        )}
      </div>

      {/* Smart input: ToggleGroup for ≤4 short categories, Select otherwise */}
      {useToggle ? (
        <ToggleGroup
          type="single"
          value={value ?? ''}
          onValueChange={(val) => {
            if (val) onChange(val, comment)
          }}
          className="justify-start flex-wrap"
        >
          {categories.map((cat) => (
            <ToggleGroupItem
              key={cat}
              value={cat}
              aria-label={cat}
              className="px-3 py-1.5 text-sm"
            >
              {cat}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>
      ) : (
        <Select value={value} onValueChange={(val) => onChange(val, comment)}>
          <SelectTrigger>
            <SelectValue placeholder="Select a value" />
          </SelectTrigger>
          <SelectContent>
            {categories.map((category) => (
              <SelectItem key={category} value={category}>
                {category}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}

      <Textarea
        placeholder="Optional comment..."
        value={comment ?? ''}
        onChange={(e) => onChange(value ?? '', e.target.value || undefined)}
        rows={2}
        className="text-sm"
      />
    </div>
  )
}

function BooleanScoreInput({
  config,
  value,
  comment,
  onChange,
}: {
  config: ScoreConfig
  value: boolean | undefined
  comment?: string
  onChange: (value: boolean, comment?: string) => void
}) {
  // Convert boolean to string for ToggleGroup
  const stringValue = value === true ? 'true' : value === false ? 'false' : ''

  return (
    <div className="space-y-3 rounded-lg border p-4">
      <div>
        <Label className="font-medium">{config.name}</Label>
        {config.description && (
          <p className="text-xs text-muted-foreground mt-0.5">{config.description}</p>
        )}
      </div>

      {/* Two-button toggle for boolean values (clearer than Switch) */}
      <ToggleGroup
        type="single"
        value={stringValue}
        onValueChange={(val) => {
          if (val === 'true') onChange(true, comment)
          else if (val === 'false') onChange(false, comment)
        }}
        className="justify-start"
      >
        <ToggleGroupItem
          value="true"
          aria-label="Yes"
          className="px-4 py-1.5 text-sm data-[state=on]:bg-green-100 data-[state=on]:text-green-700 dark:data-[state=on]:bg-green-900 dark:data-[state=on]:text-green-300"
        >
          Yes
        </ToggleGroupItem>
        <ToggleGroupItem
          value="false"
          aria-label="No"
          className="px-4 py-1.5 text-sm data-[state=on]:bg-red-100 data-[state=on]:text-red-700 dark:data-[state=on]:bg-red-900 dark:data-[state=on]:text-red-300"
        >
          No
        </ToggleGroupItem>
      </ToggleGroup>

      <Textarea
        placeholder="Optional comment..."
        value={comment ?? ''}
        onChange={(e) => onChange(value ?? false, e.target.value || undefined)}
        rows={2}
        className="text-sm"
      />
    </div>
  )
}

function getDefaultValue(config: ScoreConfig): number | string | boolean {
  switch (config.type) {
    case 'NUMERIC':
      return config.min_value ?? 0
    case 'CATEGORICAL':
      return config.categories?.[0] ?? ''
    case 'BOOLEAN':
      return false
  }
}
