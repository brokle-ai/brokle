'use client'

import { useState } from 'react'
import { Plus, FolderOpen, Settings, Users, BarChart3, DollarSign, Activity, TrendingUp } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useWorkspace } from '@/context/workspace-context'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { MetricCard } from '@/components/shared/metrics/metric-card'
import { StatsGrid } from '@/components/shared/metrics/stats-grid'
import { ProjectGrid } from '@/components/organization/project-grid'
import { cn } from '@/lib/utils'

export function OrganizationOverview() {
  const router = useRouter()
  const { currentOrganization, isLoading } = useWorkspace()
  const [selectedTab, setSelectedTab] = useState('overview')

  if (!currentOrganization) {
    return <div>Loading...</div>
  }

  // Projects come from currentOrganization
  const projects = currentOrganization.projects || []
  const activeProjects = projects.filter(p => p.status === 'active')
  const totalRequests = projects.reduce((sum, p) => sum + (p.metrics?.total_requests || 0), 0)
  const totalCost = projects.reduce((sum, p) => sum + (p.metrics?.total_cost || 0), 0)
  const avgLatency = projects.length > 0
    ? projects.reduce((sum, p) => sum + (p.metrics?.avg_latency || 0), 0) / projects.length
    : 0
  const avgErrorRate = projects.length > 0
    ? projects.reduce((sum, p) => sum + (p.metrics?.error_rate || 0), 0) / projects.length
    : 0

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
    return num.toString()
  }

  const getPlanColor = (plan: string) => {
    switch (plan) {
      case 'enterprise':
        return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300'
      case 'business':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
      case 'pro':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'free':
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'bg-green-500'
      case 'inactive':
        return 'bg-yellow-500'
      case 'archived':
        return 'bg-gray-500'
      default:
        return 'bg-gray-500'
    }
  }

  const getUsagePercentage = () => {
    if (currentOrganization.plan === 'free') return (totalRequests / 10000) * 100
    if (currentOrganization.plan === 'pro') return (totalRequests / 100000) * 100
    if (currentOrganization.plan === 'business') return (totalRequests / 1000000) * 100
    return 0 // Enterprise has unlimited
  }

  return (
    <>
      <DashboardHeader />

      <Main>
        {/* Organization Header Section */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3 min-w-0">
            <div className="min-w-0">
              <h1 className="text-2xl font-semibold text-foreground truncate">
                {currentOrganization.name}
              </h1>
              <div className="flex items-center gap-2 mt-1">
                <Badge className={cn("text-xs", getPlanColor(currentOrganization.plan))}>
                  {currentOrganization.plan.charAt(0).toUpperCase() + currentOrganization.plan.slice(1)}
                </Badge>
                <span className="text-xs text-muted-foreground">
                  {currentOrganization.projects.length} project{currentOrganization.projects.length !== 1 ? 's' : ''}
                </span>
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2 flex-shrink-0">
            <Button variant="outline" size="sm">
              <Settings className="mr-2 h-4 w-4" />
              Settings
            </Button>
            <Button variant="outline" size="sm">
              <Users className="mr-2 h-4 w-4" />
              Members
            </Button>
            <Button size="sm">
              <Plus className="mr-2 h-4 w-4" />
              New Project
            </Button>
          </div>
        </div>

        <Tabs value={selectedTab} onValueChange={setSelectedTab} className="space-y-4">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="projects">Projects ({projects.length})</TabsTrigger>
            <TabsTrigger value="analytics">Analytics</TabsTrigger>
            <TabsTrigger value="usage">Usage</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6 mt-0">
            {/* Key Metrics */}
            <StatsGrid>
              <MetricCard
                title="Total Projects"
                value={projects.length.toString()}
                description={`${activeProjects.length} active`}
                icon={FolderOpen}
                trend={{ value: 12, isPositive: true }}
              />
              <MetricCard
                title="Total Requests"
                value={formatNumber(totalRequests)}
                description="All time"
                icon={BarChart3}
                trend={{ value: 8.2, isPositive: true }}
              />
              <MetricCard
                title="Total Cost"
                value={`$${totalCost.toFixed(2)}`}
                description="All time"
                icon={DollarSign}
                trend={{ value: 3.1, isPositive: false }}
              />
              <MetricCard
                title="Avg Latency"
                value={`${Math.round(avgLatency)}ms`}
                description="Across all projects"
                icon={Activity}
                trend={{ value: 2.4, isPositive: false }}
              />
            </StatsGrid>

            {/* Usage Progress */}
            {currentOrganization.plan !== 'enterprise' && (
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <TrendingUp className="h-5 w-5" />
                    Usage This Month
                  </CardTitle>
                  <CardDescription>
                    Track your usage against your plan limits
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <div className="flex justify-between text-sm mb-2">
                      <span>API Requests</span>
                      <span>{formatNumber(currentOrganization.usage?.requests_this_month || 0)}</span>
                    </div>
                    <Progress value={getUsagePercentage()} className="h-2" />
                  </div>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <div className="text-muted-foreground">Cost this month</div>
                      <div className="font-medium">${currentOrganization.usage?.cost_this_month.toFixed(2) || '0.00'}</div>
                    </div>
                    <div>
                      <div className="text-muted-foreground">Models used</div>
                      <div className="font-medium">{currentOrganization.usage?.models_used || 0}</div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
          </TabsContent>

          <TabsContent value="projects" className="space-y-6 mt-0">
            <ProjectGrid />
          </TabsContent>

          <TabsContent value="analytics" className="space-y-6 mt-0">
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
              <Card>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Error Rate</p>
                      <p className="text-2xl font-bold">{(avgErrorRate * 100).toFixed(2)}%</p>
                    </div>
                    <Activity className="h-8 w-8 text-muted-foreground" />
                  </div>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">Avg Latency</p>
                      <p className="text-2xl font-bold">{Math.round(avgLatency)}ms</p>
                    </div>
                    <TrendingUp className="h-8 w-8 text-muted-foreground" />
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="usage" className="space-y-6 mt-0">
            <Card>
              <CardHeader>
                <CardTitle>Plan Usage</CardTitle>
                <CardDescription>
                  Monitor your usage across all projects in this organization
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-sm text-muted-foreground">Requests this month</span>
                      <span className="text-sm font-medium">
                        {formatNumber(currentOrganization.usage?.requests_this_month || 0)}
                      </span>
                    </div>
                    {currentOrganization.plan !== 'enterprise' && (
                      <Progress value={getUsagePercentage()} className="h-2" />
                    )}
                  </div>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-sm text-muted-foreground">Cost this month</span>
                      <span className="text-sm font-medium">
                        ${currentOrganization.usage?.cost_this_month.toFixed(2) || '0.00'}
                      </span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </Main>
    </>
  )
}