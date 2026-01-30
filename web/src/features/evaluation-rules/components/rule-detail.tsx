'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useSearchParams, useRouter } from 'next/navigation'
import { Pencil, Trash2, Play, Pause, Zap, Loader2, FlaskConical, BarChart3 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { PageHeader } from '@/components/layout/page-header'
import { RuleDetailProvider, useRuleDetail } from '../context/rule-detail-context'
import { RuleDetailDialogs } from './rule-detail-dialogs'
import { RuleDetailSkeleton } from './rule-detail-skeleton'
import { ScorerConfigDisplay } from './scorer-config-display'
import { RuleStatusBadge } from './rule-status-badge'
import { RuleScorerBadge } from './rule-scorer-badge'
import { RuleExecutionsTable } from './rule-executions-table'
import { RuleAnalyticsTab } from './rule-analytics-tab'
import { TestRuleDialog } from './test-rule-dialog'

interface RuleDetailProps {
  projectSlug: string
  ruleId: string
}

export function RuleDetail({ projectSlug, ruleId }: RuleDetailProps) {
  return (
    <RuleDetailProvider projectSlug={projectSlug} ruleId={ruleId}>
      <RuleDetailContent />
    </RuleDetailProvider>
  )
}

function RuleDetailContent() {
  const { rule, isLoading, projectSlug, projectId, setOpen, handleToggleStatus, handleTrigger, isToggling, isTriggering } = useRuleDetail()
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
    return <RuleDetailSkeleton />
  }

  if (!projectId) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">No project selected</p>
      </div>
    )
  }

  if (!rule) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-lg font-medium">Rule not found</p>
        <Link
          href={`/projects/${projectSlug}/evaluations/rules`}
          className="text-sm text-muted-foreground hover:underline mt-2"
        >
          Back to rules
        </Link>
      </div>
    )
  }

  return (
    <>
      <div className="space-y-6">
        <PageHeader
          title={rule.name}
          backHref={`/projects/${projectSlug}/evaluations/rules`}
          description={rule.description}
          badges={
            <>
              <RuleStatusBadge status={rule.status} />
              <RuleScorerBadge scorerType={rule.scorer_type} />
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
          </TabsContent>

          <TabsContent value="analytics" className="mt-6">
            <RuleAnalyticsTab projectId={projectId} ruleId={rule.id} />
          </TabsContent>

          <TabsContent value="executions" className="mt-6">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
                <div className="space-y-1">
                  <CardTitle>Execution History</CardTitle>
                  <CardDescription>
                    Recent rule executions and their results
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
                <RuleExecutionsTable
                  projectId={projectId}
                  projectSlug={projectSlug}
                  ruleId={rule.id}
                />
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>

      <RuleDetailDialogs />

      {/* Test Rule Dialog */}
      {projectId && rule && (
        <TestRuleDialog
          projectId={projectId}
          rule={rule}
          open={isTestDialogOpen}
          onOpenChange={setIsTestDialogOpen}
        />
      )}
    </>
  )
}
