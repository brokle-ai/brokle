'use client'

import { useState, useCallback } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Slider } from '@/components/ui/slider'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Play,
  Loader2,
  CheckCircle2,
  XCircle,
  AlertCircle,
  ChevronDown,
  ChevronRight,
  Target,
  Zap,
  TriangleAlert,
  Clock,
  FileCode,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useTestRule, type TestOptions } from '../hooks/use-test-rule'
import type { EvaluationRule, TestSummary, TestExecution, RulePreview } from '../types'

interface TestRuleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  rule: EvaluationRule
  projectId: string
}

/**
 * Dialog for testing evaluation rules before activation.
 *
 * Features:
 * - Configure sample size (1-20 spans)
 * - Time range filter
 * - Real-time progress tracking
 * - Detailed results display with per-span breakdown
 */
export function TestRuleDialog({
  open,
  onOpenChange,
  rule,
  projectId,
}: TestRuleDialogProps) {
  const [sampleLimit, setSampleLimit] = useState(5)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [timeRange, setTimeRange] = useState<'1h' | '24h' | '7d'>('24h')

  const { result, startTest, resetTest, isRunning } = useTestRule(
    projectId,
    rule.id,
    {
      defaultSampleLimit: sampleLimit,
      defaultTimeRange: timeRange,
    }
  )

  const handleRunTest = useCallback(async () => {
    const options: TestOptions = {
      sampleLimit,
      timeRange,
    }
    await startTest(options)
  }, [sampleLimit, timeRange, startTest])

  const handleClose = useCallback(() => {
    if (!isRunning) {
      resetTest()
      onOpenChange(false)
    }
  }, [isRunning, resetTest, onOpenChange])

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[700px] max-h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Zap className="h-5 w-5 text-primary" />
            Test Rule: {rule.name}
          </DialogTitle>
          <DialogDescription>
            Run a test evaluation against sample spans to verify the rule works correctly before activation.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-hidden">
          <ScrollArea className="h-full max-h-[calc(85vh-200px)]">
            <div className="space-y-6 py-4 pr-4">
              {/* Configuration Section */}
              {result.status === 'idle' && (
                <TestConfiguration
                  sampleLimit={sampleLimit}
                  onSampleLimitChange={setSampleLimit}
                  timeRange={timeRange}
                  onTimeRangeChange={setTimeRange}
                  showAdvanced={showAdvanced}
                  onShowAdvancedChange={setShowAdvanced}
                  rule={rule}
                />
              )}

              {/* Running State */}
              {result.status === 'running' && <TestRunning />}

              {/* Results */}
              {result.status === 'completed' && result.summary && (
                <TestResults
                  summary={result.summary}
                  executions={result.executions ?? []}
                  rulePreview={result.rulePreview}
                />
              )}

              {/* Error State */}
              {result.status === 'failed' && <TestError error={result.error} />}
            </div>
          </ScrollArea>
        </div>

        <DialogFooter>
          {result.status === 'idle' && (
            <>
              <Button variant="outline" onClick={handleClose}>
                Cancel
              </Button>
              <Button onClick={handleRunTest}>
                <Play className="mr-2 h-4 w-4" />
                Run Test
              </Button>
            </>
          )}

          {isRunning && (
            <Button variant="outline" disabled>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Running...
            </Button>
          )}

          {(result.status === 'completed' || result.status === 'failed') && (
            <>
              <Button variant="outline" onClick={resetTest}>
                Run Another Test
              </Button>
              <Button onClick={handleClose}>Close</Button>
            </>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

/**
 * Configuration section for test parameters
 */
function TestConfiguration({
  sampleLimit,
  onSampleLimitChange,
  timeRange,
  onTimeRangeChange,
  showAdvanced,
  onShowAdvancedChange,
  rule,
}: {
  sampleLimit: number
  onSampleLimitChange: (value: number) => void
  timeRange: '1h' | '24h' | '7d'
  onTimeRangeChange: (value: '1h' | '24h' | '7d') => void
  showAdvanced: boolean
  onShowAdvancedChange: (value: boolean) => void
  rule: EvaluationRule
}) {
  return (
    <div className="space-y-4">
      {/* Rule Summary */}
      <div className="rounded-lg border bg-muted/50 p-4 space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium">Rule Configuration</span>
          <Badge variant="outline" className="text-xs">
            {rule.scorer_type.toUpperCase()}
          </Badge>
        </div>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">Sampling Rate:</span>
            <span className="ml-2 font-medium">
              {Math.round(rule.sampling_rate * 100)}%
            </span>
          </div>
          <div>
            <span className="text-muted-foreground">Target:</span>
            <span className="ml-2 font-medium">{rule.target_scope}</span>
          </div>
          {rule.span_names && rule.span_names.length > 0 && (
            <div className="col-span-2">
              <span className="text-muted-foreground">Span Names:</span>
              <span className="ml-2 font-medium">
                {rule.span_names.join(', ')}
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Sample Size */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <Label htmlFor="sample-limit">Sample Size</Label>
          <span className="text-sm text-muted-foreground">
            {sampleLimit} spans
          </span>
        </div>
        <Slider
          id="sample-limit"
          value={[sampleLimit]}
          onValueChange={([value]) => onSampleLimitChange(value)}
          min={1}
          max={20}
          step={1}
          className="w-full"
        />
        <p className="text-xs text-muted-foreground">
          Number of matching spans to evaluate. Start with a small sample to
          verify the rule works correctly.
        </p>
      </div>

      {/* Time Range */}
      <div className="space-y-2">
        <Label htmlFor="time-range">Time Range</Label>
        <Select
          value={timeRange}
          onValueChange={(v) => onTimeRangeChange(v as typeof timeRange)}
        >
          <SelectTrigger id="time-range">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1h">Last 1 hour</SelectItem>
            <SelectItem value="24h">Last 24 hours</SelectItem>
            <SelectItem value="7d">Last 7 days</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Advanced Options */}
      <Collapsible open={showAdvanced} onOpenChange={onShowAdvancedChange}>
        <CollapsibleTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="gap-2 text-muted-foreground"
          >
            {showAdvanced ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
            Advanced Options
          </Button>
        </CollapsibleTrigger>
        <CollapsibleContent className="pt-2">
          <div className="rounded-lg border p-4 space-y-3">
            <p className="text-sm text-muted-foreground">
              Additional filtering options will be available in a future update.
            </p>
            {rule.filter && rule.filter.length > 0 && (
              <div>
                <span className="text-sm font-medium">Active Filters:</span>
                <div className="mt-2 space-y-1">
                  {rule.filter.map((f, i) => (
                    <div
                      key={i}
                      className="text-xs text-muted-foreground font-mono bg-muted px-2 py-1 rounded"
                    >
                      {f.field} {f.operator} {JSON.stringify(f.value)}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </CollapsibleContent>
      </Collapsible>
    </div>
  )
}

/**
 * Running state display
 */
function TestRunning() {
  return (
    <div className="flex flex-col items-center justify-center py-8 space-y-4">
      <div className="relative">
        <Loader2 className="h-12 w-12 animate-spin text-primary" />
      </div>
      <div className="text-center space-y-2">
        <p className="font-medium">Running test evaluation...</p>
        <p className="text-sm text-muted-foreground">
          Querying spans and executing scorer
        </p>
      </div>
    </div>
  )
}

/**
 * Test results display
 */
function TestResults({
  summary,
  executions,
  rulePreview,
}: {
  summary: TestSummary
  executions: TestExecution[]
  rulePreview?: RulePreview
}) {
  const [showExecutions, setShowExecutions] = useState(false)
  const successRate =
    summary.evaluated_spans > 0
      ? Math.round((summary.success_count / summary.evaluated_spans) * 100)
      : 0

  const hasSuccess = summary.success_count > 0

  return (
    <div className="space-y-4">
      {/* Success/Failure Banner */}
      {hasSuccess ? (
        <div className="flex items-center gap-3 p-4 rounded-lg bg-green-50 dark:bg-green-950/30 border border-green-200 dark:border-green-800">
          <CheckCircle2 className="h-6 w-6 text-green-600 dark:text-green-400" />
          <div>
            <p className="font-medium text-green-800 dark:text-green-300">
              Test Completed Successfully
            </p>
            <p className="text-sm text-green-600 dark:text-green-400">
              Your rule is ready to be activated.
            </p>
          </div>
        </div>
      ) : (
        <div className="flex items-center gap-3 p-4 rounded-lg bg-yellow-50 dark:bg-yellow-950/30 border border-yellow-200 dark:border-yellow-800">
          <TriangleAlert className="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
          <div>
            <p className="font-medium text-yellow-800 dark:text-yellow-300">
              Test Completed with Issues
            </p>
            <p className="text-sm text-yellow-600 dark:text-yellow-400">
              No spans were successfully scored. Check your rule configuration.
            </p>
          </div>
        </div>
      )}

      {/* Results Summary */}
      <div className="grid grid-cols-4 gap-3">
        <ResultCard
          label="Total Spans"
          value={summary.total_spans}
          icon={<FileCode className="h-4 w-4" />}
        />
        <ResultCard
          label="Matched"
          value={summary.matched_spans}
          icon={<Target className="h-4 w-4" />}
        />
        <ResultCard
          label="Success"
          value={summary.success_count}
          icon={<CheckCircle2 className="h-4 w-4 text-green-500" />}
        />
        <ResultCard
          label="Failed"
          value={summary.failure_count}
          icon={<XCircle className="h-4 w-4 text-red-500" />}
          className={
            summary.failure_count > 0
              ? 'border-red-200 dark:border-red-800'
              : ''
          }
        />
      </div>

      {/* Additional Stats */}
      <div className="rounded-lg border p-4 space-y-3">
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Success Rate</span>
          <span className="font-medium">{successRate}%</span>
        </div>
        {summary.average_latency_ms && (
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Average Latency</span>
            <span className="font-medium">
              {summary.average_latency_ms < 1000
                ? `${Math.round(summary.average_latency_ms)}ms`
                : `${(summary.average_latency_ms / 1000).toFixed(2)}s`}
            </span>
          </div>
        )}
        {summary.average_score !== undefined && summary.average_score !== null && (
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Average Score</span>
            <span className="font-medium">
              {typeof summary.average_score === 'number'
                ? summary.average_score.toFixed(2)
                : summary.average_score}
            </span>
          </div>
        )}
      </div>

      {/* Rule Preview */}
      {rulePreview && (
        <div className="rounded-lg border p-4 space-y-2">
          <p className="text-sm font-medium">Rule Configuration</p>
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div>
              <span className="text-muted-foreground">Scorer Type:</span>
              <span className="ml-2 font-medium">{rulePreview.scorer_type}</span>
            </div>
            {rulePreview.matching_count !== undefined && (
              <div>
                <span className="text-muted-foreground">Est. Matching:</span>
                <span className="ml-2 font-medium">
                  {rulePreview.matching_count} spans
                </span>
              </div>
            )}
          </div>
          {rulePreview.filter_description && (
            <div className="text-xs text-muted-foreground font-mono bg-muted p-2 rounded">
              {rulePreview.filter_description}
            </div>
          )}
        </div>
      )}

      {/* Executions Details */}
      {executions.length > 0 && (
        <Collapsible open={showExecutions} onOpenChange={setShowExecutions}>
          <CollapsibleTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="gap-2 text-muted-foreground w-full justify-start"
            >
              {showExecutions ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronRight className="h-4 w-4" />
              )}
              View {executions.length} Execution
              {executions.length !== 1 ? 's' : ''} Details
            </Button>
          </CollapsibleTrigger>
          <CollapsibleContent className="pt-2">
            <div className="space-y-2">
              {executions.map((exec, i) => (
                <ExecutionDetail key={i} execution={exec} />
              ))}
            </div>
          </CollapsibleContent>
        </Collapsible>
      )}

      {/* Warnings */}
      {summary.matched_spans === 0 && (
        <div className="flex items-start gap-3 p-4 rounded-lg bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-800">
          <AlertCircle className="h-5 w-5 text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5" />
          <div>
            <p className="font-medium text-blue-800 dark:text-blue-300">
              No spans matched your filters
            </p>
            <p className="text-sm text-blue-600 dark:text-blue-400">
              Try expanding your time range or adjusting the filter criteria.
            </p>
          </div>
        </div>
      )}
    </div>
  )
}

/**
 * Execution detail row
 */
function ExecutionDetail({ execution }: { execution: TestExecution }) {
  const [expanded, setExpanded] = useState(false)

  const statusIcon =
    execution.status === 'success' ? (
      <CheckCircle2 className="h-4 w-4 text-green-500" />
    ) : execution.status === 'failed' ? (
      <XCircle className="h-4 w-4 text-red-500" />
    ) : execution.status === 'filtered' ? (
      <Target className="h-4 w-4 text-gray-400" />
    ) : (
      <AlertCircle className="h-4 w-4 text-yellow-500" />
    )

  return (
    <Collapsible open={expanded} onOpenChange={setExpanded}>
      <CollapsibleTrigger asChild>
        <div className="flex items-center gap-3 p-3 rounded-lg border hover:bg-muted/50 cursor-pointer">
          {statusIcon}
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium truncate">{execution.span_name}</p>
            <p className="text-xs text-muted-foreground font-mono truncate">
              {execution.span_id}
            </p>
          </div>
          {execution.latency_ms && (
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <Clock className="h-3 w-3" />
              {execution.latency_ms}ms
            </div>
          )}
          <Badge
            variant={execution.status === 'success' ? 'default' : 'secondary'}
            className="text-xs"
          >
            {execution.status}
          </Badge>
          {expanded ? (
            <ChevronDown className="h-4 w-4 text-muted-foreground" />
          ) : (
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
          )}
        </div>
      </CollapsibleTrigger>
      <CollapsibleContent className="pt-2 pl-8 pr-4 space-y-3">
        {/* Score Results */}
        {execution.score_results.length > 0 && (
          <div>
            <p className="text-xs font-medium mb-1">Score Results</p>
            <div className="space-y-1">
              {execution.score_results.map((score, i) => (
                <div
                  key={i}
                  className="flex items-center justify-between text-sm bg-muted/50 px-3 py-2 rounded"
                >
                  <span className="font-medium">{score.score_name}</span>
                  <span className="font-mono">
                    {typeof score.value === 'boolean'
                      ? score.value
                        ? 'true'
                        : 'false'
                      : score.value}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Variables */}
        {execution.variables_resolved.length > 0 && (
          <div>
            <p className="text-xs font-medium mb-1">Resolved Variables</p>
            <div className="text-xs font-mono bg-muted p-2 rounded space-y-1">
              {execution.variables_resolved.map((v, i) => (
                <div key={i} className="flex gap-2">
                  <span className="text-muted-foreground">{v.variable_name}:</span>
                  <span className="truncate">
                    {typeof v.resolved_value === 'string'
                      ? v.resolved_value.length > 100
                        ? v.resolved_value.slice(0, 100) + '...'
                        : v.resolved_value
                      : JSON.stringify(v.resolved_value)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Error */}
        {execution.error_message && (
          <div className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-950/30 p-2 rounded">
            {execution.error_message}
          </div>
        )}

        {/* LLM Response */}
        {execution.llm_response && (
          <div>
            <p className="text-xs font-medium mb-1">LLM Response</p>
            <pre className="text-xs font-mono bg-muted p-2 rounded overflow-x-auto whitespace-pre-wrap">
              {execution.llm_response}
            </pre>
          </div>
        )}
      </CollapsibleContent>
    </Collapsible>
  )
}

/**
 * Result card component
 */
function ResultCard({
  label,
  value,
  icon,
  className,
}: {
  label: string
  value: number
  icon: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn('rounded-lg border p-3 text-center', className)}>
      <div className="flex items-center justify-center gap-1.5 text-muted-foreground mb-1">
        {icon}
        <span className="text-xs">{label}</span>
      </div>
      <p className="text-xl font-bold">{value}</p>
    </div>
  )
}

/**
 * Error state display
 */
function TestError({ error }: { error?: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-8 space-y-4">
      <div className="p-3 rounded-full bg-red-100 dark:bg-red-950">
        <XCircle className="h-8 w-8 text-red-600 dark:text-red-400" />
      </div>
      <div className="text-center space-y-2">
        <p className="font-medium text-red-600 dark:text-red-400">Test Failed</p>
        <p className="text-sm text-muted-foreground max-w-md">
          {error || 'An unexpected error occurred during the test.'}
        </p>
      </div>
    </div>
  )
}
