'use client'

import { useState } from 'react'
import Link from 'next/link'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  CheckCircle2,
  XCircle,
  Clock,
  AlertTriangle,
  ChevronDown,
  ExternalLink,
  Code,
  MessageSquare,
  Bug,
  Layers,
} from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { useEvaluatorExecutionDetailQuery } from '../hooks/use-evaluator-executions'
import type {
  EvaluatorExecutionDetail,
  SpanExecutionDetail,
  ExecutionScoreResult,
  ResolvedVariable,
  LLMMessage,
} from '../types'

interface ExecutionDetailDialogProps {
  projectId: string
  projectSlug: string
  evaluatorId: string
  executionId: string | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

function formatDuration(ms: number | undefined): string {
  if (!ms) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`
  return `${(ms / 60000).toFixed(2)}m`
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
      return <Clock className="h-5 w-5 text-blue-500 animate-pulse" />
    case 'skipped':
      return <AlertTriangle className="h-5 w-5 text-yellow-500" />
    default:
      return <Clock className="h-5 w-5 text-muted-foreground" />
  }
}

function CodeBlock({ content, language = 'json' }: { content: string; language?: string }) {
  return (
    <pre className="bg-muted rounded-md p-3 overflow-x-auto text-xs font-mono whitespace-pre-wrap break-all">
      <code className={`language-${language}`}>{content}</code>
    </pre>
  )
}

function MessageDisplay({ messages }: { messages: LLMMessage[] }) {
  return (
    <div className="space-y-3">
      {messages.map((message, index) => (
        <div key={index} className="rounded-lg border p-3">
          <div className="flex items-center gap-2 mb-2">
            <Badge
              variant={
                message.role === 'system'
                  ? 'secondary'
                  : message.role === 'user'
                    ? 'outline'
                    : 'default'
              }
            >
              {message.role}
            </Badge>
          </div>
          <pre className="text-sm whitespace-pre-wrap break-words font-mono">
            {message.content}
          </pre>
        </div>
      ))}
    </div>
  )
}

function VariableResolutionDisplay({ variables }: { variables: ResolvedVariable[] }) {
  if (variables.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No variables were resolved.</p>
    )
  }

  return (
    <div className="space-y-2">
      {variables.map((variable, index) => (
        <div
          key={index}
          className="flex items-start justify-between p-3 rounded-lg bg-muted/50 border"
        >
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <code className="text-sm font-semibold text-primary">
                {`{{${variable.variable_name}}}`}
              </code>
              <span className="text-xs text-muted-foreground">←</span>
              <span className="text-xs text-muted-foreground">
                {variable.source}
                {variable.json_path && (
                  <code className="ml-1 font-mono">[{variable.json_path}]</code>
                )}
              </span>
            </div>
            <pre className="text-xs font-mono whitespace-pre-wrap break-all bg-background rounded p-2 mt-2">
              {typeof variable.resolved_value === 'string'
                ? variable.resolved_value
                : JSON.stringify(variable.resolved_value, null, 2)}
            </pre>
          </div>
        </div>
      ))}
    </div>
  )
}

function ScoreResultDisplay({ results }: { results: ExecutionScoreResult[] }) {
  if (results.length === 0) {
    return <p className="text-sm text-muted-foreground">No scores recorded.</p>
  }

  return (
    <div className="space-y-2">
      {results.map((result, index) => (
        <div key={index} className="p-3 rounded-lg border">
          <div className="flex items-center justify-between mb-2">
            <span className="font-medium">{result.score_name}</span>
            <Badge variant="outline" className="font-mono">
              {typeof result.value === 'number'
                ? result.value.toFixed(3)
                : String(result.value)}
            </Badge>
          </div>
          {result.reasoning && (
            <p className="text-sm text-muted-foreground mt-1">{result.reasoning}</p>
          )}
          {result.confidence !== undefined && (
            <p className="text-xs text-muted-foreground mt-1">
              Confidence: {(result.confidence * 100).toFixed(1)}%
            </p>
          )}
        </div>
      ))}
    </div>
  )
}

function SpanDetailCard({
  span,
  projectSlug,
}: {
  span: SpanExecutionDetail
  projectSlug: string
}) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <Card className="mb-2">
        <CollapsibleTrigger className="w-full">
          <CardHeader className="py-3 px-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <StatusIcon status={span.status} />
                <div className="text-left">
                  <p className="font-medium text-sm">{span.span_name}</p>
                  <p className="text-xs text-muted-foreground font-mono">
                    {span.span_id.substring(0, 12)}...
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                {span.score_results.length > 0 && (
                  <Badge variant="secondary" className="font-mono text-xs">
                    {span.score_results[0].value}
                  </Badge>
                )}
                {span.latency_ms && (
                  <span className="text-xs text-muted-foreground">
                    {formatDuration(span.latency_ms)}
                  </span>
                )}
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
          <CardContent className="pt-0 pb-4 px-4 space-y-4">
            {/* Link to trace */}
            <div className="flex items-center gap-2 text-sm">
              <Link
                href={`/projects/${projectSlug}/traces/${span.trace_id}`}
                className="flex items-center gap-1 text-primary hover:underline"
              >
                View Trace
                <ExternalLink className="h-3 w-3" />
              </Link>
            </div>

            {/* Score Results */}
            {span.score_results.length > 0 && (
              <div>
                <h4 className="text-sm font-medium mb-2">Score Results</h4>
                <ScoreResultDisplay results={span.score_results} />
              </div>
            )}

            {/* Variables Resolved */}
            {span.variables_resolved.length > 0 && (
              <div>
                <h4 className="text-sm font-medium mb-2">Variables Resolved</h4>
                <VariableResolutionDisplay variables={span.variables_resolved} />
              </div>
            )}

            {/* Prompt Sent */}
            {span.prompt_sent && span.prompt_sent.length > 0 && (
              <div>
                <h4 className="text-sm font-medium mb-2">Prompt Sent</h4>
                <MessageDisplay messages={span.prompt_sent} />
              </div>
            )}

            {/* LLM Response */}
            {span.llm_response_raw && (
              <div>
                <h4 className="text-sm font-medium mb-2">LLM Response (Raw)</h4>
                <CodeBlock content={span.llm_response_raw} />
              </div>
            )}
            {span.llm_response_parsed && (
              <div>
                <h4 className="text-sm font-medium mb-2">LLM Response (Parsed)</h4>
                <CodeBlock content={JSON.stringify(span.llm_response_parsed, null, 2)} />
              </div>
            )}

            {/* Error Details */}
            {span.error_message && (
              <div className="rounded-lg border border-destructive/50 bg-destructive/5 p-3">
                <h4 className="text-sm font-medium text-destructive mb-2">Error</h4>
                <p className="text-sm">{span.error_message}</p>
                {span.error_stack && (
                  <pre className="text-xs font-mono mt-2 whitespace-pre-wrap text-muted-foreground">
                    {span.error_stack}
                  </pre>
                )}
              </div>
            )}
          </CardContent>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  )
}

function ExecutionSummary({ execution }: { execution: EvaluatorExecutionDetail }) {
  const avgScore =
    execution.spans.length > 0
      ? execution.spans
          .filter((s) => s.score_results.length > 0)
          .reduce((sum, s) => {
            const numericScore = s.score_results.find(
              (r) => typeof r.value === 'number'
            )
            return sum + (numericScore?.value as number || 0)
          }, 0) /
        Math.max(
          execution.spans.filter((s) => s.score_results.length > 0).length,
          1
        )
      : null

  return (
    <div className="flex items-center gap-4 p-4 rounded-lg bg-muted/50 mb-4">
      <StatusIcon status={execution.status} />
      <div className="flex-1 grid grid-cols-2 md:grid-cols-4 gap-4">
        <div>
          <p className="text-xs text-muted-foreground">Status</p>
          <p className="font-medium capitalize">{execution.status}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Duration</p>
          <p className="font-medium font-mono">{formatDuration(execution.duration_ms)}</p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Spans Scored</p>
          <p className="font-medium">
            {execution.spans_scored} / {execution.spans_matched}
          </p>
        </div>
        {avgScore !== null && (
          <div>
            <p className="text-xs text-muted-foreground">Avg Score</p>
            <p className="font-medium font-mono">{avgScore.toFixed(3)}</p>
          </div>
        )}
      </div>
    </div>
  )
}

function LoadingSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-24 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-64 w-full" />
    </div>
  )
}

export function ExecutionDetailDialog({
  projectId,
  projectSlug,
  evaluatorId,
  executionId,
  open,
  onOpenChange,
}: ExecutionDetailDialogProps) {
  const { data: execution, isLoading, error } = useEvaluatorExecutionDetailQuery(
    projectId,
    evaluatorId,
    executionId ?? '',
    { enabled: open && !!executionId }
  )

  // Find the first span with prompt/response for the Prompt and Response tabs
  const spanWithPrompt = execution?.spans.find((s) => s.prompt_sent && s.prompt_sent.length > 0)
  const spanWithResponse = execution?.spans.find((s) => s.llm_response_raw || s.llm_response_parsed)
  const spansWithErrors = execution?.spans.filter((s) => s.error_message) ?? []

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            Execution Detail
            {executionId && (
              <code className="text-xs font-normal text-muted-foreground">
                {executionId.substring(0, 8)}...
              </code>
            )}
          </DialogTitle>
        </DialogHeader>

        <ScrollArea className="flex-1 -mx-6 px-6">
          {isLoading ? (
            <LoadingSkeleton />
          ) : error ? (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <XCircle className="h-8 w-8 text-destructive mb-2" />
              <p className="text-sm text-destructive">Failed to load execution details</p>
              <p className="text-xs text-muted-foreground mt-1">
                {error instanceof Error ? error.message : 'Unknown error'}
              </p>
            </div>
          ) : execution ? (
            <>
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
                    {spansWithErrors.length > 0 && (
                      <Badge variant="destructive" className="ml-1 h-5 px-1">
                        {spansWithErrors.length}
                      </Badge>
                    )}
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="spans" className="mt-4">
                  {execution.spans.length === 0 ? (
                    <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
                      <Layers className="h-8 w-8 mb-2 opacity-50" />
                      <p className="text-sm">No spans were evaluated</p>
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {execution.spans.map((span) => (
                        <SpanDetailCard
                          key={span.span_id}
                          span={span}
                          projectSlug={projectSlug}
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
                          <CardTitle className="text-sm">Messages Sent to LLM</CardTitle>
                        </CardHeader>
                        <CardContent>
                          <MessageDisplay messages={spanWithPrompt.prompt_sent} />
                        </CardContent>
                      </Card>

                      {spanWithPrompt.variables_resolved.length > 0 && (
                        <Card>
                          <CardHeader className="py-3">
                            <CardTitle className="text-sm">Variable Resolution</CardTitle>
                          </CardHeader>
                          <CardContent>
                            <VariableResolutionDisplay
                              variables={spanWithPrompt.variables_resolved}
                            />
                          </CardContent>
                        </Card>
                      )}
                    </div>
                  ) : (
                    <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
                      <MessageSquare className="h-8 w-8 mb-2 opacity-50" />
                      <p className="text-sm">No prompt data available</p>
                      <p className="text-xs mt-1">
                        This execution may not be an LLM-based scorer
                      </p>
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="response" className="mt-4">
                  {spanWithResponse ? (
                    <div className="space-y-4">
                      {spanWithResponse.llm_response_raw && (
                        <Card>
                          <CardHeader className="py-3">
                            <CardTitle className="text-sm">Raw LLM Response</CardTitle>
                          </CardHeader>
                          <CardContent>
                            <CodeBlock content={spanWithResponse.llm_response_raw} />
                          </CardContent>
                        </Card>
                      )}

                      {spanWithResponse.llm_response_parsed && (
                        <Card>
                          <CardHeader className="py-3">
                            <CardTitle className="text-sm">Parsed Response</CardTitle>
                          </CardHeader>
                          <CardContent>
                            <CodeBlock
                              content={JSON.stringify(
                                spanWithResponse.llm_response_parsed,
                                null,
                                2
                              )}
                            />
                          </CardContent>
                        </Card>
                      )}

                      {spanWithResponse.score_results.length > 0 && (
                        <Card>
                          <CardHeader className="py-3">
                            <CardTitle className="text-sm">Extracted Scores</CardTitle>
                          </CardHeader>
                          <CardContent>
                            <ScoreResultDisplay results={spanWithResponse.score_results} />
                          </CardContent>
                        </Card>
                      )}
                    </div>
                  ) : (
                    <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
                      <Code className="h-8 w-8 mb-2 opacity-50" />
                      <p className="text-sm">No response data available</p>
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="debug" className="mt-4">
                  <div className="space-y-4">
                    {/* Execution metadata */}
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
                            <code className="font-mono text-xs">{execution.evaluator_id}</code>
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
                          {execution.started_at && (
                            <div>
                              <p className="text-muted-foreground">Started</p>
                              <p>{new Date(execution.started_at).toLocaleString()}</p>
                            </div>
                          )}
                          {execution.completed_at && (
                            <div>
                              <p className="text-muted-foreground">Completed</p>
                              <p>{new Date(execution.completed_at).toLocaleString()}</p>
                            </div>
                          )}
                        </div>
                      </CardContent>
                    </Card>

                    {/* Evaluator snapshot */}
                    {execution.evaluator_snapshot && (
                      <Card>
                        <CardHeader className="py-3">
                          <CardTitle className="text-sm">Evaluator Configuration (at execution time)</CardTitle>
                        </CardHeader>
                        <CardContent>
                          <CodeBlock
                            content={JSON.stringify(execution.evaluator_snapshot, null, 2)}
                          />
                        </CardContent>
                      </Card>
                    )}

                    {/* Errors */}
                    {spansWithErrors.length > 0 && (
                      <Card>
                        <CardHeader className="py-3">
                          <CardTitle className="text-sm flex items-center gap-2">
                            <AlertTriangle className="h-4 w-4 text-destructive" />
                            Errors ({spansWithErrors.length})
                          </CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-3">
                          {spansWithErrors.map((span) => (
                            <div
                              key={span.span_id}
                              className="rounded-lg border border-destructive/50 bg-destructive/5 p-3"
                            >
                              <div className="flex items-center gap-2 mb-1">
                                <code className="text-xs font-mono text-muted-foreground">
                                  {span.span_id.substring(0, 12)}...
                                </code>
                                <span className="text-sm font-medium">{span.span_name}</span>
                              </div>
                              <p className="text-sm text-destructive">{span.error_message}</p>
                              {span.error_stack && (
                                <Collapsible>
                                  <CollapsibleTrigger className="text-xs text-muted-foreground flex items-center gap-1 mt-2 hover:text-foreground">
                                    <ChevronDown className="h-3 w-3" />
                                    Show stack trace
                                  </CollapsibleTrigger>
                                  <CollapsibleContent>
                                    <pre className="text-xs font-mono mt-2 whitespace-pre-wrap text-muted-foreground bg-background rounded p-2">
                                      {span.error_stack}
                                    </pre>
                                  </CollapsibleContent>
                                </Collapsible>
                              )}
                            </div>
                          ))}
                        </CardContent>
                      </Card>
                    )}

                    {/* Execution-level error */}
                    {execution.error_message && (
                      <Card>
                        <CardHeader className="py-3">
                          <CardTitle className="text-sm flex items-center gap-2">
                            <XCircle className="h-4 w-4 text-destructive" />
                            Execution Error
                          </CardTitle>
                        </CardHeader>
                        <CardContent>
                          <p className="text-sm text-destructive">{execution.error_message}</p>
                        </CardContent>
                      </Card>
                    )}

                    {spansWithErrors.length === 0 && !execution.error_message && (
                      <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
                        <CheckCircle2 className="h-8 w-8 mb-2 text-green-500" />
                        <p className="text-sm">No errors in this execution</p>
                      </div>
                    )}
                  </div>
                </TabsContent>
              </Tabs>
            </>
          ) : (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <Clock className="h-8 w-8 mb-2 opacity-50" />
              <p className="text-sm">Select an execution to view details</p>
            </div>
          )}
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}
