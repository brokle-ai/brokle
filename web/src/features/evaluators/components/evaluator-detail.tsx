'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useSearchParams, useRouter } from 'next/navigation'
import { Pencil, Trash2, Play, Pause, Zap, Loader2, FlaskConical, BarChart3 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { PageHeader } from '@/components/layout/page-header'
import { EvaluatorDetailProvider, useEvaluatorDetail } from '../context/evaluator-detail-context'
import { EvaluatorDetailDialogs } from './evaluator-detail-dialogs'
import { EvaluatorDetailSkeleton } from './evaluator-detail-skeleton'
import { ScorerConfigDisplay } from './scorer-config-display'
import { EvaluatorStatusBadge } from './evaluator-status-badge'
import { EvaluatorScorerBadge } from './evaluator-scorer-badge'
import { EvaluatorExecutionsTable } from './evaluator-executions-table'
import { EvaluatorAnalyticsTab } from './evaluator-analytics-tab'
import { TestEvaluatorDialog } from './test-evaluator-dialog'

interface EvaluatorDetailProps {
  projectSlug: string
  evaluatorId: string
}

export function EvaluatorDetail({ projectSlug, evaluatorId }: EvaluatorDetailProps) {
  return (
    <EvaluatorDetailProvider projectSlug={projectSlug} evaluatorId={evaluatorId}>
      <EvaluatorDetailContent />
    </EvaluatorDetailProvider>
  )
}

function EvaluatorDetailContent() {
  const { evaluator, isLoading, projectSlug, projectId, setOpen, handleToggleStatus, handleTrigger, isToggling, isTriggering } = useEvaluatorDetail()
  const [isTestDialogOpen, setIsTestDialogOpen] = useState(false)
  const searchParams = useSearchParams()
  const router = useRouter()

  // Read tab from URL, default to 'overview'
  const currentTab = searchParams.get('tab') || 'overview'

  // Update URL when tab changes (keeps history, enables back button)
  const handleTabChange = (value: string) => {
    const newParams = new URLSearchParams(searchParams.toString())
    newParams.set('tab', value)
    router.push(`?${newParams.toString()}`)
  }

  if (isLoading) {
    return <EvaluatorDetailSkeleton />
  }

  if (!projectId) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">No project selected</p>
      </div>
    )
  }

  if (!evaluator) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-lg font-medium">Evaluator not found</p>
        <Link
          href={`/projects/${projectSlug}/evaluators`}
          className="text-sm text-muted-foreground hover:underline mt-2"
        >
          Back to evaluators
        </Link>
      </div>
    )
  }

  return (
    <>
      <div className="space-y-6">
        <PageHeader
          title={evaluator.name}
          backHref={`/projects/${projectSlug}/evaluators`}
          description={evaluator.description}
          badges={
            <>
              <EvaluatorStatusBadge status={evaluator.status} />
              <EvaluatorScorerBadge scorerType={evaluator.scorer_type} />
            </>
          }
        >
          <Button
            variant="outline"
            onClick={() => setIsTestDialogOpen(true)}
          >
            <FlaskConical className="mr-2 h-4 w-4" />
            Test
          </Button>
          <Button
            variant="outline"
            onClick={handleToggleStatus}
            disabled={isToggling}
          >
            {evaluator.status === 'active' ? (
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
          <Button variant="outline" onClick={() => setOpen('edit')}>
            <Pencil className="mr-2 h-4 w-4" />
            Edit
          </Button>
          <Button
            variant="outline"
            className="text-destructive hover:text-destructive hover:bg-destructive/10"
            onClick={() => setOpen('delete')}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Delete
          </Button>
        </PageHeader>

        <Tabs value={currentTab} onValueChange={handleTabChange} className="w-full">
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="analytics">
              <BarChart3 className="mr-2 h-4 w-4" />
              Analytics
            </TabsTrigger>
            <TabsTrigger value="executions">Executions</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6 mt-6">
            <div className="grid gap-6 md:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle>Configuration</CardTitle>
                  <CardDescription>Basic evaluator settings and targeting</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Trigger Type</p>
                      <p className="text-sm">{evaluator.trigger_type}</p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Target Scope</p>
                      <p className="text-sm capitalize">{evaluator.target_scope}</p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Sampling Rate</p>
                      <p className="text-sm">{Math.round(evaluator.sampling_rate * 100)}%</p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Scorer Type</p>
                      <p className="text-sm capitalize">{evaluator.scorer_type}</p>
                    </div>
                  </div>
                  {evaluator.span_names && evaluator.span_names.length > 0 && (
                    <div>
                      <p className="text-sm font-medium text-muted-foreground mb-2">Target Span Names</p>
                      <div className="flex flex-wrap gap-1">
                        {evaluator.span_names.map((name) => (
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
                    {evaluator.scorer_type === 'llm' && 'LLM-as-Judge scoring configuration'}
                    {evaluator.scorer_type === 'builtin' && 'Built-in scorer configuration'}
                    {evaluator.scorer_type === 'regex' && 'Regex pattern matching configuration'}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <ScorerConfigDisplay scorerType={evaluator.scorer_type} config={evaluator.scorer_config} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Filters</CardTitle>
                  <CardDescription>Conditions for matching spans</CardDescription>
                </CardHeader>
                <CardContent>
                  {evaluator.filter && evaluator.filter.length > 0 ? (
                    <div className="space-y-2">
                      {evaluator.filter.map((clause, index) => (
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
                  {evaluator.variable_mapping && evaluator.variable_mapping.length > 0 ? (
                    <div className="space-y-2">
                      {evaluator.variable_mapping.map((mapping, index) => (
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
                    <p className="font-medium text-muted-foreground">Evaluator ID</p>
                    <p className="font-mono">{evaluator.id}</p>
                  </div>
                  <div>
                    <p className="font-medium text-muted-foreground">Project ID</p>
                    <p className="font-mono">{evaluator.project_id}</p>
                  </div>
                  <div>
                    <p className="font-medium text-muted-foreground">Created</p>
                    <p>{new Date(evaluator.created_at).toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="font-medium text-muted-foreground">Updated</p>
                    <p>{new Date(evaluator.updated_at).toLocaleString()}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="analytics" className="mt-6">
            <EvaluatorAnalyticsTab projectId={projectId} evaluatorId={evaluator.id} />
          </TabsContent>

          <TabsContent value="executions" className="mt-6">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
                <div className="space-y-1">
                  <CardTitle>Execution History</CardTitle>
                  <CardDescription>
                    Recent evaluator executions and their results
                  </CardDescription>
                </div>
                <Button
                  onClick={handleTrigger}
                  disabled={isTriggering}
                  size="sm"
                >
                  {isTriggering ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Triggering...
                    </>
                  ) : (
                    <>
                      <Zap className="mr-2 h-4 w-4" />
                      Trigger Now
                    </>
                  )}
                </Button>
              </CardHeader>
              <CardContent>
                <EvaluatorExecutionsTable
                  projectId={projectId}
                  projectSlug={projectSlug}
                  evaluatorId={evaluator.id}
                />
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>

      <EvaluatorDetailDialogs />

      {/* Test Evaluator Dialog */}
      {projectId && evaluator && (
        <TestEvaluatorDialog
          projectId={projectId}
          evaluator={evaluator}
          open={isTestDialogOpen}
          onOpenChange={setIsTestDialogOpen}
        />
      )}
    </>
  )
}
