import { Link } from '@tanstack/react-router'
import { Badge } from '@/components/ui/badge'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import type { ScoreDataType, ScoreListItem, ScoreSource } from '../api/types'

interface ScoreDetailSheetProps {
  score: ScoreListItem | null
  orgId: string
  projectId: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

// Re-implement value formatting locally so the side panel can render
// the raw value ('—' for missing, formatted for present) without
// reaching into the table component's private helper.
function formatValue(score: ScoreListItem): string {
  if (score.type === 'CATEGORICAL') {
    return score.string_value && score.string_value.length > 0
      ? score.string_value
      : '—'
  }
  if (score.type === 'BOOLEAN') {
    if (score.value === undefined || score.value === null) return '—'
    return score.value >= 0.5 ? 'true' : 'false'
  }
  if (score.value === undefined || score.value === null) return '—'
  return Number.isInteger(score.value)
    ? score.value.toLocaleString()
    : score.value.toLocaleString(undefined, { maximumFractionDigits: 4 })
}

function TypeBadge({ type }: { type: ScoreDataType }) {
  const label =
    type === 'NUMERIC'
      ? 'Numeric'
      : type === 'BOOLEAN'
        ? 'Boolean'
        : 'Categorical'
  return <Badge variant={type === 'NUMERIC' ? 'secondary' : 'outline'}>{label}</Badge>
}

function SourceBadge({ source }: { source: ScoreSource }) {
  const label = source === 'llm' ? 'LLM' : source.charAt(0).toUpperCase() + source.slice(1)
  return <Badge variant="secondary">{label}</Badge>
}

function Field({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-1">
      <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {label}
      </div>
      <div className="text-sm">{children}</div>
    </div>
  )
}

export function ScoreDetailSheet({
  score,
  orgId,
  projectId,
  open,
  onOpenChange,
}: ScoreDetailSheetProps) {
  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-full overflow-y-auto sm:max-w-xl">
        {score && (
          <>
            <SheetHeader>
              <SheetTitle className="font-mono text-base">
                {score.name}
              </SheetTitle>
              <SheetDescription>Score record details</SheetDescription>
            </SheetHeader>

            <div className="space-y-6 px-4 pb-6">
              <div className="grid grid-cols-2 gap-4">
                <Field label="Value">
                  <span className="font-mono">{formatValue(score)}</span>
                </Field>
                <Field label="Type">
                  <TypeBadge type={score.type} />
                </Field>
                <Field label="Source">
                  <SourceBadge source={score.source} />
                </Field>
                <Field label="Recorded">
                  <span className="text-muted-foreground">
                    {formatTimestamp(score.timestamp)}
                  </span>
                </Field>
              </div>

              {(score.trace_id || score.span_id) && (
                <div className="space-y-3 border-t pt-4">
                  <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    Linked telemetry
                  </div>
                  {score.trace_id && (
                    <Field label="Trace">
                      <Link
                        to="/o/$orgId/p/$projectId/traces/$traceId"
                        params={{
                          orgId,
                          projectId,
                          traceId: score.trace_id,
                        }}
                        className="block break-all font-mono text-xs text-primary underline-offset-4 hover:underline"
                      >
                        {score.trace_id}
                      </Link>
                    </Field>
                  )}
                  {score.span_id && (
                    <Field label="Span">
                      <span className="block break-all font-mono text-xs text-muted-foreground">
                        {score.span_id}
                      </span>
                    </Field>
                  )}
                </div>
              )}

              {score.reason && (
                <div className="space-y-2 border-t pt-4">
                  <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    Reason
                  </div>
                  <p className="whitespace-pre-wrap rounded-md bg-muted p-3 text-sm">
                    {score.reason}
                  </p>
                </div>
              )}

              {(score.experiment_id || score.experiment_item_id) && (
                <div className="space-y-3 border-t pt-4">
                  <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    Experiment
                  </div>
                  {score.experiment_id && (
                    <Field label="Experiment ID">
                      <span className="block break-all font-mono text-xs text-muted-foreground">
                        {score.experiment_id}
                      </span>
                    </Field>
                  )}
                  {score.experiment_item_id && (
                    <Field label="Item ID">
                      <span className="block break-all font-mono text-xs text-muted-foreground">
                        {score.experiment_item_id}
                      </span>
                    </Field>
                  )}
                </div>
              )}

              {score.created_by && (
                <div className="border-t pt-4">
                  <Field label="Created by">
                    <span className="font-mono text-xs text-muted-foreground">
                      {score.created_by}
                    </span>
                  </Field>
                </div>
              )}

              <div className="border-t pt-4">
                <Field label="Score ID">
                  <span className="block break-all font-mono text-xs text-muted-foreground">
                    {score.id}
                  </span>
                </Field>
              </div>
            </div>
          </>
        )}
      </SheetContent>
    </Sheet>
  )
}
