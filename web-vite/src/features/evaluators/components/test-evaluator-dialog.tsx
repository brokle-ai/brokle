import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { CheckCircle2, Loader2, Play, XCircle } from 'lucide-react'
import { toast } from 'sonner'
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
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ScrollArea } from '@/components/ui/scroll-area'
import { BrokleError } from '@/lib/api/errors'
import { testEvaluator } from '../api/queries'
import type {
  TestEvaluatorRequest,
  TestEvaluatorResponse,
  TestExecution,
  TestScoreResult,
} from '../api/types'

interface TestEvaluatorDialogProps {
  projectId: string
  evaluatorId: string
  evaluatorName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

function formatScoreValue(value: TestScoreResult['value']): string {
  if (typeof value === 'number') return value.toFixed(2)
  if (typeof value === 'boolean') return value ? 'true' : 'false'
  return String(value)
}

function StatusBadge({ status }: { status: TestExecution['status'] }) {
  if (status === 'success') return <Badge variant="secondary">Success</Badge>
  if (status === 'failed') return <Badge variant="destructive">Failed</Badge>
  if (status === 'skipped') return <Badge variant="outline">Skipped</Badge>
  return <Badge variant="outline">Filtered</Badge>
}

function errorMessage(err: unknown, fallback: string): string {
  if (err instanceof BrokleError) return err.message
  if (err instanceof Error) return err.message
  return fallback
}

// Minimal test-evaluator dialog. Captures sample limit + time range,
// POSTs to /evaluators/{id}/test, and renders a summary strip + a
// scrollable execution list with per-span scores and variables.
export function TestEvaluatorDialog({
  projectId,
  evaluatorId,
  evaluatorName,
  open,
  onOpenChange,
}: TestEvaluatorDialogProps) {
  const [limit, setLimit] = useState(5)
  const [timeRange, setTimeRange] = useState<'1h' | '24h' | '7d'>('24h')
  const [traceId, setTraceId] = useState('')
  const [spanId, setSpanId] = useState('')
  const [result, setResult] = useState<TestEvaluatorResponse | null>(null)

  const mutation = useMutation({
    mutationFn: async () => {
      const payload: TestEvaluatorRequest = {}
      // Explicit span wins; trace wins over time-range sampling; else
      // fall back to a backfill-style sample.
      if (spanId.trim()) payload.span_id = spanId.trim()
      else if (traceId.trim()) payload.trace_id = traceId.trim()
      else {
        payload.time_range = timeRange
        payload.limit = limit
      }
      return testEvaluator(projectId, evaluatorId, payload)
    },
    onSuccess: (data) => setResult(data),
    onError: (err) => {
      toast.error('Failed to run test', {
        description: errorMessage(err, 'Could not run test evaluation.'),
      })
    },
  })

  const handleClose = (next: boolean) => {
    if (mutation.isPending) return
    if (!next) {
      setResult(null)
      mutation.reset()
    }
    onOpenChange(next)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="flex max-h-[85vh] flex-col sm:max-w-[720px]">
        <DialogHeader>
          <DialogTitle>Test evaluator: {evaluatorName}</DialogTitle>
          <DialogDescription>
            Run the evaluator against sample spans without persisting scores.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-hidden">
          <ScrollArea className="h-full max-h-[calc(85vh-220px)]">
            <div className="space-y-6 py-2 pr-4">
              {!result && (
                <div className="grid gap-4">
                  <div className="grid gap-4 sm:grid-cols-2">
                    <div className="space-y-2">
                      <Label htmlFor="test-limit">Sample size</Label>
                      <Input
                        id="test-limit"
                        type="number"
                        min={1}
                        max={20}
                        value={limit}
                        onChange={(e) =>
                          setLimit(Math.min(20, Math.max(1, +e.target.value)))
                        }
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="test-range">Time range</Label>
                      <Select
                        value={timeRange}
                        onValueChange={(v) =>
                          setTimeRange(v as '1h' | '24h' | '7d')
                        }
                      >
                        <SelectTrigger id="test-range">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="1h">Last hour</SelectItem>
                          <SelectItem value="24h">Last 24 hours</SelectItem>
                          <SelectItem value="7d">Last 7 days</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  <div className="grid gap-4 sm:grid-cols-2">
                    <div className="space-y-2">
                      <Label htmlFor="test-trace">Trace ID (optional)</Label>
                      <Input
                        id="test-trace"
                        value={traceId}
                        onChange={(e) => setTraceId(e.target.value)}
                        placeholder="Override sampling — test this trace"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="test-span">Span ID (optional)</Label>
                      <Input
                        id="test-span"
                        value={spanId}
                        onChange={(e) => setSpanId(e.target.value)}
                        placeholder="Override — test only this span"
                      />
                    </div>
                  </div>

                  <p className="text-xs text-muted-foreground">
                    Span ID wins, then trace ID, then the time-range sample.
                  </p>
                </div>
              )}

              {result && (
                <TestResults
                  result={result}
                  onRerun={() => {
                    setResult(null)
                    mutation.reset()
                  }}
                />
              )}
            </div>
          </ScrollArea>
        </div>

        <DialogFooter>
          {!result && (
            <>
              <Button
                variant="outline"
                onClick={() => handleClose(false)}
                disabled={mutation.isPending}
              >
                Close
              </Button>
              <Button
                onClick={() => mutation.mutate()}
                disabled={mutation.isPending}
              >
                {mutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Running…
                  </>
                ) : (
                  <>
                    <Play className="mr-2 h-4 w-4" />
                    Run test
                  </>
                )}
              </Button>
            </>
          )}
          {result && (
            <Button onClick={() => handleClose(false)}>Done</Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function TestResults({
  result,
  onRerun,
}: {
  result: TestEvaluatorResponse
  onRerun: () => void
}) {
  const { summary, executions } = result

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <SummaryStat label="Total" value={summary.total_spans} />
        <SummaryStat label="Matched" value={summary.matched_spans} />
        <SummaryStat
          label="Success"
          value={summary.success_count}
          tone="ok"
        />
        <SummaryStat
          label="Failures"
          value={summary.failure_count}
          tone={summary.failure_count > 0 ? 'err' : 'muted'}
        />
      </div>

      {summary.average_score !== undefined && (
        <div className="rounded-md border bg-muted/30 p-3 text-sm">
          <strong>Average score:</strong>{' '}
          {summary.average_score.toFixed(3)}
          {summary.average_latency_ms !== undefined && (
            <span className="ml-4 text-muted-foreground">
              avg latency: {summary.average_latency_ms.toFixed(0)}ms
            </span>
          )}
        </div>
      )}

      <div className="space-y-3">
        <h4 className="text-sm font-semibold">Executions</h4>
        {executions.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No spans were scored — none matched the evaluator filter.
          </p>
        ) : (
          <ul className="space-y-3">
            {executions.map((exec, idx) => (
              <li
                key={`${exec.span_id}-${idx}`}
                className="space-y-2 rounded-md border p-3"
              >
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    {exec.status === 'success' ? (
                      <CheckCircle2 className="h-4 w-4 text-green-600" />
                    ) : exec.status === 'failed' ? (
                      <XCircle className="h-4 w-4 text-red-600" />
                    ) : null}
                    <code className="font-mono text-xs">
                      {exec.span_name || exec.span_id}
                    </code>
                  </div>
                  <StatusBadge status={exec.status} />
                </div>

                {exec.score_results.length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    {exec.score_results.map((s, sIdx) => (
                      <Badge key={sIdx} variant="outline">
                        {s.score_name}: {formatScoreValue(s.value)}
                      </Badge>
                    ))}
                  </div>
                )}

                {exec.error_message && (
                  <p className="text-xs text-destructive">
                    {exec.error_message}
                  </p>
                )}

                {exec.latency_ms !== undefined && (
                  <p className="text-xs text-muted-foreground">
                    {exec.latency_ms}ms
                  </p>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      <Button variant="outline" size="sm" onClick={onRerun}>
        Run again
      </Button>
    </div>
  )
}

function SummaryStat({
  label,
  value,
  tone,
}: {
  label: string
  value: number
  tone?: 'ok' | 'err' | 'muted'
}) {
  const toneClass =
    tone === 'ok'
      ? 'text-green-600 dark:text-green-400'
      : tone === 'err'
        ? 'text-destructive'
        : 'text-foreground'
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className={`text-lg font-semibold ${toneClass}`}>
        {value.toLocaleString()}
      </p>
    </div>
  )
}
