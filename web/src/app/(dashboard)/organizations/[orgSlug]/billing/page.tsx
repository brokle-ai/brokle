'use client'

import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { 
  CreditCard, 
  Download, 
  Calendar, 
  DollarSign, 
  TrendingUp, 
  AlertCircle,
  Check,
  Crown,
  Zap
} from 'lucide-react'
import { useOrganization } from '@/context/org-context'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Breadcrumbs } from '@/components/layout/breadcrumbs'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'
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
import type { OrganizationParams, SubscriptionPlan } from '@/types/organization'

// Mock billing data
const MOCK_INVOICES = [
  {
    id: 'inv_001',
    date: '2024-03-01',
    amount: 29.00,
    status: 'paid',
    description: 'Pro Plan - March 2024',
    downloadUrl: '#',
  },
  {
    id: 'inv_002',
    date: '2024-02-01',
    amount: 29.00,
    status: 'paid',
    description: 'Pro Plan - February 2024',
    downloadUrl: '#',
  },
  {
    id: 'inv_003',
    date: '2024-01-01',
    amount: 29.00,
    status: 'paid',
    description: 'Pro Plan - January 2024',
    downloadUrl: '#',
  },
]

const PLAN_FEATURES = {
  free: [
    '10K requests per month',
    'Basic analytics',
    'Email support',
    '7 days data retention',
    'Up to 3 projects'
  ],
  pro: [
    '100K requests per month',
    'Advanced observability',
    'Intelligent routing',
    'Semantic caching',
    'Priority support',
    '30 days data retention',
    'Unlimited projects'
  ],
  business: [
    '1M requests per month',
    'Predictive analytics',
    'Custom dashboards',
    'Team collaboration',
    'Phone/chat support',
    '180 days data retention',
    'Model hosting (coming soon)',
    'Advanced integrations'
  ],
  enterprise: [
    'Unlimited requests',
    'Custom integrations',
    'Compliance features',
    'Dedicated support',
    'Custom data retention',
    'On-premise deployment',
    'SLA guarantees',
    'Priority feature requests'
  ]
}

