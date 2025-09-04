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
import { TopNav } from '@/components/layout/top-nav'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

export function ModelsView() {
  return (
    <>
      {/* ===== Top Heading ===== */}
      <Header>
        <TopNav links={topNav} />
        <div className='ml-auto flex items-center space-x-4'>
          <Search />
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      {/* ===== Main ===== */}
      <Main>
        <div className='mb-2 flex items-center justify-between space-y-2'>
          <h1 className='text-2xl font-bold tracking-tight'>Models</h1>
          <div className='flex items-center space-x-2'>
            <Button variant="outline">Configure Routing</Button>
            <Button>Add Model</Button>
          </div>
        </div>
        
        <Tabs defaultValue='overview' className='space-y-4'>
          <TabsList>
            <TabsTrigger value='overview'>Overview</TabsTrigger>
            <TabsTrigger value='performance'>Performance</TabsTrigger>
            <TabsTrigger value='routing'>Routing</TabsTrigger>
            <TabsTrigger value='configuration'>Configuration</TabsTrigger>
          </TabsList>
          
          <TabsContent value='overview' className='space-y-4'>
            <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Active Models
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>23</div>
                  <p className='text-muted-foreground text-xs'>
                    Across 5 providers
                  </p>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Most Used
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>GPT-4</div>
                  <p className='text-muted-foreground text-xs'>
                    45% of requests
                  </p>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Best Performance
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>Claude-3</div>
                  <p className='text-muted-foreground text-xs'>
                    198ms avg latency
                  </p>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Cost Efficiency
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>Gemini Pro</div>
                  <p className='text-muted-foreground text-xs'>
                    $0.0012/1k tokens
                  </p>
                </CardContent>
              </Card>
            </div>
            
            <div className='grid grid-cols-1 gap-4 lg:grid-cols-3'>
              <Card className='col-span-2'>
                <CardHeader>
                  <CardTitle>Model Performance Comparison</CardTitle>
                  <CardDescription>
                    Latency vs accuracy across different models
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className='h-[300px] bg-muted/50 rounded flex items-center justify-center'>
                    <p className='text-muted-foreground'>Performance comparison chart</p>
                  </div>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader>
                  <CardTitle>Usage Distribution</CardTitle>
                  <CardDescription>
                    Requests by model type
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className='h-[300px] bg-muted/50 rounded flex items-center justify-center'>
                    <p className='text-muted-foreground'>Usage pie chart</p>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
          
          <TabsContent value='performance' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>Model Performance Metrics</CardTitle>
                <CardDescription>
                  Detailed performance analysis for each model
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[400px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Performance metrics table will be implemented</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value='routing' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>Routing Configuration</CardTitle>
                <CardDescription>
                  Manage how requests are routed to different models
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[400px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Routing configuration will be implemented</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value='configuration' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>Model Configuration</CardTitle>
                <CardDescription>
                  Configure model settings and parameters
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[400px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Model configuration will be implemented</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </Main>
    </>
  )
}

const topNav = [
  {
    title: 'Models',
    href: '#',
    isActive: true,
    disabled: false,
  },
]