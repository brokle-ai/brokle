import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import {
  AlertTriangle,
  Bug,
  CheckCircle2,
  ChevronDown,
  Clock,
  Code,
  Layers,
  MessageSquare,
  XCircle,
} from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ReadonlyCode } from '@/editors/readonly-code'
import { executionDetailQueryOptions } from '../api/queries'
import type {
  EvaluatorExecutionDetail,
  ExecutionScoreResult,
  LLMMessage,
  ResolvedVariable,
  SpanExecutionDetail,
} from '../api/types'

interface ExecutionDetailViewProps {
  orgId: string
  projectId: string
  evaluatorId: string
  executionId: string
}

function formatDuration(ms: number | undefined | null): string {
  if (ms === undefined || ms === null) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(2)}s`
  return `${(ms / 60_000).toFixed(2)}m`
}

function StatusIcon({ status }: { status: string }) {
  switch (status) {
    case 'completed':
    case 'success':
      return <CheckCircle2 className="h-5 w-5 text-green-500" />
    case 'failed':
      return <XCircle className="h-5 w-5 text-destructive" />
    case 'running':
    case 'pending':
      return <Clock className="h-5 w-5 animate-pulse text-blue-500" />
    case 'skipped':
      return <AlertTriangle className="h-5 w-5 text-yellow-500" />
    default:
      return <Clock className="h-5 w-5 text-muted-foreground" />
  }
}

function MessageDisplay({ messages }: { messages: LLMMessage[] }) {
  return (
    <div className="space-y-3">
      {messages.map((m, idx) => (
        <div key={idx} className="rounded-lg border p-3">
          <div className="mb-2 flex items-center gap-2">
            <Badge
              variant={
                m.role === 'system'
                  ? 'secondary'
                  : m.role === 'user'
                    ? 'outline'
                    : 'default'
              }
            >
              {m.role}
            </Badge>
          </div>
          <pre className="whitespace-pre-wrap break-words font-mono text-sm">
            {m.content}
          </pre>
        </div>
      ))}
    </div>
  )
}

function VariableResolutionDisplay({
  variables,
}: {
  variables: ResolvedVariable[]
}) {
  if (variables.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No variables were resolved.</p>
    )
  }

  return (
    <div className="space-y-2">
      {variables.map((v, idx) => (
        <div
          key={idx}
          className="flex items-start justify-between rounded-lg border bg-muted/50 p-3"
        >
          <div className="min-w-0 flex-1">
            <div className="mb-1 flex items-center gap-2">
              <code className="text-sm font-semibold text-primary">
                {`{{${v.variable_name}}}`}
              </code>
              <span className="text-xs text-muted-foreground">←</span>
              <span className="text-xs text-muted-foreground">
                {v.source}
                {v.json_path ? (
                  <code className="ml-1 font-mono">[{v.json_path}]</code>
                ) : null}
              </span>
            </div>
            <pre className="mt-2 whitespace-pre-wrap break-all rounded bg-background p-2 font-mono text-xs">
              {typeof v.resolved_value === 'string'
                ? v.resolved_value
                : JSON.stringify(v.resolved_value, null, 2)}
            </pre>
          </div>
        </div>
      ))}
    </div>
  )
}

function ScoreResultDisplay({
  results,
}: {
  results: ExecutionScoreResult[]
}) {
  if (results.length === 0) {
    return <p className="text-sm text-muted-foreground">No scores recorded.</p>
  }

  return (
    <div className="space-y-2">
      {results.map((r, idx) => (
        <div key={idx} className="rounded-lg border p-3">
          <div className="mb-2 flex items-center justify-between">
            <span className="font-medium">{r.score_name}</span>
            <Badge variant="outline" className="font-mono">
              {typeof r.value === 'number'
                ? r.value.toFixed(3)
                : String(r.value)}
            </Badge>
          </div>
          {r.reasoning ? (
            <p className="mt-1 text-sm text-muted-foreground">{r.reasoning}</p>
          ) : null}
          {r.confidence !== undefined ? (
            <p className="mt-1 text-xs text-muted-foreground">
              Confidence: {(r.confidence * 100).toFixed(1)}%
            </p>
          ) : null}
        </div>
      ))}
    </div>
  )
}

function SpanDetailCard({
  orgId,
  projectId,
  span,
}: {
  orgId: string
  projectId: string
  span: SpanExecutionDetail
}) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <Card className="mb-2">
        <CollapsibleTrigger className="w-full">
          <CardHeader className="px-4 py-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <StatusIcon status={span.status} />
                <div className="text-left">
                  <p className="text-sm font-medium">{span.span_name}</p>
                  <p className="font-mono text-xs text-muted-foreground">
                    {span.span_id.substring(0, 12)}…
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                {span.score_results.length > 0 ? (
                  <Badge variant="secondary" className="font-mono text-xs">
                    {String(span.score_results[0].value)}
                  </Badge>
                ) : null}
                {span.latency_ms !== undefined ? (
                  <span className="text-xs text-muted-foreground">
                    {formatDuration(span.latency_ms)}
                  </span>
                ) : null}
                <ChevronDown
                  className={`h-4 w-4 transition-transform ${
                    isOpen ? 'rotate-180' : ''
                  }`}
                />
              </div>
            </div>
          </CardHeader>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <CardContent className="space-y-4 px-4 pb-4 pt-0">
            <div className="flex items-center gap-2 text-sm">
              <Link
                to="/o/$orgId/p/$projectId/traces/$traceId"
                params={{ orgId, projectId, traceId: span.trace_id }}
                className="text-primary hover:underline"
              >
                View Trace →
              </Link>
            </div>

            {span.score_results.length > 0 ? (
              <div>
                <h4 className="mb-2 text-sm font-medium">Score Results</h4>
                <ScoreResultDisplay results={span.score_results} />
              </div>
            ) : null}

            {span.variables_resolved.length > 0 ? (
              <div>
                <h4 className="mb-2 text-sm font-medium">Variables Resolved</h4>
                <VariableResolutionDisplay variables={span.variables_resolved} />
              </div>
            ) : null}

            {span.prompt_sent && span.prompt_sent.length > 0 ? (
              <div>
                <h4 className="mb-2 text-sm font-medium">Prompt Sent</h4>
                <MessageDisplay messages={span.prompt_sent} />
              </div>
            ) : null}

            {span.llm_response_raw ? (
              <div>
                <h4 className="mb-2 text-sm font-medium">LLM Response (Raw)</h4>
                <ReadonlyCode code={span.llm_response_raw} language="text" />
              </div>
            ) : null}
            {span.llm_response_parsed ? (
              <div>
                <h4 className="mb-2 text-sm font-medium">
                  LLM Response (Parsed)
                </h4>
                <ReadonlyCode
                  code={JSON.stringify(span.llm_response_parsed, null, 2)}
                  language="json"
                />
              </div>
            ) : null}

            {span.error_message ? (
              <div className="rounded-lg border border-destructive/50 bg-destructive/5 p-3">
                <h4 className="mb-2 text-sm font-medium text-destructive">
                  Error
                </h4>
                <p className="text-sm">{span.error_message}</p>
                {span.error_stack ? (
                  <pre className="mt-2 whitespace-pre-wrap font-mono text-xs text-muted-foreground">
                    {span.error_stack}
                  </pre>
                ) : null}
              </div>
            ) : null}
          </CardContent>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  )
}

function ExecutionSummary({
  execution,
}: {
  execution: EvaluatorExecutionDetail
}) {
  const scoredSpans = execution.spans.filter((s) => s.score_results.length > 0)
  const numericTotals = scoredSpans.reduce<{
    sum: number
    count: number
  }>(
    (acc, s) => {
      const numeric = s.score_results.find(
        (r) => typeof r.value === 'number',
      )
      if (numeric && typeof numeric.value === 'number') {
        return { sum: acc.sum + numeric.value, count: acc.count + 1 }
      }
      return acc
    },
    { sum: 0, count: 0 },
  )
  const avgScore =
    numericTotals.count > 0 ? numericTotals.sum / numericTotals.count : null

  return (
    <div className="mb-4 flex items-center gap-4 rounded-lg bg-muted/50 p-4">
      <StatusIcon status={execution.status} />
      <div className="grid flex-1 grid-cols-2 gap-4 md:grid-cols-4">
        <div>
          <p className="text-xs text-muted-foreground">Status</p>
          <p className="font-medium capitalize">{execution.status}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Duration</p>
          <p className="font-mono font-medium">
            {formatDuration(execution.duration_ms)}
          </p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Spans Scored</p>
          <p className="font-medium">
            {execution.spans_scored} / {execution.spans_matched}
          </p>
        </div>
        {avgScore !== null ? (
          <div>
            <p className="text-xs text-muted-foreground">Avg Score</p>
            <p className="font-mono font-medium">{avgScore.toFixed(3)}</p>
          </div>
        ) : null}
      </div>
    </div>
  )
}

export function ExecutionDetailView({
  orgId,
  projectId,
  evaluatorId,
  executionId,
}: ExecutionDetailViewProps) {
  const { data: execution } = useSuspenseQuery(
    executionDetailQueryOptions(projectId, evaluatorId, executionId),
  )

  const spanWithPrompt = execution.spans.find(
    (s) => s.prompt_sent && s.prompt_sent.length > 0,
  )
  const spanWithResponse = execution.spans.find(
    (s) => s.llm_response_raw || s.llm_response_parsed,
  )
  const spansWithErrors = execution.spans.filter((s) => s.error_message)

  return (
    <main className="mx-auto max-w-6xl space-y-6 p-6">
      <nav className="text-sm text-muted-foreground">
        <Link
          to="/o/$orgId/p/$projectId/evaluators"
          params={{ orgId, projectId }}
          search={{ page: 1, limit: 20, q: undefined }}
          className="hover:text-foreground"
        >
          Evaluators
        </Link>
        <span className="mx-2">/</span>
        <Link
          to="/o/$orgId/p/$projectId/evaluators/$evaluatorId"
          params={{ orgId, projectId, evaluatorId }}
          search={{ page: 1, limit: 20, q: undefined }}
          className="hover:text-foreground"
        >
          {evaluatorId.substring(0, 8)}…
        </Link>
        <span className="mx-2">/</span>
        <span className="text-foreground">
          Execution {executionId.substring(0, 8)}…
        </span>
      </nav>

      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-2xl font-semibold">
          Execution Detail
          <code className="font-mono text-xs font-normal text-muted-foreground">
            {executionId}
          </code>
        </h1>
      </div>

      <ExecutionSummary execution={execution} />

      <Tabs defaultValue="spans" className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="spans" className="gap-1">
            <Layers className="h-4 w-4" />
            Spans ({execution.spans.length})
          </TabsTrigger>
          <TabsTrigger value="prompt" className="gap-1">
            <MessageSquare className="h-4 w-4" />
            Prompt
          </TabsTrigger>
          <TabsTrigger value="response" className="gap-1">
            <Code className="h-4 w-4" />
            Response
          </TabsTrigger>
          <TabsTrigger value="debug" className="gap-1">
            <Bug className="h-4 w-4" />
            Debug
            {spansWithErrors.length > 0 ? (
              <Badge variant="destructive" className="ml-1 h-5 px-1">
                {spansWithErrors.length}
              </Badge>
            ) : null}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="spans" className="mt-4">
          {execution.spans.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <Layers className="mb-2 h-8 w-8 opacity-50" />
              <p className="text-sm">No spans were evaluated</p>
            </div>
          ) : (
            <div className="space-y-2">
              {execution.spans.map((s) => (
                <SpanDetailCard
                  key={s.span_id}
                  orgId={orgId}
                  projectId={projectId}
                  span={s}
                />
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="prompt" className="mt-4">
          {spanWithPrompt?.prompt_sent ? (
            <div className="space-y-4">
              <Card>
                <CardHeader className="py-3">
                  <CardTitle className="text-sm">
                    Messages Sent to LLM
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <MessageDisplay messages={spanWithPrompt.prompt_sent} />
                </CardContent>
              </Card>

              {spanWithPrompt.variables_resolved.length > 0 ? (
                <Card>
                  <CardHeader className="py-3">
                    <CardTitle className="text-sm">
                      Variable Resolution
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <VariableResolutionDisplay
                      variables={spanWithPrompt.variables_resolved}
                    />
                  </CardContent>
                </Card>
              ) : null}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <MessageSquare className="mb-2 h-8 w-8 opacity-50" />
              <p className="text-sm">No prompt data available</p>
              <p className="mt-1 text-xs">
                This execution may not be an LLM-based scorer.
              </p>
            </div>
          )}
        </TabsContent>

        <TabsContent value="response" className="mt-4">
          {spanWithResponse ? (
            <div className="space-y-4">
              {spanWithResponse.llm_response_raw ? (
                <Card>
                  <CardHeader className="py-3">
                    <CardTitle className="text-sm">Raw LLM Response</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <ReadonlyCode
                      code={spanWithResponse.llm_response_raw}
                      language="text"
                    />
                  </CardContent>
                </Card>
              ) : null}

              {spanWithResponse.llm_response_parsed ? (
                <Card>
                  <CardHeader className="py-3">
                    <CardTitle className="text-sm">Parsed Response</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <ReadonlyCode
                      code={JSON.stringify(
                        spanWithResponse.llm_response_parsed,
                        null,
                        2,
                      )}
                      language="json"
                    />
                  </CardContent>
                </Card>
              ) : null}

              {spanWithResponse.score_results.length > 0 ? (
                <Card>
                  <CardHeader className="py-3">
                    <CardTitle className="text-sm">Extracted Scores</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <ScoreResultDisplay
                      results={spanWithResponse.score_results}
                    />
                  </CardContent>
                </Card>
              ) : null}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <Code className="mb-2 h-8 w-8 opacity-50" />
              <p className="text-sm">No response data available</p>
            </div>
          )}
        </TabsContent>

        <TabsContent value="debug" className="mt-4">
          <div className="space-y-4">
            <Card>
              <CardHeader className="py-3">
                <CardTitle className="text-sm">Execution Info</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <p className="text-muted-foreground">Execution ID</p>
                    <code className="font-mono text-xs">{execution.id}</code>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Evaluator ID</p>
                    <code className="font-mono text-xs">
                      {execution.evaluator_id}
                    </code>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Trigger Type</p>
                    <p className="capitalize">{execution.trigger_type}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Created</p>
                    <p>
                      {formatDistanceToNow(new Date(execution.created_at), {
                        addSuffix: true,
                      })}
                    </p>
                  </div>
                  {execution.started_at ? (
                    <div>
                      <p className="text-muted-foreground">Started</p>
                      <p>{new Date(execution.started_at).toLocaleString()}</p>
                    </div>
                  ) : null}
                  {execution.completed_at ? (
                    <div>
                      <p className="text-muted-foreground">Completed</p>
                      <p>
                        {new Date(execution.completed_at).toLocaleString()}
                      </p>
                    </div>
                  ) : null}
                </div>
              </CardContent>
            </Card>

            {execution.evaluator_snapshot ? (
              <Card>
                <CardHeader className="py-3">
                  <CardTitle className="text-sm">
                    Evaluator Configuration (at execution time)
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <ReadonlyCode
                    code={JSON.stringify(
                      execution.evaluator_snapshot,
                      null,
                      2,
                    )}
                    language="json"
                  />
                </CardContent>
              </Card>
            ) : null}

            {spansWithErrors.length > 0 ? (
              <Card>
                <CardHeader className="py-3">
                  <CardTitle className="flex items-center gap-2 text-sm">
                    <AlertTriangle className="h-4 w-4 text-destructive" />
                    Errors ({spansWithErrors.length})
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  {spansWithErrors.map((s) => (
                    <div
                      key={s.span_id}
                      className="rounded-lg border border-destructive/50 bg-destructive/5 p-3"
                    >
                      <div className="mb-1 flex items-center gap-2">
                        <code className="font-mono text-xs text-muted-foreground">
                          {s.span_id.substring(0, 12)}…
                        </code>
                        <span className="text-sm font-medium">
                          {s.span_name}
                        </span>
                      </div>
                      <p className="text-sm text-destructive">
                        {s.error_message}
                      </p>
                      {s.error_stack ? (
                        <Collapsible>
                          <CollapsibleTrigger className="mt-2 flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground">
                            <ChevronDown className="h-3 w-3" />
                            Show stack trace
                          </CollapsibleTrigger>
                          <CollapsibleContent>
                            <pre className="mt-2 whitespace-pre-wrap rounded bg-background p-2 font-mono text-xs text-muted-foreground">
                              {s.error_stack}
                            </pre>
                          </CollapsibleContent>
                        </Collapsible>
                      ) : null}
                    </div>
                  ))}
                </CardContent>
              </Card>
            ) : null}

            {execution.error_message ? (
              <Card>
                <CardHeader className="py-3">
                  <CardTitle className="flex items-center gap-2 text-sm">
                    <XCircle className="h-4 w-4 text-destructive" />
                    Execution Error
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-destructive">
                    {execution.error_message}
                  </p>
                </CardContent>
              </Card>
            ) : null}

            {spansWithErrors.length === 0 && !execution.error_message ? (
              <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
                <CheckCircle2 className="mb-2 h-8 w-8 text-green-500" />
                <p className="text-sm">No errors in this execution</p>
              </div>
            ) : null}
          </div>
        </TabsContent>
      </Tabs>
    </main>
  )
}
