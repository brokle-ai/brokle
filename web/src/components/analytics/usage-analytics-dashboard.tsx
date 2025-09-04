'use client'

import { useState } from 'react'
import { 
  BarChart3, 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  Clock, 
  AlertCircle,
  Download,
  RefreshCw
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'

interface UsageMetrics {
  period: string
  requests: number
  cost: number
  avg_latency: number
  error_rate: number
  top_models: { name: string; requests: number; cost: number }[]
  daily_usage: { date: string; requests: number; cost: number }[]
}

interface BillingData {
  current_usage: {
    requests_this_month: number
    cost_this_month: number
    quota_limit: number
    cost_limit?: number
  }
  cost_breakdown: {
    model: string
    requests: number
    cost: number
    percentage: number
  }[]
  monthly_trends: {
    month: string
    requests: number
    cost: number
    savings: number
  }[]
}

const MOCK_USAGE_METRICS: UsageMetrics = {
  period: 'Last 30 days',
  requests: 145823,
  cost: 2847.63,
  avg_latency: 1234,
  error_rate: 2.1,
  top_models: [
    { name: 'GPT-4 Turbo', requests: 68492, cost: 1423.56 },
    { name: 'Claude 3 Opus', requests: 45312, cost: 876.23 },
    { name: 'GPT-3.5 Turbo', requests: 32019, cost: 547.84 }
  ],
  daily_usage: Array.from({ length: 30 }, (_, i) => ({
    date: new Date(Date.now() - (29 - i) * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    requests: Math.floor(Math.random() * 8000) + 2000,
    cost: Math.floor(Math.random() * 150) + 50
  }))
}

const MOCK_BILLING_DATA: BillingData = {
  current_usage: {
    requests_this_month: 89432,
    cost_this_month: 1647.32,
    quota_limit: 100000,
    cost_limit: 2000
  },
  cost_breakdown: [
    { model: 'GPT-4 Turbo', requests: 45123, cost: 987.45, percentage: 60.0 },
    { model: 'Claude 3 Opus', requests: 28943, cost: 456.78, percentage: 27.7 },
    { model: 'GPT-3.5 Turbo', requests: 15366, cost: 203.09, percentage: 12.3 }
  ],
  monthly_trends: [
    { month: 'Jan 2024', requests: 67432, cost: 1234.56, savings: 234.67 },
    { month: 'Feb 2024', requests: 72156, cost: 1345.78, savings: 298.45 },
    { month: 'Mar 2024', requests: 89432, cost: 1647.32, savings: 356.89 }
  ]
}

interface UsageAnalyticsDashboardProps {
  organizationId?: string
  projectId?: string
}

export function UsageAnalyticsDashboard({ }: UsageAnalyticsDashboardProps) {
  const [timeRange, setTimeRange] = useState<'7d' | '30d' | '90d' | '1y'>('30d')
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [selectedTab, setSelectedTab] = useState('overview')

  const handleRefresh = async () => {
    setIsRefreshing(true)
    // TODO: Implement actual data refresh
    await new Promise(resolve => setTimeout(resolve, 1000))
    setIsRefreshing(false)
  }

  const handleExport = () => {
    const exportData = {
      usage_metrics: MOCK_USAGE_METRICS,
      billing_data: MOCK_BILLING_DATA,
      exported_at: new Date().toISOString(),
      time_range: timeRange
    }
    
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `usage-analytics-${timeRange}-${new Date().toISOString().split('T')[0]}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
    return num.toString()
  }

  const getUsagePercentage = () => {
    return (MOCK_BILLING_DATA.current_usage.requests_this_month / MOCK_BILLING_DATA.current_usage.quota_limit) * 100
  }

  const getCostPercentage = () => {
    if (!MOCK_BILLING_DATA.current_usage.cost_limit) return 0
    return (MOCK_BILLING_DATA.current_usage.cost_this_month / MOCK_BILLING_DATA.current_usage.cost_limit) * 100
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-foreground">Usage Analytics</h2>
          <p className="text-muted-foreground">
            Monitor usage, costs, and performance metrics
          </p>
        </div>
        
        <div className="flex items-center gap-2">
          <Select value={timeRange} onValueChange={(value: '7d' | '30d' | '90d' | '1y') => setTimeRange(value)}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="7d">Last 7 days</SelectItem>
              <SelectItem value="30d">Last 30 days</SelectItem>
              <SelectItem value="90d">Last 90 days</SelectItem>
              <SelectItem value="1y">Last year</SelectItem>
            </SelectContent>
          </Select>
          
          <Button variant="outline" size="sm" onClick={handleRefresh} disabled={isRefreshing}>
            <RefreshCw className={cn("mr-2 h-4 w-4", isRefreshing && "animate-spin")} />
            Refresh
          </Button>
          
          <Button variant="outline" size="sm" onClick={handleExport}>
            <Download className="mr-2 h-4 w-4" />
            Export
          </Button>
        </div>
      </div>

      <Tabs value={selectedTab} onValueChange={setSelectedTab}>
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="costs">Costs</TabsTrigger>
          <TabsTrigger value="performance">Performance</TabsTrigger>
          <TabsTrigger value="models">Models</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          {/* Key Metrics */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Total Requests</p>
                    <p className="text-2xl font-bold">{formatNumber(MOCK_USAGE_METRICS.requests)}</p>
                    <div className="flex items-center gap-1 text-sm">
                      <TrendingUp className="h-3 w-3 text-green-500" />
                      <span className="text-green-500">+12.5%</span>
                    </div>
                  </div>
                  <BarChart3 className="h-8 w-8 text-blue-500" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Total Cost</p>
                    <p className="text-2xl font-bold">${MOCK_USAGE_METRICS.cost.toFixed(2)}</p>
                    <div className="flex items-center gap-1 text-sm">
                      <TrendingDown className="h-3 w-3 text-green-500" />
                      <span className="text-green-500">-8.2%</span>
                    </div>
                  </div>
                  <DollarSign className="h-8 w-8 text-green-500" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Avg Latency</p>
                    <p className="text-2xl font-bold">{MOCK_USAGE_METRICS.avg_latency}ms</p>
                    <div className="flex items-center gap-1 text-sm">
                      <TrendingDown className="h-3 w-3 text-green-500" />
                      <span className="text-green-500">-15.3%</span>
                    </div>
                  </div>
                  <Clock className="h-8 w-8 text-purple-500" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Error Rate</p>
                    <p className="text-2xl font-bold">{MOCK_USAGE_METRICS.error_rate}%</p>
                    <div className="flex items-center gap-1 text-sm">
                      <TrendingUp className="h-3 w-3 text-red-500" />
                      <span className="text-red-500">+0.3%</span>
                    </div>
                  </div>
                  <AlertCircle className="h-8 w-8 text-red-500" />
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Usage Progress */}
          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Monthly Usage</CardTitle>
                <CardDescription>
                  Current usage against your quota limits
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <div className="flex justify-between text-sm mb-2">
                    <span>API Requests</span>
                    <span>
                      {formatNumber(MOCK_BILLING_DATA.current_usage.requests_this_month)} / 
                      {formatNumber(MOCK_BILLING_DATA.current_usage.quota_limit)}
                    </span>
                  </div>
                  <Progress value={getUsagePercentage()} className="h-2" />
                  {getUsagePercentage() > 80 && (
                    <Alert className="mt-2">
                      <AlertCircle className="h-4 w-4" />
                      <AlertDescription className="text-sm">
                        You're approaching your monthly quota limit.
                      </AlertDescription>
                    </Alert>
                  )}
                </div>

                {MOCK_BILLING_DATA.current_usage.cost_limit && (
                  <div>
                    <div className="flex justify-between text-sm mb-2">
                      <span>Cost Budget</span>
                      <span>
                        ${MOCK_BILLING_DATA.current_usage.cost_this_month.toFixed(2)} / 
                        ${MOCK_BILLING_DATA.current_usage.cost_limit.toFixed(2)}
                      </span>
                    </div>
                    <Progress value={getCostPercentage()} className="h-2" />
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Top Models</CardTitle>
                <CardDescription>
                  Most used models by request volume
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {MOCK_USAGE_METRICS.top_models.map((model, index) => (
                    <div key={model.name} className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <Badge variant="secondary">{index + 1}</Badge>
                        <div>
                          <div className="font-medium">{model.name}</div>
                          <div className="text-sm text-muted-foreground">
                            {formatNumber(model.requests)} requests
                          </div>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="font-medium">${model.cost.toFixed(2)}</div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="costs" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Cost Breakdown by Model</CardTitle>
              <CardDescription>
                Detailed cost analysis across different models
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Model</TableHead>
                    <TableHead>Requests</TableHead>
                    <TableHead>Cost</TableHead>
                    <TableHead>Percentage</TableHead>
                    <TableHead>Avg Cost/1K</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {MOCK_BILLING_DATA.cost_breakdown.map((item) => (
                    <TableRow key={item.model}>
                      <TableCell className="font-medium">{item.model}</TableCell>
                      <TableCell>{formatNumber(item.requests)}</TableCell>
                      <TableCell>${item.cost.toFixed(2)}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Progress value={item.percentage} className="h-2 w-16" />
                          {item.percentage.toFixed(1)}%
                        </div>
                      </TableCell>
                      <TableCell>${(item.cost / (item.requests / 1000)).toFixed(3)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Monthly Cost Trends</CardTitle>
              <CardDescription>
                Cost trends and savings over time
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {MOCK_BILLING_DATA.monthly_trends.map((trend) => (
                  <div key={trend.month} className="flex items-center justify-between p-4 border rounded-lg">
                    <div>
                      <div className="font-medium">{trend.month}</div>
                      <div className="text-sm text-muted-foreground">
                        {formatNumber(trend.requests)} requests
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="font-medium">${trend.cost.toFixed(2)}</div>
                      <div className="text-sm text-green-600">
                        Saved ${trend.savings.toFixed(2)}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="performance" className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Response Time Distribution</CardTitle>
                <CardDescription>
                  Latency percentiles across all requests
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-sm">P50 (Median)</span>
                    <span className="font-medium">892ms</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm">P90</span>
                    <span className="font-medium">1,456ms</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm">P95</span>
                    <span className="font-medium">2,134ms</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm">P99</span>
                    <span className="font-medium">4,567ms</span>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Error Analysis</CardTitle>
                <CardDescription>
                  Error rates by type and frequency
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Rate Limits (429)</span>
                    <div className="flex items-center gap-2">
                      <Progress value={45} className="h-2 w-20" />
                      <span className="text-sm font-medium">1.2%</span>
                    </div>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Timeouts (408)</span>
                    <div className="flex items-center gap-2">
                      <Progress value={30} className="h-2 w-20" />
                      <span className="text-sm font-medium">0.7%</span>
                    </div>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Server Errors (5xx)</span>
                    <div className="flex items-center gap-2">
                      <Progress value={15} className="h-2 w-20" />
                      <span className="text-sm font-medium">0.2%</span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="models" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Model Performance Comparison</CardTitle>
              <CardDescription>
                Compare performance metrics across different models
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Model</TableHead>
                    <TableHead>Requests</TableHead>
                    <TableHead>Avg Latency</TableHead>
                    <TableHead>Success Rate</TableHead>
                    <TableHead>Cost/1K Tokens</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow>
                    <TableCell className="font-medium">GPT-4 Turbo</TableCell>
                    <TableCell>{formatNumber(68492)}</TableCell>
                    <TableCell>1,245ms</TableCell>
                    <TableCell>
                      <Badge className="bg-green-100 text-green-800">99.2%</Badge>
                    </TableCell>
                    <TableCell>$0.010</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell className="font-medium">Claude 3 Opus</TableCell>
                    <TableCell>{formatNumber(45312)}</TableCell>
                    <TableCell>1,387ms</TableCell>
                    <TableCell>
                      <Badge className="bg-green-100 text-green-800">98.8%</Badge>
                    </TableCell>
                    <TableCell>$0.015</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell className="font-medium">GPT-3.5 Turbo</TableCell>
                    <TableCell>{formatNumber(32019)}</TableCell>
                    <TableCell>892ms</TableCell>
                    <TableCell>
                      <Badge className="bg-green-100 text-green-800">99.5%</Badge>
                    </TableCell>
                    <TableCell>$0.002</TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}