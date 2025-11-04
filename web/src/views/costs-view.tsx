'use client'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'

export function CostsView() {
  return (
    <>
      {/* ===== Top Heading ===== */}
      <DashboardHeader />

      {/* ===== Main ===== */}
      <Main>
        <div className='mb-2 flex items-center justify-between space-y-2'>
          <h1 className='text-2xl font-bold tracking-tight'>Cost Optimization</h1>
          <div className='flex items-center space-x-2'>
            <Button variant="outline">Set Budget</Button>
            <Button>Export Report</Button>
          </div>
        </div>
        
        <Tabs defaultValue='overview' className='space-y-4'>
          <TabsList>
            <TabsTrigger value='overview'>Overview</TabsTrigger>
            <TabsTrigger value='breakdown'>Breakdown</TabsTrigger>
            <TabsTrigger value='optimization'>Optimization</TabsTrigger>
            <TabsTrigger value='budgets'>Budgets</TabsTrigger>
          </TabsList>
          
          <TabsContent value='overview' className='space-y-4'>
            <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    This Month
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>$1,234.56</div>
                  <p className='text-muted-foreground text-xs'>
                    +12.5% from last month
                  </p>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Today
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>$42.18</div>
                  <p className='text-muted-foreground text-xs'>
                    +5.2% from yesterday
                  </p>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Avg Cost/Request
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>$0.0027</div>
                  <p className='text-muted-foreground text-xs'>
                    -8.1% optimization
                  </p>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Savings This Month
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>$387.92</div>
                  <p className='text-muted-foreground text-xs'>
                    Through optimization
                  </p>
                </CardContent>
              </Card>
            </div>
            
            <div className='grid grid-cols-1 gap-4 lg:grid-cols-3'>
              <Card className='col-span-2'>
                <CardHeader>
                  <CardTitle>Cost Trends</CardTitle>
                  <CardDescription>
                    Daily cost breakdown over the last 30 days
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className='h-[300px] bg-muted/50 rounded flex items-center justify-center'>
                    <p className='text-muted-foreground'>Cost trends chart</p>
                  </div>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader>
                  <CardTitle>Cost by Provider</CardTitle>
                  <CardDescription>
                    Distribution across providers
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className='h-[300px] bg-muted/50 rounded flex items-center justify-center'>
                    <p className='text-muted-foreground'>Provider cost breakdown</p>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
          
          <TabsContent value='breakdown' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>Detailed Cost Breakdown</CardTitle>
                <CardDescription>
                  Analyze costs by provider, model, and time period
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[400px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Detailed breakdown table will be implemented</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value='optimization' className='space-y-4'>
            <div className='grid gap-4 sm:grid-cols-1 lg:grid-cols-2'>
              <Card>
                <CardHeader>
                  <CardTitle>Optimization Opportunities</CardTitle>
                  <CardDescription>
                    Recommendations to reduce costs
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className='h-[300px] bg-muted/50 rounded flex items-center justify-center'>
                    <p className='text-muted-foreground'>Optimization suggestions</p>
                  </div>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader>
                  <CardTitle>Savings Potential</CardTitle>
                  <CardDescription>
                    Projected savings with optimizations
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className='h-[300px] bg-muted/50 rounded flex items-center justify-center'>
                    <p className='text-muted-foreground'>Savings projection</p>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
          
          <TabsContent value='budgets' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>Budget Management</CardTitle>
                <CardDescription>
                  Set and monitor spending budgets
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[400px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Budget management will be implemented</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </Main>
    </>
  )
}

