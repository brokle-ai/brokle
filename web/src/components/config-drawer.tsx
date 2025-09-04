'use client'

import { useState } from 'react'
import { Settings, Palette } from 'lucide-react'
import { useTheme } from '@/context/theme-context'
import { useDirection } from '@/context/direction-context'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'

export function ConfigDrawer() {
  const [open, setOpen] = useState(false)
  const { theme, setTheme } = useTheme()
  const { dir, setDir } = useDirection()

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button
          variant='outline'
          size='icon'
          className='fixed bottom-4 right-4 z-50 rounded-full shadow-lg'
        >
          <Settings className='h-4 w-4' />
        </Button>
      </SheetTrigger>
      <SheetContent className='w-80 sm:w-96'>
        <SheetHeader>
          <SheetTitle>Customize Dashboard</SheetTitle>
          <SheetDescription>
            Configure the appearance of your dashboard.
          </SheetDescription>
        </SheetHeader>

        <div className='space-y-6 mt-6'>
          <div className='space-y-3'>
            <Label className='text-sm font-medium flex items-center gap-2'>
              <Palette className='h-4 w-4' />
              Color Theme
            </Label>
            <RadioGroup
              value={theme}
              onValueChange={setTheme}
              className='grid grid-cols-3 gap-3'
            >
              <div>
                <RadioGroupItem value='light' id='light' className='sr-only' />
                <Label
                  htmlFor='light'
                  className='flex flex-col items-center justify-between rounded-md border-2 border-muted bg-popover p-2 hover:bg-accent hover:text-accent-foreground [&:has([data-state=checked])]:border-primary'
                >
                  <div className='w-full h-8 rounded bg-white border mb-2' />
                  <span className='text-xs'>Light</span>
                </Label>
              </div>
              <div>
                <RadioGroupItem value='dark' id='dark' className='sr-only' />
                <Label
                  htmlFor='dark'
                  className='flex flex-col items-center justify-between rounded-md border-2 border-muted bg-popover p-2 hover:bg-accent hover:text-accent-foreground [&:has([data-state=checked])]:border-primary'
                >
                  <div className='w-full h-8 rounded bg-slate-900 border mb-2' />
                  <span className='text-xs'>Dark</span>
                </Label>
              </div>
              <div>
                <RadioGroupItem value='system' id='system' className='sr-only' />
                <Label
                  htmlFor='system'
                  className='flex flex-col items-center justify-between rounded-md border-2 border-muted bg-popover p-2 hover:bg-accent hover:text-accent-foreground [&:has([data-state=checked])]:border-primary'
                >
                  <div className='w-full h-8 rounded bg-gradient-to-r from-white to-slate-900 border mb-2' />
                  <span className='text-xs'>System</span>
                </Label>
              </div>
            </RadioGroup>
          </div>

          <Separator />

          <div className='space-y-3'>
            <Label className='text-sm font-medium'>Direction</Label>
            <div className='flex items-center justify-between'>
              <div className='space-y-0.5'>
                <Label className='text-sm font-medium'>Right-to-left (RTL)</Label>
                <p className='text-xs text-muted-foreground'>
                  Switch to RTL text direction
                </p>
              </div>
              <Switch
                checked={dir === 'rtl'}
                onCheckedChange={(checked) => setDir(checked ? 'rtl' : 'ltr')}
              />
            </div>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
}