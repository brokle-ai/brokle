'use client'

import { useEffect } from 'react'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Slider } from '@/components/ui/slider'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { ScoreSubmission } from '../types'

// TODO: This should come from the evaluation feature's score configs API
// For now, we'll use a simple placeholder
interface ScoreConfig {
  id: string
  name: string
  description?: string
  type: 'NUMERIC' | 'CATEGORICAL' | 'BOOLEAN'
  min_value?: number
  max_value?: number
  categories?: string[]
}

interface ScoreInputFormProps {
  queueId: string
  scoreConfigIds: string[]
  scores: ScoreSubmission[]
  onScoresChange: (scores: ScoreSubmission[]) => void
}

export function ScoreInputForm({
  queueId,
  scoreConfigIds,
  scores,
  onScoresChange,
}: ScoreInputFormProps) {
  // TODO: Fetch actual score configs based on scoreConfigIds
  // For now, we'll create placeholder configs for demonstration
  const scoreConfigs: ScoreConfig[] = scoreConfigIds.map((id, index) => ({
    id,
    name: `Score ${index + 1}`,
    description: 'Rate this item',
    type: 'NUMERIC',
    min_value: 0,
    max_value: 10,
  }))

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

  return (
    <div className="space-y-3 rounded-lg border p-4">
      <div>
        <Label className="font-medium">{config.name}</Label>
        {config.description && (
          <p className="text-xs text-muted-foreground mt-0.5">{config.description}</p>
        )}
      </div>
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
  const currentValue = value ?? false

  return (
    <div className="space-y-3 rounded-lg border p-4">
      <div className="flex items-center justify-between">
        <div>
          <Label className="font-medium">{config.name}</Label>
          {config.description && (
            <p className="text-xs text-muted-foreground mt-0.5">{config.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">
            {currentValue ? 'Yes' : 'No'}
          </span>
          <Switch
            checked={currentValue}
            onCheckedChange={(val) => onChange(val, comment)}
          />
        </div>
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