export default function BillingSettingsPage() {
  const params = useParams() as OrganizationParams
  const router = useRouter()
  const { 
    currentOrganization,
    isLoading,
    error,
    hasAccess,
    getUserRole
  } = useOrganization()
  
  const [selectedPlan, setSelectedPlan] = useState<SubscriptionPlan>('pro')

  useEffect(() => {
    if (isLoading) return

    if (!hasAccess(params.orgSlug)) {
      router.push('/')
      return
    }

    // Check if user has admin permissions
    const userRole = getUserRole(params.orgSlug)
    if (userRole !== 'owner' && userRole !== 'admin') {
      router.push(`/${params.orgSlug}`)
      return
    }
  }, [params.orgSlug, isLoading, hasAccess, getUserRole, router])

  if (isLoading) {
    return (
      <>
        <Header>
          <Skeleton className="h-8 w-64" />
        </Header>
        <Main className="space-y-6">
          <Skeleton className="h-6 w-96" />
          <div className="grid gap-6 md:grid-cols-2">
            <Skeleton className="h-48" />
            <Skeleton className="h-48" />
          </div>
        </Main>
      </>
    )
  }

  if (error || !currentOrganization) {
    return (
      <>
        <Header>
          <h1 className="text-2xl font-bold text-foreground">Access Denied</h1>
        </Header>
        <Main>
          <div className="text-center py-12">
            <h2 className="text-xl font-semibold mb-2">Access Denied</h2>
            <p className="text-muted-foreground mb-4">
              You don't have permission to manage billing for this organization.
            </p>
            <button 
              onClick={() => router.push(currentOrganization ? `/${currentOrganization.slug}` : '/')}
              className="text-primary hover:underline"
            >
              Go back
            </button>
          </div>
        </Main>
      </>
    )
  }

  const getPlanPrice = (plan: SubscriptionPlan) => {
    switch (plan) {
      case 'free': return 0
      case 'pro': return 29
      case 'business': return 99
      case 'enterprise': return 'Custom'
      default: return 0
    }
  }

  const getPlanColor = (plan: SubscriptionPlan) => {
    switch (plan) {
      case 'free': return 'text-gray-600'
      case 'pro': return 'text-green-600'
      case 'business': return 'text-blue-600'
      case 'enterprise': return 'text-purple-600'
      default: return 'text-gray-600'
    }
  }

  const getUsagePercentage = () => {
    const usage = currentOrganization.usage
    if (!usage) return 0
    
    switch (currentOrganization.plan) {
      case 'free': return (usage.requests_this_month / 10000) * 100
      case 'pro': return (usage.requests_this_month / 100000) * 100
      case 'business': return (usage.requests_this_month / 1000000) * 100
      case 'enterprise': return 0
      default: return 0
    }
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
    return num.toString()
  }

  return (
    <>
      <Header>
        <div className="space-y-2">
          <Breadcrumbs />
          <div>
            <h1 className="text-2xl font-bold text-foreground">
              Billing & Subscription
            </h1>
            <p className="text-muted-foreground">
              Manage your subscription and billing details for {currentOrganization.name}
            </p>
          </div>
        </div>
      </Header>

      <Main className="space-y-8">
        {/* Current Plan & Usage */}
        <div className="grid gap-6 md:grid-cols-2">
          {/* Current Plan */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <CreditCard className="h-5 w-5" />
                Current Plan
              </CardTitle>
              <CardDescription>
                Your active subscription plan and features
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className={cn("text-2xl font-bold capitalize", getPlanColor(currentOrganization.plan))}>
                      {currentOrganization.plan}
                    </h3>
                    {currentOrganization.plan === 'enterprise' && <Crown className="h-5 w-5 text-purple-500" />}
                    {currentOrganization.plan === 'pro' && <Zap className="h-5 w-5 text-green-500" />}
                  </div>
                  <p className="text-muted-foreground">
                    ${getPlanPrice(currentOrganization.plan)}{typeof getPlanPrice(currentOrganization.plan) === 'number' ? '/month' : ''}
                  </p>
                </div>
                <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                  Active
                </Badge>
              </div>

              <div className="space-y-2">
                <h4 className="font-medium">Included Features</h4>
                <ul className="space-y-1 text-sm">
                  {PLAN_FEATURES[currentOrganization.plan]?.slice(0, 3).map((feature, index) => (
                    <li key={index} className="flex items-center gap-2">
                      <Check className="h-3 w-3 text-green-500" />
                      {feature}
                    </li>
                  ))}
                </ul>
              </div>

              <Button className="w-full">Manage Subscription</Button>
            </CardContent>
          </Card>

          {/* Usage This Month */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5" />
                Usage This Month
              </CardTitle>
              <CardDescription>
                Track your current usage against plan limits
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {currentOrganization.plan !== 'enterprise' && (
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>API Requests</span>
                    <span>{formatNumber(currentOrganization.usage?.requests_this_month || 0)} / {
                      currentOrganization.plan === 'free' ? '10K' :
                      currentOrganization.plan === 'pro' ? '100K' : '1M'
                    }</span>
                  </div>
                  <Progress value={getUsagePercentage()} className="h-2" />
                  {getUsagePercentage() > 80 && (
                    <Alert>
                      <AlertCircle className="h-4 w-4" />
                      <AlertDescription>
                        You're approaching your monthly limit. Consider upgrading your plan.
                      </AlertDescription>
                    </Alert>
                  )}
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-2xl font-bold">${currentOrganization.usage?.cost_this_month.toFixed(2) || '0.00'}</div>
                  <div className="text-sm text-muted-foreground">Total cost</div>
                </div>
                <div>
                  <div className="text-2xl font-bold">{currentOrganization.usage?.models_used || 0}</div>
                  <div className="text-sm text-muted-foreground">Models used</div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Plan Upgrade Options */}
        <Card>
          <CardHeader>
            <CardTitle>Upgrade Your Plan</CardTitle>
            <CardDescription>
              Get more features and higher limits for your growing AI infrastructure needs
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
              {(['free', 'pro', 'business', 'enterprise'] as SubscriptionPlan[]).map((plan) => (
                <Card 
                  key={plan}
                  className={cn(
                    "cursor-pointer transition-all",
                    plan === currentOrganization.plan ? "ring-2 ring-primary" : "hover:shadow-md",
                    plan === selectedPlan ? "ring-2 ring-blue-500" : ""
                  )}
                  onClick={() => setSelectedPlan(plan)}
                >
                  <CardHeader className="pb-2">
                    <CardTitle className={cn("text-lg capitalize", getPlanColor(plan))}>
                      {plan}
                    </CardTitle>
                    <div className="text-2xl font-bold">
                      ${getPlanPrice(plan)}{typeof getPlanPrice(plan) === 'number' && plan !== 'free' ? '/mo' : ''}
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-2">
                    <ul className="space-y-1 text-sm">
                      {PLAN_FEATURES[plan]?.slice(0, 4).map((feature, index) => (
                        <li key={index} className="flex items-center gap-2">
                          <Check className="h-3 w-3 text-green-500" />
                          <span className="text-muted-foreground">{feature}</span>
                        </li>
                      ))}
                    </ul>
                    
                    {plan !== currentOrganization.plan ? (
                      <Button 
                        className="w-full mt-4" 
                        variant={plan > currentOrganization.plan ? "default" : "outline"}
                      >
                        {plan > currentOrganization.plan ? 'Upgrade' : 'Downgrade'}
                      </Button>
                    ) : (
                      <Badge className="w-full justify-center mt-4">
                        Current Plan
                      </Badge>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Billing History */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <Calendar className="h-5 w-5" />
                  Billing History
                </CardTitle>
                <CardDescription>
                  View and download your past invoices
                </CardDescription>
              </div>
              <Button variant="outline">
                <Download className="mr-2 h-4 w-4" />
                Download All
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Date</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="w-[100px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {MOCK_INVOICES.map((invoice) => (
                  <TableRow key={invoice.id}>
                    <TableCell>
                      {new Date(invoice.date).toLocaleDateString()}
                    </TableCell>
                    <TableCell>{invoice.description}</TableCell>
                    <TableCell>${invoice.amount.toFixed(2)}</TableCell>
                    <TableCell>
                      <Badge variant="outline" className="bg-green-50 text-green-700">
                        {invoice.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button variant="ghost" size="sm">
                        <Download className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Payment Method */}
        <Card>
          <CardHeader>
            <CardTitle>Payment Method</CardTitle>
            <CardDescription>
              Manage your payment method and billing address
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between p-4 border rounded-lg">
              <div className="flex items-center gap-3">
                <CreditCard className="h-8 w-8 text-muted-foreground" />
                <div>
                  <div className="font-medium">•••• •••• •••• 4242</div>
                  <div className="text-sm text-muted-foreground">Expires 12/25</div>
                </div>
              </div>
              <Button variant="outline">Update</Button>
            </div>
            
            <div className="text-sm text-muted-foreground">
              <p><strong>Billing Address:</strong></p>
              <p>123 AI Street<br />San Francisco, CA 94105<br />United States</p>
            </div>
          </CardContent>
        </Card>
      </Main>
    </>
  )
}