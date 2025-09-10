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
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ContextNavbar } from '@/components/layout/context-navbar'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { ProfileDropdown } from '@/components/profile-dropdown'

export function AnalyticsView() {
  return (
    <>
      {/* ===== Top Heading ===== */}
      <Header>
        <ContextNavbar />
        <div className='ml-auto flex items-center space-x-4'>
          <Search />
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      {/* ===== Main ===== */}
      <Main>
        <div className='mb-2 flex items-center justify-between space-y-2'>
          <h1 className='text-2xl font-bold tracking-tight'>Analytics</h1>
          <div className='flex items-center space-x-2'>
            <Button variant="outline">Export</Button>
            <Button>Refresh</Button>
          </div>
        </div>
        
        <Tabs defaultValue='performance' className='space-y-4'>
          <TabsList>
            <TabsTrigger value='performance'>Performance</TabsTrigger>
            <TabsTrigger value='usage'>Usage</TabsTrigger>
            <TabsTrigger value='providers'>Providers</TabsTrigger>
            <TabsTrigger value='costs'>Costs</TabsTrigger>
          </TabsList>
          
          <TabsContent value='performance' className='space-y-4'>
            <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Avg Response Time
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>234ms</div>
                  <p className='text-muted-foreground text-xs'>
                    -12ms from last hour
                  </p>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Throughput
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>1.2k/min</div>
                  <p className='text-muted-foreground text-xs'>
                    +5% from last hour
                  </p>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Success Rate
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>99.77%</div>
                  <p className='text-muted-foreground text-xs'>
                    +0.05% from last hour
                  </p>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Cache Hit Rate
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>84.2%</div>
                  <p className='text-muted-foreground text-xs'>
                    +2.1% from last hour
                  </p>
                </CardContent>
              </Card>
            </div>
            
            <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
              <Card>
                <CardHeader>
                  <CardTitle>Response Time Trends</CardTitle>
                  <CardDescription>
                    Average response time over the last 24 hours
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className='h-[300px] bg-muted/50 rounded flex items-center justify-center'>
                    <p className='text-muted-foreground'>Chart will be implemented with shared components</p>
                  </div>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader>
                  <CardTitle>Request Volume</CardTitle>
                  <CardDescription>
                    Requests per minute over time
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className='h-[300px] bg-muted/50 rounded flex items-center justify-center'>
                    <p className='text-muted-foreground'>Chart will be implemented with shared components</p>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
          
          <TabsContent value='usage' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>Usage Analytics</CardTitle>
                <CardDescription>
                  Detailed usage patterns and statistics
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[400px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Usage analytics will be implemented</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value='providers' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>Provider Analytics</CardTitle>
                <CardDescription>
                  Performance comparison across providers
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[400px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Provider analytics will be implemented</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value='costs' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>Cost Analytics</CardTitle>
                <CardDescription>
                  Cost breakdown and optimization insights
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[400px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Cost analytics will be implemented</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </Main>
    </>
  )
}

