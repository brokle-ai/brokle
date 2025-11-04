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
import { Overview } from '@/features/dashboard/components/overview'
import { RecentSales } from '@/features/dashboard/components/recent-sales'
import { HelpCircle, Download } from 'lucide-react'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

export function DashboardView() {
  return (
    <Tabs defaultValue='overview' className='flex-1'>
      {/* ===== Top Heading ===== */}
      <DashboardHeader />

      {/* ===== Main Content ===== */}
      <Main>
        <div className='flex items-center justify-between mb-3'>
          <div className='flex items-center gap-2'>
            <h1 className='text-2xl font-medium text-foreground/90'>Dashboard</h1>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button 
                    variant="ghost" 
                    size="sm" 
                    className="h-5 w-5 p-0 text-muted-foreground hover:text-foreground"
                  >
                    <HelpCircle className="h-3.5 w-3.5" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="bottom">
                  <p>View your AI platform overview and metrics</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          <Button size="sm" className="gap-1.5 text-xs">
            <Download className="h-3.5 w-3.5" />
            Export Data
          </Button>
        </div>
        
        {/* Professional Tab Navigation */}
        <div className='mb-4 border-b border-border/50 pb-1'>
          <TabsList className='h-auto p-0 bg-transparent gap-6 w-auto inline-flex'>
            {[
              { value: 'overview', label: 'Overview', disabled: false },
              { value: 'analytics', label: 'Analytics', disabled: true },
              { value: 'reports', label: 'Reports', disabled: true },
              { value: 'notifications', label: 'Notifications', disabled: true }
            ].map((tab) => (
              <TabsTrigger 
                key={tab.value}
                value={tab.value} 
                disabled={tab.disabled}
                className={cn(
                  'relative rounded-none bg-transparent px-0 pb-2.5 pt-0 h-auto',
                  'text-sm font-normal text-muted-foreground/80 transition-all duration-200',
                  'hover:text-muted-foreground data-[state=active]:text-foreground data-[state=active]:font-medium',
                  'data-[state=active]:bg-transparent data-[state=active]:shadow-none',
                  'disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:text-muted-foreground/80',
                  'after:absolute after:bottom-0 after:left-0 after:right-0 after:h-[3px] after:rounded-t-sm',
                  'after:bg-transparent data-[state=active]:after:bg-primary after:transition-all after:duration-200'
                )}
              >
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>
        
        {/* Tab Content */}
        <TabsContent value='overview' className='space-y-4 focus-visible:outline-none'>
          {/* Metrics Grid */}
          <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Total Requests
                  </CardTitle>
                  <svg
                    xmlns='http://www.w3.org/2000/svg'
                    viewBox='0 0 24 24'
                    fill='none'
                    stroke='currentColor'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                    strokeWidth='2'
                    className='text-muted-foreground h-4 w-4'
                  >
                    <path d='M22 12h-4l-3 9L9 3l-3 9H2' />
                  </svg>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>45,231</div>
                  <p className='text-muted-foreground text-xs'>
                    +20.1% from last month
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Active Models
                  </CardTitle>
                  <svg
                    xmlns='http://www.w3.org/2000/svg'
                    viewBox='0 0 24 24'
                    fill='none'
                    stroke='currentColor'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                    strokeWidth='2'
                    className='text-muted-foreground h-4 w-4'
                  >
                    <path d='M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2' />
                    <circle cx='9' cy='7' r='4' />
                    <path d='M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75' />
                  </svg>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>+23</div>
                  <p className='text-muted-foreground text-xs'>
                    +180.1% from last month
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>Avg Latency</CardTitle>
                  <svg
                    xmlns='http://www.w3.org/2000/svg'
                    viewBox='0 0 24 24'
                    fill='none'
                    stroke='currentColor'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                    strokeWidth='2'
                    className='text-muted-foreground h-4 w-4'
                  >
                    <path d='M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6' />
                  </svg>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>234ms</div>
                  <p className='text-muted-foreground text-xs'>
                    -19% from last month
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Error Rate
                  </CardTitle>
                  <svg
                    xmlns='http://www.w3.org/2000/svg'
                    viewBox='0 0 24 24'
                    fill='none'
                    stroke='currentColor'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                    strokeWidth='2'
                    className='text-muted-foreground h-4 w-4'
                  >
                    <path d='M22 12h-4l-3 9L9 3l-3 9H2' />
                  </svg>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>0.23%</div>
                  <p className='text-muted-foreground text-xs'>
                    -0.05% since last hour
                  </p>
                </CardContent>
              </Card>
          </div>
          
          {/* Charts Grid */}
          <div className='grid grid-cols-1 gap-4 lg:grid-cols-7'>
              <Card className='col-span-1 lg:col-span-4'>
                <CardHeader>
                  <CardTitle>Request Volume</CardTitle>
                </CardHeader>
                <CardContent className='pl-2'>
                  <Overview />
                </CardContent>
              </Card>
              <Card className='col-span-1 lg:col-span-3'>
                <CardHeader>
                  <CardTitle>Recent Activity</CardTitle>
                  <CardDescription>
                    Latest infrastructure events and requests.
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <RecentSales />
                </CardContent>
              </Card>
            </div>
          </TabsContent>
      </Main>
    </Tabs>
  )
}

