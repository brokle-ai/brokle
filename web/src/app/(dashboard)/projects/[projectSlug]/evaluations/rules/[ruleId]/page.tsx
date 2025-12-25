'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { ArrowLeft, Edit2, Trash2, Play, Pause } from 'lucide-react'
import Link from 'next/link'

import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { useProjectOnly } from '@/features/projects'
import {
  useEvaluationRuleQuery,
  useDeleteEvaluationRuleMutation,
  useActivateEvaluationRuleMutation,
  useDeactivateEvaluationRuleMutation,
  RuleStatusBadge,
  RuleScorerBadge,
  EditRuleDialog,
} from '@/features/evaluation-rules'
import type { LLMScorerConfig, BuiltinScorerConfig, RegexScorerConfig } from '@/features/evaluation-rules'

export default function RuleDetailPage() {
  const params = useParams<{ projectSlug: string; ruleId: string }>()
  const router = useRouter()
  const { currentProject, isLoading: projectLoading } = useProjectOnly()

  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

  const projectId = currentProject?.id ?? ''

  const {
    data: rule,
    isLoading: ruleLoading,
    error,
  } = useEvaluationRuleQuery(projectId || undefined, params.ruleId)

  const deleteMutation = useDeleteEvaluationRuleMutation(projectId)
  const activateMutation = useActivateEvaluationRuleMutation(projectId)
  const deactivateMutation = useDeactivateEvaluationRuleMutation(projectId)

  const isLoading = projectLoading || ruleLoading

  const handleDelete = async () => {
    if (!rule) return
    await deleteMutation.mutateAsync({ ruleId: rule.id, ruleName: rule.name })
    router.push(`/projects/${params.projectSlug}/evaluations/rules`)
  }

  const handleToggleStatus = async () => {
    if (!rule) return
    if (rule.status === 'active') {
      await deactivateMutation.mutateAsync({ ruleId: rule.id, ruleName: rule.name })
    } else {
      await activateMutation.mutateAsync({ ruleId: rule.id, ruleName: rule.name })
    }
  }

  if (isLoading) {
    return <RuleDetailSkeleton />
  }

  if (error || !rule) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="text-center py-12">
            <h1 className="text-2xl font-bold mb-2">Rule Not Found</h1>
            <p className="text-muted-foreground mb-4">
              The evaluation rule you&apos;re looking for doesn&apos;t exist or has been deleted.
            </p>
            <Button asChild>
              <Link href={`/projects/${params.projectSlug}/evaluations/rules`}>
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Rules
              </Link>
            </Button>
          </div>
        </Main>
      </>
    )
  }

  return (
    <>
      <DashboardHeader />
      <Main>
        <div className="space-y-6">
          <div className="flex items-start justify-between">
            <div className="space-y-1">
              <div className="flex items-center gap-3">
                <Link
                  href={`/projects/${params.projectSlug}/evaluations/rules`}
                  className="text-muted-foreground hover:text-foreground"
                >
                  <ArrowLeft className="h-5 w-5" />
                </Link>
                <h1 className="text-2xl font-bold tracking-tight">{rule.name}</h1>
                <RuleStatusBadge status={rule.status} />
                <RuleScorerBadge scorerType={rule.scorer_type} />
              </div>
              {rule.description && (
                <p className="text-muted-foreground ml-8">{rule.description}</p>
              )}
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                onClick={handleToggleStatus}
                disabled={activateMutation.isPending || deactivateMutation.isPending}
              >
                {rule.status === 'active' ? (
                  <>
                    <Pause className="mr-2 h-4 w-4" />
                    Deactivate
                  </>
                ) : (
                  <>
                    <Play className="mr-2 h-4 w-4" />
                    Activate
                  </>
                )}
              </Button>
              <Button variant="outline" onClick={() => setEditDialogOpen(true)}>
                <Edit2 className="mr-2 h-4 w-4" />
                Edit
              </Button>
              <Button variant="destructive" onClick={() => setDeleteDialogOpen(true)}>
                <Trash2 className="mr-2 h-4 w-4" />
                Delete
              </Button>
            </div>
          </div>

          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Configuration</CardTitle>
                <CardDescription>Basic rule settings and targeting</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Trigger Type</p>
                    <p className="text-sm">{rule.trigger_type}</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Target Scope</p>
                    <p className="text-sm capitalize">{rule.target_scope}</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Sampling Rate</p>
                    <p className="text-sm">{Math.round(rule.sampling_rate * 100)}%</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Scorer Type</p>
                    <p className="text-sm capitalize">{rule.scorer_type}</p>
                  </div>
                </div>
                {rule.span_names && rule.span_names.length > 0 && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground mb-2">Target Span Names</p>
                    <div className="flex flex-wrap gap-1">
                      {rule.span_names.map((name) => (
                        <span
                          key={name}
                          className="px-2 py-0.5 bg-muted rounded text-xs font-mono"
                        >
                          {name}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Scorer Configuration</CardTitle>
                <CardDescription>
                  {rule.scorer_type === 'llm' && 'LLM-as-Judge scoring configuration'}
                  {rule.scorer_type === 'builtin' && 'Built-in scorer configuration'}
                  {rule.scorer_type === 'regex' && 'Regex pattern matching configuration'}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ScorerConfigDisplay scorerType={rule.scorer_type} config={rule.scorer_config} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Filters</CardTitle>
                <CardDescription>Conditions for matching spans</CardDescription>
              </CardHeader>
              <CardContent>
                {rule.filter && rule.filter.length > 0 ? (
                  <div className="space-y-2">
                    {rule.filter.map((clause, index) => (
                      <div
                        key={index}
                        className="flex items-center gap-2 px-3 py-2 bg-muted rounded-md text-sm font-mono"
                      >
                        <span className="text-muted-foreground">{clause.field}</span>
                        <span className="text-primary">{clause.operator}</span>
                        <span>{JSON.stringify(clause.value)}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">No filters configured. All spans will be evaluated.</p>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Variable Mapping</CardTitle>
                <CardDescription>How span data maps to scorer variables</CardDescription>
              </CardHeader>
              <CardContent>
                {rule.variable_mapping && rule.variable_mapping.length > 0 ? (
                  <div className="space-y-2">
                    {rule.variable_mapping.map((mapping, index) => (
                      <div
                        key={index}
                        className="flex items-center justify-between px-3 py-2 bg-muted rounded-md text-sm"
                      >
                        <span className="font-mono text-primary">{`{${mapping.variable_name}}`}</span>
                        <span className="text-muted-foreground">
                          {mapping.source}
                          {mapping.json_path && <span className="font-mono ml-1">[{mapping.json_path}]</span>}
                        </span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">
                    No variable mapping configured. Default span input/output will be used.
                  </p>
                )}
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Metadata</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div>
                  <p className="font-medium text-muted-foreground">Rule ID</p>
                  <p className="font-mono">{rule.id}</p>
                </div>
                <div>
                  <p className="font-medium text-muted-foreground">Project ID</p>
                  <p className="font-mono">{rule.project_id}</p>
                </div>
                <div>
                  <p className="font-medium text-muted-foreground">Created</p>
                  <p>{new Date(rule.created_at).toLocaleString()}</p>
                </div>
                <div>
                  <p className="font-medium text-muted-foreground">Updated</p>
                  <p>{new Date(rule.updated_at).toLocaleString()}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </Main>

      <EditRuleDialog
        projectId={projectId}
        rule={rule}
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
      />

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Evaluation Rule</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{rule.name}&quot;? This action cannot be undone.
              The rule will stop evaluating spans immediately.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}

function RuleDetailSkeleton() {
  return (
    <>
      <DashboardHeader />
      <Main>
        <div className="space-y-6">
          <div className="flex items-center gap-4">
            <Skeleton className="h-10 w-10" />
            <Skeleton className="h-8 w-64" />
          </div>
          <Skeleton className="h-4 w-48" />
          <div className="grid gap-6 md:grid-cols-2">
            {[1, 2, 3, 4].map((i) => (
              <Card key={i}>
                <CardHeader>
                  <Skeleton className="h-6 w-32" />
                  <Skeleton className="h-4 w-48" />
                </CardHeader>
                <CardContent>
                  <Skeleton className="h-24 w-full" />
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </Main>
    </>
  )
}

interface ScorerConfigDisplayProps {
  scorerType: 'llm' | 'builtin' | 'regex'
  config: LLMScorerConfig | BuiltinScorerConfig | RegexScorerConfig
}

function ScorerConfigDisplay({ scorerType, config }: ScorerConfigDisplayProps) {
  if (scorerType === 'llm') {
    const llmConfig = config as LLMScorerConfig
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Model</p>
            <p className="text-sm font-mono">{llmConfig.model}</p>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">Temperature</p>
            <p className="text-sm">{llmConfig.temperature}</p>
          </div>
        </div>
        {llmConfig.messages && llmConfig.messages.length > 0 && (
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-2">Messages</p>
            <div className="space-y-2">
              {llmConfig.messages.map((msg, index) => (
                <div key={index} className="p-2 bg-muted rounded text-sm">
                  <span className="font-medium capitalize">{msg.role}:</span>
                  <p className="text-muted-foreground mt-1 whitespace-pre-wrap">{msg.content}</p>
                </div>
              ))}
            </div>
          </div>
        )}
        {llmConfig.output_schema && llmConfig.output_schema.length > 0 && (
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-2">Output Schema</p>
            <div className="space-y-1">
              {llmConfig.output_schema.map((field, index) => (
                <div key={index} className="flex items-center gap-2 text-sm">
                  <span className="font-mono">{field.name}</span>
                  <span className="text-muted-foreground">({field.type})</span>
                  {field.description && (
                    <span className="text-muted-foreground">- {field.description}</span>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    )
  }

  if (scorerType === 'builtin') {
    const builtinConfig = config as BuiltinScorerConfig
    return (
      <div className="space-y-4">
        <div>
          <p className="text-sm font-medium text-muted-foreground">Scorer Name</p>
          <p className="text-sm font-mono">{builtinConfig.scorer_name}</p>
        </div>
        {Object.keys(builtinConfig.config || {}).length > 0 && (
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-2">Configuration</p>
            <pre className="p-2 bg-muted rounded text-xs overflow-auto">
              {JSON.stringify(builtinConfig.config, null, 2)}
            </pre>
          </div>
        )}
      </div>
    )
  }

  if (scorerType === 'regex') {
    const regexConfig = config as RegexScorerConfig
    return (
      <div className="space-y-4">
        <div>
          <p className="text-sm font-medium text-muted-foreground">Pattern</p>
          <p className="text-sm font-mono bg-muted px-2 py-1 rounded">{regexConfig.pattern}</p>
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Score Name</p>
            <p className="text-sm">{regexConfig.score_name}</p>
          </div>
          {regexConfig.capture_group !== undefined && (
            <div>
              <p className="text-sm font-medium text-muted-foreground">Capture Group</p>
              <p className="text-sm">{regexConfig.capture_group}</p>
            </div>
          )}
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Match Score</p>
            <p className="text-sm">{regexConfig.match_score ?? 1}</p>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">No Match Score</p>
            <p className="text-sm">{regexConfig.no_match_score ?? 0}</p>
          </div>
        </div>
      </div>
    )
  }

  return null
}
