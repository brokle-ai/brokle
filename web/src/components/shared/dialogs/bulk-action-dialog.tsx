'use client'

import { useState } from 'react'
import { AlertTriangle, CheckCircle2, Info, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'

interface BulkActionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  actionType: 'info' | 'warning' | 'destructive'
  itemCount: number
  itemType: string
  onConfirm: () => Promise<void>
  confirmText?: string
  cancelText?: string
}

export function BulkActionDialog({
  open,
  onOpenChange,
  title,
  description,
  actionType,
  itemCount,
  itemType,
  onConfirm,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
}: BulkActionDialogProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [progress, setProgress] = useState(0)

  const handleConfirm = async () => {
    setIsLoading(true)
    setProgress(0)

    try {
      // Simulate progress
      const interval = setInterval(() => {
        setProgress((prev) => {
          if (prev >= 90) {
            clearInterval(interval)
            return 90
          }
          return prev + 10
        })
      }, 200)

      await onConfirm()
      
      clearInterval(interval)
      setProgress(100)
      
      // Brief delay to show completion
      setTimeout(() => {
        onOpenChange(false)
        setIsLoading(false)
        setProgress(0)
      }, 500)
    } catch (error) {
      setIsLoading(false)
      setProgress(0)
    }
  }

  const getIcon = () => {
    if (isLoading) {
      return <div className='h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent' />
    }
    
    switch (actionType) {
      case 'destructive':
        return <AlertTriangle className='h-6 w-6 text-destructive' />
      case 'warning':
        return <AlertTriangle className='h-6 w-6 text-orange-500' />
      default:
        return <Info className='h-6 w-6 text-primary' />
    }
  }

  const getVariant = () => {
    switch (actionType) {
      case 'destructive':
        return 'destructive'
      default:
        return 'default'
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[425px]'>
        <DialogHeader>
          <div className='flex items-center gap-3'>
            {getIcon()}
            <div className='flex-1'>
              <DialogTitle className='text-left'>{title}</DialogTitle>
              <div className='flex items-center gap-2 mt-1'>
                <Badge variant='secondary' className='text-xs'>
                  {itemCount} {itemType}{itemCount !== 1 ? 's' : ''}
                </Badge>
              </div>
            </div>
          </div>
          <DialogDescription className='text-left pt-2'>
            {description}
          </DialogDescription>
        </DialogHeader>

        {isLoading && (
          <div className='space-y-2'>
            <div className='flex items-center justify-between text-sm'>
              <span>Processing...</span>
              <span>{progress}%</span>
            </div>
            <Progress value={progress} className='h-2' />
          </div>
        )}

        <DialogFooter className='gap-2 sm:gap-0'>
          <Button
            variant='outline'
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
          >
            {isLoading ? 'Processing...' : cancelText}
          </Button>
          <Button
            variant={getVariant()}
            onClick={handleConfirm}
            disabled={isLoading}
          >
            {isLoading ? (
              <div className='flex items-center gap-2'>
                <div className='h-4 w-4 animate-spin rounded-full border-2 border-background border-t-transparent' />
                Processing
              </div>
            ) : (
              confirmText
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}