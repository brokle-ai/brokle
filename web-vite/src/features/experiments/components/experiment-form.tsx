import { useState, type FormEvent } from 'react'
import { useQuery } from '@tanstack/react-query'
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
import { datasetListQueryOptions } from '@/features/datasets/api/queries'

export interface ExperimentFormState {
  name: string
  description: string
  // Dataset is optional — backend accepts experiments without a
  // dataset (e.g. trace-sourced experiments created from the trace
  // viewer). Empty string means "no dataset".
  datasetId: string
}

export const EMPTY_EXPERIMENT_FORM: ExperimentFormState = {
  name: '',
  description: '',
  datasetId: '',
}

interface ExperimentFormProps {
  projectId: string
  initial: ExperimentFormState
  onSubmit: (state: ExperimentFormState) => void
  isSubmitting?: boolean
  error?: string | null
  onCancel?: () => void
}

// The sentinel value for "no dataset" in the Select — Radix Select
// cannot represent an empty string, so we use a literal that no UUID
// will ever collide with.
const NO_DATASET = '__none__'

export function ExperimentForm({
  projectId,
  initial,
  onSubmit,
  isSubmitting,
  error,
  onCancel,
}: ExperimentFormProps) {
  const [state, setState] = useState<ExperimentFormState>(initial)

  // Load the full first page of datasets. We cap limit at 100 (the
  // backend's max) — projects with more than 100 datasets are rare
  // and a searchable combobox can replace this if we hit it.
  const datasetsQuery = useQuery(
    datasetListQueryOptions(projectId, { page: 1, limit: 100 }),
  )

  function update<K extends keyof ExperimentFormState>(
    key: K,
    value: ExperimentFormState[K],
  ) {
    setState((prev) => ({ ...prev, [key]: value }))
  }

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    onSubmit(state)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Basics</CardTitle>
          <CardDescription>
            Evaluators, model configuration, and "run immediately" are
            not yet exposed on the dashboard — the backend wizard
            endpoint isn't wired. For now, create the experiment and
            attach runs via the SDK or the legacy dashboard.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="experiment-name">Name</Label>
            <Input
              id="experiment-name"
              value={state.name}
              onChange={(e) => update('name', e.target.value)}
              disabled={isSubmitting}
              required
              maxLength={255}
              placeholder="summarization-quality-2026-q2"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="experiment-description">Description</Label>
            <Textarea
              id="experiment-description"
              value={state.description}
              onChange={(e) => update('description', e.target.value)}
              disabled={isSubmitting}
              rows={3}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="experiment-dataset">Dataset (optional)</Label>
            <Select
              value={state.datasetId === '' ? NO_DATASET : state.datasetId}
              onValueChange={(v) =>
                update('datasetId', v === NO_DATASET ? '' : v)
              }
              disabled={isSubmitting || datasetsQuery.isLoading}
            >
              <SelectTrigger id="experiment-dataset">
                <SelectValue
                  placeholder={
                    datasetsQuery.isLoading
                      ? 'Loading datasets…'
                      : 'No dataset (attach later)'
                  }
                />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={NO_DATASET}>
                  No dataset (attach later)
                </SelectItem>
                {datasetsQuery.data?.data.map((ds) => (
                  <SelectItem key={ds.id} value={ds.id}>
                    {ds.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {datasetsQuery.isError ? (
              <p className="text-xs text-destructive">
                Failed to load datasets. You can still create the
                experiment without one.
              </p>
            ) : null}
          </div>
        </CardContent>
      </Card>

      {error ? (
        <div className="rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
          {error}
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
          {isSubmitting ? 'Creating…' : 'Create experiment'}
        </Button>
      </div>
    </form>
  )
}
