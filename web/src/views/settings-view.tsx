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

export function SettingsView() {
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
          <h1 className='text-2xl font-bold tracking-tight'>Settings</h1>
          <div className='flex items-center space-x-2'>
            <Button variant="outline">Reset</Button>
            <Button>Save Changes</Button>
          </div>
        </div>
        
        <Tabs defaultValue='account' className='space-y-4'>
          <TabsList>
            <TabsTrigger value='account'>Account</TabsTrigger>
            <TabsTrigger value='platform'>Platform</TabsTrigger>
            <TabsTrigger value='notifications'>Notifications</TabsTrigger>
            <TabsTrigger value='security'>Security</TabsTrigger>
          </TabsList>
          
          <TabsContent value='account' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>Account Information</CardTitle>
                <CardDescription>
                  Manage your account details and preferences
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[300px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Account settings form will be implemented</p>
                </div>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader>
                <CardTitle>Organization Settings</CardTitle>
                <CardDescription>
                  Configure organization-wide settings
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[200px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Organization settings will be implemented</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value='platform' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>Platform Configuration</CardTitle>
                <CardDescription>
                  Configure platform behavior and integrations
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[400px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Platform configuration will be implemented</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value='notifications' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>Notification Preferences</CardTitle>
                <CardDescription>
                  Configure how and when you receive notifications
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className='h-[400px] bg-muted/50 rounded flex items-center justify-center'>
                  <p className='text-muted-foreground'>Notification settings will be implemented</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value='security' className='space-y-4'>
            <div className='grid gap-4'>
              <Card>
                <CardHeader>
                  <CardTitle>API Keys</CardTitle>
                  <CardDescription>
                    Manage your API keys and access tokens
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className='h-[200px] bg-muted/50 rounded flex items-center justify-center'>
                    <p className='text-muted-foreground'>API key management will be implemented</p>
                  </div>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader>
                  <CardTitle>Security Settings</CardTitle>
                  <CardDescription>
                    Configure security and authentication settings
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className='h-[200px] bg-muted/50 rounded flex items-center justify-center'>
                    <p className='text-muted-foreground'>Security settings will be implemented</p>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
        </Tabs>
      </Main>
    </>
  )
}

const topNav = [
  {
    title: 'Settings',
    href: '#',
    isActive: true,
    disabled: false,
  },
]